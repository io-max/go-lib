# Go 集成库 Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development to implement this plan task-by-task.

**Goal:** 创建一个 Go 集成库，封装 Gin 中间件、Viper 配置、Zap 日志等常用框架，减少项目重复代码。

**Architecture:** 采用分层设计，每个模块独立可插拔。middleware/ 提供 Gin 中间件，config/ 封装 Viper 配置加载，log/ 封装 Zap 日志，gincrud/ 提供泛型 CRUD 封装。所有模块通过统一的 ErrorCode 和 Response 进行错误处理和响应格式化。

**Tech Stack:** Go 1.21+, Gin v1.9, GORM v1.25, Viper v1.18, Zap v1.26, go-redis/v9, golang-jwt/jwt/v5, validator/v10, lumberjack.v2

---

## Phase 1: 项目骨架搭建

### Task 1: 创建 go.mod 和基础目录结构

**Files:**
- Create: `go.mod`
- Create: `middleware/.gitkeep`
- Create: `config/.gitkeep`
- Create: `log/.gitkeep`
- Create: `gincrud/.gitkeep`
- Create: `examples/basic/.gitkeep`
- Create: `examples/full/.gitkeep`

**Step 1: 创建 go.mod**

```go
module github.com/your-org/go-lib

go 1.21

require (
	github.com/gin-gonic/gin v1.9.1
	gorm.io/gorm v1.25.5
	gorm.io/driver/mysql v1.5.2
	github.com/spf13/viper v1.18.2
	go.uber.org/zap v1.26.0
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
	github.com/golang-jwt/jwt/v5 v5.2.0
	github.com/redis/go-redis/v9 v9.3.0
	github.com/go-playground/validator/v10 v10.15.5
	github.com/google/uuid v1.5.0
)
```

**Step 2: 创建目录结构**

Run:
```bash
mkdir -p middleware config log gincrud examples/basic examples/full docs
touch middleware/.gitkeep config/.gitkeep log/.gitkeep gincrud/.gitkeep
```

Expected: 所有目录创建成功

**Step 3: 验证目录结构**

Run: `ls -la`

Expected: 看到 middleware/, config/, log/, gincrud/, examples/, docs/

**Step 4: 提交**

```bash
git add go.mod middleware/.gitkeep config/.gitkeep log/.gitkeep gincrud/.gitkeep
git commit -m "feat: 初始化项目骨架"
```

---

## Phase 2: 统一响应模块（middleware/response.go）

### Task 2: 实现 ErrorCode 接口和 StandardErrorCode

**Files:**
- Create: `middleware/response.go`
- Create: `middleware/errors.go`

**Step 1: 编写测试**

Create `middleware/response_test.go`:

```go
package middleware

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestNewCode(t *testing.T) {
	err := NewCode(400001, "Invalid parameter")
	assert.Equal(t, 400001, err.Code())
	assert.Equal(t, "Invalid parameter", err.Message())
}

func TestNewCodeDuplicate(t *testing.T) {
	NewCode(500001, "First error")
	assert.Panics(t, func() {
		NewCode(500001, "Duplicate error")
	})
}

func TestStandardErrorCode(t *testing.T) {
	err := NewCode(404001, "Not found")
	impl, ok := err.(*StandardErrorCode)
	assert.True(t, ok)
	assert.Equal(t, 404001, impl.Code())
	assert.Equal(t, "Not found", impl.Message())
}
```

**Step 2: 运行测试验证失败**

Run: `go test ./middleware -run TestNewCode -v`
Expected: FAIL with "undefined: NewCode"

**Step 3: 实现代码**

Create `middleware/errors.go`:

```go
package middleware

import (
	"fmt"
	"sync"
)

// ErrorCode 错误码接口
type ErrorCode interface {
	Code() int
	Message() string
}

// StandardErrorCode 标准错误码实现
type StandardErrorCode struct {
	code    int
	message string
}

func (e *StandardErrorCode) Code() int    { return e.code }
func (e *StandardErrorCode) Message() string { return e.message }

// errorCodeRegistry 错误码注册表
var (
	errorCodeRegistry = make(map[int]string)
	errorCodeMutex    sync.Mutex
)

// NewCode 创建标准错误码
func NewCode(code int, message string) ErrorCode {
	errorCodeMutex.Lock()
	defer errorCodeMutex.Unlock()

	if existingMsg, ok := errorCodeRegistry[code]; ok {
		panic(fmt.Sprintf("duplicate error code %d: %s vs %s", code, existingMsg, message))
	}
	errorCodeRegistry[code] = message

	return &StandardErrorCode{code: code, message: message}
}

// NewCodef 创建标准错误码（支持格式化）
func NewCodef(code int, format string, args ...interface{}) ErrorCode {
	return NewCode(code, fmt.Sprintf(format, args...))
}
```

**Step 4: 运行测试验证通过**

Run: `go test ./middleware -run TestNewCode -v`
Expected: PASS

**Step 5: 提交**

```bash
git add middleware/errors.go middleware/response_test.go
git commit -m "feat(middleware): 实现 ErrorCode 接口和 StandardErrorCode"
```

---

### Task 3: 实现 Response 结构和 RespondError/RespondSuccess

