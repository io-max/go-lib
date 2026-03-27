package config

import "github.com/io-max/go-lib/middleware"

// 配置预定义错误码
var (
	CfgErrLoadFailed      = middleware.NewCode(500101, "Failed to load configuration")
	CfgErrUnmarshalFailed = middleware.NewCode(500102, "Failed to unmarshal configuration")
	CfgErrValidateFailed  = middleware.NewCode(500103, "Configuration validation failed")
	CfgErrFileNotFound    = middleware.NewCode(400101, "Configuration file not found")
)
