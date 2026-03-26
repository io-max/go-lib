package log

import (
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Level 日志级别
type Level zapcore.Level

const (
	DebugLevel Level = Level(zapcore.DebugLevel)
	InfoLevel  Level = Level(zapcore.InfoLevel)
	WarnLevel  Level = Level(zapcore.WarnLevel)
	ErrorLevel Level = Level(zapcore.ErrorLevel)
	FatalLevel Level = Level(zapcore.FatalLevel)
)

// ParseLevel 解析日志级别
func ParseLevel(level string) Level {
	switch strings.ToLower(level) {
	case "debug":
		return DebugLevel
	case "info":
		return InfoLevel
	case "warn", "warning":
		return WarnLevel
	case "error":
		return ErrorLevel
	case "fatal":
		return FatalLevel
	default:
		return InfoLevel
	}
}

// Config 日志配置
type Config struct {
	Level       Level
	Format      string // json, console
	OutputFile  string // 日志文件路径（为空则输出到 stdout）
	MaxSize     int    // 单文件最大大小 (MB)
	MaxBackups  int    // 最大保留的旧日志文件数
	MaxAge      int    // 日志最大保留时间 (小时)
	Compress    bool   // 是否压缩旧日志
	Caller      bool   // 是否显示调用者信息
}

// DefaultConfig 默认配置
func DefaultConfig() Config {
	return Config{
		Level:      InfoLevel,
		Format:     "json",
		MaxSize:    100,
		MaxBackups: 7,
		MaxAge:     720,
		Compress:   true,
		Caller:     true,
	}
}

// Logger 日志包装器
type Logger struct {
	zap *zap.Logger
}

// Load 加载日志
func Load(cfg Config) (*Logger, error) {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "time"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.LevelKey = "level"
	encoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder
	encoderConfig.MessageKey = "msg"
	encoderConfig.CallerKey = "caller"

	var encoder zapcore.Encoder
	if cfg.Format == "console" {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	var writeSyncer zapcore.WriteSyncer
	if cfg.OutputFile != "" {
		writeSyncer = zapcore.AddSync(&lumberjack.Logger{
			Filename:   cfg.OutputFile,
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,
			Compress:   cfg.Compress,
		})
	} else {
		writeSyncer = zapcore.AddSync(os.Stdout)
	}

	core := zapcore.NewCore(encoder, writeSyncer, zapcore.Level(cfg.Level))

	var zlogger *zap.Logger
	if cfg.Caller {
		zlogger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	} else {
		zlogger = zap.New(core)
	}

	return &Logger{zap: zlogger}, nil
}

// MustLoad 加载日志（失败则 panic）
func MustLoad(cfg Config) *Logger {
	l, err := Load(cfg)
	if err != nil {
		panic(err)
	}
	return l
}

// Debug 日志
func (l *Logger) Debug(msg string, keys ...interface{}) {
	l.zap.Sugar().Debugw(msg, keys...)
}

// Info 日志
func (l *Logger) Info(msg string, keys ...interface{}) {
	l.zap.Sugar().Infow(msg, keys...)
}

// Warn 日志
func (l *Logger) Warn(msg string, keys ...interface{}) {
	l.zap.Sugar().Warnw(msg, keys...)
}

// Error 日志
func (l *Logger) Error(msg string, keys ...interface{}) {
	l.zap.Sugar().Errorw(msg, keys...)
}

// Fatal 日志
func (l *Logger) Fatal(msg string, keys ...interface{}) {
	l.zap.Sugar().Fatalw(msg, keys...)
}

// With 创建带字段的 logger
func (l *Logger) With(keys ...interface{}) *Logger {
	return &Logger{zap: l.zap.With(fieldsToZapFields(keys...)...)}
}

func fieldsToZapFields(keys ...interface{}) []zap.Field {
	var fields []zap.Field
	for i := 0; i < len(keys); i += 2 {
		if i+1 >= len(keys) {
			break
		}
		key, ok := keys[i].(string)
		if !ok {
			continue
		}
		fields = append(fields, zap.Any(key, keys[i+1]))
	}
	return fields
}

// 全局 logger
var globalLogger *Logger

// Global 获取全局 logger
func Global() *Logger {
	if globalLogger == nil {
		globalLogger, _ = Load(DefaultConfig())
	}
	return globalLogger
}

// SetGlobal 设置全局 logger
func SetGlobal(l *Logger) {
	globalLogger = l
}

// 全局快捷方法
func Debug(msg string, keys ...interface{})    { Global().Debug(msg, keys...) }
func Info(msg string, keys ...interface{})     { Global().Info(msg, keys...) }
func Warn(msg string, keys ...interface{})     { Global().Warn(msg, keys...) }
func Error(msg string, keys ...interface{})    { Global().Error(msg, keys...) }
func Fatal(msg string, keys ...interface{})    { Global().Fatal(msg, keys...) }
func With(keys ...interface{}) *Logger         { return Global().With(keys...) }
