package main

import (
	"github.com/gin-gonic/gin"
	"github.com/your-org/go-lib/middleware"
	"github.com/your-org/go-lib/log"
)

func main() {
	logger := log.MustLoad(log.Config{
		Level:      log.InfoLevel,
		Format:     "json",
		OutputFile: "/var/log/app/app.log",
		MaxSize:    100,
		MaxBackups: 7,
		Compress:   true,
	})
	log.SetGlobal(logger)

	r := gin.New()
	r.Use(middleware.RecoveryMiddleware(middleware.RecoveryMiddlewareConfig{}))
	r.Use(middleware.RequestIDMiddleware(middleware.RequestIDMiddlewareConfig{}))
	r.Use(middleware.LoggerMiddleware(middleware.LoggerMiddlewareConfig{
		Logger: logger,
	}))

	r.GET("/hello", func(c *gin.Context) {
		middleware.RespondSuccessWithData(c, gin.H{"message": "Hello, World!"})
	})

	r.Run(":8080")
}
