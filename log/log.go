package log

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	defaultLogger *Logger
	once          sync.Once
)

// Logger 封装的日志器
type Logger struct {
	*zap.Logger
	config *LogConfig
	sugar  *zap.SugaredLogger
}

// NewLogger 创建新的日志器实例
func NewLogger(config *LogConfig) (*Logger, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// 验证配置
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// 创建日志核心
	cores := make([]zapcore.Core, 0, 2)

	// 控制台输出
	if config.LogToStdout {
		consoleCore := createConsoleCore(config)
		cores = append(cores, consoleCore)
	}

	// 文件输出
	if config.LogToFile {
		fileCore, err := createFileCore(config)
		if err != nil {
			return nil, fmt.Errorf("failed to create file core: %w", err)
		}
		cores = append(cores, fileCore)
	}

	if len(cores) == 0 {
		return nil, fmt.Errorf("no output configured")
	}

	// 合并核心
	core := zapcore.NewTee(cores...)

	// 创建选项
	options := []zap.Option{}
	if config.EnableCaller {
		options = append(options, zap.AddCaller())
	}
	if config.EnableStack {
		options = append(options, zap.AddStacktrace(zapcore.ErrorLevel))
	}

	// 创建 zap logger
	zapLogger := zap.New(core, options...)

	logger := &Logger{
		Logger: zapLogger,
		config: config,
		sugar:  zapLogger.Sugar(),
	}

	return logger, nil
}

// InitDefault 初始化默认日志器
func InitDefault(config *LogConfig) error {
	var err error
	once.Do(func() {
		defaultLogger, err = NewLogger(config)
	})
	return err
}

// GetDefault 获取默认日志器
func GetDefault() *Logger {
	if defaultLogger == nil {
		// 如果没有初始化，使用默认配置
		_ = InitDefault(nil)
	}
	return defaultLogger
}

// Sugar 返回 SugaredLogger
func (l *Logger) Sugar() *zap.SugaredLogger {
	return l.sugar
}

// WithFields 添加字段
func (l *Logger) WithFields(fields ...zap.Field) *Logger {
	return &Logger{
		Logger: l.Logger.With(fields...),
		config: l.config,
		sugar:  l.Logger.With(fields...).Sugar(),
	}
}

// WithContext 添加上下文字段
func (l *Logger) WithContext(key string, value interface{}) *Logger {
	return l.WithFields(zap.Any(key, value))
}

// Sync 同步日志
func (l *Logger) Sync() error {
	return l.Logger.Sync()
}

// Close 关闭日志器
func (l *Logger) Close() error {
	return l.Sync()
}

// createConsoleCore 创建控制台核心
func createConsoleCore(config *LogConfig) zapcore.Core {
	encoderConfig := getEncoderConfig(config)

	var encoder zapcore.Encoder
	if config.Format == "console" {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	level := parseLogLevel(config.Level)
	return zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), level)
}

