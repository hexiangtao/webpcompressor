package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"webpcompressor/internal/config"
	"webpcompressor/internal/domain"
	"webpcompressor/pkg/logger"
)

// MockToolExecutor 模拟工具执行器
type MockToolExecutor struct {
	commands []string
	outputs  map[string]string
	errors   map[string]error
}

func NewMockToolExecutor() *MockToolExecutor {
	return &MockToolExecutor{
		commands: make([]string, 0),
		outputs:  make(map[string]string),
		errors:   make(map[string]error),
	}
}

func (m *MockToolExecutor) ExecuteCommand(ctx context.Context, toolName string, args ...string) error {
	key := toolName + " " + strings.Join(args, " ")
	m.commands = append(m.commands, key)
	if err, exists := m.errors[key]; exists {
		return err
	}
	return nil
}

func (m *MockToolExecutor) ExecuteCommandWithOutput(ctx context.Context, toolName string, args ...string) (string, error) {
	key := toolName + " " + strings.Join(args, " ")
	m.commands = append(m.commands, key)
	if err, exists := m.errors[key]; exists {
		return "", err
	}
	if output, exists := m.outputs[key]; exists {
		return output, nil
	}
	return "", nil
}

func (m *MockToolExecutor) GetToolPath(toolName string) string {
	return toolName + ".exe"
}

func (m *MockToolExecutor) IsToolAvailable(toolName string) bool {
	return true
}

func (m *MockToolExecutor) SetMockOutput(command, output string) {
	m.outputs[command] = output
}

func (m *MockToolExecutor) SetMockError(command string, err error) {
	m.errors[command] = err
}

// MockFileManager 模拟文件管理器
type MockFileManager struct {
	files     map[string]bool
	fileSizes map[string]int64
	tempDirs  []string
}

func NewMockFileManager() *MockFileManager {
	return &MockFileManager{
		files:     make(map[string]bool),
		fileSizes: make(map[string]int64),
		tempDirs:  make([]string, 0),
	}
}

func (m *MockFileManager) CreateTempDir(prefix string) (string, error) {
	dir := filepath.Join(os.TempDir(), prefix+"_test")
	m.tempDirs = append(m.tempDirs, dir)
	return dir, nil
}

func (m *MockFileManager) CleanupTempDir(path string) error {
	return nil
}

func (m *MockFileManager) GetFileSize(path string) (int64, error) {
	if size, exists := m.fileSizes[path]; exists {
		return size, nil
	}
	return 1024, nil // 默认大小
}

func (m *MockFileManager) FileExists(path string) bool {
	if exists, ok := m.files[path]; ok {
		return exists
	}
	return true // 默认存在
}

func (m *MockFileManager) CopyFile(src, dst string) error {
	return nil
}

func (m *MockFileManager) SetFileExists(path string, exists bool) {
	m.files[path] = exists
}

func (m *MockFileManager) SetFileSize(path string, size int64) {
	m.fileSizes[path] = size
}

func createTestWebPService() *WebPService {
	cfg := config.DefaultConfig()
	logger := logger.NewDefaultLogger()
	toolExecutor := NewMockToolExecutor()
	fileManager := NewMockFileManager()

	return NewWebPService(cfg, toolExecutor, fileManager, logger)
}

