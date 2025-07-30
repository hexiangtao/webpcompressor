package web

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"webpcompressor/internal/domain"
	"webpcompressor/internal/infrastructure"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// healthCheck 健康检查
func (s *Server) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now(),
		"version":   s.config.App.Version,
	})
}

// getSystemInfo 获取系统信息
func (s *Server) getSystemInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"app": gin.H{
			"name":    s.config.App.Name,
			"version": s.config.App.Version,
		},
		"server": gin.H{
			"max_file_size":        s.config.Web.MaxFileSize,
			"max_concurrent_tasks": s.config.Web.MaxConcurrentTasks,
			"task_timeout":         s.config.Web.TaskTimeout,
		},
		"processing": gin.H{
			"enable_parallel": s.config.Processing.EnableParallel,
			"max_workers":     s.config.Processing.MaxWorkers,
			"default_preset":  s.config.Processing.DefaultPreset,
		},
		"presets":          s.config.Advanced.CompressionPresets,
		"quality_profiles": s.config.Advanced.QualityProfiles,
	})
}

// getStats 获取统计信息
func (s *Server) getStats(c *gin.Context) {
	if taskManager, ok := s.taskManager.(*infrastructure.MemoryTaskManager); ok {
		stats := taskManager.GetStats()
		c.JSON(http.StatusOK, stats)
	} else {
		c.JSON(http.StatusOK, &domain.WebPStats{})
	}
}

// uploadFile 文件上传
func (s *Server) uploadFile(c *gin.Context) {
	startTime := time.Now()
	s.logger.Info("开始文件上传处理", 
		"content_type", c.GetHeader("Content-Type"),
		"content_length", c.GetHeader("Content-Length"),
		"method", c.Request.Method)

	// 设置临时目录为当前工作目录的temp文件夹，避免系统临时目录性能问题
	tempDir := "./temp"
	os.MkdirAll(tempDir, 0755)
	os.Setenv("TMPDIR", tempDir)
	os.Setenv("TMP", tempDir) 
	os.Setenv("TEMP", tempDir)
	
	s.logger.Info("开始获取FormFile", "temp_dir", tempDir)
	
	// 直接使用ShouldBind会更快，避免ParseMultipartForm
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		s.logger.Error("获取FormFile失败", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "文件上传失败",
			"details": err.Error(),
		})
		return
	}
	defer file.Close()

	s.logger.Info("获取FormFile完成", "耗时", time.Since(startTime))

	// 验证文件类型
	validateStart := time.Now()
	if !s.isValidWebPFile(header.Filename) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "仅支持WebP文件格式",
		})
		return
	}
	s.logger.Info("文件验证完成", "耗时", time.Since(validateStart))

	// 生成唯一文件名
	filename := fmt.Sprintf("%s_%s", uuid.New().String(), header.Filename)
	uploadPath := filepath.Join(s.config.Web.UploadDir, filename)
	s.logger.Info("生成文件路径", "path", uploadPath)

	// 创建目标文件
	dst, err := os.Create(uploadPath)
	if err != nil {
		s.logger.Error("创建上传文件失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "保存文件失败",
		})
		return
	}
	defer dst.Close()

	// 复制文件内容 (使用更大的缓冲区提高性能)
	copyStart := time.Now()
	buffer := make([]byte, 1024*1024) // 1MB缓冲区
	copied, err := io.CopyBuffer(dst, file, buffer)
	if err != nil {
		s.logger.Error("复制文件内容失败", "error", err)
		os.Remove(uploadPath) // 清理失败的文件
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "保存文件失败",
		})
		return
	}
	copyDuration := time.Since(copyStart)
	speed := float64(copied) / copyDuration.Seconds() / 1024 / 1024 // MB/s
	s.logger.Info("文件复制完成", "大小", copied, "耗时", copyDuration, "速度", fmt.Sprintf("%.2f MB/s", speed))

	// 获取文件信息
	fileInfo, _ := dst.Stat()

	uploadInfo := &domain.FileUploadInfo{
		Filename:     filename,
		OriginalName: header.Filename,
		Size:         fileInfo.Size(),
		ContentType:  header.Header.Get("Content-Type"),
		UploadedAt:   time.Now(),
		Path:         uploadPath,
	}

	totalDuration := time.Since(startTime)
	s.logger.Info("文件上传成功",
		"filename", filename,
		"original_name", header.Filename,
		"总耗时", totalDuration,
		"size", fileInfo.Size(),
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "文件上传成功",
		"file":    uploadInfo,
	})
}

