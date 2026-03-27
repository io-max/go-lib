package middleware

import (
	"fmt"
	"sync"
)

// =============================================================================
// ErrorCode 接口定义
// =============================================================================

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

// =============================================================================
// 错误码工厂函数
// =============================================================================

// errorCodeRegistry 错误码注册表（防止重复）
var (
	errorCodeRegistry = make(map[int]string)
	errorCodeMutex    sync.Mutex
)

// NewCode 创建标准错误码
// 如果错误码已存在，会 panic
func NewCode(code int, message string) ErrorCode {
	errorCodeMutex.Lock()
	defer errorCodeMutex.Unlock()

	if existingMsg, ok := errorCodeRegistry[code]; ok {
		panic(fmt.Sprintf("duplicate error code %d: %s vs %s", code, existingMsg, message))
	}
	errorCodeRegistry[code] = message

	return &StandardErrorCode{code: code, message: message}
}

// NewCodef 创建标准错误码（支持格式化消息）
func NewCodef(code int, format string, args ...interface{}) ErrorCode {
	return NewCode(code, fmt.Sprintf(format, args...))
}

// =============================================================================
// 基础错误码（HTTP 5xx - 服务器错误）
// =============================================================================

var (
	ErrInternalServer     = NewCode(500001, "Internal server error")
	ErrServiceUnavailable = NewCode(503001, "Service unavailable")
	ErrBadGateway         = NewCode(502001, "Bad gateway")
)

// =============================================================================
// 客户端错误码（HTTP 4xx）
// =============================================================================

var (
	ErrInvalidParam    = NewCode(400001, "Invalid parameter")
	ErrUnauthorized    = NewCode(401001, "Unauthorized")
	ErrForbidden       = NewCode(403001, "Forbidden")
	ErrNotFound        = NewCode(404001, "Resource not found")
	ErrConflict        = NewCode(409001, "Resource conflict")
	ErrTooManyRequests = NewCode(429001, "Too many requests")
)

// =============================================================================
// 认证相关错误码（HTTP 401）
// =============================================================================

var (
	ErrMissingToken    = NewCode(401011, "Missing token")
	ErrInvalidToken    = NewCode(401012, "Invalid token")
	ErrTokenExpired    = NewCode(401013, "Token expired")
	ErrTokenBlocked    = NewCode(401014, "Token has been blocked")
	ErrUserNotFound    = NewCode(404002, "User not found")
)

// =============================================================================
// 限流相关错误码（HTTP 429）
// =============================================================================

var (
	ErrRateLimitExceeded = NewCode(429101, "Rate limit exceeded")
)

// =============================================================================
// 数据库相关错误码（HTTP 5xx）
// =============================================================================

var (
	ErrDatabaseOperation = NewCode(500101, "Database operation failed")
	ErrRecordNotFound    = NewCode(404101, "Record not found")
	ErrDuplicateEntry    = NewCode(409101, "Duplicate entry")
)

// =============================================================================
// 验证相关错误码（HTTP 4xx）
// =============================================================================

var (
	ErrValidationFailed  = NewCode(400101, "Validation failed")
	ErrMissingField      = NewCode(400102, "Missing required field")
	ErrInvalidFieldType  = NewCode(400103, "Invalid field type")
	ErrInvalidFieldRange = NewCode(400104, "Invalid field range")
)
