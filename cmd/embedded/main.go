package main

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"webpcompressor/internal/config"
	"webpcompressor/internal/domain"
	"webpcompressor/internal/infrastructure"
	"webpcompressor/internal/service"
	"webpcompressor/pkg/logger"
)

// 嵌入所有WebP工具二进制文件
//
//go:embed embedded/webpmux.exe
var webpmuxBin []byte

//go:embed embedded/cwebp.exe
var cwebpBin []byte

//go:embed embedded/dwebp.exe
var dwebpBin []byte

//go:embed embedded/gif2webp.exe
var gif2webpBin []byte

//go:embed embedded/webpinfo.exe
var webpinfoBin []byte

//go:embed embedded/anim_diff.exe
var animDiffBin []byte

//go:embed embedded/anim_dump.exe
var animDumpBin []byte

//go:embed embedded/get_disto.exe
var getDistoBin []byte

//go:embed embedded/img2webp.exe
var img2webpBin []byte

//go:embed embedded/webp_quality.exe
var webpQualityBin []byte

//go:embed embedded/vwebp.exe
var vwebpBin []byte

//go:embed embedded/freeglut.dll
var freeglutDLL []byte

// EmbeddedTool 嵌入工具定义
type EmbeddedTool struct {
	name string
	data []byte
	desc string
}

// 嵌入工具列表
var embeddedTools = []EmbeddedTool{
	{"webpmux.exe", webpmuxBin, "WebP动画信息解析和处理"},
	{"cwebp.exe", cwebpBin, "将图像转换为WebP格式"},
	{"dwebp.exe", dwebpBin, "将WebP格式转换为其他图像格式"},
	{"gif2webp.exe", gif2webpBin, "将GIF动画转换为WebP动画"},
	{"webpinfo.exe", webpinfoBin, "显示WebP文件详细信息"},
	{"anim_diff.exe", animDiffBin, "比较两个WebP动画的差异"},
	{"anim_dump.exe", animDumpBin, "从WebP动画中提取帧"},
	{"get_disto.exe", getDistoBin, "计算失真度量"},
	{"img2webp.exe", img2webpBin, "将多个图像合成WebP动画"},
	{"webp_quality.exe", webpQualityBin, "评估WebP图像质量"},
	{"vwebp.exe", vwebpBin, "WebP图像查看器"},
	{"freeglut.dll", freeglutDLL, "OpenGL实用工具库"},
}

// EmbeddedApplication 嵌入式应用程序
type EmbeddedApplication struct {
	config         *config.Config
	logger         logger.Logger
	webpService    *service.WebPService
	tempDirManager *infrastructure.TempDirManager
	tempDir        string
}

// NewEmbeddedApplication 创建嵌入式应用程序
func NewEmbeddedApplication() (*EmbeddedApplication, error) {
	// 加载配置
	cfg := config.DefaultConfig()
	cfg.LoadFromEnv()
	cfg.Tools.UseEmbedded = true // 强制使用嵌入模式
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}

	// 初始化日志
	appLogger, err := logger.NewLogger(&cfg.Logging)
	if err != nil {
		appLogger = logger.NewDefaultLogger()
		appLogger.Warn("使用默认日志配置", "error", err)
	}

	// 提取嵌入的工具到临时目录
	tempDir, err := extractEmbeddedTools(appLogger)
	if err != nil {
		return nil, fmt.Errorf("提取嵌入工具失败: %w", err)
	}

	// 创建工厂
	toolFactory := infrastructure.NewToolExecutorFactory(cfg, appLogger)
	fileFactory := infrastructure.NewFileManagerFactory(cfg, appLogger)

	// 创建基础组件（使用嵌入模式）
	toolExecutor := toolFactory.CreateExecutor(true, tempDir)
	fileManager := fileFactory.CreateFileManager(true)

	// 验证工具可用性
	if err := toolFactory.ValidateTools(toolExecutor); err != nil {
		return nil, fmt.Errorf("工具验证失败: %w", err)
	}

	// 创建临时目录管理器
	tempDirManager := infrastructure.NewTempDirManager(fileManager, appLogger)

	// 创建服务
	webpService := service.NewWebPService(cfg, toolExecutor, fileManager, appLogger)

	return &EmbeddedApplication{
		config:         cfg,
		logger:         appLogger,
		webpService:    webpService,
		tempDirManager: tempDirManager,
		tempDir:        tempDir,
	}, nil
}

// extractEmbeddedTools 提取嵌入的工具到临时目录
func extractEmbeddedTools(logger logger.Logger) (string, error) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "webptools_*")
	if err != nil {
		return "", fmt.Errorf("创建临时目录失败: %w", err)
	}

	logger.Info("提取嵌入工具", "temp_dir", tempDir, "tools_count", len(embeddedTools))

	// 提取所有工具
	for _, tool := range embeddedTools {
		toolPath := filepath.Join(tempDir, tool.name)

		if err := os.WriteFile(toolPath, tool.data, 0755); err != nil {
			return "", fmt.Errorf("写入工具文件失败 %s: %w", tool.name, err)
		}

		logger.Debug("提取工具文件", "name", tool.name, "size", len(tool.data))
	}

	logger.Info("所有嵌入工具提取完成", "temp_dir", tempDir)
	return tempDir, nil
}