**Files:**
- Modify: `middleware/response.go`
- Test: `middleware/response_test.go`

**Step 1: 编写测试**

Append to `middleware/response_test.go`:

```go
func TestRespondError(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	err := NewCode(400001, "Invalid parameter")
	RespondError(c, err)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), `"code":400001`)
	assert.Contains(t, w.Body.String(), `"message":"Invalid parameter"`)
}

func TestRespondErrorWithMessage(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	err := NewCode(400001, "Invalid parameter")
	RespondError(c, err, "username is required")

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), `"message":"username is required"`)
}

func TestRespondSuccess(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	RespondSuccess(c, gin.H{"data": "test"})

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"code":200`)
	assert.Contains(t, w.Body.String(), `"data":"test"`)
}
```

**Step 2: 运行测试验证失败**

Run: `go test ./middleware -run TestRespondError -v`
Expected: FAIL with "undefined: RespondError"

**Step 3: 实现代码**

Create `middleware/response.go`:

```go
package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	TraceID string      `json:"trace_id,omitempty"`
}

// Context Keys
const (
	RequestIDKey = "request_id"
)

// RespondError 返回错误响应 (ErrorCode)
func RespondError(c *gin.Context, err ErrorCode) {
	httpStatus := httpStatusFromCode(err.Code())
	c.JSON(httpStatus, Response{
		Code:    err.Code(),
		Message: err.Message(),
		TraceID: c.GetString(RequestIDKey),
	})
}

// RespondError 返回错误响应 (ErrorCode + 自定义消息)
func RespondError(c *gin.Context, err ErrorCode, message string) {
	httpStatus := httpStatusFromCode(err.Code())
	c.JSON(httpStatus, Response{
		Code:    err.Code(),
		Message: message,
		TraceID: c.GetString(RequestIDKey),
	})
}

// RespondError 返回错误响应 (ErrorCode + 格式化消息)
func RespondError(c *gin.Context, err ErrorCode, format string, args ...interface{}) {
	httpStatus := httpStatusFromCode(err.Code())
	c.JSON(httpStatus, Response{
		Code:    err.Code(),
		Message: fmt.Sprintf(format, args...),
		TraceID: c.GetString(RequestIDKey),
	})
}

// RespondError 返回错误响应 (HTTP 状态码 + ErrorCode)
func RespondError(c *gin.Context, httpStatus int, err ErrorCode) {
	c.JSON(httpStatus, Response{
		Code:    err.Code(),
		Message: err.Message(),
		TraceID: c.GetString(RequestIDKey),
	})
}

// RespondError 返回错误响应 (error + ErrorCode)
func RespondError(c *gin.Context, err error, code ErrorCode) {
	httpStatus := httpStatusFromCode(code.Code())
	c.JSON(httpStatus, Response{
		Code:    code.Code(),
		Message: code.Message(),
		Data: map[string]interface{}{
			"error": err.Error(),
		},
		TraceID: c.GetString(RequestIDKey),
	})
}

// RespondError 返回错误响应 (HTTP 状态码 + 消息)
func RespondError(c *gin.Context, httpStatus int, message string) {
	code := httpStatus * 1000
	c.JSON(httpStatus, Response{
		Code:    code,
		Message: message,
		TraceID: c.GetString(RequestIDKey),
	})
}

// RespondSuccess 返回成功响应 (无数据)
func RespondSuccess(c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Code:    http.StatusOK,
		Message: "success",
		TraceID: c.GetString(RequestIDKey),
	})
}

// RespondSuccess 返回成功响应 (带数据)
func RespondSuccess(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    http.StatusOK,
		Message: "success",
		Data:    data,
		TraceID: c.GetString(RequestIDKey),
	})
}

// RespondSuccess 返回成功响应 (带数据和消息)
func RespondSuccess(c *gin.Context, data interface{}, message string) {
	c.JSON(http.StatusOK, Response{
		Code:    http.StatusOK,
		Message: message,
		Data:    data,
		TraceID: c.GetString(RequestIDKey),
	})
}

// httpStatusFromCode 根据错误码获取 HTTP 状态码
func httpStatusFromCode(code int) int {
	switch {
	case code >= 400000 && code < 401000:
		return http.StatusBadRequest
	case code >= 401000 && code < 403000:
		return http.StatusUnauthorized
	case code >= 403000 && code < 404000:
		return http.StatusForbidden
	case code >= 404000 && code < 429000:
		return http.StatusNotFound
	case code >= 429000 && code < 500000:
		return http.StatusTooManyRequests
	default:
		return http.StatusInternalServerError
	}
}
```

注意：需要添加 `import "fmt"`

**Step 4: 运行测试验证通过**

Run: `go test ./middleware -run TestRespond -v`
Expected: PASS

**Step 5: 提交**

```bash
git add middleware/response.go middleware/response_test.go
git commit -m "feat(middleware): 实现 Response 和 RespondError/RespondSuccess"
```

---

### Task 4: 实现预定义错误码

**Files:**
- Create: `middleware/std_errors.go`
- Create: `middleware/jwt_errors.go`
- Create: `middleware/rate_limit_errors.go`
- Create: `middleware/config_errors.go`
- Create: `middleware/log_errors.go`

**Step 1: 实现基础错误码**

Create `middleware/std_errors.go`:

```go
package middleware

