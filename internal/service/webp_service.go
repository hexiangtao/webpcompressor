package service

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"webpcompressor/internal/config"
	"webpcompressor/internal/domain"
	"webpcompressor/pkg/errors"
	"webpcompressor/pkg/logger"
)

// WebPService WebP处理服务
type WebPService struct {
	config       *config.Config
	toolExecutor domain.ToolExecutor
	fileManager  domain.FileManager
	logger       logger.Logger
}

// NewWebPService 创建WebP服务
func NewWebPService(
	cfg *config.Config,
	toolExecutor domain.ToolExecutor,
	fileManager domain.FileManager,
	logger logger.Logger,
) *WebPService {
	return &WebPService{
		config:       cfg,
		toolExecutor: toolExecutor,
		fileManager:  fileManager,
		logger:       logger,
	}
}

// CompressAnimation 压缩WebP动画
func (s *WebPService) CompressAnimation(ctx context.Context, inputPath, outputPath string, config *domain.CompressionConfig) (*domain.CompressResult, error) {
	opLogger := logger.NewOperationLogger(s.logger, "WebP动画压缩").
		WithContext("input", inputPath).
		WithContext("output", outputPath).
		WithContext("quality", config.Quality).
		WithContext("parallel", config.EnableParallel)

	opLogger.Start()
	startTime := time.Now()

	// 验证输入
	if err := s.validateInput(inputPath, outputPath, config); err != nil {
		opLogger.Error(err)
		return nil, err
	}

	// 获取原始文件大小
	originalSize, err := s.fileManager.GetFileSize(inputPath)
	if err != nil {
		err = errors.Wrap(err, errors.ErrorTypeIO, "GET_FILE_SIZE", "获取文件大小失败")
		opLogger.Error(err)
		return nil, err
	}

	// 解析动画信息
	animInfo, err := s.ParseAnimation(ctx, inputPath)
	if err != nil {
		opLogger.Error(err)
		return nil, err
	}

	// 创建临时目录
	tempDir, err := s.fileManager.CreateTempDir("webp_compress")
	if err != nil {
		err = errors.Wrap(err, errors.ErrorTypeIO, "CREATE_TEMP_DIR", "创建临时目录失败")
		opLogger.Error(err)
		return nil, err
	}
	defer s.fileManager.CleanupTempDir(tempDir)

	// 提取帧
	if err := s.ExtractFrames(ctx, inputPath, tempDir, animInfo.Frames); err != nil {
		opLogger.Error(err)
		return nil, err
	}

	// 压缩帧
	if err := s.CompressFrames(ctx, animInfo.Frames, config); err != nil {
		opLogger.Error(err)
		return nil, err
	}

	// 重新组装动画
	if err := s.AssembleAnimation(ctx, animInfo.Frames, outputPath); err != nil {
		opLogger.Error(err)
		return nil, err
	}

	// 获取压缩后文件大小
	compressedSize, err := s.fileManager.GetFileSize(outputPath)
	if err != nil {
		s.logger.Warn("获取压缩后文件大小失败", "error", err)
		compressedSize = 0
	}

	// 计算使用的并行工作者数量
	parallelWorkers := 1 // 默认顺序处理
	if config.EnableParallel && len(animInfo.Frames) > 1 {
		maxWorkers := config.MaxConcurrency
		if maxWorkers <= 0 {
			maxWorkers = s.config.App.MaxConcurrency
		}
		if maxWorkers > len(animInfo.Frames) {
			maxWorkers = len(animInfo.Frames)
		}
		parallelWorkers = maxWorkers
	}

	result := &domain.CompressResult{
		OriginalSize:    originalSize,
		CompressedSize:  compressedSize,
		ProcessingTime:  time.Since(startTime),
		FramesProcessed: len(animInfo.Frames),
		ParallelWorkers: parallelWorkers,
	}
	result.CalculateCompressionRatio()

	opLogger.Success()

	s.logger.Info("压缩完成",
		"original_size", formatFileSize(originalSize),
		"compressed_size", formatFileSize(compressedSize),
		"compression_ratio", fmt.Sprintf("%.1f%%", result.CompressionRatio),
		"frames", result.FramesProcessed,
		"duration", result.ProcessingTime,
		"parallel_workers", result.ParallelWorkers,
	)

	return result, nil
}

