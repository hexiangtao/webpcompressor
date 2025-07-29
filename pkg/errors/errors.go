package errors

import (
	"fmt"
	"runtime"
	"strings"
)

// ErrorType 错误类型
type ErrorType string

const (
	ErrorTypeValidation    ErrorType = "VALIDATION"
	ErrorTypeIO            ErrorType = "IO"
	ErrorTypeExecution     ErrorType = "EXECUTION"
	ErrorTypeConfiguration ErrorType = "CONFIGURATION"
	ErrorTypeInternal      ErrorType = "INTERNAL"
	ErrorTypeExternal      ErrorType = "EXTERNAL"
)

// AppError 应用程序错误
type AppError struct {
	Type       ErrorType              `json:"type"`
	Code       string                 `json:"code"`
	Message    string                 `json:"message"`
	Details    string                 `json:"details,omitempty"`
	Cause      error                  `json:"-"`
	StackTrace string                 `json:"stack_trace,omitempty"`
	Context    map[string]interface{} `json:"context,omitempty"`
}

// Error 实现error接口
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s:%s] %s: %v", e.Type, e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s:%s] %s", e.Type, e.Code, e.Message)
}

// Unwrap 支持errors.Unwrap
func (e *AppError) Unwrap() error {
	return e.Cause
}

// WithContext 添加上下文信息
func (e *AppError) WithContext(key string, value interface{}) *AppError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// WithDetails 添加详细信息
func (e *AppError) WithDetails(details string) *AppError {
	e.Details = details
	return e
}

// New 创建新的应用程序错误
func New(errorType ErrorType, code, message string) *AppError {
	return &AppError{
		Type:       errorType,
		Code:       code,
		Message:    message,
		StackTrace: getStackTrace(),
	}
}

// Wrap 包装现有错误
func Wrap(err error, errorType ErrorType, code, message string) *AppError {
	return &AppError{
		Type:       errorType,
		Code:       code,
		Message:    message,
		Cause:      err,
		StackTrace: getStackTrace(),
	}
}

// Wrapf 格式化包装错误
func Wrapf(err error, errorType ErrorType, code, format string, args ...interface{}) *AppError {
	return Wrap(err, errorType, code, fmt.Sprintf(format, args...))
}

// getStackTrace 获取调用栈
func getStackTrace() string {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])

	var sb strings.Builder
	frames := runtime.CallersFrames(pcs[:n])

	for {
		frame, more := frames.Next()
		sb.WriteString(fmt.Sprintf("%s:%d %s\n", frame.File, frame.Line, frame.Function))
		if !more {
			break
		}
	}

	return sb.String()
}

// 预定义常见错误
var (
	// 验证错误
	ErrInvalidQuality = New(ErrorTypeValidation, "INVALID_QUALITY", "质量参数必须在0-100之间")
	ErrInvalidInput   = New(ErrorTypeValidation, "INVALID_INPUT", "输入参数无效")
	ErrEmptyInput     = New(ErrorTypeValidation, "EMPTY_INPUT", "输入不能为空")

	// IO错误
	ErrFileNotFound      = New(ErrorTypeIO, "FILE_NOT_FOUND", "文件不存在")
	ErrFileNotReadable   = New(ErrorTypeIO, "FILE_NOT_READABLE", "文件不可读")
	ErrFileNotWritable   = New(ErrorTypeIO, "FILE_NOT_WRITABLE", "文件不可写")
	ErrDirectoryCreation = New(ErrorTypeIO, "DIRECTORY_CREATION", "无法创建目录")

	// 执行错误
	ErrToolNotFound     = New(ErrorTypeExecution, "TOOL_NOT_FOUND", "工具不存在")
	ErrCommandFailed    = New(ErrorTypeExecution, "COMMAND_FAILED", "命令执行失败")
	ErrTimeout          = New(ErrorTypeExecution, "TIMEOUT", "操作超时")
	ErrProcessingFailed = New(ErrorTypeExecution, "PROCESSING_FAILED", "处理失败")

	// 配置错误
	ErrConfigInvalid  = New(ErrorTypeConfiguration, "CONFIG_INVALID", "配置无效")
	ErrConfigNotFound = New(ErrorTypeConfiguration, "CONFIG_NOT_FOUND", "配置文件不存在")

	// 内部错误
	ErrInternal       = New(ErrorTypeInternal, "INTERNAL", "内部错误")
	ErrNotImplemented = New(ErrorTypeInternal, "NOT_IMPLEMENTED", "功能未实现")
)

// IsType 检查错误类型
func IsType(err error, errorType ErrorType) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Type == errorType
	}
	return false
}

// IsCode 检查错误代码
func IsCode(err error, code string) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code == code
	}
	return false
}

// GetType 获取错误类型
func GetType(err error) ErrorType {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Type
	}
	return ErrorTypeInternal
}

// GetCode 获取错误代码
func GetCode(err error) string {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code
	}
	return "UNKNOWN"
}
