package errors

import (
	"fmt"
	"testing"
)

func TestNewError(t *testing.T) {
	err := New(ErrorTypeValidation, "INVALID_INPUT", "输入参数无效")

	if err.Type != ErrorTypeValidation {
		t.Errorf("Expected type %v, got %v", ErrorTypeValidation, err.Type)
	}

	if err.Code != "INVALID_INPUT" {
		t.Errorf("Expected code 'INVALID_INPUT', got '%s'", err.Code)
	}

	if err.Message != "输入参数无效" {
		t.Errorf("Expected message '输入参数无效', got '%s'", err.Message)
	}
}

func TestWrapError(t *testing.T) {
	originalErr := fmt.Errorf("原始错误")
	wrappedErr := Wrap(originalErr, ErrorTypeIO, "FILE_READ", "文件读取失败")

	if wrappedErr.Type != ErrorTypeIO {
		t.Errorf("Expected type %v, got %v", ErrorTypeIO, wrappedErr.Type)
	}

	if wrappedErr.Cause != originalErr {
		t.Errorf("Expected cause to be original error")
	}

	// 测试Unwrap功能
	unwrapped := wrappedErr.Unwrap()
	if unwrapped != originalErr {
		t.Errorf("Expected unwrapped error to be original error")
	}
}

func TestWrapfError(t *testing.T) {
	originalErr := fmt.Errorf("原始错误")
	wrappedErr := Wrapf(originalErr, ErrorTypeExecution, "COMMAND_FAILED", "命令执行失败: %s", "cwebp")

	expectedMessage := "命令执行失败: cwebp"
	if wrappedErr.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, wrappedErr.Message)
	}
}

func TestErrorWithContext(t *testing.T) {
	err := New(ErrorTypeValidation, "INVALID_QUALITY", "质量参数无效")

	errWithContext := err.WithContext("input_file", "test.webp").WithContext("quality", 150)

	if errWithContext.Context == nil {
		t.Error("Expected context to be set")
	}

	// 检查上下文值
	if inputFile := errWithContext.Context["input_file"]; inputFile != "test.webp" {
		t.Errorf("Expected input_file 'test.webp', got '%v'", inputFile)
	}

	if quality := errWithContext.Context["quality"]; quality != 150 {
		t.Errorf("Expected quality 150, got '%v'", quality)
	}
}

func TestErrorWithDetails(t *testing.T) {
	err := New(ErrorTypeIO, "FILE_NOT_FOUND", "文件未找到")

	detailsText := "文件路径: /path/to/file.webp, 大小: 1024, 存在: false"

	errWithDetails := err.WithDetails(detailsText)

	if errWithDetails.Details == "" {
		t.Error("Expected details to be set")
	}

	if errWithDetails.Details != detailsText {
		t.Errorf("Expected details '%s', got '%s'", detailsText, errWithDetails.Details)
	}
}

func TestErrorString(t *testing.T) {
	err := New(ErrorTypeValidation, "INVALID_QUALITY", "质量参数无效")

	expected := "[VALIDATION:INVALID_QUALITY] 质量参数无效"
	if err.Error() != expected {
		t.Errorf("Expected error string '%s', got '%s'", expected, err.Error())
	}
}

func TestErrorWithCause(t *testing.T) {
	originalErr := fmt.Errorf("系统错误")
	wrappedErr := Wrap(originalErr, ErrorTypeExecution, "TOOL_FAILED", "工具执行失败")

	errorString := wrappedErr.Error()
	if !contains(errorString, "工具执行失败") {
		t.Errorf("Expected error string to contain '工具执行失败', got '%s'", errorString)
	}

	if !contains(errorString, "系统错误") {
		t.Errorf("Expected error string to contain cause '系统错误', got '%s'", errorString)
	}
}

func TestErrorTypes(t *testing.T) {
	testCases := []struct {
		errorType ErrorType
		expected  string
	}{
		{ErrorTypeValidation, "VALIDATION"},
		{ErrorTypeIO, "IO"},
		{ErrorTypeExecution, "EXECUTION"},
		{ErrorTypeConfiguration, "CONFIGURATION"},
		{ErrorTypeInternal, "INTERNAL"},
		{ErrorTypeExternal, "EXTERNAL"},
	}

	for _, tc := range testCases {
		err := New(tc.errorType, "TEST_CODE", "测试消息")
		if !contains(err.Error(), tc.expected) {
			t.Errorf("Expected error type %s in error string, got %s", tc.expected, err.Error())
		}
	}
}

func TestCommonErrors(t *testing.T) {
	// 测试预定义的常见错误
	testCases := []struct {
		name     string
		errorVar *AppError
	}{
		{"ErrInvalidInput", ErrInvalidInput},
		{"ErrFileNotFound", ErrFileNotFound},
		{"ErrToolNotFound", ErrToolNotFound},
		{"ErrProcessingFailed", ErrProcessingFailed},
		{"ErrTimeout", ErrTimeout},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.errorVar == nil {
				t.Errorf("Expected %s to be defined", tc.name)
			}

			if tc.errorVar.Message == "" {
				t.Errorf("Expected %s to have a message", tc.name)
			}

			if tc.errorVar.Code == "" {
				t.Errorf("Expected %s to have a code", tc.name)
			}
		})
	}
}

func TestErrorChaining(t *testing.T) {
	// 测试错误链
	err1 := fmt.Errorf("底层错误")
	err2 := Wrap(err1, ErrorTypeIO, "IO_ERROR", "IO错误")
	err3 := Wrap(err2, ErrorTypeExecution, "EXEC_ERROR", "执行错误")

	// 检查错误链
	current := error(err3)
	depth := 0
	for current != nil {
		depth++
		if appErr, ok := current.(*AppError); ok {
			current = appErr.Unwrap()
		} else {
			break
		}

		if depth > 10 { // 防止无限循环
			t.Error("Error chain too deep, possible circular reference")
			break
		}
	}

	if depth != 3 {
		t.Errorf("Expected error chain depth 3, got %d", depth)
	}
}

func TestUtilityFunctions(t *testing.T) {
	err := New(ErrorTypeValidation, "TEST_CODE", "测试消息")

	// 测试 IsType
	if !IsType(err, ErrorTypeValidation) {
		t.Error("Expected IsType to return true for correct type")
	}

	if IsType(err, ErrorTypeIO) {
		t.Error("Expected IsType to return false for incorrect type")
	}

	// 测试 IsCode
	if !IsCode(err, "TEST_CODE") {
		t.Error("Expected IsCode to return true for correct code")
	}

	if IsCode(err, "WRONG_CODE") {
		t.Error("Expected IsCode to return false for incorrect code")
	}

	// 测试 GetType
	if GetType(err) != ErrorTypeValidation {
		t.Errorf("Expected GetType to return %v, got %v", ErrorTypeValidation, GetType(err))
	}

	// 测试 GetCode
	if GetCode(err) != "TEST_CODE" {
		t.Errorf("Expected GetCode to return 'TEST_CODE', got '%s'", GetCode(err))
	}
}

func BenchmarkNewError(b *testing.B) {
	for i := 0; i < b.N; i++ {
		New(ErrorTypeValidation, "BENCH_TEST", "基准测试错误")
	}
}

func BenchmarkWrapError(b *testing.B) {
	originalErr := fmt.Errorf("原始错误")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Wrap(originalErr, ErrorTypeIO, "BENCH_WRAP", "包装测试错误")
	}
}

// 辅助函数
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