// createTask 创建压缩任务
func (s *Server) createTask(c *gin.Context) {
	var req struct {
		InputFile  string `json:"input_file" binding:"required"`
		Quality    int    `json:"quality" binding:"required,min=0,max=100"`
		Preset     string `json:"preset,omitempty"`
		Lossless   bool   `json:"lossless,omitempty"`
		Parallel   bool   `json:"parallel,omitempty"`
		OutputName string `json:"output_name,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求参数错误",
			"details": err.Error(),
		})
		return
	}

	// 验证输入文件是否存在
	inputPath := filepath.Join(s.config.Web.UploadDir, req.InputFile)
	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "输入文件不存在",
		})
		return
	}

	// 生成输出文件名
	outputFile := req.OutputName
	if outputFile == "" {
		ext := filepath.Ext(req.InputFile)
		base := strings.TrimSuffix(req.InputFile, ext)
		outputFile = fmt.Sprintf("%s_compressed_%d%s", base, req.Quality, ext)
	}
	outputPath := filepath.Join(s.config.Web.OutputDir, outputFile)

	// 创建压缩配置
	config := domain.DefaultCompressionConfig(req.Quality)
	if req.Preset != "" {
		config.Preset = req.Preset
	}
	config.Lossless = req.Lossless
	config.EnableParallel = req.Parallel && s.config.Processing.EnableParallel

	// 创建任务
	task, err := s.taskManager.CreateTask(inputPath, outputPath, config)
	if err != nil {
		s.logger.Error("创建任务失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "创建任务失败",
			"details": err.Error(),
		})
		return
	}

	// 提交任务到工作池
	s.workerPool.SubmitTask(task, s.webpService, s.taskManager, s.progressReporter)

	c.JSON(http.StatusCreated, gin.H{
		"message": "任务创建成功",
		"task":    task,
	})
}

// getTask 获取任务信息
func (s *Server) getTask(c *gin.Context) {
	taskID := c.Param("id")

	task, err := s.taskManager.GetTask(taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "任务不存在",
		})
		return
	}

	c.JSON(http.StatusOK, task)
}

// listTasks 列出任务
func (s *Server) listTasks(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100 // 限制最大值
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	tasks, err := s.taskManager.ListTasks(limit, offset)
	if err != nil {
		s.logger.Error("获取任务列表失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取任务列表失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tasks":  tasks,
		"limit":  limit,
		"offset": offset,
		"count":  len(tasks),
	})
}

// deleteTask 删除任务
func (s *Server) deleteTask(c *gin.Context) {
	taskID := c.Param("id")

	// 获取任务信息
	task, err := s.taskManager.GetTask(taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "任务不存在",
		})
		return
	}

	// 检查任务状态
	if task.Status == domain.TaskStatusProcessing {
		c.JSON(http.StatusConflict, gin.H{
			"error": "无法删除正在处理的任务",
		})
		return
	}

	// 删除任务
	if err := s.taskManager.DeleteTask(taskID); err != nil {
		s.logger.Error("删除任务失败", "task_id", taskID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "删除任务失败",
		})
		return
	}

	// 清理相关文件
	s.cleanupTaskFiles(task)

	c.JSON(http.StatusOK, gin.H{
		"message": "任务删除成功",
	})
}

// cancelTask 取消任务
func (s *Server) cancelTask(c *gin.Context) {
	taskID := c.Param("id")

	task, err := s.taskManager.GetTask(taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "任务不存在",
		})
		return
	}

	if task.Status != domain.TaskStatusPending && task.Status != domain.TaskStatusProcessing {
		c.JSON(http.StatusConflict, gin.H{
			"error": "任务无法取消",
		})
		return
	}

	// 取消任务
	task.Cancel()
	if err := s.taskManager.UpdateTask(task); err != nil {
		s.logger.Error("更新任务状态失败", "task_id", taskID, "error", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "任务已取消",
		"task":    task,
	})
}

// downloadFile 文件下载
func (s *Server) downloadFile(c *gin.Context) {
	filename := c.Param("filename")

	// 安全检查：防止路径遍历攻击
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的文件名",
		})
		return
	}

	filePath := filepath.Join(s.config.Web.OutputDir, filename)

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "文件不存在",
		})
		return
	}

	s.logger.Info("文件下载", "filename", filename)

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.File(filePath)
}

// isValidWebPFile 验证是否为有效的WebP文件
func (s *Server) isValidWebPFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".webp"
}

// cleanupTaskFiles 清理任务相关文件
func (s *Server) cleanupTaskFiles(task *domain.TaskInfo) {
	// 清理输出文件
	if task.OutputFile != "" {
		if err := os.Remove(task.OutputFile); err != nil && !os.IsNotExist(err) {
			s.logger.Warn("清理输出文件失败", "file", task.OutputFile, "error", err)
		}
	}
}
