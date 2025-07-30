package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// Config 应用程序配置
type Config struct {
	App        AppConfig        `json:"app"`
	Tools      ToolsConfig      `json:"tools"`
	Processing ProcessingConfig `json:"processing"`
	Logging    LoggingConfig    `json:"logging"`
	Advanced   AdvancedConfig   `json:"advanced"`
	Web        WebConfig        `json:"web"`
}

// AppConfig 应用程序基础配置
type AppConfig struct {
	Name           string        `json:"name"`
	Version        string        `json:"version"`
	MaxConcurrency int           `json:"max_concurrency"`
	TempDirPrefix  string        `json:"temp_dir_prefix"`
	DefaultQuality int           `json:"default_quality"`
	TempDir        string        `json:"temp_dir"` // 临时目录
	Timeout        time.Duration `json:"timeout"`  // 操作超时时间
}

// ToolsConfig 工具配置
type ToolsConfig struct {
	ToolsPath      string            `json:"tools_path"`
	WebpmuxPath    string            `json:"webpmux_path"`
	CwebpPath      string            `json:"cwebp_path"`
	DwebpPath      string            `json:"dwebp_path"`
	CommandTimeout int               `json:"command_timeout"` // 秒
	UseEmbedded    bool              `json:"use_embedded"`    // 是否使用嵌入式工具
	ToolPaths      map[string]string `json:"tool_paths"`      // 工具路径映射
}

// ProcessingConfig 处理配置
type ProcessingConfig struct {
	EnableParallel     bool   `json:"enable_parallel"`
	MaxWorkers         int    `json:"max_workers"`
	ChunkSize          int    `json:"chunk_size"`
	PreserveMetadata   bool   `json:"preserve_metadata"`
	DefaultPreset      string `json:"default_preset"`
	EnableProgressBar  bool   `json:"enable_progress_bar"`
	EnableOptimization bool   `json:"enable_optimization"`
	MaxFileSize        int64  `json:"max_file_size"` // 最大文件大小限制
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level      string `json:"level"`
	Format     string `json:"format"`
	OutputFile string `json:"output_file,omitempty"`
	MaxSize    int    `json:"max_size"` // MB
	MaxBackups int    `json:"max_backups"`
	MaxAge     int    `json:"max_age"` // 天
}

// AdvancedConfig 高级配置
type AdvancedConfig struct {
	CompressionPresets map[string]CompressionPreset `json:"compression_presets"`
	QualityProfiles    map[string]QualityProfile    `json:"quality_profiles"`
	OptimizationRules  OptimizationRules            `json:"optimization_rules"`
	PerformanceConfig  PerformanceConfig            `json:"performance"`
}

// CompressionPreset 压缩预设
type CompressionPreset struct {
	Name           string `json:"name"`
	Description    string `json:"description"`
	Quality        int    `json:"quality"`
	Method         int    `json:"method"`
	FilterStrength int    `json:"filter_strength"`
	Preset         string `json:"preset"`
	AlphaQuality   int    `json:"alpha_quality"`
	Lossless       bool   `json:"lossless"`
	NearLossless   int    `json:"near_lossless"` // 0-100, 0=disabled
	Sharpness      int    `json:"sharpness"`     // 0-7
	SNS            int    `json:"sns"`           // 0-100
	Segments       int    `json:"segments"`      // 1-4
	Pass           int    `json:"pass"`          // 1-10
	TargetSize     int    `json:"target_size"`   // bytes, 0=disabled
}

// QualityProfile 质量配置文件
type QualityProfile struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	MinQuality  int    `json:"min_quality"`
	MaxQuality  int    `json:"max_quality"`
	UseCase     string `json:"use_case"`
}

// OptimizationRules 优化规则
type OptimizationRules struct {
	EnableAutoQuality   bool    `json:"enable_auto_quality"`
	MaxFileSize         int64   `json:"max_file_size"`         // bytes
	TargetSizeReduction float64 `json:"target_size_reduction"` // 0.0-1.0
	EnableSmartPreset   bool    `json:"enable_smart_preset"`
}

