package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
)

// LoggerMiddlewareConfig 日志中间件配置
type LoggerMiddlewareConfig struct {
	Logger    Logger
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
