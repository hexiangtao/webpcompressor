package domain

import (
	"context"
	"sync"
	"time"
)

// FrameInfo 表示WebP动画帧信息
type FrameInfo struct {
	Index    int           `json:"index"`
	X        int           `json:"x"`
	Y        int           `json:"y"`
	Duration time.Duration `json:"duration"`
	Dispose  DisposeMethod `json:"dispose"`
	Blend    BlendMethod   `json:"blend"`
	Path     string        `json:"path"`
}

// DisposeMethod 表示帧处理方式
type DisposeMethod int

const (
	DisposeNone       DisposeMethod = 0 // 不处理
	DisposeBackground DisposeMethod = 1 // 背景色处理
)

// BlendMethod 表示帧混合方式
type BlendMethod int

const (
	BlendNo  BlendMethod = 0 // 不混合
	BlendYes BlendMethod = 1 // 混合
)

// AnimationInfo 表示动画信息
type AnimationInfo struct {
	Width      int          `json:"width"`
	Height     int          `json:"height"`
	FrameCount int          `json:"frame_count"`
	LoopCount  int          `json:"loop_count"`
	Frames     []*FrameInfo `json:"frames"`
}

// CompressionConfig 表示压缩配置
type CompressionConfig struct {
	Quality        int    `json:"quality"`         // 质量 0-100
	Method         int    `json:"method"`          // 压缩方法 0-6
	FilterStrength int    `json:"filter_strength"` // 滤波强度 0-100
	Preset         string `json:"preset"`          // 预设
	Lossless       bool   `json:"lossless"`        // 无损压缩
	AlphaQuality   int    `json:"alpha_quality"`   // Alpha质量
	EnableParallel bool   `json:"enable_parallel"` // 启用并行处理
	MaxConcurrency int    `json:"max_concurrency"` // 最大并发数
}

// DefaultCompressionConfig 返回默认压缩配置
func DefaultCompressionConfig(quality int) *CompressionConfig {
	return &CompressionConfig{
		Quality:        quality,
		Method:         6,   // 最高压缩方法
		FilterStrength: 100, // 最大滤波强度
		Preset:         "photo",
		Lossless:       false,
		AlphaQuality:   quality / 2,
		EnableParallel: true, // 默认启用并行处理
		MaxConcurrency: 4,    // 默认4个并发
	}
}

// CompressResult 表示压缩结果
type CompressResult struct {
	OriginalSize     int64         `json:"original_size"`
	CompressedSize   int64         `json:"compressed_size"`
	CompressionRatio float64       `json:"compression_ratio"`
	ProcessingTime   time.Duration `json:"processing_time"`
	FramesProcessed  int           `json:"frames_processed"`
	ParallelWorkers  int           `json:"parallel_workers"` // 使用的并行工作者数量
}

// CalculateCompressionRatio 计算压缩率
func (r *CompressResult) CalculateCompressionRatio() {
	if r.OriginalSize > 0 {
		r.CompressionRatio = float64(r.CompressedSize) / float64(r.OriginalSize) * 100
	}
}

// ParallelProcessor 并行处理器接口
type ParallelProcessor interface {
	// ProcessFramesParallel 并行处理帧
	ProcessFramesParallel(ctx context.Context, frames []*FrameInfo, processor FrameProcessor, maxWorkers int) error
}

// FrameProcessor 帧处理器函数类型
type FrameProcessor func(ctx context.Context, frame *FrameInfo) error

// WorkerPool 工作池
type WorkerPool struct {
	maxWorkers int
	jobs       chan *FrameInfo
	results    chan error
	wg         sync.WaitGroup
}

// NewWorkerPool 创建工作池
func NewWorkerPool(maxWorkers int) *WorkerPool {
	return &WorkerPool{
		maxWorkers: maxWorkers,
		jobs:       make(chan *FrameInfo, maxWorkers*2),
		results:    make(chan error, maxWorkers*2),
	}
}

// Start 启动工作池
func (wp *WorkerPool) Start(ctx context.Context, processor FrameProcessor) {
	for i := 0; i < wp.maxWorkers; i++ {
		wp.wg.Add(1)
		go wp.worker(ctx, processor)
	}
}

// Submit 提交任务
func (wp *WorkerPool) Submit(frame *FrameInfo) {
	wp.jobs <- frame
}

// Close 关闭工作池
func (wp *WorkerPool) Close() {
	close(wp.jobs)
}

// Wait 等待所有任务完成
func (wp *WorkerPool) Wait() []error {
	wp.wg.Wait()
	close(wp.results)

	var errors []error
	for err := range wp.results {
		if err != nil {
			errors = append(errors, err)
		}
	}
	return errors
}

// worker 工作者
func (wp *WorkerPool) worker(ctx context.Context, processor FrameProcessor) {
	defer wp.wg.Done()

	for frame := range wp.jobs {
		select {
		case <-ctx.Done():
			wp.results <- ctx.Err()
			return
		default:
			err := processor(ctx, frame)
			wp.results <- err
		}
	}
}

// BatchProcessor 批量处理器接口
type BatchProcessor interface {
	// ProcessBatch 批量处理多个文件
	ProcessBatch(ctx context.Context, inputFiles []string, config *CompressionConfig) ([]*CompressResult, error)

	// ProcessBatchWithProgress 批量处理并提供进度回调
	ProcessBatchWithProgress(ctx context.Context, inputFiles []string, config *CompressionConfig, progressCallback ProgressCallback) ([]*CompressResult, error)
}

// ProgressCallback 进度回调函数类型
type ProgressCallback func(completed, total int, currentFile string)

// WebPProcessor 定义WebP处理接口
type WebPProcessor interface {
	// ParseAnimation 解析WebP动画信息
	ParseAnimation(ctx context.Context, inputPath string) (*AnimationInfo, error)

	// ExtractFrames 提取动画帧
	ExtractFrames(ctx context.Context, inputPath string, outputDir string, frames []*FrameInfo) error

	// CompressFrames 压缩帧
	CompressFrames(ctx context.Context, frames []*FrameInfo, config *CompressionConfig) error

	// CompressFramesParallel 并行压缩帧
	CompressFramesParallel(ctx context.Context, frames []*FrameInfo, config *CompressionConfig) error

	// AssembleAnimation 重新组装动画
	AssembleAnimation(ctx context.Context, frames []*FrameInfo, outputPath string) error

	// CompressAnimation 完整的动画压缩流程
	CompressAnimation(ctx context.Context, inputPath, outputPath string, config *CompressionConfig) (*CompressResult, error)
}

// ToolExecutor 定义工具执行接口
type ToolExecutor interface {
	// ExecuteCommand 执行命令
	ExecuteCommand(ctx context.Context, toolName string, args ...string) error

	// ExecuteCommandWithOutput 执行命令并返回输出
	ExecuteCommandWithOutput(ctx context.Context, toolName string, args ...string) (string, error)

	// GetToolPath 获取工具路径
	GetToolPath(toolName string) string

	// IsToolAvailable 检查工具是否可用
	IsToolAvailable(toolName string) bool
}

// FileManager 定义文件管理接口
type FileManager interface {
	// CreateTempDir 创建临时目录
	CreateTempDir(prefix string) (string, error)

	// CleanupTempDir 清理临时目录
	CleanupTempDir(path string) error

	// GetFileSize 获取文件大小
	GetFileSize(path string) (int64, error)

	// FileExists 检查文件是否存在
	FileExists(path string) bool

	// CopyFile 复制文件
	CopyFile(src, dst string) error
}
