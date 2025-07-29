package domain

import (
	"context"
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
	}
}

// CompressResult 表示压缩结果
type CompressResult struct {
	OriginalSize     int64         `json:"original_size"`
	CompressedSize   int64         `json:"compressed_size"`
	CompressionRatio float64       `json:"compression_ratio"`
	ProcessingTime   time.Duration `json:"processing_time"`
	FramesProcessed  int           `json:"frames_processed"`
}

// CalculateCompressionRatio 计算压缩率
func (r *CompressResult) CalculateCompressionRatio() {
	if r.OriginalSize > 0 {
		r.CompressionRatio = float64(r.CompressedSize) / float64(r.OriginalSize) * 100
	}
}

// WebPProcessor 定义WebP处理接口
type WebPProcessor interface {
	// ParseAnimation 解析WebP动画信息
	ParseAnimation(ctx context.Context, inputPath string) (*AnimationInfo, error)

	// ExtractFrames 提取动画帧
	ExtractFrames(ctx context.Context, inputPath string, outputDir string, frames []*FrameInfo) error

	// CompressFrames 压缩帧
	CompressFrames(ctx context.Context, frames []*FrameInfo, config *CompressionConfig) error

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
