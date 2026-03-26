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
