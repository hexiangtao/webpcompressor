package infrastructure

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"webpcompressor/internal/config"
	"webpcompressor/internal/domain"
	"webpcompressor/pkg/errors"
	"webpcompressor/pkg/logger"
)

// LocalToolExecutor 本地工具执行器
type LocalToolExecutor struct {
	config    *config.Config
	logger    logger.Logger
	toolPaths map[string]string
}

// NewLocalToolExecutor 创建本地工具执行器
func NewLocalToolExecutor(cfg *config.Config, logger logger.Logger) *LocalToolExecutor {
	executor := &LocalToolExecutor{
		config:    cfg,
		logger:    logger,
		toolPaths: make(map[string]string),
	}

	// 初始化工具路径
	executor.initializeToolPaths()

	return executor
}

// initializeToolPaths 初始化工具路径
func (e *LocalToolExecutor) initializeToolPaths() {
	for toolName, toolPath := range e.config.Tools.ToolPaths {
		e.toolPaths[toolName] = toolPath
	}

	// 验证工具可用性
	for toolName := range e.toolPaths {
		if e.IsToolAvailable(toolName) {
			e.logger.Debug("工具可用", "tool", toolName, "path", e.GetToolPath(toolName))
		} else {
			e.logger.Warn("工具不可用", "tool", toolName, "path", e.GetToolPath(toolName))
		}
	}
}

// ExecuteCommand 执行命令
func (e *LocalToolExecutor) ExecuteCommand(ctx context.Context, toolName string, args ...string) error {
	_, err := e.executeCommand(ctx, toolName, false, args...)
	return err
}

// ExecuteCommandWithOutput 执行命令并返回输出
func (e *LocalToolExecutor) ExecuteCommandWithOutput(ctx context.Context, toolName string, args ...string) (string, error) {
	return e.executeCommand(ctx, toolName, true, args...)
}

// executeCommand 执行命令的核心逻辑
func (e *LocalToolExecutor) executeCommand(ctx context.Context, toolName string, captureOutput bool, args ...string) (string, error) {
	toolPath := e.GetToolPath(toolName)

	// 创建带超时的上下文
	timeoutCtx, cancel := context.WithTimeout(ctx, e.config.App.Timeout)
	defer cancel()

	// 创建命令
	cmd := exec.CommandContext(timeoutCtx, toolPath, args...)

	// 设置工作目录
	if wd, err := os.Getwd(); err == nil {
		cmd.Dir = wd
	}

	e.logger.Debug("执行命令",
		"tool", toolName,
		"path", toolPath,
		"args", strings.Join(args, " "),
		"timeout", e.config.App.Timeout,
	)

	startTime := time.Now()

	var output string
	var err error

	if captureOutput {
		// 捕获输出
		outputBytes, execErr := cmd.Output()
		output = string(outputBytes)
		err = execErr

		// 如果出错，尝试获取标准错误输出
		if err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				stderr := string(exitError.Stderr)
				if stderr != "" {
					e.logger.Error("命令标准错误输出", "tool", toolName, "stderr", stderr)
					output = stderr // 将错误信息作为输出返回
				}
			}
		}
	} else {
		// 捕获标准错误以便调试
		var stderr strings.Builder
		cmd.Stderr = &stderr

		// 执行命令
		err = cmd.Run()

		// 如果出错，记录标准错误
		if err != nil && stderr.Len() > 0 {
			stderrOutput := stderr.String()
			e.logger.Error("命令标准错误输出", "tool", toolName, "stderr", stderrOutput)
		}
	}

	duration := time.Since(startTime)

	if err != nil {
		// 检查是否是超时错误
		if timeoutCtx.Err() == context.DeadlineExceeded {
			e.logger.Error("命令执行超时",
				"tool", toolName,
				"timeout", e.config.App.Timeout,
				"duration", duration,
			)
			return output, errors.Wrap(err, errors.ErrorTypeExecution, "COMMAND_TIMEOUT", "命令执行超时")
		}

		// 检查是否是工具不存在
		if isToolNotFoundError(err) {
			e.logger.Error("工具不存在",
				"tool", toolName,
				"path", toolPath,
			)
			return output, errors.Wrap(err, errors.ErrorTypeExecution, "TOOL_NOT_FOUND", "工具不存在")
		}

		e.logger.Error("命令执行失败",
			"tool", toolName,
			"error", err,
			"duration", duration,
		)
		return output, errors.Wrap(err, errors.ErrorTypeExecution, "COMMAND_FAILED", "命令执行失败")
	}

	e.logger.Debug("命令执行成功",
		"tool", toolName,
		"duration", duration,
	)

	return output, nil
}