// ParseAnimation 解析WebP动画信息
func (s *WebPService) ParseAnimation(ctx context.Context, inputPath string) (*domain.AnimationInfo, error) {
	s.logger.Debug("开始解析动画信息", "file", inputPath)

	output, err := s.toolExecutor.ExecuteCommandWithOutput(ctx, "webpmux", "-info", inputPath)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeExecution, "PARSE_ANIMATION", "执行webpmux失败")
	}

	return s.parseWebpmuxOutput(output)
}

// ExtractFrames 提取动画帧
func (s *WebPService) ExtractFrames(ctx context.Context, inputPath string, outputDir string, frames []*domain.FrameInfo) error {
	s.logger.Info("开始提取帧", "total_frames", len(frames))

	progressLogger := logger.NewProgressLogger(s.logger, len(frames), "提取帧")

	for i, frame := range frames {
		frameOutput := filepath.Join(outputDir, fmt.Sprintf("frame_%d.webp", frame.Index))

		err := s.toolExecutor.ExecuteCommand(ctx, "webpmux",
			"-get", "frame", strconv.Itoa(frame.Index),
			"-o", frameOutput, inputPath)

		if err != nil {
			return errors.Wrapf(err, errors.ErrorTypeExecution, "EXTRACT_FRAME",
				"提取第%d帧失败", frame.Index)
		}

		// 检查文件是否成功创建
		if !s.fileManager.FileExists(frameOutput) {
			return errors.New(errors.ErrorTypeExecution, "FRAME_NOT_CREATED",
				fmt.Sprintf("第%d帧文件未成功创建: %s", frame.Index, frameOutput))
		}

		frame.Path = frameOutput
		s.logger.Debug("提取帧成功",
			"index", frame.Index,
			"output", frameOutput,
		)
		progressLogger.Update(i + 1)
	}

	progressLogger.Finish()
	return nil
}

// CompressFrames 压缩帧
func (s *WebPService) CompressFrames(ctx context.Context, frames []*domain.FrameInfo, config *domain.CompressionConfig) error {
	if config.EnableParallel && len(frames) > 1 {
		return s.CompressFramesParallel(ctx, frames, config)
	}
	return s.compressFramesSequential(ctx, frames, config)
}

// CompressFramesParallel 并行压缩帧
func (s *WebPService) CompressFramesParallel(ctx context.Context, frames []*domain.FrameInfo, config *domain.CompressionConfig) error {
	s.logger.Info("开始并行压缩帧",
		"total_frames", len(frames),
		"quality", config.Quality,
		"max_concurrency", config.MaxConcurrency,
	)

	// 限制并发数
	maxWorkers := config.MaxConcurrency
	if maxWorkers <= 0 {
		maxWorkers = s.config.App.MaxConcurrency
	}
	if maxWorkers > len(frames) {
		maxWorkers = len(frames)
	}

	// 创建工作池
	workerPool := domain.NewWorkerPool(maxWorkers)

	// 创建帧处理器
	frameProcessor := func(ctx context.Context, frame *domain.FrameInfo) error {
		return s.compressFrame(ctx, frame, config)
	}

	// 启动工作池
	workerPool.Start(ctx, frameProcessor)

	// 提交所有帧任务
	for _, frame := range frames {
		workerPool.Submit(frame)
	}

	// 关闭任务队列
	workerPool.Close()

	// 等待所有任务完成
	errors := workerPool.Wait()

	// 检查是否有错误
	if len(errors) > 0 {
		s.logger.Error("并行压缩出现错误", "error_count", len(errors))
		return errors[0] // 返回第一个错误
	}

	s.logger.Info("并行压缩完成",
		"workers", maxWorkers,
		"frames", len(frames),
	)

	return nil
}

// compressFramesSequential 顺序压缩帧（原有逻辑）
func (s *WebPService) compressFramesSequential(ctx context.Context, frames []*domain.FrameInfo, config *domain.CompressionConfig) error {
	s.logger.Info("开始顺序压缩帧", "total_frames", len(frames), "quality", config.Quality)

	progressLogger := logger.NewProgressLogger(s.logger, len(frames), "压缩帧")

	for i, frame := range frames {
		if err := s.compressFrame(ctx, frame, config); err != nil {
			return err
		}
		progressLogger.Update(i + 1)
	}

	progressLogger.Finish()
	return nil
}