// 基础错误码
var (
	StdErrInternalServer = NewCode(500001, "Internal server error")
)
```

Create `middleware/jwt_errors.go`:

```go
package middleware

// JWT 预定义错误码
var (
	JwtErrMissingToken   = NewCode(401001, "Missing token")
	JwtErrInvalidToken   = NewCode(401002, "Invalid token")
	JwtErrTokenBlocked   = NewCode(401003, "Token has been blocked")
	JwtErrUserNotFound   = NewCode(404001, "User not found")
)
```

Create `middleware/rate_limit_errors.go`:

```go
package middleware

// 限流预定义错误码
var (
	RateLimitErrExceeded = NewCode(429001, "Rate limit exceeded")
)
```

**Step 2: 提交**

```bash
git add middleware/std_errors.go middleware/jwt_errors.go middleware/rate_limit_errors.go
git commit -m "feat(middleware): 添加预定义错误码"
```

---

## Phase 3: Zap 日志封装（log/log.go）

### Task 5: 实现 Logger 结构和 Load 函数

**Files:**
- Create: `log/log.go`
- Create: `log/log_test.go`

**Step 1: 编写测试**

Create `log/log_test.go`:

```go
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
```

**Step 2: 运行测试验证失败**

Run: `go test ./log -v`
Expected: FAIL with "undefined: Load"

**Step 3: 实现代码**

Create `log/log.go`:

```go
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
	Format      string  // json, console
	OutputFile  string  // 日志文件路径（为空则输出到 stdout）
	MaxSize     int     // 单文件最大大小 (MB)
	MaxBackups  int     // 最大保留的旧日志文件数
	MaxAge      int     // 日志最大保留时间 (小时)
	Compress    bool    // 是否压缩旧日志
	Caller      bool    // 是否显示调用者信息
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
	l.zap.Sugar().Debugw(msg, toZapFields(keys...)...)
}

// Info 日志
func (l *Logger) Info(msg string, keys ...interface{}) {
	l.zap.Sugar().Infow(msg, toZapFields(keys...)...)
}

// Warn 日志
func (l *Logger) Warn(msg string, keys ...interface{}) {
	l.zap.Sugar().Warnw(msg, toZapFields(keys...)...)
}

// Error 日志
func (l *Logger) Error(msg string, keys ...interface{}) {
	l.zap.Sugar().Errorw(msg, toZapFields(keys...)...)
}

// Fatal 日志
func (l *Logger) Fatal(msg string, keys ...interface{}) {
	l.zap.Sugar().Fatalw(msg, toZapFields(keys...)...)
}

// With 创建带字段的 logger
func (l *Logger) With(keys ...interface{}) *Logger {
	return &Logger{zap: l.zap.With(toZapFields(keys...)...)}
}

func toZapFields(keys ...interface{}) []zap.Field {
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
```

**Step 4: 运行测试验证通过**

Run: `go test ./log -v`
Expected: PASS

**Step 5: 提交**

```bash
git add log/log.go log/log_test.go
git commit -m "feat(log): 实现 Zap 日志封装"
```

---

### Task 6: 实现日志错误码

**Files:**
- Create: `log/errors.go`

**Step 1: 实现代码**

Create `log/errors.go`:

```go
package log

import "your-project/lib/middleware"

// 日志预定义错误码
var (
	LogErrInitFailed = middleware.NewCode(500201, "Failed to initialize logger")
)
```

**Step 2: 提交**

```bash
git add log/errors.go
git commit -m "feat(log): 添加日志错误码"
```

---

## Phase 4: Viper 配置封装（config/config.go）

### Task 7: 实现 ConfigLoader 和 Load 函数

**Files:**
- Create: `config/config.go`
- Create: `config/config_test.go`

**Step 1: 编写测试**

Create `config/config_test.go`:

```go
package config

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

type TestConfig struct {
	Server ServerConfig `mapstructure:"server" validate:"required"`
}

type ServerConfig struct {
	Port int `mapstructure:"port" validate:"required,min=1,max=65535"`
}

func TestLoad(t *testing.T) {
	// 需要配置文件支持
	cfg, err := Load[TestConfig](
		WithName("test-config"),
		WithType("yaml"),
		WithPaths("./testdata"),
		WithValidate(true),
	)
	assert.NoError(t, err)
	assert.Equal(t, 8080, cfg.Server.Port)
}

func TestMustLoad(t *testing.T) {
	cfg := MustLoad[TestConfig](
		WithName("test-config"),
		WithType("yaml"),
		WithPaths("./testdata"),
	)
	assert.NotNil(t, cfg)
}
```

**Step 2: 运行测试验证失败**

Run: `go test ./config -v`
Expected: FAIL with "undefined: Load"

**Step 3: 实现代码**

Create `config/config.go`:

```go
package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"github.com/go-playground/validator/v10"
)

// ConfigLoader 配置加载器
type ConfigLoader struct {
	name        string
	env         string
	configType  string
	paths       []string
	envPrefix   string
	envReplacer *strings.Replacer
	watch       bool
	validate    bool
	onChange    func(ConfigChangeEvent)
}

// ConfigChangeEvent 配置变更事件
type ConfigChangeEvent struct {
	Path string
}

