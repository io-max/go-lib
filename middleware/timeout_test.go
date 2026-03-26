package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestTimeoutMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 测试默认配置（30 秒超时）
	t.Run("DefaultConfig", func(t *testing.T) {
		router := gin.New()
		router.Use(TimeoutMiddleware(TimeoutMiddlewareConfig{}))

		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "ok")
	})

	// 测试自定义超时时间
	t.Run("CustomTimeout", func(t *testing.T) {
		router := gin.New()
		router.Use(TimeoutMiddleware(TimeoutMiddlewareConfig{
			Timeout: 5 * time.Second,
		}))

		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "ok")
	})

	// 测试超时情况
	t.Run("TimeoutOccurred", func(t *testing.T) {
		router := gin.New()
		router.Use(TimeoutMiddleware(TimeoutMiddlewareConfig{
			Timeout: 100 * time.Millisecond,
		}))

		var contextErr error
		router.GET("/test", func(c *gin.Context) {
			// 模拟长时间运行的操作
			select {
			case <-time.After(200 * time.Millisecond):
				c.JSON(http.StatusOK, gin.H{"status": "completed"})
			case <-c.Request.Context().Done():
				// 超时被取消
				contextErr = c.Request.Context().Err()
				c.Abort()
				return
			}
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		router.ServeHTTP(w, req)

		// 验证请求被超时取消
		assert.Equal(t, context.DeadlineExceeded, contextErr)
	})

	// 测试请求上下文被正确传递
	t.Run("ContextPassedCorrectly", func(t *testing.T) {
		router := gin.New()
		router.Use(TimeoutMiddleware(TimeoutMiddlewareConfig{
			Timeout: 5 * time.Second,
		}))

		var capturedContext context.Context
		router.GET("/test", func(c *gin.Context) {
			capturedContext = c.Request.Context()
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.NotNil(t, capturedContext)
		// 验证上下文有超时设置
		_, hasDeadline := capturedContext.Deadline()
		assert.True(t, hasDeadline)
	})

	// 测试负数超时时间使用默认值
	t.Run("NegativeTimeoutUsesDefault", func(t *testing.T) {
		router := gin.New()
		router.Use(TimeoutMiddleware(TimeoutMiddlewareConfig{
			Timeout: -1 * time.Second,
		}))

		router.GET("/test", func(c *gin.Context) {
			_, hasDeadline := c.Request.Context().Deadline()
			assert.True(t, hasDeadline)
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	// 测试零超时时间使用默认值
	t.Run("ZeroTimeoutUsesDefault", func(t *testing.T) {
		router := gin.New()
		router.Use(TimeoutMiddleware(TimeoutMiddlewareConfig{
			Timeout: 0,
		}))

		router.GET("/test", func(c *gin.Context) {
			_, hasDeadline := c.Request.Context().Deadline()
			assert.True(t, hasDeadline)
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	// 测试超时后上下文被取消
	t.Run("ContextCancelledOnTimeout", func(t *testing.T) {
		router := gin.New()
		router.Use(TimeoutMiddleware(TimeoutMiddlewareConfig{
			Timeout: 50 * time.Millisecond,
		}))

		var contextErr error
		router.GET("/test", func(c *gin.Context) {
			// 等待超时
			<-c.Request.Context().Done()
			contextErr = c.Request.Context().Err()
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		start := time.Now()
		router.ServeHTTP(w, req)
		elapsed := time.Since(start)

		// 验证上下文被取消
		assert.Equal(t, context.DeadlineExceeded, contextErr)
		// 验证超时时间大约在 50ms 左右（允许一些误差）
		assert.True(t, elapsed >= 50*time.Millisecond && elapsed < 200*time.Millisecond)
	})
}
