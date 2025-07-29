package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"webpcompressor/internal/config"
)

// Logger 接口定义
type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Fatal(msg string, args ...interface{})

	With(args ...interface{}) Logger
	WithError(err error) Logger
	WithContext(ctx map[string]interface{}) Logger
}

// StructuredLogger 结构化日志记录器
type StructuredLogger struct {
	logger *slog.Logger
	level  slog.Level
}

// NewLogger 创建新的日志记录器
func NewLogger(cfg *config.LoggingConfig) (Logger, error) {
	level := parseLogLevel(cfg.Level)

	var writer io.Writer
	if cfg.OutputFile == "" {
		writer = os.Stdout
	} else {
		// 文件输出
		file, err := openLogFile(cfg.OutputFile)
		if err != nil {
			return nil, fmt.Errorf("打开日志文件失败: %w", err)
		}
		writer = file
	}

	// 创建带有时间戳和格式化的处理器
	handler := slog.NewTextHandler(writer, &slog.HandlerOptions{
		Level:     level,
		AddSource: true,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// 自定义时间格式
			if a.Key == slog.TimeKey {
				a.Value = slog.StringValue(time.Now().Format("2006-01-02 15:04:05"))
			}
			return a
		},
	})

	return &StructuredLogger{
		logger: slog.New(handler),
		level:  level,
	}, nil
}

// NewDefaultLogger 创建默认日志记录器
func NewDefaultLogger() Logger {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: false,
	})

	return &StructuredLogger{
		logger: slog.New(handler),
		level:  slog.LevelInfo,
	}
}

// Debug 记录调试信息
func (l *StructuredLogger) Debug(msg string, args ...interface{}) {
	l.logger.Debug(msg, args...)
}

// Info 记录一般信息
func (l *StructuredLogger) Info(msg string, args ...interface{}) {
	l.logger.Info(msg, args...)
}

// Warn 记录警告信息
func (l *StructuredLogger) Warn(msg string, args ...interface{}) {
	l.logger.Warn(msg, args...)
}

// Error 记录错误信息
func (l *StructuredLogger) Error(msg string, args ...interface{}) {
	l.logger.Error(msg, args...)
}

// Fatal 记录致命错误并退出
func (l *StructuredLogger) Fatal(msg string, args ...interface{}) {
	l.logger.Error(msg, args...)
	os.Exit(1)
}

// With 添加结构化字段
func (l *StructuredLogger) With(args ...interface{}) Logger {
	return &StructuredLogger{
		logger: l.logger.With(args...),
		level:  l.level,
	}
}

// WithError 添加错误字段
func (l *StructuredLogger) WithError(err error) Logger {
	return l.With("error", err.Error())
}

// WithContext 添加上下文字段
func (l *StructuredLogger) WithContext(ctx map[string]interface{}) Logger {
	args := make([]interface{}, 0, len(ctx)*2)
	for k, v := range ctx {
		args = append(args, k, v)
	}
	return l.With(args...)
}

// parseLogLevel 解析日志级别
func parseLogLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// openLogFile 打开日志文件
func openLogFile(path string) (*os.File, error) {
	// 确保目录存在
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	// 打开或创建文件
	return os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
}

// ProgressLogger 进度日志记录器
type ProgressLogger struct {
	logger  Logger
	total   int
	current int
	prefix  string
}

// NewProgressLogger 创建进度日志记录器
func NewProgressLogger(logger Logger, total int, prefix string) *ProgressLogger {
	return &ProgressLogger{
		logger: logger,
		total:  total,
		prefix: prefix,
	}
}

// Update 更新进度
func (p *ProgressLogger) Update(current int) {
	p.current = current
	percentage := float64(current) / float64(p.total) * 100

	p.logger.Info(fmt.Sprintf("%s进度", p.prefix),
		"current", current,
		"total", p.total,
		"percentage", fmt.Sprintf("%.1f%%", percentage),
	)
}

// Increment 增加进度
func (p *ProgressLogger) Increment() {
	p.Update(p.current + 1)
}

// Finish 完成进度
func (p *ProgressLogger) Finish() {
	p.logger.Info(fmt.Sprintf("%s完成", p.prefix),
		"total", p.total,
	)
}

// LogOperation 记录操作日志的辅助函数
type OperationLogger struct {
	logger    Logger
	operation string
	startTime time.Time
	context   map[string]interface{}
}

// NewOperationLogger 创建操作日志记录器
func NewOperationLogger(logger Logger, operation string) *OperationLogger {
	return &OperationLogger{
		logger:    logger,
		operation: operation,
		startTime: time.Now(),
		context:   make(map[string]interface{}),
	}
}

// WithContext 添加上下文
func (o *OperationLogger) WithContext(key string, value interface{}) *OperationLogger {
	o.context[key] = value
	return o
}

// Start 开始操作
func (o *OperationLogger) Start() {
	o.logger.WithContext(o.context).Info(fmt.Sprintf("开始%s", o.operation))
}

// Success 操作成功
func (o *OperationLogger) Success() {
	duration := time.Since(o.startTime)
	ctx := make(map[string]interface{})
	for k, v := range o.context {
		ctx[k] = v
	}
	ctx["duration"] = duration.String()

	o.logger.WithContext(ctx).Info(fmt.Sprintf("完成%s", o.operation))
}

// Error 操作失败
func (o *OperationLogger) Error(err error) {
	duration := time.Since(o.startTime)
	ctx := make(map[string]interface{})
	for k, v := range o.context {
		ctx[k] = v
	}
	ctx["duration"] = duration.String()
	ctx["error"] = err.Error()

	o.logger.WithContext(ctx).Error(fmt.Sprintf("失败%s", o.operation))
}