// GetToolPath 获取工具路径
func (e *LocalToolExecutor) GetToolPath(toolName string) string {
	if path, exists := e.toolPaths[toolName]; exists {
		return path
	}
	return e.config.GetToolPath(toolName)
}

// IsToolAvailable 检查工具是否可用
func (e *LocalToolExecutor) IsToolAvailable(toolName string) bool {
	toolPath := e.GetToolPath(toolName)

	// 检查文件是否存在
	if _, err := os.Stat(toolPath); err == nil {
		return true
	}

	// 检查是否在PATH中
	if _, err := exec.LookPath(toolPath); err == nil {
		return true
	}

	return false
}

// isToolNotFoundError 检查是否是工具不存在错误
func isToolNotFoundError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	return strings.Contains(errStr, "not found") ||
		strings.Contains(errStr, "executable file not found") ||
		strings.Contains(errStr, "cannot find")
}

// EmbeddedToolExecutor 嵌入式工具执行器
type EmbeddedToolExecutor struct {
	*LocalToolExecutor
	tempDir   string
	extracted bool
}

// NewEmbeddedToolExecutor 创建嵌入式工具执行器
func NewEmbeddedToolExecutor(cfg *config.Config, logger logger.Logger, tempDir string) *EmbeddedToolExecutor {
	return &EmbeddedToolExecutor{
		LocalToolExecutor: NewLocalToolExecutor(cfg, logger),
		tempDir:           tempDir,
		extracted:         false,
	}
}

// GetToolPath 获取嵌入工具路径
func (e *EmbeddedToolExecutor) GetToolPath(toolName string) string {
	if e.tempDir != "" {
		// 构建临时目录中的工具路径
		toolFileName := e.config.GetToolPath(toolName)
		return filepath.Join(e.tempDir, filepath.Base(toolFileName))
	}

	// 回退到本地工具
	return e.LocalToolExecutor.GetToolPath(toolName)
}

// IsToolAvailable 检查嵌入工具是否可用
func (e *EmbeddedToolExecutor) IsToolAvailable(toolName string) bool {
	toolPath := e.GetToolPath(toolName)

	// 检查临时目录中的工具
	if e.tempDir != "" {
		if _, err := os.Stat(toolPath); err == nil {
			return true
		}
	}

	// 回退到本地工具检查
	return e.LocalToolExecutor.IsToolAvailable(toolName)
}

// ToolExecutorFactory 工具执行器工厂
type ToolExecutorFactory struct {
	config *config.Config
	logger logger.Logger
}

// NewToolExecutorFactory 创建工具执行器工厂
func NewToolExecutorFactory(cfg *config.Config, logger logger.Logger) *ToolExecutorFactory {
	return &ToolExecutorFactory{
		config: cfg,
		logger: logger,
	}
}

// CreateExecutor 创建工具执行器
func (f *ToolExecutorFactory) CreateExecutor(useEmbedded bool, tempDir string) domain.ToolExecutor {
	if useEmbedded && tempDir != "" {
		f.logger.Info("使用嵌入式工具执行器", "temp_dir", tempDir)
		return NewEmbeddedToolExecutor(f.config, f.logger, tempDir)
	}

	f.logger.Info("使用本地工具执行器")
	return NewLocalToolExecutor(f.config, f.logger)
}

// ValidateTools 验证工具可用性
func (f *ToolExecutorFactory) ValidateTools(executor domain.ToolExecutor) error {
	requiredTools := []string{"webpmux", "cwebp"}
	var missingTools []string

	for _, tool := range requiredTools {
		if !executor.IsToolAvailable(tool) {
			missingTools = append(missingTools, tool)
		}
	}

	if len(missingTools) > 0 {
		return errors.New(errors.ErrorTypeConfiguration, "TOOLS_MISSING",
			fmt.Sprintf("缺少必需的工具: %s", strings.Join(missingTools, ", ")))
	}

	f.logger.Info("所有必需工具都可用")
	return nil
}
