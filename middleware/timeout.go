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
