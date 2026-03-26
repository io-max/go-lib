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