// LoadOption 配置加载选项
type LoadOption func(*ConfigLoader)

// WithName 配置文件名
func WithName(name string) LoadOption {
	return func(l *ConfigLoader) {
		l.name = name
	}
}

// WithEnv 运行环境
func WithEnv(env string) LoadOption {
	return func(l *ConfigLoader) {
		l.env = env
	}
}

// WithType 配置文件类型
func WithType(t string) LoadOption {
	return func(l *ConfigLoader) {
		l.configType = t
	}
}

// WithPaths 配置文件搜索路径
func WithPaths(paths ...string) LoadOption {
	return func(l *ConfigLoader) {
		l.paths = paths
	}
}

// WithEnvPrefix 环境变量前缀
func WithEnvPrefix(prefix string) LoadOption {
	return func(l *ConfigLoader) {
		l.envPrefix = prefix
	}
}

// WithEnvReplacer 环境变量替换规则
func WithEnvReplacer(replacer *strings.Replacer) LoadOption {
	return func(l *ConfigLoader) {
		l.envReplacer = replacer
	}
}

// WithWatch 启用配置热重载
func WithWatch(watch bool) LoadOption {
	return func(l *ConfigLoader) {
		l.watch = watch
	}
}

// WithValidate 启用配置验证
func WithValidate(validate bool) LoadOption {
	return func(l *ConfigLoader) {
		l.validate = validate
	}
}

// WithOnChange 配置变更回调
func WithOnChange(onChange func(ConfigChangeEvent)) LoadOption {
	return func(l *ConfigLoader) {
		l.onChange = onChange
	}
}

// Load 加载配置
func Load[T any](opts ...LoadOption) (*T, error) {
	loader := &ConfigLoader{
		name:       "config",
		env:        os.Getenv("APP_ENV"),
		configType: "yaml",
		paths:      []string{".", "./configs"},
	}

	for _, opt := range opts {
		opt(loader)
	}

	v := viper.New()

	configName := loader.name
	if loader.env != "" {
		configName = fmt.Sprintf("%s.%s", loader.name, loader.env)
	}

	v.SetConfigName(configName)
	v.SetConfigType(loader.configType)

	for _, path := range loader.paths {
		v.AddConfigPath(path)
	}

	if loader.envPrefix != "" {
		v.SetEnvPrefix(loader.envPrefix)
		v.AutomaticEnv()
		if loader.envReplacer != nil {
			v.SetEnvKeyReplacer(loader.envReplacer)
		}
	}

	if err := v.ReadInConfig(); err != nil {
		var configNotFound viper.ConfigFileNotFoundError
		if !errors.As(err, &configNotFound) {
			return nil, fmt.Errorf("failed to load config: %w", err)
		}
		return nil, fmt.Errorf("config file not found: %s.%s", configName, loader.configType)
	}

	if loader.watch {
		v.WatchConfig()
		if loader.onChange != nil {
			v.OnConfigChange(func(e fsnotify.Event) {
				loader.onChange(ConfigChangeEvent{Path: e.Name})
			})
		}
	}

	var cfg T
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if loader.validate {
		validate := validator.New()
		if err := validate.Struct(&cfg); err != nil {
			return nil, fmt.Errorf("config validation failed: %w", err)
		}
	}

	return &cfg, nil
}

// MustLoad 加载配置（失败则 panic）
func MustLoad[T any](opts ...LoadOption) *T {
	cfg, err := Load[T](opts...)
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}
	return cfg
}
```

**Step 4: 运行测试验证通过**

Run: `go test ./config -v`
Expected: PASS（需要创建 testdata/test-config.yaml）

**Step 5: 提交**

```bash
git add config/config.go config/config_test.go
git commit -m "feat(config): 实现 Viper 配置封装"
```

---

### Task 8: 实现配置错误码

**Files:**
- Create: `config/errors.go`

**Step 1: 实现代码**

Create `config/errors.go`:

```go
package config

import "your-project/lib/middleware"

// 配置预定义错误码
var (
	CfgErrLoadFailed      = middleware.NewCode(500101, "Failed to load configuration")
	CfgErrUnmarshalFailed = middleware.NewCode(500102, "Failed to unmarshal configuration")
	CfgErrValidateFailed  = middleware.NewCode(500103, "Configuration validation failed")
	CfgErrFileNotFound    = middleware.NewCode(400101, "Configuration file not found")
)
```

**Step 2: 提交**

```bash
git add config/errors.go
git commit -m "feat(config): 添加配置错误码"
```

---

## Phase 5: Gin 中间件实现

### Task 9: 实现 Recovery 中间件

**Files:**
- Create: `middleware/recovery.go`
- Test: `middleware/recovery_test.go`

**Step 1: 编写测试**

Create `middleware/recovery_test.go`:

```go
package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRecoveryMiddleware(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	handler := RecoveryMiddleware(RecoveryMiddlewareConfig{})(func(c *gin.Context) {
		panic("test panic")
	})

	handler(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Internal server error")
}
```

**Step 2: 实现代码**

Create `middleware/recovery.go`:

```go
package middleware

import (
	"runtime/debug"
	"github.com/gin-gonic/gin"
)

// RecoveryMiddlewareConfig Recovery 中间件配置
type RecoveryMiddlewareConfig struct {
	Logger *Logger
}

// Recovery 恢复中间件
func RecoveryMiddleware(cfg RecoveryMiddlewareConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				if cfg.Logger != nil {
					cfg.Logger.Error("panic recovered",
						"error", err,
						"stack", string(debug.Stack()),
					)
				}
				RespondError(c, StdErrInternalServer)
				c.Abort()
			}
		}()
		c.Next()
	}
}
```

注意：需要导入 `log` 包或定义本地 Logger 接口

**Step 3: 提交**

```bash
git add middleware/recovery.go middleware/recovery_test.go
git commit -m "feat(middleware): 实现 Recovery 中间件"
```

---

### Task 10: 实现 RequestID 中间件

**Files:**
- Create: `middleware/request_id.go`
- Test: `middleware/request_id_test.go`

**Step 1: 实现代码**

Create `middleware/request_id.go`:

```go
package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestIDMiddlewareConfig RequestID 配置
type RequestIDMiddlewareConfig struct {
	HeaderName string
}

