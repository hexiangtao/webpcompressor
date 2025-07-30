package web

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"webpcompressor/internal/config"
	"webpcompressor/internal/domain"
	"webpcompressor/internal/infrastructure"
	"webpcompressor/internal/service"
	"webpcompressor/pkg/logger"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// Server Web服务器
type Server struct {
	config           *config.Config
	logger           logger.Logger
	router           *gin.Engine
	httpServer       *http.Server
	webpService      *service.WebPService
	taskManager      domain.TaskManager
	progressReporter *infrastructure.ProgressReporter
	workerPool       *WorkerPool
}

// NewServer 创建Web服务器
func NewServer(
	cfg *config.Config,
	webpService *service.WebPService,
	taskManager domain.TaskManager,
	logger logger.Logger,
) *Server {
	// 设置Gin模式
	if cfg.Logging.Level == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	progressReporter := infrastructure.NewProgressReporter(logger)

	server := &Server{
		config:           cfg,
		logger:           logger,
		router:           router,
		webpService:      webpService,
		taskManager:      taskManager,
		progressReporter: progressReporter,
		workerPool:       NewWorkerPool(cfg.Web.MaxConcurrentTasks, logger),
	}

	server.setupMiddleware()
	server.setupRoutes()

	return server
}

// setupMiddleware 设置中间件
func (s *Server) setupMiddleware() {
	// 日志中间件
	s.router.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		Output: os.Stdout,
		Formatter: func(param gin.LogFormatterParams) string {
			return fmt.Sprintf("[GIN] %v | %3d | %13v | %15s | %-7s %s\n",
				param.TimeStamp.Format("2006/01/02 - 15:04:05"),
				param.StatusCode,
				param.Latency,
				param.ClientIP,
				param.Method,
				param.Path,
			)
		},
	}))

	// 恢复中间件
	s.router.Use(gin.Recovery())

	// CORS中间件
	if s.config.Web.EnableCORS {
		corsConfig := cors.DefaultConfig()
		if s.config.Web.AllowedOrigins == "*" {
			corsConfig.AllowAllOrigins = true
		} else {
			corsConfig.AllowOrigins = []string{s.config.Web.AllowedOrigins}
		}
		corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
		corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
		s.router.Use(cors.New(corsConfig))
	}

	// 认证中间件
	if s.config.Web.EnableAuth {
		s.router.Use(s.authMiddleware())
	}

	// 文件大小限制中间件
	s.router.Use(s.fileSizeLimitMiddleware())
}

// setupRoutes 设置路由
func (s *Server) setupRoutes() {
	// 健康检查
	s.router.GET("/health", s.healthCheck)

	// 静态文件服务
	s.router.Static("/static", "./web/static")
	s.router.StaticFile("/", "./web/index.html")
	s.router.StaticFile("/favicon.ico", "./web/favicon.ico")

	// API路由组
	api := s.router.Group("/api/v1")
	{
		// 系统信息
		api.GET("/info", s.getSystemInfo)
		api.GET("/stats", s.getStats)

		// 文件上传
		api.POST("/upload", s.uploadFile)

		// 任务管理
		tasks := api.Group("/tasks")
		{
			tasks.POST("", s.createTask)
			tasks.GET("", s.listTasks)
			tasks.GET("/:id", s.getTask)
			tasks.DELETE("/:id", s.deleteTask)
			tasks.POST("/:id/cancel", s.cancelTask)
		}

		// 文件下载
		api.GET("/download/:filename", s.downloadFile)

		// WebSocket进度推送
		api.GET("/progress/:taskId", s.handleProgressWebSocket)
	}
}

// authMiddleware 认证中间件
func (s *Server) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			token = c.Query("token")
		}

		if token != s.config.Web.AuthToken {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "未授权访问",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// fileSizeLimitMiddleware 文件大小限制中间件
func (s *Server) fileSizeLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, s.config.Web.MaxFileSize)
		c.Next()
	}
}

// Start 启动服务器
func (s *Server) Start() error {
	// 确保上传和输出目录存在
	if err := s.ensureDirectories(); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	// 启动工作池
	s.workerPool.Start()

	// 启动定期清理任务
	go s.startCleanupRoutine()

	address := fmt.Sprintf("%s:%d", s.config.Web.Host, s.config.Web.Port)

	s.httpServer = &http.Server{
		Addr:         address,
		Handler:      s.router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: time.Duration(s.config.Web.TaskTimeout) * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	s.logger.Info("启动Web服务器",
		"address", address,
		"tls", s.config.Web.EnableTLS,
		"auth", s.config.Web.EnableAuth,
		"cors", s.config.Web.EnableCORS,
	)

	// 启动服务器
	var err error
	if s.config.Web.EnableTLS {
		err = s.httpServer.ListenAndServeTLS(s.config.Web.TLSCert, s.config.Web.TLSKey)
	} else {
		err = s.httpServer.ListenAndServe()
	}

	if err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("启动服务器失败: %w", err)
	}

	return nil
}

// Stop 停止服务器
func (s *Server) Stop() error {
	s.logger.Info("正在停止Web服务器...")

	// 停止工作池
	s.workerPool.Stop()

	// 清理进度报告器
	s.progressReporter.CleanupSubscribers()

	// 停止HTTP服务器
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("停止服务器失败: %w", err)
	}

	s.logger.Info("Web服务器已停止")
	return nil
}

// ensureDirectories 确保目录存在
func (s *Server) ensureDirectories() error {
	dirs := []string{
		s.config.Web.UploadDir,
		s.config.Web.OutputDir,
		"web/static",
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("创建目录 %s 失败: %w", dir, err)
		}
	}

	return nil
}

// startCleanupRoutine 启动清理例程
func (s *Server) startCleanupRoutine() {
	ticker := time.NewTicker(time.Duration(s.config.Web.CleanupInterval) * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		// 清理旧任务
		olderThan := 24 * time.Hour // 清理24小时前的任务
		if err := s.taskManager.CleanupOldTasks(olderThan); err != nil {
			s.logger.Error("清理旧任务失败", "error", err)
		}

		// 清理临时文件
		s.cleanupTempFiles()
	}
}

// cleanupTempFiles 清理临时文件
func (s *Server) cleanupTempFiles() {
	// 清理上传目录中的旧文件
	cutoff := time.Now().Add(-24 * time.Hour)

	err := filepath.Walk(s.config.Web.UploadDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // 忽略错误，继续清理
		}

		if !info.IsDir() && info.ModTime().Before(cutoff) {
			if removeErr := os.Remove(path); removeErr != nil {
				s.logger.Warn("删除临时文件失败", "file", path, "error", removeErr)
			}
		}

		return nil
	})

	if err != nil {
		s.logger.Warn("清理临时文件失败", "error", err)
	}
}

// RunWithGracefulShutdown 运行服务器并支持优雅关闭
func (s *Server) RunWithGracefulShutdown() error {
	// 启动服务器
	errChan := make(chan error, 1)
	go func() {
		errChan <- s.Start()
	}()

	// 等待停止信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errChan:
		return err
	case sig := <-sigChan:
		s.logger.Info("接收到停止信号", "signal", sig)
		return s.Stop()
	}
}
