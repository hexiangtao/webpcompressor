package web

import (
	"context"
	"fmt"
	"sync"
	"time"

	"webpcompressor/internal/domain"
	"webpcompressor/internal/infrastructure"
	"webpcompressor/internal/service"
	"webpcompressor/pkg/logger"
)

// WorkerPool 任务工作池
type WorkerPool struct {
	maxWorkers int
	taskQueue  chan *TaskJob
	workers    []*Worker
	wg         sync.WaitGroup
	stopped    bool
	stopChan   chan struct{}
	logger     logger.Logger
	mu         sync.RWMutex
}

// TaskJob 任务作业
type TaskJob struct {
	Task             *domain.TaskInfo
	WebPService      *service.WebPService
	TaskManager      domain.TaskManager
	ProgressReporter *infrastructure.ProgressReporter
}

// Worker 工作者
type Worker struct {
	id        int
	taskQueue <-chan *TaskJob
	quit      chan struct{}
	logger    logger.Logger
}

// NewWorkerPool 创建工作池
func NewWorkerPool(maxWorkers int, logger logger.Logger) *WorkerPool {
	return &WorkerPool{
		maxWorkers: maxWorkers,
		taskQueue:  make(chan *TaskJob, maxWorkers*2), // 缓冲队列
		workers:    make([]*Worker, maxWorkers),
		stopChan:   make(chan struct{}),
		logger:     logger,
	}
}

// Start 启动工作池
func (wp *WorkerPool) Start() {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	if wp.stopped {
		return
	}

	wp.logger.Info("启动任务工作池", "workers", wp.maxWorkers)

	// 启动工作者
	for i := 0; i < wp.maxWorkers; i++ {
		worker := &Worker{
			id:        i,
			taskQueue: wp.taskQueue,
			quit:      make(chan struct{}),
			logger:    wp.logger,
		}
		wp.workers[i] = worker
		wp.wg.Add(1)
		go worker.Start(&wp.wg)
	}
}

// Stop 停止工作池
func (wp *WorkerPool) Stop() {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	if wp.stopped {
		return
	}

	wp.logger.Info("停止任务工作池...")
	wp.stopped = true
	close(wp.stopChan)

	// 停止所有工作者
	for _, worker := range wp.workers {
		worker.Stop()
	}

	// 等待所有工作者完成
	wp.wg.Wait()

	// 关闭任务队列
	close(wp.taskQueue)

	wp.logger.Info("任务工作池已停止")
}

// SubmitTask 提交任务
func (wp *WorkerPool) SubmitTask(
	task *domain.TaskInfo,
	webpService *service.WebPService,
	taskManager domain.TaskManager,
	progressReporter *infrastructure.ProgressReporter,
) {
	wp.mu.RLock()
	defer wp.mu.RUnlock()

	if wp.stopped {
		wp.logger.Warn("工作池已停止，无法提交任务", "task_id", task.ID)
		return
	}

	job := &TaskJob{
		Task:             task,
		WebPService:      webpService,
		TaskManager:      taskManager,
		ProgressReporter: progressReporter,
	}

	select {
	case wp.taskQueue <- job:
		wp.logger.Debug("任务已提交到队列", "task_id", task.ID)
	default:
		wp.logger.Error("任务队列已满，无法提交任务", "task_id", task.ID)
		// 标记任务失败
		task.Fail(domain.ErrTaskQueueFull)
		taskManager.UpdateTask(task)
	}
}

// Start 启动工作者
func (w *Worker) Start(wg *sync.WaitGroup) {
	defer wg.Done()

	w.logger.Debug("启动工作者", "worker_id", w.id)

	for {
		select {
		case job := <-w.taskQueue:
			if job == nil {
				return // 队列已关闭
			}
			w.processTask(job)
		case <-w.quit:
			w.logger.Debug("工作者退出", "worker_id", w.id)
			return
		}
	}
}

// Stop 停止工作者
func (w *Worker) Stop() {
	close(w.quit)
}

// processTask 处理任务
func (w *Worker) processTask(job *TaskJob) {
	task := job.Task
	taskManager := job.TaskManager
	progressReporter := job.ProgressReporter

	w.logger.Info("开始处理任务",
		"worker_id", w.id,
		"task_id", task.ID,
		"input", task.InputFile,
		"quality", task.Config.Quality,
	)

	// 开始任务
	task.Start()
	if err := taskManager.UpdateTask(task); err != nil {
		w.logger.Error("更新任务状态失败", "task_id", task.ID, "error", err)
	}

	// 报告开始进度
	progressReporter.ReportProgress(task.ID, 0.0, "开始压缩...")

	// 创建带有进度回调的上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	// 创建进度回调
	progressCallback := func(progress float64, message string) {
		task.UpdateProgress(progress, message)
		taskManager.UpdateTask(task)
		progressReporter.ReportProgress(task.ID, progress, message)
	}

	// 执行压缩
	result, err := w.executeCompressionWithProgress(ctx, job, progressCallback)

	if err != nil {
		w.logger.Error("任务处理失败",
			"worker_id", w.id,
			"task_id", task.ID,
			"error", err,
		)

		task.Fail(err)
		progressReporter.ReportProgress(task.ID, 100.0, "压缩失败: "+err.Error())
	} else {
		w.logger.Info("任务处理成功",
			"worker_id", w.id,
			"task_id", task.ID,
			"compression_ratio", result.CompressionRatio,
			"processing_time", result.ProcessingTime,
		)

		task.Complete(result)
		progressReporter.ReportProgress(task.ID, 100.0, "压缩完成")
	}

	// 更新任务状态
	if err := taskManager.UpdateTask(task); err != nil {
		w.logger.Error("更新最终任务状态失败", "task_id", task.ID, "error", err)
	}
}