// Cleanup 清理资源
func (app *EmbeddedApplication) Cleanup() {
	// 清理临时目录管理器管理的目录
	app.tempDirManager.CleanupAll()

	// 清理嵌入工具的临时目录
	if app.tempDir != "" {
		if err := os.RemoveAll(app.tempDir); err != nil {
			app.logger.Warn("清理嵌入工具临时目录失败", "dir", app.tempDir, "error", err)
		} else {
			app.logger.Info("清理嵌入工具临时目录成功", "dir", app.tempDir)
		}
	}
}

// Run 运行应用程序
func (app *EmbeddedApplication) Run(args []string) error {
	// 确保清理资源
	defer app.Cleanup()

	if len(args) < 2 {
		app.showUsage()
		return nil
	}

	command := args[1]

	switch command {
	case "compress", "压缩":
		return app.handleCompress(args[2:])
	case "info", "信息":
		return app.handleInfo(args[2:])
	case "help", "帮助":
		app.showDetailedHelp()
		return nil
	case "version", "版本":
		fmt.Printf("WebP工具集 v%s (嵌入版)\n", app.config.App.Version)
		return nil
	default:
		fmt.Printf("❌ 未知命令: %s\n", command)
		app.showUsage()
		return fmt.Errorf("未知命令: %s", command)
	}
}

// handleCompress 处理压缩命令
func (app *EmbeddedApplication) handleCompress(args []string) error {
	if len(args) < 3 {
		fmt.Println("用法: webptools compress <input.webp> <quality[0-100]> <output.webp>")
		return fmt.Errorf("参数不足")
	}

	inputFile := args[0]
	quality, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("无效的质量参数: %s", args[1])
	}
	outputFile := args[2]

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
		"mode", "embedded",
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

// handleInfo 处理信息命令
func (app *EmbeddedApplication) handleInfo(args []string) error {
	if len(args) < 1 {
		fmt.Println("用法: webptools info <input.webp>")
		return fmt.Errorf("参数不足")
	}

	inputFile := args[0]

	// 创建上下文
	ctx, cancel := context.WithTimeout(context.Background(), app.config.App.Timeout)
	defer cancel()

	// 解析动画信息
	animInfo, err := app.webpService.ParseAnimation(ctx, inputFile)
	if err != nil {
		return fmt.Errorf("解析WebP文件失败: %w", err)
	}

	// 显示信息
	fmt.Printf("📄 WebP文件信息: %s\n", inputFile)
	fmt.Printf("📐 画布大小: %dx%d\n", animInfo.Width, animInfo.Height)
	fmt.Printf("🎞️  总帧数: %d\n", len(animInfo.Frames))
	fmt.Printf("🔄 循环次数: %d\n", animInfo.LoopCount)

	if len(animInfo.Frames) > 0 {
		fmt.Printf("\n📋 帧详情:\n")
		for i, frame := range animInfo.Frames {
			if i >= 5 { // 只显示前5帧的详情
				fmt.Printf("  ... 还有 %d 帧\n", len(animInfo.Frames)-5)
				break
			}
			fmt.Printf("  帧 %d: 位置(%d,%d) 持续时间=%dms\n",
				frame.Index, frame.X, frame.Y, int(frame.Duration/time.Millisecond))
		}
	}

	return nil
}

// showUsage 显示使用说明
func (app *EmbeddedApplication) showUsage() {
	fmt.Printf(`WebP工具集 v%s (嵌入版) - 内置所有WebP工具

🎯 主要命令:
  compress    压缩WebP动画
  info        显示WebP文件信息
  help        显示详细帮助
  version     显示版本信息

💡 快速开始:
  webptools compress input.webp 40 output.webp
  webptools info animation.webp

🔧 完整用法:
  webptools help     查看详细帮助和所有功能

✨ 特性:
  • 内置12个WebP工具，无需外部依赖
  • 高性能动画压缩
  • 智能错误处理和进度显示
  • 支持环境变量配置

`, app.config.App.Version)
}

// showDetailedHelp 显示详细帮助
func (app *EmbeddedApplication) showDetailedHelp() {
	fmt.Printf(`WebP工具集 v%s (嵌入版) - 详细帮助

🎯 主要功能:

1. compress/压缩 - 压缩WebP动画
   用法: webptools compress <input.webp> <quality[0-100]> <output.webp>
   示例: webptools compress animation.webp 40 compressed.webp

2. info/信息 - 显示WebP文件详细信息
   用法: webptools info <input.webp>
   示例: webptools info animation.webp

🛠️ 内置工具 (%d个):
`, app.config.App.Version, len(embeddedTools))

	for _, tool := range embeddedTools {
		fmt.Printf("  • %-15s - %s\n", tool.name, tool.desc)
	}

	fmt.Printf(`
🔧 环境变量配置:
  WEBP_LOG_LEVEL       日志级别 (debug|info|warn|error)
  WEBP_TEMP_DIR        临时目录路径
  WEBP_MAX_CONCURRENCY 最大并发数
  WEBP_TIMEOUT         操作超时时间
  WEBP_MAX_FILE_SIZE   最大文件大小限制

💡 使用提示:
  • 压缩质量: 0-100 (0=最小文件,100=最高质量)
  • 建议质量: 30-50 获得最佳压缩效果
  • 所有工具都已内置，无需外部依赖
  • 工具会自动提取到临时目录并在程序结束时清理

更多信息请访问: https://github.com/webmproject/libwebp
`)
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
	// 创建嵌入式应用程序
	app, err := NewEmbeddedApplication()
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