// compressFrame 压缩单个帧
func (s *WebPService) compressFrame(ctx context.Context, frame *domain.FrameInfo, config *domain.CompressionConfig) error {
	// 检查输入文件是否存在
	if !s.fileManager.FileExists(frame.Path) {
		return errors.New(errors.ErrorTypeIO, "INPUT_FRAME_NOT_FOUND",
			fmt.Sprintf("输入帧文件不存在: %s", frame.Path))
	}

	compressedPath := strings.Replace(frame.Path, "frame_", "frame_compressed_", 1)

	args := s.buildCompressionArgs(config, frame.Path, compressedPath)

	err := s.toolExecutor.ExecuteCommand(ctx, "cwebp", args...)
	if err != nil {
		return errors.Wrapf(err, errors.ErrorTypeExecution, "COMPRESS_FRAME",
			"压缩第%d帧失败", frame.Index)
	}

	// 检查压缩后的文件是否成功创建
	if !s.fileManager.FileExists(compressedPath) {
		return errors.New(errors.ErrorTypeExecution, "COMPRESSED_FRAME_NOT_CREATED",
			fmt.Sprintf("第%d帧压缩文件未成功创建: %s", frame.Index, compressedPath))
	}

	frame.Path = compressedPath

	s.logger.Debug("压缩帧成功",
		"index", frame.Index,
		"output", compressedPath,
	)

	return nil
}

// AssembleAnimation 重新组装动画
func (s *WebPService) AssembleAnimation(ctx context.Context, frames []*domain.FrameInfo, outputPath string) error {
	s.logger.Info("开始重新组装动画", "output", outputPath)

	// 确保输出目录存在
	outputDir := filepath.Dir(outputPath)
	if outputDir != "." && outputDir != "" {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return errors.Wrap(err, errors.ErrorTypeIO, "CREATE_OUTPUT_DIR",
				fmt.Sprintf("创建输出目录失败: %s", outputDir))
		}
		s.logger.Debug("创建输出目录", "dir", outputDir)
	}

	// 验证所有帧文件是否存在
	for _, frame := range frames {
		if !s.fileManager.FileExists(frame.Path) {
			return errors.New(errors.ErrorTypeIO, "FRAME_FILE_NOT_FOUND",
				fmt.Sprintf("帧文件不存在: %s (索引: %d)", frame.Path, frame.Index))
		}

		// 检查文件大小
		if size, err := s.fileManager.GetFileSize(frame.Path); err != nil {
			s.logger.Warn("无法获取帧文件大小", "file", frame.Path, "error", err)
		} else if size == 0 {
			return errors.New(errors.ErrorTypeIO, "EMPTY_FRAME_FILE",
				fmt.Sprintf("帧文件为空: %s (索引: %d)", frame.Path, frame.Index))
		} else {
			s.logger.Debug("帧文件验证通过",
				"index", frame.Index,
				"path", frame.Path,
				"size", size,
			)
		}
	}

	args := []string{}
	for _, frame := range frames {
		blendStr := "-b"
		if frame.Blend == domain.BlendYes {
			blendStr = "+b"
		}

		// 正确的webpmux格式：file_i +di+xi+yi+mi+bi
		// 文件路径和参数应该分别作为独立的参数
		frameParams := fmt.Sprintf("+%d+%d+%d+%d%s",
			int(frame.Duration/time.Millisecond),
			frame.X, frame.Y,
			int(frame.Dispose), blendStr)

		args = append(args, "-frame", frame.Path, frameParams)

		// 添加调试信息
		s.logger.Debug("添加帧参数",
			"index", frame.Index,
			"path", frame.Path,
			"frame_params", frameParams,
			"duration_ms", int(frame.Duration/time.Millisecond),
			"x", frame.X,
			"y", frame.Y,
			"dispose", int(frame.Dispose),
			"blend", blendStr,
		)
	}
	args = append(args, "-loop", "0", "-o", outputPath)

	// 记录完整的命令
	s.logger.Info("执行webpmux命令",
		"args", strings.Join(args, " "),
		"total_frames", len(frames),
	)

	err := s.toolExecutor.ExecuteCommand(ctx, "webpmux", args...)
	if err != nil {
		return errors.Wrap(err, errors.ErrorTypeExecution, "ASSEMBLE_ANIMATION", "重新组装动画失败")
	}

	return nil
}