// RequestIDMiddleware 请求 ID 中间件
func RequestIDMiddleware(cfg RequestIDMiddlewareConfig) gin.HandlerFunc {
	if cfg.HeaderName == "" {
		cfg.HeaderName = "X-Request-ID"
	}

	return func(c *gin.Context) {
		requestID := c.GetHeader(cfg.HeaderName)
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set(RequestIDKey, requestID)
		c.Writer.Header().Set(cfg.HeaderName, requestID)
		c.Next()
	}
}
```

**Step 2: 提交**

```bash
git add middleware/request_id.go
git commit -m "feat(middleware): 实现 RequestID 中间件"
```

---

### Task 11: 实现 CORS 中间件

**Files:**
- Create: `middleware/cors.go`
- Test: `middleware/cors_test.go`

**Step 1: 实现代码**

Create `middleware/cors.go`:

```go
package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
)

// CorsMiddlewareConfig CORS 配置
type CorsMiddlewareConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	AllowCredentials bool
}

// CorsMiddleware CORS 中间件
func CorsMiddleware(cfg CorsMiddlewareConfig) gin.HandlerFunc {
	if len(cfg.AllowOrigins) == 0 {
		cfg.AllowOrigins = []string{"*"}
	}
	if len(cfg.AllowMethods) == 0 {
		cfg.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	}
	if len(cfg.AllowHeaders) == 0 {
		cfg.AllowHeaders = []string{"Content-Type", "Authorization"}
	}

	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", cfg.AllowOrigins[0])
		c.Writer.Header().Set("Access-Control-Allow-Credentials", strconv.FormatBool(cfg.AllowCredentials))
		c.Writer.Header().Set("Access-Control-Allow-Headers", strings.Join(cfg.AllowHeaders, ", "))
		c.Writer.Header().Set("Access-Control-Allow-Methods", strings.Join(cfg.AllowMethods, ", "))

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
```

**Step 2: 提交**

```bash
git add middleware/cors.go
git commit -m "feat(middleware): 实现 CORS 中间件"
```

---

### Task 12: 实现 Timeout 中间件

**Files:**
- Create: `middleware/timeout.go`
- Test: `middleware/timeout_test.go`

**Step 1: 实现代码**

Create `middleware/timeout.go`:

```go
package middleware

import (
	"context"
	"github.com/gin-gonic/gin"
	"time"
)

// TimeoutMiddlewareConfig 超时配置
type TimeoutMiddlewareConfig struct {
	Timeout time.Duration
}

// TimeoutMiddleware 超时中间件
func TimeoutMiddleware(cfg TimeoutMiddlewareConfig) gin.HandlerFunc {
	if cfg.Timeout <= 0 {
		cfg.Timeout = 30 * time.Second
	}

	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), cfg.Timeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
```

**Step 2: 提交**

```bash
git add middleware/timeout.go
git commit -m "feat(middleware): 实现 Timeout 中间件"
```

---

### Task 13: 实现 JWT 中间件

**Files:**
- Create: `middleware/jwt.go`
- Test: `middleware/jwt_test.go`

**Step 1: 实现代码**

Create `middleware/jwt.go`:

```go
package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JwtValidator JWT 验证接口
type JwtValidator interface {
	ValidateToken(c *gin.Context, claims jwt.Claims) bool
	GetUser(c *gin.Context, claims jwt.Claims) bool
}

// JwtConfig JWT 中间件配置
type JwtConfig struct {
	Secret        []byte
	SigningMethod jwt.SigningMethod
	TokenHeader   string
	TokenPrefix   string
	Validator     JwtValidator
}

// JwtClaimsFunc Claims 工厂函数
type JwtClaimsFunc func() jwt.Claims

// JwtAuth 创建 JWT 认证中间件
func JwtAuth(cfg JwtConfig, claimsFunc JwtClaimsFunc) gin.HandlerFunc {
	if cfg.SigningMethod == nil {
		cfg.SigningMethod = jwt.SigningMethodHS256
	}
	if cfg.TokenHeader == "" {
		cfg.TokenHeader = "Authorization"
	}
	if cfg.TokenPrefix == "" {
		cfg.TokenPrefix = "Bearer "
	}

	return func(c *gin.Context) {
		tokenStr := c.GetHeader(cfg.TokenHeader)
		if tokenStr == "" {
			RespondError(c, JwtErrMissingToken)
			c.Abort()
			return
		}

		if cfg.TokenPrefix != "" && len(tokenStr) > len(cfg.TokenPrefix) {
			tokenStr = tokenStr[len(cfg.TokenPrefix):]
		}

		claims := claimsFunc()
		if claims == nil {
			claims = &jwt.RegisteredClaims{}
		}

		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return cfg.Secret, nil
		})

		if err != nil || !token.Valid {
			RespondError(c, JwtErrInvalidToken)
			c.Abort()
			return
		}

		if cfg.Validator != nil {
			if !cfg.Validator.ValidateToken(c, claims) {
				RespondError(c, JwtErrTokenBlocked)
				c.Abort()
				return
			}
			if !cfg.Validator.GetUser(c, claims) {
				RespondError(c, JwtErrUserNotFound)
				c.Abort()
				return
			}
		}

		c.Set("jwt_claims", claims)
		c.Next()
	}
}

