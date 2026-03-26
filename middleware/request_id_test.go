package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRequestIDMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 测试默认配置（无 Request ID）
	t.Run("DefaultConfigWithoutRequestID", func(t *testing.T) {
		router := gin.New()
		router.Use(RequestIDMiddleware(RequestIDMiddlewareConfig{}))

		router.GET("/test", func(c *gin.Context) {
			requestID := c.GetString(RequestIDKey)
			c.JSON(http.StatusOK, gin.H{"request_id": requestID})
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		// 验证生成了新的 Request ID
		requestID := w.Header().Get("X-Request-ID")
		assert.NotEmpty(t, requestID)
		// 验证 Request ID 是有效的 UUID 格式
		assert.True(t, strings.Contains(requestID, "-"))
	})

	// 测试自定义 Header 名称
	t.Run("CustomHeaderName", func(t *testing.T) {
		router := gin.New()
		router.Use(RequestIDMiddleware(RequestIDMiddlewareConfig{
			HeaderName: "X-Custom-Request-ID",
		}))

		router.GET("/test", func(c *gin.Context) {
			requestID := c.GetString(RequestIDKey)
			c.JSON(http.StatusOK, gin.H{"request_id": requestID})
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		// 验证使用了自定义 Header 名称
		customRequestID := w.Header().Get("X-Custom-Request-ID")
		assert.NotEmpty(t, customRequestID)
		// 验证默认 Header 不存在
		assert.Empty(t, w.Header().Get("X-Request-ID"))
	})

	// 测试已有 Request ID 的情况
	t.Run("ExistingRequestID", func(t *testing.T) {
		router := gin.New()
		router.Use(RequestIDMiddleware(RequestIDMiddlewareConfig{}))

		existingID := "test-existing-request-id-12345"

		router.GET("/test", func(c *gin.Context) {
			requestID := c.GetString(RequestIDKey)
			c.JSON(http.StatusOK, gin.H{"request_id": requestID})
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("X-Request-ID", existingID)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		// 验证使用了已有的 Request ID
		responseID := w.Header().Get("X-Request-ID")
		assert.Equal(t, existingID, responseID)
	})

	// 测试已有 Request ID（自定义 Header）
	t.Run("ExistingRequestIDWithCustomHeader", func(t *testing.T) {
		customHeader := "X-My-Request-ID"
		existingID := "custom-existing-request-id-67890"

		router := gin.New()
		router.Use(RequestIDMiddleware(RequestIDMiddlewareConfig{
			HeaderName: customHeader,
		}))

		router.GET("/test", func(c *gin.Context) {
			requestID := c.GetString(RequestIDKey)
			c.JSON(http.StatusOK, gin.H{"request_id": requestID})
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set(customHeader, existingID)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		// 验证使用了已有的 Request ID（自定义 Header）
		responseID := w.Header().Get(customHeader)
		assert.Equal(t, existingID, responseID)
	})

	// 测试 RequestID 被设置到 Context 中
	t.Run("RequestIDSetInContext", func(t *testing.T) {
		router := gin.New()
		router.Use(RequestIDMiddleware(RequestIDMiddlewareConfig{}))

		var capturedRequestID string
		router.GET("/test", func(c *gin.Context) {
			if v, ok := c.Get(RequestIDKey); ok {
				capturedRequestID = v.(string)
			}
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		// 验证 Request ID 被设置到 Context 中
		assert.NotEmpty(t, capturedRequestID)
	})

	// 测试连续请求生成不同的 Request ID
	t.Run("DifferentRequestIDsForDifferentRequests", func(t *testing.T) {
		router := gin.New()
		router.Use(RequestIDMiddleware(RequestIDMiddlewareConfig{}))

		router.GET("/test", func(c *gin.Context) {
			requestID := c.GetString(RequestIDKey)
			c.JSON(http.StatusOK, gin.H{"request_id": requestID})
		})

		// 第一个请求
		w1 := httptest.NewRecorder()
		req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
		router.ServeHTTP(w1, req1)

		// 第二个请求
		w2 := httptest.NewRecorder()
		req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
		router.ServeHTTP(w2, req2)

		id1 := w1.Header().Get("X-Request-ID")
		id2 := w2.Header().Get("X-Request-ID")

		// 验证两个请求有不同的 Request ID
		assert.NotEmpty(t, id1)
		assert.NotEmpty(t, id2)
		assert.NotEqual(t, id1, id2)
	})
}
