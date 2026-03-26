package main

import (
	"github.com/gin-gonic/gin"
	"github.com/your-org/go-lib/log"
	"github.com/your-org/go-lib/middleware"
)

func main() {
	// 初始化日志
	logger := log.MustLoad(log.Config{
		Level:  log.InfoLevel,
		Format: "json",
	})
	log.SetGlobal(logger)

	r := gin.New()

	// 基础中间件
	r.Use(middleware.RecoveryMiddleware(middleware.RecoveryMiddlewareConfig{
		Logger: logger,
	}))
	r.Use(middleware.RequestIDMiddleware(middleware.RequestIDMiddlewareConfig{}))
	r.Use(middleware.CorsMiddleware(middleware.CorsMiddlewareConfig{}))

	// 路由
	r.GET("/hello", func(c *gin.Context) {
		middleware.RespondSuccessWithData(c, gin.H{"message": "Hello, World!"})
	})

	r.GET("/health", func(c *gin.Context) {
		middleware.RespondSuccess(c)
	})

	// 日志中间件
	r.Use(middleware.LoggerMiddleware(middleware.LoggerMiddlewareConfig{
		Logger:    logger,
		SkipPaths: []string{"/health"},
	}))

	r.Run(":8080")
}
