# go-lib

Go 项目集成库，封装常用框架，减少重复代码。

## 特性

- 🚀 **Gin 中间件**：JWT、限流、CORS、Recovery、RequestID 等
- 📦 **统一响应**：标准化的错误处理和响应格式
- 🔧 **Viper 配置**：多环境支持、热重载、结构体验证
- 📝 **Zap 日志**：JSON 格式、日志轮转、结构化输出
- 🗄️ **CRUD 泛型**：Entity/QueryDTO/PageResult 基础类型

## 快速开始

```bash
go get github.com/your-org/go-lib
```

### 最简示例

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/your-org/go-lib/middleware"
    "github.com/your-org/go-lib/log"
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

    // 路由
    r.GET("/hello", func(c *gin.Context) {
        middleware.RespondSuccessWithData(c, gin.H{"message": "Hello!"})
    })

    r.Run(":8080")
}
```

## 目录结构

```
go-lib/
├── middleware/    # Gin 中间件
├── config/        # Viper 配置
├── log/           # Zap 日志
├── gincrud/       # CRUD 基础类型
└── examples/      # 示例代码
```

## Repository 使用示例

### 基础使用

```go
type User struct {
    gincrud.BaseEntity
    Username string `json:"username"`
    Email    string `json:"email"`
}

func main() {
    db := initDB()
    repo := gincrud.NewRepository[User](db)

    // 创建
    user := &User{Username: "alice", Email: "alice@example.com"}
    repo.Create(ctx, user)

    // 查询
    found, _ := repo.GetByID(ctx, user.ID)

    // 条件查询
    cond := gincrud.NewQuery().
        WhereEq("status", 1).
        WhereLike("username", "%admin%").
        OrderBy("created_at", true).
        Page(1).
        PageSize(10)

    users, total, _ := repo.FindPage(ctx, cond)
}
```

### 自定义 Repository

```go
type UserRepository struct {
    *gincrud.Repository[User]
}

func NewUserRepository(db *gorm.DB) *UserRepository {
    return &UserRepository{
        Repository: gincrud.NewRepository[User](db),
    }
}

func (r *UserRepository) GetByUsername(username string) (*User, error) {
    // 自定义查询逻辑
}
```

## License

MIT
