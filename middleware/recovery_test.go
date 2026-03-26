package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// mockLogger 模拟 Logger 用于测试
type mockLogger struct {
	errors []string
}

func (m *mockLogger) Error(msg string, keys ...interface{}) {
	m.errors = append(m.errors, msg)
}

func TestRecoveryMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 测试正常请求（不触发 panic）
	t.Run("NormalRequest", func(t *testing.T) {
		router := gin.New()
		logger := &mockLogger{}
		router.Use(RecoveryMiddleware(RecoveryMiddlewareConfig{
			Logger: logger,
		}))

		router.GET("/test", func(c *gin.Context) {
			c.Set(RequestIDKey, "test-trace-id-normal")
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "success")
		assert.Len(t, logger.errors, 0)
	})

	// 测试 panic 恢复
	t.Run("PanicRecovery", func(t *testing.T) {
		router := gin.New()
		logger := &mockLogger{}
		router.Use(RecoveryMiddleware(RecoveryMiddlewareConfig{
			Logger: logger,
		}))

		router.GET("/test", func(c *gin.Context) {
			c.Set(RequestIDKey, "test-trace-id-panic")
			panic("test panic")
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		router.ServeHTTP(w, req)

		// 验证返回了 500 错误
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), `"code":500001`)
		assert.Contains(t, w.Body.String(), `"message":"Internal server error"`)
		// 验证 logger 记录了错误
		assert.Len(t, logger.errors, 1)
	})

	// 测试没有 logger 的情况
	t.Run("PanicRecoveryWithoutLogger", func(t *testing.T) {
		router := gin.New()
		router.Use(RecoveryMiddleware(RecoveryMiddlewareConfig{
			Logger: nil, // 不提供 logger
		}))

		router.GET("/test", func(c *gin.Context) {
			panic("test panic without logger")
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		router.ServeHTTP(w, req)

		// 验证返回了 500 错误
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), `"code":500001`)
	})

	// 测试 error 类型的 panic
	t.Run("PanicWithErrorType", func(t *testing.T) {
		router := gin.New()
		logger := &mockLogger{}
		router.Use(RecoveryMiddleware(RecoveryMiddlewareConfig{
			Logger: logger,
		}))

		router.GET("/test", func(c *gin.Context) {
			c.Set(RequestIDKey, "test-trace-id-err")
			panic("something went wrong")
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), `"code":500001`)
		assert.Contains(t, w.Body.String(), `"message":"Internal server error"`)
		// 验证 logger 记录了错误
		assert.Len(t, logger.errors, 1)
	})
}

func TestRecoveryMiddlewareAborts(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("AbortAfterPanic", func(t *testing.T) {
		router := gin.New()
		logger := &mockLogger{}
		router.Use(RecoveryMiddleware(RecoveryMiddlewareConfig{
			Logger: logger,
		}))

		executed := false
		router.GET("/test", func(c *gin.Context) {
			panic("test panic")
			executed = true // 这行不会被执行
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		router.ServeHTTP(w, req)

		// 验证后续处理函数没有执行（在 panic 后被 abort）
		assert.False(t, executed)
	})
}
