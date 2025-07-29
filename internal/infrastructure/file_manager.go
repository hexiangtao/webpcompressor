package infrastructure

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"webpcompressor/internal/config"
	"webpcompressor/internal/domain"
	"webpcompressor/pkg/errors"
	"webpcompressor/pkg/logger"
)

// LocalFileManager 本地文件管理器
type LocalFileManager struct {
	config *config.Config
	logger logger.Logger
}

// NewLocalFileManager 创建本地文件管理器
func NewLocalFileManager(cfg *config.Config, logger logger.Logger) domain.FileManager {
	return &LocalFileManager{
		config: cfg,
		logger: logger,
	}
}

// CreateTempDir 创建临时目录
func (f *LocalFileManager) CreateTempDir(prefix string) (string, error) {
	// 使用配置的临时目录或系统临时目录
	baseDir := f.config.App.TempDir
	if baseDir == "" {
		baseDir = os.TempDir()
	}

	// 确保基础目录存在
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return "", errors.Wrap(err, errors.ErrorTypeIO, "CREATE_BASE_DIR", "创建基础目录失败")
	}

	// 创建唯一的临时目录
	tempDir, err := os.MkdirTemp(baseDir, prefix+"_*")
	if err != nil {
		return "", errors.Wrap(err, errors.ErrorTypeIO, "CREATE_TEMP_DIR", "创建临时目录失败")
	}

	f.logger.Debug("创建临时目录", "path", tempDir)
	return tempDir, nil
}

// CleanupTempDir 清理临时目录
func (f *LocalFileManager) CleanupTempDir(path string) error {
	if path == "" {
		return nil
	}

	// 安全检查：确保删除的是临时目录
	if !f.isTempDir(path) {
		f.logger.Warn("拒绝删除非临时目录", "path", path)
		return errors.New(errors.ErrorTypeValidation, "NOT_TEMP_DIR", "拒绝删除非临时目录")
	}

	err := os.RemoveAll(path)
	if err != nil {
		f.logger.Warn("清理临时目录失败", "path", path, "error", err)
		return errors.Wrap(err, errors.ErrorTypeIO, "CLEANUP_TEMP_DIR", "清理临时目录失败")
	}

	f.logger.Debug("清理临时目录成功", "path", path)
	return nil
}

// GetFileSize 获取文件大小
func (f *LocalFileManager) GetFileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, errors.ErrFileNotFound.WithContext("file", path)
		}
		return 0, errors.Wrap(err, errors.ErrorTypeIO, "GET_FILE_INFO", "获取文件信息失败")
	}

	if info.IsDir() {
		return 0, errors.New(errors.ErrorTypeValidation, "IS_DIRECTORY", "路径是目录而不是文件")
	}

	return info.Size(), nil
}

// FileExists 检查文件是否存在
func (f *LocalFileManager) FileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// CopyFile 复制文件
func (f *LocalFileManager) CopyFile(src, dst string) error {
	// 检查源文件
	if !f.FileExists(src) {
		return errors.ErrFileNotFound.WithContext("file", src)
	}

	// 打开源文件
	srcFile, err := os.Open(src)
	if err != nil {
		return errors.Wrap(err, errors.ErrorTypeIO, "OPEN_SOURCE", "打开源文件失败")
	}
	defer srcFile.Close()

	// 确保目标目录存在
	dstDir := filepath.Dir(dst)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return errors.Wrap(err, errors.ErrorTypeIO, "CREATE_DST_DIR", "创建目标目录失败")
	}

	// 创建目标文件
	dstFile, err := os.Create(dst)
	if err != nil {
		return errors.Wrap(err, errors.ErrorTypeIO, "CREATE_DST_FILE", "创建目标文件失败")
	}
	defer dstFile.Close()

	// 复制文件内容
	written, err := io.Copy(dstFile, srcFile)
	if err != nil {
		return errors.Wrap(err, errors.ErrorTypeIO, "COPY_CONTENT", "复制文件内容失败")
	}

	f.logger.Debug("复制文件成功",
		"src", src,
		"dst", dst,
		"size", written,
	)

	return nil
}

// isTempDir 检查是否是临时目录
func (f *LocalFileManager) isTempDir(path string) bool {
	// 检查是否在配置的临时目录下
	if f.config.App.TempDir != "" {
		absConfigTemp, err := filepath.Abs(f.config.App.TempDir)
		if err == nil {
			absPath, err := filepath.Abs(path)
			if err == nil && strings.HasPrefix(absPath, absConfigTemp) {
				return true
			}
		}
	}

	// 检查是否在系统临时目录下
	sysTempDir := os.TempDir()
	absSysTemp, err := filepath.Abs(sysTempDir)
	if err == nil {
		absPath, err := filepath.Abs(path)
		if err == nil && strings.HasPrefix(absPath, absSysTemp) {
			return true
		}
	}

	// 检查目录名是否包含临时目录特征
	base := filepath.Base(path)
	return strings.Contains(base, "temp") ||
		strings.Contains(base, "tmp") ||
		strings.Contains(base, "webp")
}

