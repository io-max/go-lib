package middleware

import (
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
			if err := recover(); err != nil {
				if cfg.Logger != nil {
					cfg.Logger.Error("panic recovered",
						"error", err,
						"stack", string(debug.Stack()),
					)
				}
				RespondErrorWithCode(c, ErrInternalServer)
				c.Abort()
			}
		}()
		c.Next()
	}
}
