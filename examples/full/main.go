package main

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/your-org/go-lib/config"
	"github.com/your-org/go-lib/log"
	"github.com/your-org/go-lib/middleware"
)

// AppConfig 应用配置
type AppConfig struct {
	Server ServerConfig `mapstructure:"server" validate:"required"`
	Log    LogConfig    `mapstructure:"log" validate:"required"`
	JWT    JWTConfig    `mapstructure:"jwt" validate:"required"`
}

type ServerConfig struct {
	Port int `mapstructure:"port" validate:"required,min=1,max=65535"`
}

type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

type JWTConfig struct {
	Secret string `mapstructure:"secret"`
}

func main() {
	// 1. 加载配置
	cfg, err := config.Load[AppConfig](
		config.WithName("config"),
		config.WithType("yaml"),
		config.WithPaths(".", "./examples/full"),
		config.WithEnvPrefix("APP"),
		config.WithValidate(true),
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	// 2. 初始化日志
	logger := log.MustLoad(log.Config{
		Level:      log.ParseLevel(cfg.Log.Level),
		Format:     cfg.Log.Format,
		OutputFile: "", // stdout
		Caller:     true,
	})
	log.SetGlobal(logger)

	logger.Info("Server starting", "port", cfg.Server.Port)

	// 3. 初始化 Gin
	r := gin.New()

	// 4. 注册中间件
	r.Use(middleware.ChainMiddleware(
		middleware.RecoveryMiddleware(middleware.RecoveryMiddlewareConfig{
			Logger: logger,
		}),
		middleware.RequestIDMiddleware(middleware.RequestIDMiddlewareConfig{}),
		middleware.CorsMiddleware(middleware.CorsMiddlewareConfig{}),
	))

	// 5. 公开路由
	r.GET("/hello", func(c *gin.Context) {
		middleware.RespondSuccessWithData(c, gin.H{"message": "Hello, World!"})
	})

	r.GET("/health", func(c *gin.Context) {
		middleware.RespondSuccess(c)
	})

	// 6. JWT 认证组
	if cfg.JWT.Secret != "" {
		auth := r.Group("/api")
		auth.Use(middleware.JwtAuth(middleware.JwtConfig{
			Secret: []byte(cfg.JWT.Secret),
		}, nil))
		{
			auth.GET("/profile", func(c *gin.Context) {
				middleware.RespondSuccessWithData(c, gin.H{
					"user": "authenticated",
				})
			})
		}
	}

	// 7. 日志中间件
	r.Use(middleware.LoggerMiddleware(middleware.LoggerMiddlewareConfig{
		Logger:    logger,
		SkipPaths: []string{"/health"},
	}))

	// 8. 启动服务
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	logger.Info("Server listening", "addr", addr)
	if err := r.Run(addr); err != nil {
		logger.Error("Server failed", "error", err)
		os.Exit(1)
	}
}
