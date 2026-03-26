package middleware

import (
	"fmt"
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

// RespondErrorWithCode 返回错误响应 (ErrorCode)
func RespondErrorWithCode(c *gin.Context, err ErrorCode) {
	httpStatus := httpStatusFromCode(err.Code())
	c.JSON(httpStatus, Response{
		Code:    err.Code(),
		Message: err.Message(),
		TraceID: c.GetString(RequestIDKey),
	})
}

// RespondErrorWithMessage 返回错误响应 (ErrorCode + 自定义消息)
func RespondErrorWithMessage(c *gin.Context, err ErrorCode, message string) {
	httpStatus := httpStatusFromCode(err.Code())
	c.JSON(httpStatus, Response{
		Code:    err.Code(),
		Message: message,
		TraceID: c.GetString(RequestIDKey),
	})
}

// RespondErrorWithFormat 返回错误响应 (ErrorCode + 格式化消息)
func RespondErrorWithFormat(c *gin.Context, err ErrorCode, format string, args ...interface{}) {
	httpStatus := httpStatusFromCode(err.Code())
	c.JSON(httpStatus, Response{
		Code:    err.Code(),
		Message: fmt.Sprintf(format, args...),
		TraceID: c.GetString(RequestIDKey),
	})
}

// RespondErrorWithHTTPStatus 返回错误响应 (HTTP 状态码 + ErrorCode)
func RespondErrorWithHTTPStatus(c *gin.Context, httpStatus int, err ErrorCode) {
	c.JSON(httpStatus, Response{
		Code:    err.Code(),
		Message: err.Message(),
		TraceID: c.GetString(RequestIDKey),
	})
}

// RespondErrorWithDetails 返回错误响应 (error + ErrorCode)
func RespondErrorWithDetails(c *gin.Context, err error, code ErrorCode) {
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

// RespondErrorWithHTTPStatusAndMessage 返回错误响应 (HTTP 状态码 + 消息)
func RespondErrorWithHTTPStatusAndMessage(c *gin.Context, httpStatus int, message string) {
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

// RespondSuccessWithData 返回成功响应 (带数据)
func RespondSuccessWithData(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    http.StatusOK,
		Message: "success",
		Data:    data,
		TraceID: c.GetString(RequestIDKey),
	})
}

// RespondSuccessWithMessage 返回成功响应 (带数据和消息)
func RespondSuccessWithMessage(c *gin.Context, data interface{}, message string) {
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
