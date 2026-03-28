package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

// Logger Logger 接口定义（避免循环依赖）
type Logger interface {
	Error(msg string, keys ...interface{})
	Info(msg string, keys ...interface{})
}

// RecoveryMiddlewareConfig Recovery 中间件配置
type RecoveryMiddlewareConfig struct {
	Logger Logger
}

// Recovery 恢复中间件
func RecoveryMiddleware(cfg RecoveryMiddlewareConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				var code int
				var message string
				var httpStatus int

				// 情况 1: ErrorCode → 业务错误
				if e, ok := r.(ErrorCode); ok {
					code = e.Code()
					message = e.Message()
					httpStatus = httpStatusFromCode(code)
					if cfg.Logger != nil {
						cfg.Logger.Error("business error",
							"code", code,
							"message", message,
						)
					}
				} else if e, ok := r.(error); ok {
					// 情况 2: error → 系统错误
					code = 500001
					message = "Internal server error"
					httpStatus = http.StatusInternalServerError
					if cfg.Logger != nil {
						cfg.Logger.Error("system error",
							"error", e,
							"stack", string(debug.Stack()),
						)
					}
				} else {
					// 情况 3: 未知 panic 值
					code = 500001
					message = "Internal server error"
					httpStatus = http.StatusInternalServerError
					if cfg.Logger != nil {
						cfg.Logger.Error("unknown panic",
							"value", r,
							"stack", string(debug.Stack()),
						)
					}
				}

				// 统一响应
				c.JSON(httpStatus, Response{
					Code:    code,
					Message: message,
					TraceID: c.GetString(RequestIDKey),
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}
