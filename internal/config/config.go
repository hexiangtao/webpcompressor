package config

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Config 应用程序配置
type Config struct {
	App        AppConfig        `yaml:"app"`
	Tools      ToolsConfig      `yaml:"tools"`
	Processing ProcessingConfig `yaml:"processing"`
	Logging    LoggingConfig    `yaml:"logging"`
}

// AppConfig 应用程序基本配置
type AppConfig struct {
	Name           string        `yaml:"name"`
	Version        string        `yaml:"version"`
	TempDir        string        `yaml:"temp_dir"`
	MaxConcurrency int           `yaml:"max_concurrency"`
	Timeout        time.Duration `yaml:"timeout"`
}

// ToolsConfig 工具配置
type ToolsConfig struct {
	WebpmuxPath string            `yaml:"webpmux_path"`
	CwebpPath   string            `yaml:"cwebp_path"`
	DwebpPath   string            `yaml:"dwebp_path"`
	ToolPaths   map[string]string `yaml:"tool_paths"`
	UseEmbedded bool              `yaml:"use_embedded"`
}

// ProcessingConfig 处理配置
type ProcessingConfig struct {
	DefaultQuality   int   `yaml:"default_quality"`
	MaxFileSize      int64 `yaml:"max_file_size"`
	EnableParallel   bool  `yaml:"enable_parallel"`
	PreserveMetadata bool  `yaml:"preserve_metadata"`
	AutoCleanup      bool  `yaml:"auto_cleanup"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level      string `yaml:"level"`
	Output     string `yaml:"output"`
	MaxSize    int    `yaml:"max_size"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAge     int    `yaml:"max_age"`
	Compress   bool   `yaml:"compress"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		App: AppConfig{
			Name:           "WebP Compressor",
			Version:        "2.0.0",
			TempDir:        filepath.Join(os.TempDir(), "webpcompressor"),
			MaxConcurrency: 4,
			Timeout:        5 * time.Minute,
		},
		Tools: ToolsConfig{
			WebpmuxPath: "webpmux",
			CwebpPath:   "cwebp",
			DwebpPath:   "dwebp",
			ToolPaths: map[string]string{
				"webpmux":  "webpmux.exe",
				"cwebp":    "cwebp.exe",
				"dwebp":    "dwebp.exe",
				"gif2webp": "gif2webp.exe",
				"webpinfo": "webpinfo.exe",
			},
			UseEmbedded: false,
		},
		Processing: ProcessingConfig{
			DefaultQuality:   50,
			MaxFileSize:      100 * 1024 * 1024, // 100MB
			EnableParallel:   true,
			PreserveMetadata: false,
			AutoCleanup:      true,
		},
		Logging: LoggingConfig{
			Level:      "info",
			Output:     "stdout",
			MaxSize:    10, // 10MB
			MaxBackups: 5,
			MaxAge:     30, // 30 days
			Compress:   true,
		},
	}
}

// LoadFromEnv 从环境变量加载配置
func (c *Config) LoadFromEnv() {
	// App配置
	if name := os.Getenv("WEBP_APP_NAME"); name != "" {
		c.App.Name = name
	}
	if version := os.Getenv("WEBP_APP_VERSION"); version != "" {
		c.App.Version = version
	}
	if tempDir := os.Getenv("WEBP_TEMP_DIR"); tempDir != "" {
		c.App.TempDir = tempDir
	}
	if concurrency := os.Getenv("WEBP_MAX_CONCURRENCY"); concurrency != "" {
		if val, err := strconv.Atoi(concurrency); err == nil {
			c.App.MaxConcurrency = val
		}
	}
	if timeout := os.Getenv("WEBP_TIMEOUT"); timeout != "" {
		if val, err := time.ParseDuration(timeout); err == nil {
			c.App.Timeout = val
		}
	}

	// Tools配置
	if webpmux := os.Getenv("WEBP_WEBPMUX_PATH"); webpmux != "" {
		c.Tools.WebpmuxPath = webpmux
	}
	if cwebp := os.Getenv("WEBP_CWEBP_PATH"); cwebp != "" {
		c.Tools.CwebpPath = cwebp
	}
	if dwebp := os.Getenv("WEBP_DWEBP_PATH"); dwebp != "" {
		c.Tools.DwebpPath = dwebp
	}
	if embedded := os.Getenv("WEBP_USE_EMBEDDED"); embedded != "" {
		c.Tools.UseEmbedded = strings.ToLower(embedded) == "true"
	}

	// Processing配置
	if quality := os.Getenv("WEBP_DEFAULT_QUALITY"); quality != "" {
		if val, err := strconv.Atoi(quality); err == nil {
			c.Processing.DefaultQuality = val
		}
	}
	if maxSize := os.Getenv("WEBP_MAX_FILE_SIZE"); maxSize != "" {
		if val, err := strconv.ParseInt(maxSize, 10, 64); err == nil {
			c.Processing.MaxFileSize = val
		}
	}
	if parallel := os.Getenv("WEBP_ENABLE_PARALLEL"); parallel != "" {
		c.Processing.EnableParallel = strings.ToLower(parallel) == "true"
	}

	// Logging配置
	if level := os.Getenv("WEBP_LOG_LEVEL"); level != "" {
		c.Logging.Level = level
	}
	if output := os.Getenv("WEBP_LOG_OUTPUT"); output != "" {
		c.Logging.Output = output
	}
}

// Validate 验证配置
func (c *Config) Validate() error {
	// 验证质量范围
	if c.Processing.DefaultQuality < 0 || c.Processing.DefaultQuality > 100 {
		c.Processing.DefaultQuality = 50
	}

	// 验证并发数
	if c.App.MaxConcurrency <= 0 {
		c.App.MaxConcurrency = 1
	}

	// 验证文件大小限制
	if c.Processing.MaxFileSize <= 0 {
		c.Processing.MaxFileSize = 100 * 1024 * 1024 // 100MB
	}

	// 确保临时目录存在
	if c.App.TempDir != "" {
		os.MkdirAll(c.App.TempDir, 0755)
	}

	return nil
}

// GetToolPath 获取工具路径
func (c *Config) GetToolPath(toolName string) string {
	if path, exists := c.Tools.ToolPaths[toolName]; exists {
		return path
	}

	// 回退到默认路径
	switch toolName {
	case "webpmux":
		return c.Tools.WebpmuxPath
	case "cwebp":
		return c.Tools.CwebpPath
	case "dwebp":
		return c.Tools.DwebpPath
	default:
		return toolName + ".exe"
	}
}

// IsProduction 是否为生产环境
func (c *Config) IsProduction() bool {
	return strings.ToLower(os.Getenv("WEBP_ENV")) == "production"
}

// IsDevelopment 是否为开发环境
func (c *Config) IsDevelopment() bool {
	env := strings.ToLower(os.Getenv("WEBP_ENV"))
	return env == "" || env == "development" || env == "dev"
}
