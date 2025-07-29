package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"webpcompressor/internal/config"
	"webpcompressor/internal/domain"
	"webpcompressor/internal/infrastructure"
	"webpcompressor/internal/service"
	"webpcompressor/pkg/logger"
)

// Application 应用程序结构
type Application struct {
	config         *config.Config
	logger         logger.Logger
	webpService    *service.WebPService
	tempDirManager *infrastructure.TempDirManager
}

// NewApplication 创建应用程序实例
func NewApplication() (*Application, error) {
	// 加载配置
	cfg := config.DefaultConfig()
	cfg.LoadFromEnv()
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}

	// 初始化日志
	appLogger, err := logger.NewLogger(&cfg.Logging)
	if err != nil {
		// 回退到默认日志
		appLogger = logger.NewDefaultLogger()
		appLogger.Warn("使用默认日志配置", "error", err)
	}

	// 创建工厂
	toolFactory := infrastructure.NewToolExecutorFactory(cfg, appLogger)
	fileFactory := infrastructure.NewFileManagerFactory(cfg, appLogger)

	// 创建基础组件
	toolExecutor := toolFactory.CreateExecutor(cfg.Tools.UseEmbedded, "")
	fileManager := fileFactory.CreateFileManager(true) // 使用安全模式

	// 验证工具可用性
	if err := toolFactory.ValidateTools(toolExecutor); err != nil {
		return nil, fmt.Errorf("工具验证失败: %w", err)
	}

	// 创建临时目录管理器
	tempDirManager := infrastructure.NewTempDirManager(fileManager, appLogger)

	// 创建服务
	webpService := service.NewWebPService(cfg, toolExecutor, fileManager, appLogger)

	return &Application{
		config:         cfg,
		logger:         appLogger,
		webpService:    webpService,
		tempDirManager: tempDirManager,
	}, nil
}

// Run 运行应用程序
func (app *Application) Run(args []string) error {
	// 确保清理临时文件
	defer app.tempDirManager.CleanupAll()

	// 解析命令行参数
	if len(args) < 4 {
		app.showUsage()
		return fmt.Errorf("参数不足")
	}

	inputFile := args[1]
	quality, err := strconv.Atoi(args[2])
	if err != nil {
		return fmt.Errorf("无效的质量参数: %s", args[2])
	}
	outputFile := args[3]

	// 创建压缩配置
	compressionConfig := domain.DefaultCompressionConfig(quality)

	// 创建上下文
	ctx, cancel := context.WithTimeout(context.Background(), app.config.App.Timeout)
	defer cancel()

	// 记录开始
	app.logger.Info("开始WebP压缩",
		"input", inputFile,
		"output", outputFile,
		"quality", quality,
		"version", app.config.App.Version,
	)

	startTime := time.Now()

	// 执行压缩
	result, err := app.webpService.CompressAnimation(ctx, inputFile, outputFile, compressionConfig)
	if err != nil {
		app.logger.Error("压缩失败", "error", err)
		return err
	}

	// 记录结果
	app.logger.Info("压缩成功",
		"duration", time.Since(startTime),
		"original_size", result.OriginalSize,
		"compressed_size", result.CompressedSize,
		"compression_ratio", fmt.Sprintf("%.1f%%", result.CompressionRatio),
		"frames_processed", result.FramesProcessed,
	)

	// 显示用户友好的结果
	fmt.Printf("✅ 压缩完成！\n")
	fmt.Printf("📊 压缩效果: %s -> %s (%.1f%%)\n",
		formatFileSize(result.OriginalSize),
		formatFileSize(result.CompressedSize),
		result.CompressionRatio)
	fmt.Printf("⏱️  处理时间: %v\n", result.ProcessingTime)
	fmt.Printf("🎞️  处理帧数: %d\n", result.FramesProcessed)

	return nil
}

// showUsage 显示使用说明
func (app *Application) showUsage() {
	fmt.Printf(`WebP Compressor v%s - 高性能WebP动画压缩工具

用法: %s <input.webp> <quality[0-100]> <output.webp>

参数:
  input.webp    输入的WebP动画文件
  quality       压缩质量(0-100)，建议30-50获得更好的压缩效果
  output.webp   输出的压缩文件

示例:
  %s animation.webp 40 compressed.webp

环境变量配置:
  WEBP_LOG_LEVEL       日志级别 (debug|info|warn|error)
  WEBP_TEMP_DIR        临时目录路径
  WEBP_MAX_CONCURRENCY 最大并发数
  WEBP_TIMEOUT         操作超时时间
  WEBP_MAX_FILE_SIZE   最大文件大小限制

更多信息请访问: https://github.com/webmproject/libwebp
`,
		app.config.App.Version,
		os.Args[0],
		os.Args[0])
}

// formatFileSize 格式化文件大小
func formatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// main 主函数
func main() {
	// 创建应用程序
	app, err := NewApplication()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ 初始化失败: %v\n", err)
		os.Exit(1)
	}

	// 运行应用程序
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "❌ 运行失败: %v\n", err)
		os.Exit(1)
	}
}