// SafeFileManager 安全文件管理器包装器
type SafeFileManager struct {
	domain.FileManager
	logger logger.Logger
	config *config.Config
}

// NewSafeFileManager 创建安全文件管理器
func NewSafeFileManager(fm domain.FileManager, cfg *config.Config, logger logger.Logger) domain.FileManager {
	return &SafeFileManager{
		FileManager: fm,
		logger:      logger,
		config:      cfg,
	}
}

// GetFileSize 安全获取文件大小
func (s *SafeFileManager) GetFileSize(path string) (int64, error) {
	// 验证路径安全性
	if err := s.validatePath(path); err != nil {
		return 0, err
	}

	size, err := s.FileManager.GetFileSize(path)
	if err != nil {
		return 0, err
	}

	// 检查文件大小限制
	if size > s.config.Processing.MaxFileSize {
		s.logger.Warn("文件大小超过限制",
			"file", path,
			"size", size,
			"limit", s.config.Processing.MaxFileSize,
		)
	}

	return size, nil
}

// CopyFile 安全复制文件
func (s *SafeFileManager) CopyFile(src, dst string) error {
	// 验证路径安全性
	if err := s.validatePath(src); err != nil {
		return errors.Wrap(err, errors.ErrorTypeValidation, "INVALID_SRC_PATH", "源路径无效")
	}
	if err := s.validatePath(dst); err != nil {
		return errors.Wrap(err, errors.ErrorTypeValidation, "INVALID_DST_PATH", "目标路径无效")
	}

	// 检查文件大小
	size, err := s.FileManager.GetFileSize(src)
	if err != nil {
		return err
	}

	if size > s.config.Processing.MaxFileSize {
		return errors.New(errors.ErrorTypeValidation, "FILE_TOO_LARGE",
			"文件大小超过复制限制")
	}

	return s.FileManager.CopyFile(src, dst)
}

// validatePath 验证路径安全性
func (s *SafeFileManager) validatePath(path string) error {
	// 清理路径
	cleanPath := filepath.Clean(path)

	// 检查路径遍历攻击
	if strings.Contains(cleanPath, "..") {
		return errors.New(errors.ErrorTypeValidation, "PATH_TRAVERSAL", "检测到路径遍历攻击")
	}

	// 检查绝对路径（可选：根据需求决定是否允许）
	if filepath.IsAbs(cleanPath) {
		s.logger.Debug("使用绝对路径", "path", cleanPath)
	}

	return nil
}

// FileManagerFactory 文件管理器工厂
type FileManagerFactory struct {
	config *config.Config
	logger logger.Logger
}

// NewFileManagerFactory 创建文件管理器工厂
func NewFileManagerFactory(cfg *config.Config, logger logger.Logger) *FileManagerFactory {
	return &FileManagerFactory{
		config: cfg,
		logger: logger,
	}
}

// CreateFileManager 创建文件管理器
func (f *FileManagerFactory) CreateFileManager(safe bool) domain.FileManager {
	baseManager := NewLocalFileManager(f.config, f.logger)

	if safe {
		f.logger.Debug("创建安全文件管理器")
		return NewSafeFileManager(baseManager, f.config, f.logger)
	}

	f.logger.Debug("创建标准文件管理器")
	return baseManager
}

// TempDirManager 临时目录管理器
type TempDirManager struct {
	fileManager domain.FileManager
	tempDirs    []string
	logger      logger.Logger
}

// NewTempDirManager 创建临时目录管理器
func NewTempDirManager(fm domain.FileManager, logger logger.Logger) *TempDirManager {
	return &TempDirManager{
		fileManager: fm,
		tempDirs:    make([]string, 0),
		logger:      logger,
	}
}

// CreateTempDir 创建临时目录并记录
func (t *TempDirManager) CreateTempDir(prefix string) (string, error) {
	dir, err := t.fileManager.CreateTempDir(prefix)
	if err != nil {
		return "", err
	}

	t.tempDirs = append(t.tempDirs, dir)
	t.logger.Debug("记录临时目录", "path", dir, "total", len(t.tempDirs))

	return dir, nil
}

// CleanupAll 清理所有临时目录
func (t *TempDirManager) CleanupAll() {
	t.logger.Info("开始清理所有临时目录", "count", len(t.tempDirs))

	cleaned := 0
	for _, dir := range t.tempDirs {
		if err := t.fileManager.CleanupTempDir(dir); err != nil {
			t.logger.Warn("清理临时目录失败", "path", dir, "error", err)
		} else {
			cleaned++
		}
	}

	t.logger.Info("临时目录清理完成",
		"total", len(t.tempDirs),
		"cleaned", cleaned,
		"failed", len(t.tempDirs)-cleaned,
	)

	// 清空记录
	t.tempDirs = t.tempDirs[:0]
}