// createFileCore 创建文件核心
func createFileCore(config *LogConfig) (zapcore.Core, error) {
	// 确保日志目录存在
	if err := os.MkdirAll(config.Path, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	var writer io.Writer
	var err error

	// 选择日志轮转方式
	if config.RotationTime > 0 {
		writer, err = createRotateLogsWriter(config)
	} else {
		writer = createLumberjackWriter(config)
	}

	if err != nil {
		return nil, err
	}

	encoderConfig := getEncoderConfig(config)
	encoder := zapcore.NewJSONEncoder(encoderConfig)
	level := parseLogLevel(config.Level)

	// 添加缓冲
	if config.BufferSize > 0 {
		writer = &zapcore.BufferedWriteSyncer{
			WS:            zapcore.AddSync(writer),
			Size:          config.BufferSize,
			FlushInterval: config.FlushTimeout,
		}
	}

	return zapcore.NewCore(encoder, zapcore.AddSync(writer), level), nil
}

// createRotateLogsWriter 创建基于时间的日志轮转器
func createRotateLogsWriter(config *LogConfig) (io.Writer, error) {
	// 使用自定义时间格式模板，如果为空则使用默认格式
	pattern := config.FilePattern
	if pattern == "" {
		pattern = ".%Y%m%d%H%M"
	}

	// 构建完整的日志文件路径模板
	// 如果pattern以/开头，说明包含目录结构，需要特殊处理
	var logPath string
	var linkName string
	if strings.HasPrefix(pattern, "/") {
		// pattern包含目录结构，如: /%Y_%m/%d/task_%H.log
		logPath = filepath.Join(config.Path, pattern[1:]) // 去掉开头的/
		// 对于目录结构，latest.log 放在根目录
		linkName = filepath.Join(config.Path, "latest.log")
	} else {
		// pattern只是文件名后缀，如: .%Y%m%d%H%M
		logPath = filepath.Join(config.Path, config.FileName+pattern)
		// 对于简单格式，latest.log 也放在根目录
		linkName = filepath.Join(config.Path, "latest.log")
	}

	options := []rotatelogs.Option{
		rotatelogs.WithMaxAge(config.FileAge),
		rotatelogs.WithRotationTime(config.RotationTime),
		rotatelogs.WithLinkName(linkName), // 始终创建符号链接
	}

	return rotatelogs.New(
		logPath,
		options...,
	)
}

// createLumberjackWriter 创建基于大小的日志轮转器
func createLumberjackWriter(config *LogConfig) io.Writer {
	return &lumberjack.Logger{
		Filename:   filepath.Join(config.Path, config.FileName),
		MaxSize:    config.MaxSize,
		MaxBackups: config.MaxBackups,
		MaxAge:     int(config.FileAge.Hours() / 24),
		Compress:   true,
	}
}

// getEncoderConfig 获取编码器配置
func getEncoderConfig(config *LogConfig) zapcore.EncoderConfig {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 根据格式调整编码器
	if config.Format == "console" {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")
	}

	return encoderConfig
}

// parseLogLevel 解析日志级别
func parseLogLevel(level string) zapcore.Level {
	switch strings.ToLower(level) {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn", "warning":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "fatal":
		return zapcore.FatalLevel
	case "panic":
		return zapcore.PanicLevel
	default:
		return zapcore.InfoLevel
	}
}

// validateConfig 验证配置
func validateConfig(config *LogConfig) error {
	if !config.LogToStdout && !config.LogToFile {
		return fmt.Errorf("at least one output must be enabled")
	}

	if config.LogToFile {
		if config.Path == "" {
			return fmt.Errorf("log path cannot be empty when file logging is enabled")
		}
		if config.FileName == "" {
			return fmt.Errorf("log filename cannot be empty when file logging is enabled")
		}
	}

	return nil
}

// 便捷方法 - 使用默认日志器
func Debug(msg string, fields ...zap.Field) {
	GetDefault().Debug(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	GetDefault().Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	GetDefault().Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	GetDefault().Error(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	GetDefault().Fatal(msg, fields...)
}

func Panic(msg string, fields ...zap.Field) {
	GetDefault().Panic(msg, fields...)
}

// Sugar 便捷方法
func Debugf(template string, args ...interface{}) {
	GetDefault().Sugar().Debugf(template, args...)
}

func Infof(template string, args ...interface{}) {
	GetDefault().Sugar().Infof(template, args...)
}

func Warnf(template string, args ...interface{}) {
	GetDefault().Sugar().Warnf(template, args...)
}

func Errorf(template string, args ...interface{}) {
	GetDefault().Sugar().Errorf(template, args...)
}

func Fatalf(template string, args ...interface{}) {
	GetDefault().Sugar().Fatalf(template, args...)
}

func Panicf(template string, args ...interface{}) {
	GetDefault().Sugar().Panicf(template, args...)
}

// WithField 添加单个字段
func WithField(key string, value interface{}) *Logger {
	return GetDefault().WithFields(zap.Any(key, value))
}

// WithFields 添加多个字段
func WithFields(fields ...zap.Field) *Logger {
	return GetDefault().WithFields(fields...)
}

// Sync 同步默认日志器
func Sync() error {
	return GetDefault().Sync()
}
