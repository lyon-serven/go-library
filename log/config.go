package log

import "time"

// LogConfig 日志配置结构
type LogConfig struct {
	// 基础配置
	LogToStdout bool   `yaml:"log_to_stdout" json:"log_to_stdout"`
	LogToFile   bool   `yaml:"log_to_file" json:"log_to_file"`
	Level       string `yaml:"level" json:"level"`

	// 文件配置
	Path         string        `yaml:"path" json:"path"`
	FileName     string        `yaml:"file_name" json:"file_name"`
	FilePattern  string        `yaml:"file_pattern" json:"file_pattern"` // 自定义时间格式模板
	RotationTime time.Duration `yaml:"rotation_time" json:"rotation_time"`
	FileAge      time.Duration `yaml:"file_age" json:"file_age"`
	MaxSize      int           `yaml:"max_size" json:"max_size"` // MB
	MaxBackups   int           `yaml:"max_backups" json:"max_backups"`

	// 格式配置
	Format       string `yaml:"format" json:"format"` // json, console
	EnableCaller bool   `yaml:"enable_caller" json:"enable_caller"`
	EnableStack  bool   `yaml:"enable_stack" json:"enable_stack"`

	// 性能配置
	BufferSize   int           `yaml:"buffer_size" json:"buffer_size"`
	FlushTimeout time.Duration `yaml:"flush_timeout" json:"flush_timeout"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() *LogConfig {
	return &LogConfig{
		LogToStdout:  true,
		LogToFile:    true,
		Level:        "info",
		Path:         "./logs",
		FileName:     "app.log",
		FilePattern:  ".%Y%m%d%H%M", // 默认时间格式
		RotationTime: time.Hour * 24,
		FileAge:      time.Hour * 24 * 7,
		MaxSize:      100,
		MaxBackups:   10,
		Format:       "json",
		EnableCaller: true,
		EnableStack:  true,
		BufferSize:   256 * 1024, // 256KB
		FlushTimeout: time.Second * 5,
	}
}
