package web

import (
	"context"
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
	startTime := time.Now()

	// 进度回调：开始压缩
	progressCallback(10.0, "开始压缩...")
	time.Sleep(100 * time.Millisecond) // 让前端有时间显示进度

	// 进度回调：解析文件
	progressCallback(20.0, "解析WebP文件...")
	time.Sleep(100 * time.Millisecond)

	// 进度回调：处理中
	progressCallback(40.0, "压缩处理中...")
	time.Sleep(100 * time.Millisecond)

	// 创建一个带取消的context
	compressionCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// 启动进度更新协程
	progressDone := make(chan bool, 1)
	go func() {
		defer func() {
			select {
			case progressDone <- true:
			default:
			}
		}()

		currentProgress := 40.0
		ticker := time.NewTicker(1 * time.Second) // 每1秒更新一次进度，更频繁
		defer ticker.Stop()

		for {
			select {
			case <-compressionCtx.Done():
				return
			case <-ticker.C:
				// 逐渐增加进度，但不超过95%
				if currentProgress < 95.0 {
					// 动态调整进度增量，越接近95%增量越小
					increment := 3.0
					if currentProgress > 80.0 {
						increment = 1.0 // 80%以后每次增加1%
					} else if currentProgress > 70.0 {
						increment = 2.0 // 70-80%每次增加2%
					}

					currentProgress += increment
					if currentProgress > 95.0 {
						currentProgress = 95.0
					}

					// 根据进度显示不同的状态信息
					var status string
					if currentProgress <= 50.0 {
						status = "压缩处理中..."
					} else if currentProgress <= 70.0 {
						status = "压缩帧处理中..."
					} else if currentProgress <= 90.0 {
						status = "优化压缩效果..."
					} else {
						status = "生成最终文件..."
					}

					progressCallback(currentProgress, status)
				}
			}
		}
	}()

	// 调用原始的压缩服务
	result, err := job.WebPService.CompressAnimation(
		compressionCtx,
		job.Task.InputFile,
		job.Task.OutputFile,
		job.Task.Config,
	)

	// 停止进度更新协程
	cancel()

	// 等待进度协程结束或超时
	select {
	case <-progressDone:
	case <-time.After(100 * time.Millisecond):
	}

	if err != nil {
		return nil, err
	}

	// 进度回调：完成
	progressCallback(90.0, "生成输出文件...")
	time.Sleep(100 * time.Millisecond)

	progressCallback(100.0, "压缩完成")

	// 更新处理时间
	result.ProcessingTime = time.Since(startTime)

	return result, nil
}
