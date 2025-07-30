package main

import (
	"fmt"
	"os"

	"webpcompressor/internal/config"
	"webpcompressor/internal/infrastructure"
	"webpcompressor/internal/service"
	"webpcompressor/internal/web"
	"webpcompressor/pkg/logger"
)

// WebApplication Web应用程序结构
type WebApplication struct {
	config      *config.Config
	logger      logger.Logger
	webpService *service.WebPService
	taskManager *infrastructure.MemoryTaskManager
	server      *web.Server
}

// NewWebApplication 创建Web应用程序实例
func NewWebApplication() (*WebApplication, error) {
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

	// 创建服务
	webpService := service.NewWebPService(cfg, toolExecutor, fileManager, appLogger)
	taskManager := infrastructure.NewMemoryTaskManager(appLogger)

	// 创建Web服务器
	server := web.NewServer(cfg, webpService, taskManager, appLogger)

	return &WebApplication{
		config:      cfg,
		logger:      appLogger,
		webpService: webpService,
		taskManager: taskManager,
		server:      server,
	}, nil
}

// Run 运行Web应用程序
func (app *WebApplication) Run() error {
	app.logger.Info("启动WebP压缩Web服务",
		"version", app.config.App.Version,
		"host", app.config.Web.Host,
		"port", app.config.Web.Port,
	)

	// 运行服务器并支持优雅关闭
	return app.server.RunWithGracefulShutdown()
}

// showUsage 显示使用说明
func showUsage() {
	fmt.Printf(`WebP压缩器 Web服务 v2.0

🌐 Web服务模式：
  提供现代化的Web界面进行WebP压缩

🚀 启动服务：
  webpserver

🔧 配置环境变量：
  WEBP_WEB_HOST=0.0.0.0          # 服务器地址
  WEBP_WEB_PORT=8080             # 服务器端口
  WEBP_WEB_MAX_FILE_SIZE=100MB   # 最大文件大小
  WEBP_WEB_ENABLE_AUTH=false     # 是否启用认证
  WEBP_WEB_AUTH_TOKEN=your_token # 认证令牌
  WEBP_LOG_LEVEL=info            # 日志级别

🌐 访问地址：
  http://localhost:8080          # Web界面
  http://localhost:8080/health   # 健康检查
  http://localhost:8080/api/v1   # API文档

✨ 特性：
  • 拖拽上传WebP文件
  • 实时压缩进度显示
  • 多种压缩预设
  • 批量处理支持
  • RESTful API接口
  • WebSocket实时通讯

`)
}

// main 主函数
func main() {
	// 检查命令行参数
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "help", "--help", "-h":
			showUsage()
			return
		case "version", "--version", "-v":
			fmt.Println("WebP压缩器 Web服务 v2.0.0")
			return
		}
	}

	// 创建Web应用程序
	app, err := NewWebApplication()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ 初始化失败: %v\n", err)
		os.Exit(1)
	}

	// 运行应用程序
	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "❌ 运行失败: %v\n", err)
		os.Exit(1)
	}
}
