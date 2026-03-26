package middleware

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"time"
)

// RateLimitSlidingConfig 滑动窗口限流配置
type RateLimitSlidingConfig struct {
	Rate    int
	Window  time.Duration
	KeyFunc RateLimitKeyFunc
}

// RateLimitSliding 滑动窗口限流中间件
func RateLimitSliding(client RedisClient, cfg RateLimitSlidingConfig) gin.HandlerFunc {
	if cfg.Rate <= 0 {
		cfg.Rate = 100
	}
	if cfg.Window == 0 {
		cfg.Window = time.Second
	}
	if cfg.KeyFunc == nil {
		cfg.KeyFunc = func(c *gin.Context) string {
			return c.ClientIP()
		}
	}

	return func(c *gin.Context) {
		key := fmt.Sprintf("ratelimit:sliding:%s", cfg.KeyFunc(c))

		allowed, err := runSlidingLimitScript(c, client, key, cfg.Rate, cfg.Window)
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

const slidingLimitScript = `
local key = KEYS[1]
local rate = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local now = tonumber(ARGV[3])

redis.call('ZREMRANGEBYSCORE', key, 0, now - window)

local count = redis.call('ZCARD', key)

if count >= rate then
    return 0
end

redis.call('ZADD', key, now, now .. '-' .. math.random(100000))
redis.call('EXPIRE', key, math.ceil(window))

return 1
`

func runSlidingLimitScript(c *gin.Context, client RedisClient, key string, rate int, window time.Duration) (bool, error) {
	now := time.Now().UnixMilli()

	result, err := client.Eval(
		context.Background(),
		slidingLimitScript,
		[]string{key},
		rate, int(window.Milliseconds()), now,
	).Int()

	if err != nil {
		return false, err
	}

	return result == 1, nil
}