func TestParseAnimation_Success(t *testing.T) {
	service := createTestWebPService()
	mockToolExecutor := service.toolExecutor.(*MockToolExecutor)

	// 模拟webpmux -info输出
	mockOutput := `Canvas size: 288 x 288
Features present: animation transparency
Background color : 0xFFFFFFFF
Loop Count : 0
Number of frames: 3
No.: width height alpha x_offset y_offset duration dispose blend image_size compression
  1:    172      1   yes        58      284       50    none    no        172      lossy
  2:    183     12   yes        52      276       50 background   yes        518      lossy
  3:    180     20   yes        54      268       50    none    no       1116      lossy`

	mockToolExecutor.SetMockOutput("webpmux -info test.webp", mockOutput)

	ctx := context.Background()
	animInfo, err := service.ParseAnimation(ctx, "test.webp")

	if err != nil {
		t.Fatalf("ParseAnimation failed: %v", err)
	}

	if animInfo.Width != 288 {
		t.Errorf("Expected width 288, got %d", animInfo.Width)
	}

	if animInfo.Height != 288 {
		t.Errorf("Expected height 288, got %d", animInfo.Height)
	}

	if len(animInfo.Frames) != 3 {
		t.Errorf("Expected 3 frames, got %d", len(animInfo.Frames))
	}

	// 测试第一帧
	frame1 := animInfo.Frames[0]
	if frame1.Index != 1 {
		t.Errorf("Expected frame index 1, got %d", frame1.Index)
	}
	if frame1.X != 58 {
		t.Errorf("Expected X offset 58, got %d", frame1.X)
	}
	if frame1.Y != 284 {
		t.Errorf("Expected Y offset 284, got %d", frame1.Y)
	}
	if frame1.Duration != 50*time.Millisecond {
		t.Errorf("Expected duration 50ms, got %v", frame1.Duration)
	}
	if frame1.Dispose != domain.DisposeNone {
		t.Errorf("Expected dispose none, got %v", frame1.Dispose)
	}
	if frame1.Blend != domain.BlendNo {
		t.Errorf("Expected blend no, got %v", frame1.Blend)
	}
}

func TestParseAnimation_NoFrames(t *testing.T) {
	service := createTestWebPService()
	mockToolExecutor := service.toolExecutor.(*MockToolExecutor)

	// 模拟没有帧的输出
	mockOutput := `Canvas size: 288 x 288
Features present: animation transparency
Background color : 0xFFFFFFFF
Loop Count : 0
Number of frames: 0`

	mockToolExecutor.SetMockOutput("webpmux -info test.webp", mockOutput)

	ctx := context.Background()
	_, err := service.ParseAnimation(ctx, "test.webp")

	if err == nil {
		t.Error("Expected error for no frames, got nil")
	}
}

func TestCompressFrames_Success(t *testing.T) {
	service := createTestWebPService()
	mockFileManager := service.fileManager.(*MockFileManager)

	// 创建测试帧
	frames := []*domain.FrameInfo{
		{Index: 1, Path: "frame_1.webp"},
		{Index: 2, Path: "frame_2.webp"},
	}

	// 设置文件存在
	for _, frame := range frames {
		mockFileManager.SetFileExists(frame.Path, true)
	}

	config := domain.DefaultCompressionConfig(50)
	ctx := context.Background()

	err := service.CompressFrames(ctx, frames, config)

	if err != nil {
		t.Fatalf("CompressFrames failed: %v", err)
	}

	// 检查帧路径是否已更新
	for _, frame := range frames {
		if !strings.Contains(frame.Path, "compressed") {
			t.Errorf("Frame path not updated: %s", frame.Path)
		}
	}
}

func TestValidateInput_InvalidQuality(t *testing.T) {
	service := createTestWebPService()

	config := &domain.CompressionConfig{Quality: 150} // 无效质量

	err := service.validateInput("test.webp", "output.webp", config)

	if err == nil {
		t.Error("Expected error for invalid quality, got nil")
	}
}

func TestValidateInput_FileNotExists(t *testing.T) {
	service := createTestWebPService()
	mockFileManager := service.fileManager.(*MockFileManager)

	mockFileManager.SetFileExists("nonexistent.webp", false)

	config := domain.DefaultCompressionConfig(50)

	err := service.validateInput("nonexistent.webp", "output.webp", config)

	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

func BenchmarkParseAnimation(b *testing.B) {
	service := createTestWebPService()
	mockToolExecutor := service.toolExecutor.(*MockToolExecutor)

	// 模拟大型动画输出
	mockOutput := `Canvas size: 1920 x 1080
Number of frames: 100
No.: width height alpha x_offset y_offset duration dispose blend image_size compression`

	for i := 1; i <= 100; i++ {
		mockOutput += fmt.Sprintf("\n%3d:   1920   1080   yes         0        0       33    none    no     %d      lossy", i, 1000+i)
	}

	mockToolExecutor.SetMockOutput("webpmux -info test.webp", mockOutput)

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.ParseAnimation(ctx, "test.webp")
		if err != nil {
			b.Fatalf("ParseAnimation failed: %v", err)
		}
	}
}