// GetJwtClaims 从 context 获取 Claims
func GetJwtClaims[T jwt.Claims](c *gin.Context) T {
	claims, _ := c.Get("jwt_claims")
	var zero T
	if claims == nil {
		return zero
	}
	return claims.(T)
}
```

**Step 2: 提交**

```bash
git add middleware/jwt.go
git commit -m "feat(middleware): 实现 JWT 中间件"
```

---

### Task 14: 实现固定窗口限流中间件

**Files:**
- Create: `middleware/rate_limit_fixed.go`
- Test: `middleware/rate_limit_fixed_test.go`

**Step 1: 实现代码**

Create `middleware/rate_limit_fixed.go`:

```go
package middleware

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"time"
)

// RateLimitFixedConfig 固定窗口限流配置
type RateLimitFixedConfig struct {
	Rate       int
	Burst      int
	KeyFunc    RateLimitKeyFunc
	Expiration time.Duration
}

// RateLimitKeyFunc 限流 Key 生成函数
type RateLimitKeyFunc func(c *gin.Context) string

// RateLimitFixed 固定窗口限流中间件
func RateLimitFixed(client *redis.Client, cfg RateLimitFixedConfig) gin.HandlerFunc {
	if cfg.Rate <= 0 {
		cfg.Rate = 100
	}
	if cfg.Burst <= 0 {
		cfg.Burst = cfg.Rate
	}
	if cfg.Expiration == 0 {
		cfg.Expiration = time.Hour
	}
	if cfg.KeyFunc == nil {
		cfg.KeyFunc = func(c *gin.Context) string {
			return c.ClientIP()
		}
	}

	return func(c *gin.Context) {
		key := fmt.Sprintf("ratelimit:fixed:%s", cfg.KeyFunc(c))

		allowed, err := runFixedLimitScript(c, client, key, cfg.Burst, cfg.Expiration)
		if err != nil {
			c.Next()
			return
		}

		if !allowed {
			RespondError(c, RateLimitErrExceeded)
			c.Abort()
			return
		}

		c.Next()
	}
}

const fixedLimitScript = `
local key = KEYS[1]
local burst = tonumber(ARGV[1])
local expiration = tonumber(ARGV[2])

local current = redis.call('INCR', key)
if current == 1 then
    redis.call('EXPIRE', key, expiration)
end

if current > burst then
    return 0
end

return 1
`

func runFixedLimitScript(c *gin.Context, client *redis.Client, key string, burst int, expiration time.Duration) (bool, error) {
	result, err := client.Eval(
		context.Background(),
		fixedLimitScript,
		[]string{key},
		burst, int(expiration.Seconds()),
	).Int()

	if err != nil {
		return false, err
	}

	return result == 1, nil
}
```

**Step 2: 提交**

```bash
git add middleware/rate_limit_fixed.go
git commit -m "feat(middleware): 实现固定窗口限流中间件"
```

---

### Task 15: 实现滑动窗口限流中间件

**Files:**
- Create: `middleware/rate_limit_sliding.go`
- Test: `middleware/rate_limit_sliding_test.go`

**Step 1: 实现代码**

Create `middleware/rate_limit_sliding.go`:

```go
package middleware

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"time"
)

// RateLimitSlidingConfig 滑动窗口限流配置
type RateLimitSlidingConfig struct {
	Rate    int
	Window  time.Duration
	KeyFunc RateLimitKeyFunc
}

// RateLimitSliding 滑动窗口限流中间件
func RateLimitSliding(client *redis.Client, cfg RateLimitSlidingConfig) gin.HandlerFunc {
	if cfg.Rate <= 0 {
		cfg.Rate = 100
	}
	if cfg.Window == 0 {
		cfg.Window = time.Second
	}
	if cfg.KeyFunc == nil {
		cfg.KeyFunc = func(c *gin.Context) string {
			return c.ClientIP()
		}
	}

	return func(c *gin.Context) {
		key := fmt.Sprintf("ratelimit:sliding:%s", cfg.KeyFunc(c))

		allowed, err := runSlidingLimitScript(c, client, key, cfg.Rate, cfg.Window)
		if err != nil {
			c.Next()
			return
		}

		if !allowed {
			RespondError(c, RateLimitErrExceeded)
			c.Abort()
			return
		}

		c.Next()
	}
}