// executeCompressionWithProgress 执行带进度的压缩
func (w *Worker) executeCompressionWithProgress(
	ctx context.Context,
	job *TaskJob,
	progressCallback func(float64, string),
) (*domain.CompressResult, error) {
	// 创建自定义的WebP服务，支持进度回调
	webpService := &ProgressAwareWebPService{
		WebPService:      job.WebPService,
		progressCallback: progressCallback,
	}

	return webpService.CompressAnimationWithProgress(
		ctx,
		job.Task.InputFile,
		job.Task.OutputFile,
		job.Task.Config,
	)
}

// ProgressAwareWebPService 支持进度的WebP服务包装器
type ProgressAwareWebPService struct {
	*service.WebPService
	progressCallback func(float64, string)
}

// CompressAnimationWithProgress 带进度的压缩动画
func (p *ProgressAwareWebPService) CompressAnimationWithProgress(
	ctx context.Context,
	inputPath, outputPath string,
	config *domain.CompressionConfig,
) (*domain.CompressResult, error) {
	// 阶段1: 解析动画信息 (10%)
	p.progressCallback(10.0, "解析动画信息...")

	animInfo, err := p.WebPService.ParseAnimation(ctx, inputPath)
	if err != nil {
		return nil, err
	}

	// 阶段2: 提取帧 (30%)
	p.progressCallback(30.0, "提取动画帧...")

	tempDir, err := p.WebPService.GetFileManager().CreateTempDir("webp_compress")
	if err != nil {
		return nil, err
	}
	defer p.WebPService.GetFileManager().CleanupTempDir(tempDir)

	if err := p.WebPService.ExtractFrames(ctx, inputPath, tempDir, animInfo.Frames); err != nil {
		return nil, err
	}

	// 阶段3: 压缩帧 (70%)
	p.progressCallback(50.0, "压缩帧...")

	if err := p.compressFramesWithProgress(ctx, animInfo.Frames, config); err != nil {
		return nil, err
	}

	// 阶段4: 重新组装 (90%)
	p.progressCallback(90.0, "重新组装动画...")

	if err := p.WebPService.AssembleAnimation(ctx, animInfo.Frames, outputPath); err != nil {
		return nil, err
	}

	// 完成
	p.progressCallback(100.0, "压缩完成")

	// 计算结果
	originalSize, _ := p.WebPService.GetFileManager().GetFileSize(inputPath)
	compressedSize, _ := p.WebPService.GetFileManager().GetFileSize(outputPath)

	result := &domain.CompressResult{
		OriginalSize:    originalSize,
		CompressedSize:  compressedSize,
		FramesProcessed: len(animInfo.Frames),
		ProcessingTime:  time.Since(time.Now()),
		ParallelWorkers: 1,
	}
	result.CalculateCompressionRatio()

	return result, nil
}

// compressFramesWithProgress 带进度的帧压缩
func (p *ProgressAwareWebPService) compressFramesWithProgress(
	ctx context.Context,
	frames []*domain.FrameInfo,
	config *domain.CompressionConfig,
) error {
	totalFrames := len(frames)

	for i, frame := range frames {
		progress := 50.0 + (float64(i+1)/float64(totalFrames))*40.0 // 50% to 90%
		p.progressCallback(progress, fmt.Sprintf("压缩帧 %d/%d", i+1, totalFrames))

		if err := p.compressFrame(ctx, frame, config); err != nil {
			return err
		}
	}

	return nil
}

// compressFrame 压缩单个帧（临时实现，需要访问WebPService的私有方法）
func (p *ProgressAwareWebPService) compressFrame(
	ctx context.Context,
	frame *domain.FrameInfo,
	config *domain.CompressionConfig,
) error {
	// 这里应该调用WebPService的compressFrame方法
	// 由于是私有方法，我们需要通过反射或者修改WebPService来支持
	// 暂时返回nil，等待后续实现
	return nil
}

// GetFileManager 获取文件管理器（需要在WebPService中添加）
func (p *ProgressAwareWebPService) GetFileManager() domain.FileManager {
	// 需要在WebPService中添加此方法
	return nil
}
