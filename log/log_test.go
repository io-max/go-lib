package log

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	logger, err := Load(Config{
		Level:  InfoLevel,
		Format: "json",
	})
	assert.NoError(t, err)
	assert.NotNil(t, logger)
}

func TestMustLoad(t *testing.T) {
	logger := MustLoad(Config{
		Level:  InfoLevel,
		Format: "json",
	})
	assert.NotNil(t, logger)
}

func TestLoggerInfo(t *testing.T) {
	logger := MustLoad(Config{
		Level:  InfoLevel,
		Format: "json",
	})
	// 应该不 panic
	logger.Info("test message", "key", "value")
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected Level
	}{
		{"debug", DebugLevel},
		{"DEBUG", DebugLevel},
		{"info", InfoLevel},
		{"Info", InfoLevel},
		{"warn", WarnLevel},
		{"warning", WarnLevel},
		{"error", ErrorLevel},
		{"fatal", FatalLevel},
		{"unknown", InfoLevel},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, ParseLevel(tt.input))
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	assert.Equal(t, InfoLevel, cfg.Level)
	assert.Equal(t, "json", cfg.Format)
	assert.Equal(t, 100, cfg.MaxSize)
	assert.Equal(t, 7, cfg.MaxBackups)
	assert.Equal(t, 720, cfg.MaxAge)
	assert.True(t, cfg.Compress)
	assert.True(t, cfg.Caller)
}

func TestLoggerMethods(t *testing.T) {
	logger := MustLoad(Config{
		Level:  DebugLevel,
		Format: "console",
		Caller: false,
	})

	// 测试所有日志方法不 panic
	logger.Debug("debug message", "key1", "value1")
	logger.Info("info message", "key2", "value2")
	logger.Warn("warn message", "key3", "value3")
	logger.Error("error message", "key4", "value4")
}

func TestLoggerWith(t *testing.T) {
	logger := MustLoad(Config{
		Level:  InfoLevel,
		Format: "json",
	})

	// 测试 With 方法
	loggerWith := logger.With("module", "test", "id", 123)
	assert.NotNil(t, loggerWith)
	loggerWith.Info("message with fields")
}

func TestGlobalLogger(t *testing.T) {
	// 测试 Global 函数
	global := Global()
	assert.NotNil(t, global)

	// 测试 SetGlobal 函数
	newLogger := MustLoad(Config{
		Level:  DebugLevel,
		Format: "console",
	})
	SetGlobal(newLogger)
	assert.Equal(t, newLogger, Global())

	// 恢复默认
	SetGlobal(nil)
}

func TestGlobalFunctions(t *testing.T) {
	// 测试全局快捷方法
	Debug("global debug", "key", "value")
	Info("global info", "key", "value")
	Warn("global warn", "key", "value")
	Error("global error", "key", "value")

	// 测试全局 With
	logger := With("global", "true")
	assert.NotNil(t, logger)
}

func TestToZapFields(t *testing.T) {
	// 测试偶数个参数
	fields := fieldsToZapFields("key1", "value1", "key2", 123)
	assert.Len(t, fields, 2)

	// 测试奇数个参数（最后一个被忽略）
	fields = fieldsToZapFields("key1", "value1", "key2")
	assert.Len(t, fields, 1)

	// 测试空参数
	fields = fieldsToZapFields()
	assert.Len(t, fields, 0)

	// 测试非 string key（被忽略）
	fields = fieldsToZapFields(123, "value", "key", "value")
	assert.Len(t, fields, 1)
}
