package middleware

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"time"
)

// RateLimitFixedConfig 固定窗口限流配置
type RateLimitFixedConfig struct {
	Rate       int
	Burst      int
	KeyFunc    RateLimitKeyFunc
	Expiration time.Duration
}

// RateLimitKeyFunc 限流 Key 生成函数
type RateLimitKeyFunc func(c *gin.Context) string

// RedisClient 定义 Redis 客户端接口
type RedisClient interface {
	Eval(ctx context.Context, script string, keys []string, args ...interface{}) *redis.Cmd
}

// RateLimitFixed 固定窗口限流中间件
func RateLimitFixed(client RedisClient, cfg RateLimitFixedConfig) gin.HandlerFunc {
	if cfg.Rate <= 0 {
		cfg.Rate = 100
	}
	if cfg.Burst <= 0 {
		cfg.Burst = cfg.Rate
	}
	if cfg.Expiration == 0 {
		cfg.Expiration = time.Hour
	}
	if cfg.KeyFunc == nil {
		cfg.KeyFunc = func(c *gin.Context) string {
			return c.ClientIP()
		}
	}

	return func(c *gin.Context) {
		key := fmt.Sprintf("ratelimit:fixed:%s", cfg.KeyFunc(c))

		allowed, err := runFixedLimitScript(c, client, key, cfg.Burst, cfg.Expiration)
		if err != nil {
			c.Next()
			return
		}

		if !allowed {
			RespondError(c, RateLimitErrExceeded)
			c.Abort()
			return
		}

		c.Next()
	}
}

const fixedLimitScript = `
local key = KEYS[1]
local burst = tonumber(ARGV[1])
local expiration = tonumber(ARGV[2])

local current = redis.call('INCR', key)
if current == 1 then
    redis.call('EXPIRE', key, expiration)
end

if current > burst then
    return 0
end

return 1
`

func runFixedLimitScript(c *gin.Context, client RedisClient, key string, burst int, expiration time.Duration) (bool, error) {
	result, err := client.Eval(
		context.Background(),
		fixedLimitScript,
		[]string{key},
		burst, int(expiration.Seconds()),
	).Int()

	if err != nil {
		return false, err
	}

	return result == 1, nil
}