// parseWebpmuxOutput 解析webpmux输出
func (s *WebPService) parseWebpmuxOutput(output string) (*domain.AnimationInfo, error) {
	scanner := bufio.NewScanner(strings.NewReader(output))

	animInfo := &domain.AnimationInfo{
		Frames: make([]*domain.FrameInfo, 0),
	}

	startReading := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// 解析画布大小
		if strings.HasPrefix(line, "Canvas size:") {
			if _, err := fmt.Sscanf(line, "Canvas size: %d x %d", &animInfo.Width, &animInfo.Height); err != nil {
				s.logger.Warn("解析画布大小失败", "line", line)
			}
			continue
		}

		// 解析帧数
		if strings.HasPrefix(line, "Number of frames:") {
			if _, err := fmt.Sscanf(line, "Number of frames: %d", &animInfo.FrameCount); err != nil {
				s.logger.Warn("解析帧数失败", "line", line)
			}
			continue
		}

		// 检测表头
		if strings.HasPrefix(line, "No.") && strings.Contains(line, "duration") {
			startReading = true
			continue
		}

		// 解析帧信息
		if startReading {
			if line == "" {
				break
			}

			frame, err := s.parseFrameLine(line)
			if err != nil {
				s.logger.Warn("解析帧信息失败", "line", line, "error", err)
				continue
			}

			animInfo.Frames = append(animInfo.Frames, frame)
		}
	}

	if len(animInfo.Frames) == 0 {
		return nil, errors.New(errors.ErrorTypeValidation, "NO_FRAMES", "未能解析到任何帧")
	}

	s.logger.Debug("解析动画信息成功",
		"width", animInfo.Width,
		"height", animInfo.Height,
		"frames", len(animInfo.Frames),
	)

	return animInfo, nil
}

// parseFrameLine 解析单行帧信息
func (s *WebPService) parseFrameLine(line string) (*domain.FrameInfo, error) {
	fields := strings.Fields(line)
	if len(fields) < 9 {
		return nil, fmt.Errorf("字段数量不足: %d", len(fields))
	}

	// 解析各字段
	indexStr := strings.TrimSuffix(fields[0], ":")
	index, _ := strconv.Atoi(indexStr)
	x, _ := strconv.Atoi(fields[4])          // x_offset
	y, _ := strconv.Atoi(fields[5])          // y_offset
	durationMs, _ := strconv.Atoi(fields[6]) // duration

	// 处理dispose字段
	dispose := domain.DisposeNone
	if fields[7] == "background" {
		dispose = domain.DisposeBackground
	}

	// 处理blend字段
	blend := domain.BlendNo
	if fields[8] == "yes" {
		blend = domain.BlendYes
	}

	return &domain.FrameInfo{
		Index:    index,
		X:        x,
		Y:        y,
		Duration: time.Duration(durationMs) * time.Millisecond,
		Dispose:  dispose,
		Blend:    blend,
	}, nil
}

// buildCompressionArgs 构建压缩参数
func (s *WebPService) buildCompressionArgs(config *domain.CompressionConfig, inputPath, outputPath string) []string {
	args := []string{
		"-q", strconv.Itoa(config.Quality),
		"-m", strconv.Itoa(config.Method),
		"-preset", config.Preset,
		"-mt", // 多线程
		"-f", strconv.Itoa(config.FilterStrength),
		"-sharpness", "0",
		"-sns", "100",
		"-segments", "4",
		"-pass", "10",
		"-alpha_q", strconv.Itoa(config.AlphaQuality),
		"-size", "0",
		"-metadata", "none",
		inputPath,
		"-o", outputPath,
	}

	if config.Lossless {
		args = append([]string{"-lossless"}, args...)
	}

	return args
}

// GetFileManager 获取文件管理器
func (s *WebPService) GetFileManager() domain.FileManager {
	return s.fileManager
}

// validateInput 验证输入参数
func (s *WebPService) validateInput(inputPath, outputPath string, config *domain.CompressionConfig) error {
	// 检查输入文件
	if !s.fileManager.FileExists(inputPath) {
		return errors.ErrFileNotFound.WithContext("file", inputPath)
	}

	// 检查文件大小
	if size, err := s.fileManager.GetFileSize(inputPath); err == nil {
		if size > s.config.Advanced.OptimizationRules.MaxFileSize {
			return errors.New(errors.ErrorTypeValidation, "FILE_TOO_LARGE",
				fmt.Sprintf("文件大小超过限制: %s > %s",
					formatFileSize(size),
					formatFileSize(s.config.Advanced.OptimizationRules.MaxFileSize)))
		}
	}

	// 验证质量参数
	if config.Quality < 0 || config.Quality > 100 {
		return errors.ErrInvalidQuality.WithContext("quality", config.Quality)
	}

	// 验证输出路径目录
	outputDir := filepath.Dir(outputPath)
	if outputDir != "." && outputDir != "" {
		// 这里可以添加目录创建逻辑
	}

	return nil
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
