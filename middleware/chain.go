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