// PerformanceConfig 性能配置
type PerformanceConfig struct {
	IOBufferSize        int  `json:"io_buffer_size"` // bytes
	EnableMemoryLimit   bool `json:"enable_memory_limit"`
	MaxMemoryUsage      int  `json:"max_memory_usage"` // MB
	EnableCPUThrottling bool `json:"enable_cpu_throttling"`
	CPUUsageLimit       int  `json:"cpu_usage_limit"` // 0-100%
}

// WebConfig Web服务配置
type WebConfig struct {
	Host               string `json:"host"`
	Port               int    `json:"port"`
	EnableTLS          bool   `json:"enable_tls"`
	TLSCert            string `json:"tls_cert"`
	TLSKey             string `json:"tls_key"`
	MaxFileSize        int64  `json:"max_file_size"` // bytes
	UploadDir          string `json:"upload_dir"`
	OutputDir          string `json:"output_dir"`
	MaxConcurrentTasks int    `json:"max_concurrent_tasks"`
	TaskTimeout        int    `json:"task_timeout"`     // seconds
	CleanupInterval    int    `json:"cleanup_interval"` // minutes
	EnableAuth         bool   `json:"enable_auth"`
	AuthToken          string `json:"auth_token"`
	EnableCORS         bool   `json:"enable_cors"`
	AllowedOrigins     string `json:"allowed_origins"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		App: AppConfig{
			Name:           "WebP Compressor",
			Version:        "2.0.0",
			MaxConcurrency: runtime.NumCPU(),
			TempDirPrefix:  "webpcompressor",
			DefaultQuality: 75,
			TempDir:        "",
			Timeout:        5 * time.Minute,
		},
		Tools: ToolsConfig{
			ToolsPath:      ".",
			WebpmuxPath:    "webpmux",
			CwebpPath:      "cwebp",
			DwebpPath:      "dwebp",
			CommandTimeout: 300, // 5分钟
			UseEmbedded:    false,
			ToolPaths:      make(map[string]string),
		},
		Processing: ProcessingConfig{
			EnableParallel:     true,
			MaxWorkers:         runtime.NumCPU(),
			ChunkSize:          10,
			PreserveMetadata:   true,
			DefaultPreset:      "photo",
			EnableProgressBar:  true,
			EnableOptimization: true,
			MaxFileSize:        100 * 1024 * 1024, // 100MB
		},
		Logging: LoggingConfig{
			Level:      "info",
			Format:     "text",
			MaxSize:    10,
			MaxBackups: 3,
			MaxAge:     7,
		},
		Advanced: AdvancedConfig{
			CompressionPresets: getDefaultCompressionPresets(),
			QualityProfiles:    getDefaultQualityProfiles(),
			OptimizationRules: OptimizationRules{
				EnableAutoQuality:   true,
				MaxFileSize:         100 * 1024 * 1024, // 100MB
				TargetSizeReduction: 0.3,               // 30%
				EnableSmartPreset:   true,
			},
			PerformanceConfig: PerformanceConfig{
				IOBufferSize:        64 * 1024, // 64KB
				EnableMemoryLimit:   true,
				MaxMemoryUsage:      1024, // 1GB
				EnableCPUThrottling: false,
				CPUUsageLimit:       80,
			},
		},
		Web: WebConfig{
			Host:               "0.0.0.0",
			Port:               8080,
			EnableTLS:          false,
			TLSCert:            "",
			TLSKey:             "",
			MaxFileSize:        10 * 1024 * 1024, // 10MB
			UploadDir:          "./uploads",
			OutputDir:          "./outputs",
			MaxConcurrentTasks: 10,
			TaskTimeout:        300, // 5分钟
			CleanupInterval:    60,  // 1小时
			EnableAuth:         false,
			AuthToken:          "",
			EnableCORS:         true,
			AllowedOrigins:     "*",
		},
	}
}

// getDefaultCompressionPresets 获取默认压缩预设
func getDefaultCompressionPresets() map[string]CompressionPreset {
	return map[string]CompressionPreset{
		"fast": {
			Name:           "快速",
			Description:    "快速压缩，适合批量处理",
			Quality:        60,
			Method:         0,
			FilterStrength: 60,
			Preset:         "default", // 使用cwebp支持的预设
			AlphaQuality:   30,
			Lossless:       false,
		},
		"balanced": {
			Name:           "平衡",
			Description:    "质量与速度平衡",
			Quality:        75,
			Method:         4,
			FilterStrength: 80,
			Preset:         "photo", // 使用cwebp支持的预设
			AlphaQuality:   50,
			Lossless:       false,
		},
		"quality": {
			Name:           "高质量",
			Description:    "最佳质量，处理时间较长",
			Quality:        90,
			Method:         6,
			FilterStrength: 100,
			Preset:         "photo", // 使用cwebp支持的预设
			AlphaQuality:   80,
			Lossless:       false,
			Sharpness:      2,
			SNS:            80,
			Segments:       4,
			Pass:           6,
		},
		"lossless": {
			Name:           "无损",
			Description:    "无损压缩，文件较大",
			Quality:        100,
			Method:         6,
			FilterStrength: 100,
			Preset:         "default", // 使用cwebp支持的预设
			AlphaQuality:   100,
			Lossless:       true,
		},
		"web": {
			Name:           "网页优化",
			Description:    "适合网页使用的优化设置",
			Quality:        70,
			Method:         4,
			FilterStrength: 75,
			Preset:         "picture", // 使用cwebp支持的预设
			AlphaQuality:   40,
			Lossless:       false,
			TargetSize:     512 * 1024, // 512KB
		},
	}
}

// MapWebPresetToCwebpPreset 将Web界面预设映射到cwebp预设
func (c *Config) MapWebPresetToCwebpPreset(webPreset string) string {
	// 获取预设配置
	if preset, exists := c.Advanced.CompressionPresets[webPreset]; exists {
		return preset.Preset
	}

	// 如果直接是cwebp支持的预设，直接返回
	validCwebpPresets := []string{"default", "photo", "picture", "drawing", "icon", "text"}
	for _, validPreset := range validCwebpPresets {
		if webPreset == validPreset {
			return webPreset
		}
	}

	// 默认返回photo
	return "photo"
}

// getDefaultQualityProfiles 获取默认质量配置文件
func getDefaultQualityProfiles() map[string]QualityProfile {
	return map[string]QualityProfile{
		"low": {
			Name:        "低质量",
			Description: "高压缩，适合网络传输",
			MinQuality:  10,
			MaxQuality:  40,
			UseCase:     "network",
		},
		"medium": {
			Name:        "中等质量",
			Description: "平衡压缩，日常使用",
			MinQuality:  40,
			MaxQuality:  70,
			UseCase:     "general",
		},
		"high": {
			Name:        "高质量",
			Description: "低压缩，保持细节",
			MinQuality:  70,
			MaxQuality:  90,
			UseCase:     "archive",
		},
		"premium": {
			Name:        "顶级质量",
			Description: "最佳质量，专业用途",
			MinQuality:  90,
			MaxQuality:  100,
			UseCase:     "professional",
		},
	}
}

// LoadFromEnv 从环境变量加载配置
func (c *Config) LoadFromEnv() {
	// 应用配置
	if val := os.Getenv("WEBP_MAX_CONCURRENCY"); val != "" {
		if num, err := strconv.Atoi(val); err == nil && num > 0 {
			c.App.MaxConcurrency = num
			c.Processing.MaxWorkers = num
		}
	}

	if val := os.Getenv("WEBP_DEFAULT_QUALITY"); val != "" {
		if num, err := strconv.Atoi(val); err == nil && num >= 0 && num <= 100 {
			c.App.DefaultQuality = num
		}
	}

	// 应用程序高级配置
	if val := os.Getenv("WEBP_TEMP_DIR"); val != "" {
		c.App.TempDir = val
	}

	if val := os.Getenv("WEBP_TIMEOUT"); val != "" {
		if duration, err := time.ParseDuration(val); err == nil {
			c.App.Timeout = duration
		}
	}

	// 工具配置
	if val := os.Getenv("WEBP_TOOLS_PATH"); val != "" {
		c.Tools.ToolsPath = val
	}

	if val := os.Getenv("WEBP_COMMAND_TIMEOUT"); val != "" {
		if num, err := strconv.Atoi(val); err == nil && num > 0 {
			c.Tools.CommandTimeout = num
		}
	}

	if val := os.Getenv("WEBP_USE_EMBEDDED"); val != "" {
		c.Tools.UseEmbedded = strings.ToLower(val) == "true"
	}

	// 处理配置
	if val := os.Getenv("WEBP_ENABLE_PARALLEL"); val != "" {
		c.Processing.EnableParallel = strings.ToLower(val) == "true"
	}

	if val := os.Getenv("WEBP_PRESERVE_METADATA"); val != "" {
		c.Processing.PreserveMetadata = strings.ToLower(val) == "true"
	}

	if val := os.Getenv("WEBP_DEFAULT_PRESET"); val != "" {
		c.Processing.DefaultPreset = val
	}

	if val := os.Getenv("WEBP_MAX_FILE_SIZE"); val != "" {
		if num, err := strconv.ParseInt(val, 10, 64); err == nil && num > 0 {
			c.Processing.MaxFileSize = num
		}
	}

	// 日志配置
	if val := os.Getenv("WEBP_LOG_LEVEL"); val != "" {
		c.Logging.Level = val
	}

	if val := os.Getenv("WEBP_LOG_FILE"); val != "" {
		c.Logging.OutputFile = val
	}

	// 性能配置
	if val := os.Getenv("WEBP_MAX_MEMORY"); val != "" {
		if num, err := strconv.Atoi(val); err == nil && num > 0 {
			c.Advanced.PerformanceConfig.MaxMemoryUsage = num
		}
	}

	// Web配置
	if val := os.Getenv("WEBP_WEB_HOST"); val != "" {
		c.Web.Host = val
	}
	if val := os.Getenv("WEBP_WEB_PORT"); val != "" {
		if num, err := strconv.Atoi(val); err == nil && num > 0 {
			c.Web.Port = num
		}
	}
	if val := os.Getenv("WEBP_WEB_ENABLE_TLS"); val != "" {
		c.Web.EnableTLS = strings.ToLower(val) == "true"
	}
	if val := os.Getenv("WEBP_WEB_TLS_CERT"); val != "" {
		c.Web.TLSCert = val
	}
	if val := os.Getenv("WEBP_WEB_TLS_KEY"); val != "" {
		c.Web.TLSKey = val
	}
	if val := os.Getenv("WEBP_WEB_MAX_FILE_SIZE"); val != "" {
		if num, err := strconv.ParseInt(val, 10, 64); err == nil && num > 0 {
			c.Web.MaxFileSize = num
		}
	}
	if val := os.Getenv("WEBP_WEB_UPLOAD_DIR"); val != "" {
		c.Web.UploadDir = val
	}
	if val := os.Getenv("WEBP_WEB_OUTPUT_DIR"); val != "" {
		c.Web.OutputDir = val
	}
	if val := os.Getenv("WEBP_WEB_MAX_CONCURRENT_TASKS"); val != "" {
		if num, err := strconv.Atoi(val); err == nil && num > 0 {
			c.Web.MaxConcurrentTasks = num
		}
	}
	if val := os.Getenv("WEBP_WEB_TASK_TIMEOUT"); val != "" {
		if num, err := strconv.Atoi(val); err == nil && num > 0 {
			c.Web.TaskTimeout = num
		}
	}
	if val := os.Getenv("WEBP_WEB_CLEANUP_INTERVAL"); val != "" {
		if num, err := strconv.Atoi(val); err == nil && num > 0 {
			c.Web.CleanupInterval = num
		}
	}
	if val := os.Getenv("WEBP_WEB_ENABLE_AUTH"); val != "" {
		c.Web.EnableAuth = strings.ToLower(val) == "true"
	}
	if val := os.Getenv("WEBP_WEB_AUTH_TOKEN"); val != "" {
		c.Web.AuthToken = val
	}
	if val := os.Getenv("WEBP_WEB_ENABLE_CORS"); val != "" {
		c.Web.EnableCORS = strings.ToLower(val) == "true"
	}
	if val := os.Getenv("WEBP_WEB_ALLOWED_ORIGINS"); val != "" {
		c.Web.AllowedOrigins = val
	}
}

// Validate 验证配置
func (c *Config) Validate() error {
	// 验证质量范围
	if c.App.DefaultQuality < 0 || c.App.DefaultQuality > 100 {
		return fmt.Errorf("默认质量必须在0-100之间，当前值: %d", c.App.DefaultQuality)
	}

	// 验证并发数
	if c.App.MaxConcurrency <= 0 {
		return fmt.Errorf("最大并发数必须大于0，当前值: %d", c.App.MaxConcurrency)
	}

	// 验证工具路径
	if c.Tools.ToolsPath == "" {
		return fmt.Errorf("工具路径不能为空")
	}

	// 验证超时时间
	if c.Tools.CommandTimeout <= 0 {
		return fmt.Errorf("命令超时时间必须大于0，当前值: %d", c.Tools.CommandTimeout)
	}

	// 验证日志级别
	validLogLevels := []string{"debug", "info", "warn", "error"}
	levelValid := false
	for _, level := range validLogLevels {
		if c.Logging.Level == level {
			levelValid = true
			break
		}
	}
	if !levelValid {
		return fmt.Errorf("无效的日志级别: %s，支持的级别: %v", c.Logging.Level, validLogLevels)
	}

	// 验证预设
	validPresets := []string{"default", "photo", "picture", "drawing", "icon", "text"}
	presetValid := false
	for _, preset := range validPresets {
		if c.Processing.DefaultPreset == preset {
			presetValid = true
			break
		}
	}
	if !presetValid {
		return fmt.Errorf("无效的默认预设: %s，支持的预设: %v", c.Processing.DefaultPreset, validPresets)
	}

	// 验证Web配置
	if c.Web.Port <= 0 {
		return fmt.Errorf("Web服务端口必须大于0，当前值: %d", c.Web.Port)
	}
	if c.Web.MaxFileSize <= 0 {
		return fmt.Errorf("Web服务最大文件大小必须大于0，当前值: %d", c.Web.MaxFileSize)
	}
	if c.Web.MaxConcurrentTasks <= 0 {
		return fmt.Errorf("Web服务最大并发任务数必须大于0，当前值: %d", c.Web.MaxConcurrentTasks)
	}
	if c.Web.TaskTimeout <= 0 {
		return fmt.Errorf("Web服务任务超时时间必须大于0，当前值: %d", c.Web.TaskTimeout)
	}
	if c.Web.CleanupInterval <= 0 {
		return fmt.Errorf("Web服务清理间隔必须大于0，当前值: %d", c.Web.CleanupInterval)
	}

	return nil
}

// GetCompressionPreset 获取压缩预设
func (c *Config) GetCompressionPreset(name string) (CompressionPreset, bool) {
	preset, exists := c.Advanced.CompressionPresets[name]
	return preset, exists
}

// GetQualityProfile 获取质量配置文件
func (c *Config) GetQualityProfile(name string) (QualityProfile, bool) {
	profile, exists := c.Advanced.QualityProfiles[name]
	return profile, exists
}

// IsParallelEnabled 检查是否启用并行处理
func (c *Config) IsParallelEnabled() bool {
	return c.Processing.EnableParallel && c.Processing.MaxWorkers > 1
}

// GetToolPath 获取工具路径
func (c *Config) GetToolPath(toolName string) string {
	// 首先检查工具路径映射
	if path, exists := c.Tools.ToolPaths[toolName]; exists && path != "" {
		return path
	}

	// 然后检查具体的工具路径配置
	switch toolName {
	case "webpmux":
		if c.Tools.WebpmuxPath != "" {
			return c.Tools.WebpmuxPath
		}
	case "cwebp":
		if c.Tools.CwebpPath != "" {
			return c.Tools.CwebpPath
		}
	case "dwebp":
		if c.Tools.DwebpPath != "" {
			return c.Tools.DwebpPath
		}
	}

	// 最后回退到基础路径
	if c.Tools.ToolsPath != "" && c.Tools.ToolsPath != "." {
		return filepath.Join(c.Tools.ToolsPath, toolName)
	}

	return toolName
}

// GetEffectiveWorkers 获取有效的工作者数量
func (c *Config) GetEffectiveWorkers(taskCount int) int {
	maxWorkers := c.Processing.MaxWorkers
	if maxWorkers <= 0 {
		maxWorkers = runtime.NumCPU()
	}

	// 不要超过任务数量
	if taskCount < maxWorkers {
		return taskCount
	}

	return maxWorkers
}