const slidingLimitScript = `
local key = KEYS[1]
local rate = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local now = tonumber(ARGV[3])

redis.call('ZREMRANGEBYSCORE', key, 0, now - window)

local count = redis.call('ZCARD', key)

if count >= rate then
    return 0
end

redis.call('ZADD', key, now, now .. '-' .. math.random(100000))
redis.call('EXPIRE', key, math.ceil(window))

return 1
`

func runSlidingLimitScript(c *gin.Context, client *redis.Client, key string, rate int, window time.Duration) (bool, error) {
	now := time.Now().UnixMilli()

	result, err := client.Eval(
		context.Background(),
		slidingLimitScript,
		[]string{key},
		rate, int(window.Milliseconds()), now,
	).Int()

	if err != nil {
		return false, err
	}

	return result == 1, nil
}
```

**Step 2: 提交**

```bash
git add middleware/rate_limit_sliding.go
git commit -m "feat(middleware): 实现滑动窗口限流中间件"
```

---

### Task 16: 实现 Logger 中间件和 Chain 中间件

**Files:**
- Create: `middleware/logger.go`
- Create: `middleware/chain.go`

**Step 1: 实现代码**

Create `middleware/logger.go`:

```go
package middleware

import (
	"time"
	"github.com/gin-gonic/gin"
)

// LoggerMiddlewareConfig 日志中间件配置
type LoggerMiddlewareConfig struct {
	Logger    *Logger
	SkipPaths []string
}

// LoggerMiddleware 请求日志中间件
func LoggerMiddleware(cfg LoggerMiddlewareConfig) gin.HandlerFunc {
	skipMap := make(map[string]bool)
	for _, path := range cfg.SkipPaths {
		skipMap[path] = true
	}

	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		if skipMap[c.Request.URL.Path] {
			return
		}

		if cfg.Logger != nil {
			cfg.Logger.Info("request",
				"method", c.Request.Method,
				"path", c.Request.URL.Path,
				"status", c.Writer.Status(),
				"duration", time.Since(start),
				"trace_id", c.GetString(RequestIDKey),
			)
		}
	}
}
```

Create `middleware/chain.go`:

```go
package middleware

import "github.com/gin-gonic/gin"

// ChainMiddleware 链式组合中间件
func ChainMiddleware(handlers ...gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		for _, h := range handlers {
			h(c)
			if c.IsAborted() {
				return
			}
		}
	}
}
```

**Step 2: 提交**

```bash
git add middleware/logger.go middleware/chain.go
git commit -m "feat(middleware): 实现 Logger 和 Chain 中间件"
```

---

## Phase 6: CRUD 基础类型

### Task 17: 实现 Entity 接口和 BaseEntity

**Files:**
- Create: `gincrud/entity.go`
- Test: `gincrud/entity_test.go`

**Step 1: 实现代码**

Create `gincrud/entity.go`:

```go
package gincrud

import "time"

// Entity 实体接口
type Entity interface {
	GetID() int64
	SetID(id int64)
	GetDeleted() int64
	SetDeleted(ts int64)
	GetCreatedAt() time.Time
	SetCreatedAt(t time.Time)
	GetUpdatedAt() time.Time
	SetUpdatedAt(t time.Time)
}

// BaseEntity 基础实体
type BaseEntity struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	Deleted   int64     `gorm:"default:0;index" json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (e *BaseEntity) GetID() int64        { return e.ID }
func (e *BaseEntity) SetID(id int64)      { e.ID = id }
func (e *BaseEntity) GetDeleted() int64   { return e.Deleted }
func (e *BaseEntity) SetDeleted(ts int64) { e.Deleted = ts }
func (e *BaseEntity) GetCreatedAt() time.Time { return e.CreatedAt }
func (e *BaseEntity) SetCreatedAt(t time.Time) { e.CreatedAt = t }
func (e *BaseEntity) GetUpdatedAt() time.Time { return e.UpdatedAt }
func (e *BaseEntity) SetUpdatedAt(t time.Time) { e.UpdatedAt = t }

// IsDeleted 检查是否已删除
func (e *BaseEntity) IsDeleted() bool {
	return e.Deleted > 0
}

// MarkDeleted 标记为删除
func (e *BaseEntity) MarkDeleted() {
	e.Deleted = time.Now().Unix()
}

// AuditEntity 审计实体
type AuditEntity struct {
	BaseEntity
	CreatedBy int64 `gorm:"default:0" json:"created_by"`
	UpdatedBy int64 `gorm:"default:0" json:"updated_by"`
}

func (e *AuditEntity) GetCreatedBy() int64 { return e.CreatedBy }
func (e *AuditEntity) SetCreatedBy(id int64) { e.CreatedBy = id }
func (e *AuditEntity) GetUpdatedBy() int64 { return e.UpdatedBy }
func (e *AuditEntity) SetUpdatedBy(id int64) { e.UpdatedBy = id }
```

**Step 2: 提交**

```bash
git add gincrud/entity.go
git commit -m "feat(gincrud): 实现 Entity 接口和 BaseEntity"
```

