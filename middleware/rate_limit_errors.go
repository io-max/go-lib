package middleware

// 限流预定义错误码
var (
	RateLimitErrExceeded = NewCode(429001, "Rate limit exceeded")
)
