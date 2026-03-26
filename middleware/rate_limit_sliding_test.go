package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestRateLimitSlidingMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("DefaultConfig", func(t *testing.T) {
		callCount := 0
		mockRedis := &mockRedisClient{
			evalFunc: func(ctx context.Context, script string, keys []string, args ...interface{}) *redis.Cmd {
				callCount++
				assert.Contains(t, keys[0], "ratelimit:sliding:")
				return redis.NewCmdResult(int64(1), nil)
			},
		}

		router := gin.New()
		router.Use(RateLimitSliding(mockRedis, RateLimitSlidingConfig{}))

		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, 1, callCount)
	})

	t.Run("AllowedRequest", func(t *testing.T) {
		mockRedis := &mockRedisClient{
			evalFunc: func(ctx context.Context, script string, keys []string, args ...interface{}) *redis.Cmd {
				assert.Equal(t, "ratelimit:sliding:192.168.1.1", keys[0])
				return redis.NewCmdResult(int64(1), nil)
			},
		}

		router := gin.New()
		router.Use(RateLimitSliding(mockRedis, RateLimitSlidingConfig{
			Rate:   10,
			Window: time.Second,
			KeyFunc: func(c *gin.Context) string {
				return "192.168.1.1"
			},
		}))

		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("RateLimitedRequest", func(t *testing.T) {
		mockRedis := &mockRedisClient{
			evalFunc: func(ctx context.Context, script string, keys []string, args ...interface{}) *redis.Cmd {
				return redis.NewCmdResult(int64(0), nil)
			},
		}

		router := gin.New()
		router.Use(RateLimitSliding(mockRedis, RateLimitSlidingConfig{
			Rate:   10,
			Window: time.Second,
			KeyFunc: func(c *gin.Context) string {
				return "10.0.0.1"
			},
		}))

		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusTooManyRequests, w.Code)
	})

	t.Run("RedisError", func(t *testing.T) {
		mockRedis := &mockRedisClient{
			evalFunc: func(ctx context.Context, script string, keys []string, args ...interface{}) *redis.Cmd {
				return redis.NewCmdResult(int64(0), context.Canceled)
			},
		}

		router := gin.New()
		router.Use(RateLimitSliding(mockRedis, RateLimitSlidingConfig{
			Rate:   100,
			Window: time.Second,
			KeyFunc: func(c *gin.Context) string {
				return "error-test"
			},
		}))

		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		router.ServeHTTP(w, req)

		// Redis 错误时应该继续处理请求
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("DefaultKeyFuncUsesClientIP", func(t *testing.T) {
		var capturedKey string
		mockRedis := &mockRedisClient{
			evalFunc: func(ctx context.Context, script string, keys []string, args ...interface{}) *redis.Cmd {
				capturedKey = keys[0]
				return redis.NewCmdResult(int64(1), nil)
			},
		}

		router := gin.New()
		router.Use(RateLimitSliding(mockRedis, RateLimitSlidingConfig{
			Rate:   50,
			Window: time.Second,
		}))

		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "172.16.0.1:12345"
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, capturedKey, "172.16.0.1")
	})

	t.Run("ZeroRateDefaults", func(t *testing.T) {
		var capturedArgs []interface{}
		mockRedis := &mockRedisClient{
			evalFunc: func(ctx context.Context, script string, keys []string, args ...interface{}) *redis.Cmd {
				capturedArgs = args
				return redis.NewCmdResult(int64(1), nil)
			},
		}

		router := gin.New()
		router.Use(RateLimitSliding(mockRedis, RateLimitSlidingConfig{}))

		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		// Rate 为 0 时应该使用默认值 100
		assert.Equal(t, int(100), capturedArgs[0])
	})

	t.Run("ZeroWindowDefaults", func(t *testing.T) {
		var capturedArgs []interface{}
		mockRedis := &mockRedisClient{
			evalFunc: func(ctx context.Context, script string, keys []string, args ...interface{}) *redis.Cmd {
				capturedArgs = args
				return redis.NewCmdResult(int64(1), nil)
			},
		}

		router := gin.New()
		router.Use(RateLimitSliding(mockRedis, RateLimitSlidingConfig{
			Rate: 100,
		}))

		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		// Window 为 0 时应该使用默认值 time.Second (1000 毫秒)
		assert.Equal(t, int(1000), capturedArgs[1])
	})

	t.Run("CustomKeyFunc", func(t *testing.T) {
		var capturedKey string
		mockRedis := &mockRedisClient{
			evalFunc: func(ctx context.Context, script string, keys []string, args ...interface{}) *redis.Cmd {
				capturedKey = keys[0]
				return redis.NewCmdResult(int64(1), nil)
			},
		}

		router := gin.New()
		router.Use(RateLimitSliding(mockRedis, RateLimitSlidingConfig{
			Rate:   10,
			Window: 5 * time.Second,
			KeyFunc: func(c *gin.Context) string {
				return "user:12345"
			},
		}))

		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "ratelimit:sliding:user:12345", capturedKey)
	})
}

func TestRunSlidingLimitScript(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("ScriptReturnsAllowed", func(t *testing.T) {
		mockRedis := &mockRedisClient{
			evalFunc: func(ctx context.Context, script string, keys []string, args ...interface{}) *redis.Cmd {
				return redis.NewCmdResult(int64(1), nil)
			},
		}

		ctx := &gin.Context{}
		allowed, err := runSlidingLimitScript(ctx, mockRedis, "test-key", 10, time.Second)

		assert.NoError(t, err)
		assert.True(t, allowed)
	})

	t.Run("ScriptReturnsDenied", func(t *testing.T) {
		mockRedis := &mockRedisClient{
			evalFunc: func(ctx context.Context, script string, keys []string, args ...interface{}) *redis.Cmd {
				return redis.NewCmdResult(int64(0), nil)
			},
		}

		ctx := &gin.Context{}
		allowed, err := runSlidingLimitScript(ctx, mockRedis, "test-key", 5, time.Second)

		assert.NoError(t, err)
		assert.False(t, allowed)
	})

	t.Run("ScriptReturnsError", func(t *testing.T) {
		mockRedis := &mockRedisClient{
			evalFunc: func(ctx context.Context, script string, keys []string, args ...interface{}) *redis.Cmd {
				return redis.NewCmdResult(int64(0), context.DeadlineExceeded)
			},
		}

		ctx := &gin.Context{}
		allowed, err := runSlidingLimitScript(ctx, mockRedis, "test-key", 10, time.Second)

		assert.Error(t, err)
		assert.False(t, allowed)
	})
}