---

### Task 18: 实现 QueryDTO 接口和 BaseQueryDTO

**Files:**
- Create: `gincrud/query.go`
- Test: `gincrud/query_test.go`

**Step 1: 实现代码**

Create `gincrud/query.go`:

```go
package gincrud

// QueryDTO 查询参数接口
type QueryDTO interface {
	GetPage() int
	SetPage(page int)
	GetPageSize() int
	SetPageSize(size int)
	GetSortBy() string
	SetSortBy(field string)
	GetSortOrder() string
	SetSortOrder(order string)
	Normalize()
	Offset() int
	Limit() int
}

// BaseQueryDTO 基础查询参数
type BaseQueryDTO struct {
	Page      int    `form:"page" binding:"min=1"`
	PageSize  int    `form:"page_size" binding:"min=1,max=100"`
	SortBy    string `form:"sort_by"`
	SortOrder string `form:"sort_order" binding:"oneof=asc desc"`
}

func (q *BaseQueryDTO) GetPage() int      { return q.Page }
func (q *BaseQueryDTO) SetPage(page int)  { q.Page = page }
func (q *BaseQueryDTO) GetPageSize() int  { return q.PageSize }
func (q *BaseQueryDTO) SetPageSize(size int) { q.PageSize = size }
func (q *BaseQueryDTO) GetSortBy() string { return q.SortBy }
func (q *BaseQueryDTO) SetSortBy(field string) { q.SortBy = field }
func (q *BaseQueryDTO) GetSortOrder() string { return q.SortOrder }
func (q *BaseQueryDTO) SetSortOrder(order string) { q.SortOrder = order }

func (q *BaseQueryDTO) Normalize() {
	if q.Page < 1 { q.Page = 1 }
	if q.PageSize < 1 { q.PageSize = 10 }
	if q.PageSize > 100 { q.PageSize = 100 }
	if q.SortBy == "" { q.SortBy = "id" }
	if q.SortOrder == "" { q.SortOrder = "desc" }
}

func (q *BaseQueryDTO) Offset() int { return (q.Page - 1) * q.PageSize }
func (q *BaseQueryDTO) Limit() int {
	if q.PageSize > 100 { return 100 }
	return q.PageSize
}
```

**Step 2: 提交**

```bash
git add gincrud/query.go
git commit -m "feat(gincrud): 实现 QueryDTO 接口和 BaseQueryDTO"
```

---

### Task 19: 实现 PageResult

**Files:**
- Create: `gincrud/page.go`
- Test: `gincrud/page_test.go`

**Step 1: 实现代码**

Create `gincrud/page.go`:

```go
package gincrud

// PageResult 分页响应
type PageResult[T any] struct {
	List     []T   `json:"list"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
}

// NewPageResult 创建分页响应
func NewPageResult[T any](list []T, total int64, query QueryDTO) *PageResult[T] {
	return &PageResult[T]{
		List:     list,
		Total:    total,
		Page:     query.GetPage(),
		PageSize: query.GetPageSize(),
	}
}
```

**Step 2: 提交**

```bash
git add gincrud/page.go
git commit -m "feat(gincrud): 实现 PageResult"
```

---

## Phase 7: 示例代码和文档

### Task 20: 创建 examples/basic 示例

**Files:**
- Create: `examples/basic/main.go`
- Create: `examples/basic/config.yaml`

**Step 1: 创建示例代码**

Create `examples/basic/main.go`:

```go
package main

import (
	"github.com/gin-gonic/gin"
	"your-project/lib/middleware"
	"your-project/lib/log"
)

func main() {
	logger := log.MustLoad(log.Config{
		Level:      log.InfoLevel,
		Format:     "json",
		OutputFile: "/var/log/app/app.log",
		MaxSize:    100,
		MaxBackups: 7,
		Compress:   true,
	})
	log.SetGlobal(logger)

	r := gin.New()
	r.Use(middleware.RecoveryMiddleware(middleware.RecoveryMiddlewareConfig{}))
	r.Use(middleware.RequestIDMiddleware(middleware.RequestIDMiddlewareConfig{}))
	r.Use(middleware.LoggerMiddleware(middleware.LoggerMiddlewareConfig{
		Logger: logger,
	}))

	r.GET("/hello", func(c *gin.Context) {
		middleware.RespondSuccess(c, gin.H{"message": "Hello, World!"})
	})

	r.Run(":8080")
}
```

**Step 2: 提交**

```bash
git add examples/basic/
git commit -m "docs: 添加基础示例"
```

---

### Task 21: 创建 README.md

**Files:**
- Create: `README.md`

**Step 1: 创建文档**

参考前面讨论的 README 内容创建。

**Step 2: 提交**

```bash
git add README.md
git commit -m "docs: 添加 README"
```

---

## 总结

本计划共 21 个 Task，涵盖：
1. 项目骨架搭建
2. 统一响应模块
3. Zap 日志封装
4. Viper 配置封装
5. Gin 中间件（Recovery, RequestID, CORS, Timeout, JWT, RateLimit, Logger, Chain）
6. CRUD 基础类型（Entity, QueryDTO, PageResult）
7. 示例代码和文档

预计完成时间：约 2-4 小时（取决于测试覆盖率和代码审查）
