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

// mockRedisClient 是一个简单的 mock Redis 客户端用于测试
type mockRedisClient struct {
	evalFunc func(ctx context.Context, script string, keys []string, args ...interface{}) *redis.Cmd
}

func (m *mockRedisClient) Eval(ctx context.Context, script string, keys []string, args ...interface{}) *redis.Cmd {
	if m.evalFunc != nil {
		return m.evalFunc(ctx, script, keys, args...)
	}
	return redis.NewCmdResult(int64(1), nil)
}

// 实现 redis.UniversalClient 接口所需的其他方法
func (m *mockRedisClient) Do(ctx context.Context, args ...interface{}) *redis.Cmd {
	return redis.NewCmdResult(nil, nil)
}
func (m *mockRedisClient) Process(ctx context.Context, cmd redis.Cmder) error { return nil }
func (m *mockRedisClient) AddHook(hook redis.Hook)                            {}
func (m *mockRedisClient) WrapProcess(fn func(oldProcess func(cmd redis.Cmder) error) func(cmd redis.Cmder) error) {
}
func (m *mockRedisClient) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	return nil
}
func (m *mockRedisClient) PSubscribe(ctx context.Context, channels ...string) *redis.PubSub {
	return nil
}
func (m *mockRedisClient) Close() error { return nil }

func TestRateLimitFixedMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("DefaultConfig", func(t *testing.T) {
		callCount := 0
		mockRedis := &mockRedisClient{
			evalFunc: func(ctx context.Context, script string, keys []string, args ...interface{}) *redis.Cmd {
				callCount++
				assert.Contains(t, keys[0], "ratelimit:fixed:")
				return redis.NewCmdResult(int64(1), nil)
			},
		}

		router := gin.New()
		router.Use(RateLimitFixed(mockRedis, RateLimitFixedConfig{}))

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
				assert.Equal(t, "ratelimit:fixed:192.168.1.1", keys[0])
				return redis.NewCmdResult(int64(1), nil)
			},
		}

		router := gin.New()
		router.Use(RateLimitFixed(mockRedis, RateLimitFixedConfig{
			Rate:  10,
			Burst: 5,
			KeyFunc: func(c *gin.Context) string {
				return "192.168.1.1"
			},
			Expiration: time.Hour,
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
		router.Use(RateLimitFixed(mockRedis, RateLimitFixedConfig{
			Rate:  10,
			Burst: 3,
			KeyFunc: func(c *gin.Context) string {
				return "10.0.0.1"
			},
			Expiration: time.Minute,
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
		router.Use(RateLimitFixed(mockRedis, RateLimitFixedConfig{
			Rate:  100,
			Burst: 100,
			KeyFunc: func(c *gin.Context) string {
				return "error-test"
			},
			Expiration: time.Hour,
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
		router.Use(RateLimitFixed(mockRedis, RateLimitFixedConfig{
			Rate:  50,
			Burst: 50,
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
		router.Use(RateLimitFixed(mockRedis, RateLimitFixedConfig{}))

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

	t.Run("ZeroExpirationDefaults", func(t *testing.T) {
		var capturedArgs []interface{}
		mockRedis := &mockRedisClient{
			evalFunc: func(ctx context.Context, script string, keys []string, args ...interface{}) *redis.Cmd {
				capturedArgs = args
				return redis.NewCmdResult(int64(1), nil)
			},
		}

		router := gin.New()
		router.Use(RateLimitFixed(mockRedis, RateLimitFixedConfig{
			Rate:  100,
			Burst: 100,
		}))

		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		// Expiration 为 0 时应该使用默认值 time.Hour (3600 秒)
		assert.Equal(t, int(3600), capturedArgs[1])
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
		router.Use(RateLimitFixed(mockRedis, RateLimitFixedConfig{
			Rate:       10,
			Burst:      10,
			Expiration: 5 * time.Minute,
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
		assert.Equal(t, "ratelimit:fixed:user:12345", capturedKey)
	})
}

func TestRunFixedLimitScript(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("ScriptReturnsAllowed", func(t *testing.T) {
		mockRedis := &mockRedisClient{
			evalFunc: func(ctx context.Context, script string, keys []string, args ...interface{}) *redis.Cmd {
				return redis.NewCmdResult(int64(1), nil)
			},
		}

		ctx := &gin.Context{}
		allowed, err := runFixedLimitScript(ctx, mockRedis, "test-key", 10, time.Minute)

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
		allowed, err := runFixedLimitScript(ctx, mockRedis, "test-key", 5, time.Minute)

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
		allowed, err := runFixedLimitScript(ctx, mockRedis, "test-key", 10, time.Minute)

		assert.Error(t, err)
		assert.False(t, allowed)
	})
}
