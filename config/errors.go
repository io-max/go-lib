package config

import "github.com/io-max/go-lib/middleware"

// 配置预定义错误码
var (
	CfgErrLoadFailed      = middleware.NewCode(500201, "Failed to load configuration")
	CfgErrUnmarshalFailed = middleware.NewCode(500202, "Failed to unmarshal configuration")
	CfgErrValidateFailed  = middleware.NewCode(500203, "Configuration validation failed")
	CfgErrFileNotFound    = middleware.NewCode(400201, "Configuration file not found")
)
