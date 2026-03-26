package log

import "github.com/your-org/go-lib/middleware"

// 日志预定义错误码
var (
	LogErrInitFailed = middleware.NewCode(500201, "Failed to initialize logger")
)
