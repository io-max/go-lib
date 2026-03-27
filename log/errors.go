package log

import "github.com/io-max/go-lib/middleware"

// 日志预定义错误码
var (
	LogErrInitFailed = middleware.NewCode(500301, "Failed to initialize logger")
)
