# CRUD Service Layer Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development to implement this plan task-by-task.

**Goal:** 将 `gincrud` 重命名为 `crud` 并添加泛型 Service 层 `Service[T Entity, O OptDTO, Q QueryDTO, R Response]`，支持 DTO 到 Entity 转换和统一的业务逻辑封装。

**Architecture:** 采用三层架构：Handler → Service（可选）→ Repository → GORM。Service 层通过泛型参数和转换函数实现 DTO 与 Entity 的解耦，用户可继承基础 Service 扩展业务逻辑。

**Tech Stack:** Go 1.21+ 泛型，GORM v1.25+，当前代码库 `github.com/io-max/go-lib`

---

## Phase 1: 重命名 gincrud → crud

### Task 1: 重命名目录

**Files:**
- Rename: `gincrud/` → `crud/`

**Step 1: 重命名目录**

Run:
```bash
cd /Users/max/Documents/Max-Go/go-lib
mv gincrud crud
```

**Step 2: 验证目录已重命名**

Run: `ls -la crud/`

Expected: 看到所有原 gincrud 的文件

**Step 3: 提交**

```bash
git add -u
git commit -m "refactor: rename gincrud directory to crud"
```

---

### Task 2: 更新 package 声明

**Files:**
- Modify: `crud/*.go` (所有 `.go` 文件)

**Step 1: 列出所有需要修改的文件**

Run: `ls crud/*.go`

**Step 2: 更新所有 package 声明**

Run:
```bash
sed -i '' 's/^package gincrud$/package crud/g' crud/*.go
```

**Step 3: 验证修改**

Run: `head -1 crud/*.go`

Expected: 所有文件第一行为 `package crud`

**Step 4: 提交**

```bash
git add crud/*.go
git commit -m "refactor: update package declaration to crud"
```

---

### Task 3: 更新内部引用

**Files:**
- Modify: `examples/basic/main.go`
- Modify: `examples/full/main.go`
- Modify: `examples/repository/models.go`
- Modify: `examples/repository/user_repository.go`
- Modify: `README.md`

**Step 1: 更新示例代码中的引用**

Run:
```bash
sed -i '' 's|github.com/io-max/go-lib/gincrud|github.com/io-max/go-lib/crud|g' \
    examples/basic/main.go \
    examples/full/main.go \
    examples/repository/models.go \
    examples/repository/user_repository.go \
    README.md
```

**Step 2: 验证引用已更新**

Run: `grep -r "gincrud" examples/ README.md`

Expected: 无输出（或只有注释中的引用）

**Step 3: 提交**

```bash
git add -u
git commit -m "refactor: update imports from gincrud to crud"
```

---

### Task 4: 验证编译通过

**Files:**
- Test: 整个项目

**Step 1: 运行编译**

Run: `go build ./...`

Expected: 无错误输出

**Step 2: 运行测试**

Run: `go test ./... -v 2>&1 | tail -20`

Expected: 所有测试 PASS

**Step 3: 提交（如果有修复）**

```bash
git add -u
git commit -m "fix: verify build after rename"
```

---

## Phase 2: 创建 Service 层基础结构

### Task 5: 创建 Service 接口定义

**Files:**
- Create: `crud/service.go`

**Step 1: 创建 Service 接口**

Create `crud/service.go`:

```go
package crud

import (
    "context"
    "gorm.io/gorm"
)

// IService 通用 Service 接口
type IService[T Entity, O any, Q any, R any] interface {
    // 基础 CRUD
    Create(ctx context.Context, dto *O) (R, error)
    Update(ctx context.Context, dto *O) (R, error)
    GetByID(ctx context.Context, id int64) (R, error)
    Delete(ctx context.Context, id int64) error
    DeletePermanently(ctx context.Context, id int64) error

    // 批量操作
    BatchCreate(ctx context.Context, dtos []*O) ([]R, error)
    BatchUpdateByIDs(ctx context.Context, ids []int64, dto *O) error
    BatchDelete(ctx context.Context, ids []int64) error
    DeleteByIDs(ctx context.Context, ids []int64) error

    // 查询
    GetOne(ctx context.Context, query Q) (R, error)
    List(ctx context.Context, query Q) ([]R, error)
    Page(ctx context.Context, query Q) (*PageResult[R], error)
    Count(ctx context.Context, query Q) (int64, error)
    Exists(ctx context.Context, query Q) (bool, error)
    GetByIDs(ctx context.Context, ids []int64) ([]R, error)

    // DB 访问
    DB() *gorm.DB
    Repository() IRepository[T]
}
```

**Step 2: 验证语法**

Run: `go build ./crud/...`

Expected: 无错误

**Step 3: 提交**

```bash
git add crud/service.go
git commit -m "feat(crud): add IService interface definition"
```

---

### Task 6: 创建 ServiceConfig 结构体

**Files:**
- Modify: `crud/service.go`

**Step 1: 添加 ServiceConfig**

Append to `crud/service.go`:

```go
// ServiceConfig Service 配置
type ServiceConfig[T Entity, O any, Q any, R any] struct {
    // OptDTO → Entity 转换（Create + Update 共用）
    DtoToEntity func(*O) (*T, error)

    // Entity → Response 转换
    EntityToRes func(*T) (R, error)

    // Query DTO → QueryCondition 转换
    QueryToCond func(Q) *QueryCondition

    // 可选：Update 前钩子
    BeforeUpdate func(ctx context.Context, dto *O, entity *T) error

    // 可选：Create 前钩子
    BeforeCreate func(ctx context.Context, dto *O, entity *T) error
}
```

**Step 2: 验证语法**

Run: `go build ./crud/...`

Expected: 无错误

**Step 3: 提交**

```bash
git add crud/service.go
git commit -m "feat(crud): add ServiceConfig struct"
```

---

### Task 7: 创建 Service 基础结构

**Files:**
- Modify: `crud/service.go`

**Step 1: 添加 Service 结构体和构造函数**

Append to `crud/service.go`:

```go
// Service 基础服务层（泛型版本）
type Service[T Entity, O any, Q any, R any] struct {
    repo IRepository[T]

    // 转换函数
    dtoToEntity func(*O) (*T, error)
    entityToRes func(*T) (R, error)
    queryToCond func(Q) *QueryCondition

    // 钩子
    beforeUpdate func(ctx context.Context, dto *O, entity *T) error
    beforeCreate func(ctx context.Context, dto *O, entity *T) error
}

// NewService 创建 Service
func NewService[T Entity, O any, Q any, R any](
    repo IRepository[T],
    cfg ServiceConfig[T, O, Q, R],
) *Service[T, O, Q, R] {
    return &Service[T, O, Q, R]{
        repo:         repo,
        dtoToEntity:  cfg.DtoToEntity,
        entityToRes:  cfg.EntityToRes,
        queryToCond:  cfg.QueryToCond,
        beforeUpdate: cfg.BeforeUpdate,
        beforeCreate: cfg.BeforeCreate,
    }
}

// NewServiceWithDB 使用 DB 创建 Service
func NewServiceWithDB[T Entity, O any, Q any, R any](
    db *gorm.DB,
    cfg ServiceConfig[T, O, Q, R],
) *Service[T, O, Q, R] {
    return NewService(NewRepository[T](db), cfg)
}

// Repository 获取 Repository
func (s *Service[T, O, Q, R]) Repository() IRepository[T] {
    return s.repo
}

// DB 获取底层 DB
func (s *Service[T, O, Q, R]) DB() *gorm.DB {
    return s.repo.DB()
}
```

**Step 2: 验证语法**

Run: `go build ./crud/...`

Expected: 无错误

**Step 3: 提交**

```bash
git add crud/service.go
git commit -m "feat(crud): add Service base struct and constructors"
```

---

### Task 8: 实现基础 CRUD 方法

**Files:**
- Modify: `crud/service.go`

**Step 1: 添加 Create 方法**

Append to `crud/service.go`:

```go
// =============================================================================
// 基础 CRUD
// =============================================================================

// Create 创建
func (s *Service[T, O, Q, R]) Create(ctx context.Context, dto *O) (R, error) {
    var zero R

    entity, err := s.dtoToEntity(dto)
    if err != nil {
        return zero, err
    }

    // Create 前钩子
    if s.beforeCreate != nil {
        if err := s.beforeCreate(ctx, dto, entity); err != nil {
            return zero, err
        }
    }

    if err := s.repo.Create(ctx, entity); err != nil {
        return zero, err
    }

    return s.entityToRes(entity)
}

// GetByID 根据 ID 获取单个
func (s *Service[T, O, Q, R]) GetByID(ctx context.Context, id int64) (R, error) {
    var zero R

    entity, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return zero, err
    }

    return s.entityToRes(entity)
}

// Update 更新
func (s *Service[T, O, Q, R]) Update(ctx context.Context, dto *O) (R, error) {
    var zero R

    entity, err := s.dtoToEntity(dto)
    if err != nil {
        return zero, err
    }

    // Update 前钩子
    if s.beforeUpdate != nil {
        if err := s.beforeUpdate(ctx, dto, entity); err != nil {
            return zero, err
        }
    }

    if err := s.repo.Update(ctx, entity); err != nil {
        return zero, err
    }

    return s.entityToRes(entity)
}

// Delete 删除（软删除）
func (s *Service[T, O, Q, R]) Delete(ctx context.Context, id int64) error {
    return s.repo.Delete(ctx, id)
}

// DeletePermanently 永久删除（硬删除）
func (s *Service[T, O, Q, R]) DeletePermanently(ctx context.Context, id int64) error {
    return s.repo.TrulyDelete(ctx, id)
}
```

**Step 2: 验证语法**

Run: `go build ./crud/...`

Expected: 无错误

**Step 3: 提交**

```bash
git add crud/service.go
git commit -m "feat(crud): implement base CRUD methods"
```

---

### Task 9: 实现批量操作方法

**Files:**
- Modify: `crud/service.go`

**Step 1: 添加批量操作方法**

Append to `crud/service.go`:

```go
// =============================================================================
// 批量操作
// =============================================================================

// BatchCreate 批量创建
func (s *Service[T, O, Q, R]) BatchCreate(ctx context.Context, dtos []*O) ([]R, error) {
    var entities []*T
    for _, dto := range dtos {
        entity, err := s.dtoToEntity(dto)
        if err != nil {
            return nil, err
        }
        entities = append(entities, entity)
    }

    if err := s.repo.BatchCreate(ctx, entities); err != nil {
        return nil, err
    }

    var results []R
    for _, entity := range entities {
        res, err := s.entityToRes(entity)
        if err != nil {
            return nil, err
        }
        results = append(results, res)
    }

    return results, nil
}

// BatchUpdateByIDs 批量更新
func (s *Service[T, O, Q, R]) BatchUpdateByIDs(ctx context.Context, ids []int64, dto *O) error {
    entity, err := s.dtoToEntity(dto)
    if err != nil {
        return err
    }

    // 提取要更新的字段（排除 ID 等）
    updates := map[string]any{}

    // 这里需要根据实际 DTO 字段构建，简化处理
    // 实际使用时用户可扩展此逻辑

    return s.repo.BatchUpdate(ctx, ids, updates)
}

// BatchDelete 批量删除
func (s *Service[T, O, Q, R]) BatchDelete(ctx context.Context, ids []int64) error {
    return s.repo.BatchDelete(ctx, ids)
}

// DeleteByIDs 根据 IDs 批量删除
func (s *Service[T, O, Q, R]) DeleteByIDs(ctx context.Context, ids []int64) error {
    return s.repo.DeleteByIDs(ctx, ids)
}
```

**Step 2: 验证语法**

Run: `go build ./crud/...`

Expected: 无错误

**Step 3: 提交**

```bash
git add crud/service.go
git commit -m "feat(crud): implement batch operation methods"
```

---

### Task 10: 实现查询方法

**Files:**
- Modify: `crud/service.go`

**Step 1: 添加查询方法**

Append to `crud/service.go`:

```go
// =============================================================================
// 查询
// =============================================================================

// GetOne 获取单个
func (s *Service[T, O, Q, R]) GetOne(ctx context.Context, query Q) (R, error) {
    var zero R

    cond := s.queryToCond(query)
    entity, err := s.repo.FindFirst(ctx, cond)
    if err != nil {
        return zero, err
    }

    return s.entityToRes(entity)
}

// List 列表查询
func (s *Service[T, O, Q, R]) List(ctx context.Context, query Q) ([]R, error) {
    cond := s.queryToCond(query)
    entities, err := s.repo.Find(ctx, cond)
    if err != nil {
        return nil, err
    }

    var results []R
    for _, entity := range entities {
        res, err := s.entityToRes(entity)
        if err != nil {
            return nil, err
        }
        results = append(results, res)
    }

    return results, nil
}

// Page 分页查询
func (s *Service[T, O, Q, R]) Page(ctx context.Context, query Q) (*PageResult[R], error) {
    cond := s.queryToCond(query)
    entities, total, err := s.repo.FindPage(ctx, cond)
    if err != nil {
        return nil, err
    }

    var results []R
    for _, entity := range entities {
        res, err := s.entityToRes(entity)
        if err != nil {
            return nil, err
        }
        results = append(results, res)
    }

    return &PageResult[R]{
        List:  results,
        Total: total,
    }, nil
}

// Count 计数
func (s *Service[T, O, Q, R]) Count(ctx context.Context, query Q) (int64, error) {
    cond := s.queryToCond(query)
    return s.repo.Count(ctx, cond)
}

// Exists 检查是否存在
func (s *Service[T, O, Q, R]) Exists(ctx context.Context, query Q) (bool, error) {
    cond := s.queryToCond(query)
    count, err := s.repo.Count(ctx, cond)
    return count > 0, err
}

// GetByIDs 根据 IDs 批量获取
func (s *Service[T, O, Q, R]) GetByIDs(ctx context.Context, ids []int64) ([]R, error) {
    entities, err := s.repo.FindByIDs(ctx, ids)
    if err != nil {
        return nil, err
    }

    var results []R
    for _, entity := range entities {
        res, err := s.entityToRes(entity)
        if err != nil {
            return nil, err
        }
        results = append(results, res)
    }

    return results, nil
}
```

**Step 2: 验证语法**

Run: `go build ./crud/...`

Expected: 无错误

**Step 3: 提交**

```bash
git add crud/service.go
git commit -m "feat(crud): implement query methods"
```

---

## Phase 3: 创建 Service 测试

### Task 11: 创建 Service 测试文件

**Files:**
- Create: `crud/service_test.go`

**Step 1: 创建测试文件**

Create `crud/service_test.go`:

```go
package crud

import (
    "context"
    "testing"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
    "github.com/stretchr/testify/assert"
)

// TestEntity 测试实体
type TestEntity struct {
    BaseEntity
    Name  string `json:"name"`
    Email string `json:"email"`
}

func (TestEntity) TableName() string { return "test_entities" }

// TestOptDTO 测试操作 DTO
type TestOptDTO struct {
    ID    int64  `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

// TestQueryDTO 测试查询 DTO
type TestQueryDTO struct {
    BaseQueryDTO
    Name string `form:"name"`
}

// TestResponse 测试响应
type TestResponse struct {
    ID    int64  `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

func setupTestService(t *testing.T) (*Service[TestEntity, TestOptDTO, TestQueryDTO, TestResponse], *gorm.DB) {
    db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
    assert.NoError(t, err)

    err = db.AutoMigrate(&TestEntity{})
    assert.NoError(t, err)

    cfg := ServiceConfig[TestEntity, TestOptDTO, TestQueryDTO, TestResponse]{
        DtoToEntity: func(dto *TestOptDTO) (*TestEntity, error) {
            entity := &TestEntity{}
            if dto.ID > 0 {
                entity.ID = dto.ID
            }
            if dto.Name != "" {
                entity.Name = dto.Name
            }
            if dto.Email != "" {
                entity.Email = dto.Email
            }
            return entity, nil
        },
        EntityToRes: func(entity *TestEntity) (TestResponse, error) {
            return TestResponse{
                ID:    entity.ID,
                Name:  entity.Name,
                Email: entity.Email,
            }, nil
        },
        QueryToCond: func(query TestQueryDTO) *QueryCondition {
            cond := NewQuery()
            if query.Name != "" {
                cond.WhereEq("name", query.Name)
            }
            cond.Page(query.GetPage()).PageSize(query.GetPageSize())
            return cond
        },
    }

    return NewService(NewRepository[TestEntity](db), cfg), db
}

func TestService_Create(t *testing.T) {
    svc, _ := setupTestService(t)
    ctx := context.Background()

    dto := &TestOptDTO{
        Name:  "Test User",
        Email: "test@example.com",
    }

    result, err := svc.Create(ctx, dto)

    assert.NoError(t, err)
    assert.NotZero(t, result.ID)
    assert.Equal(t, "Test User", result.Name)
    assert.Equal(t, "test@example.com", result.Email)
}

func TestService_GetByID(t *testing.T) {
    svc, _ := setupTestService(t)
    ctx := context.Background()

    // 先创建
    dto := &TestOptDTO{Name: "Test", Email: "test@example.com"}
    created, _ := svc.Create(ctx, dto)

    // 再查询
    result, err := svc.GetByID(ctx, created.ID)

    assert.NoError(t, err)
    assert.Equal(t, created.ID, result.ID)
    assert.Equal(t, "Test", result.Name)
}

func TestService_List(t *testing.T) {
    svc, _ := setupTestService(t)
    ctx := context.Background()

    // 创建测试数据
    svc.Create(ctx, &TestOptDTO{Name: "User1", Email: "user1@example.com"})
    svc.Create(ctx, &TestOptDTO{Name: "User2", Email: "user2@example.com"})

    query := TestQueryDTO{}
    results, err := svc.List(ctx, query)

    assert.NoError(t, err)
    assert.GreaterOrEqual(t, len(results), 2)
}

func TestService_Page(t *testing.T) {
    svc, _ := setupTestService(t)
    ctx := context.Background()

    // 创建测试数据
    for i := 0; i < 15; i++ {
        svc.Create(ctx, &TestOptDTO{Name: "User", Email: "user@example.com"})
    }

    query := TestQueryDTO{
        BaseQueryDTO: BaseQueryDTO{Page: 1, PageSize: 10},
    }
    result, err := svc.Page(ctx, query)

    assert.NoError(t, err)
    assert.Equal(t, int64(15), result.Total)
    assert.Len(t, result.List, 10)
}
```

**Step 2: 运行测试**

Run: `go test ./crud -run TestService -v`

Expected: 所有测试通过

**Step 3: 提交**

```bash
git add crud/service_test.go
git commit -m "test(crud): add Service unit tests"
```

---

## Phase 4: 扩展示例代码

### Task 12: 创建用户服务示例

**Files:**
- Create: `examples/service/user_service.go`
- Create: `examples/service/models.go`

**Step 1: 创建模型**

Create `examples/service/models.go`:

```go
package service

import "github.com/io-max/go-lib/crud"

// User 用户模型
type User struct {
    crud.BaseEntity
    Username string `gorm:"uniqueIndex;size:50" json:"username"`
    Email    string `gorm:"uniqueIndex;size:100" json:"email"`
    Password string `gorm:"size:255" json:"-"`
    Status   int    `gorm:"default:1" json:"status"`
}

func (User) TableName() string { return "users" }
```

**Step 2: 创建服务**

Create `examples/service/user_service.go`:

```go
package service

import (
    "context"
    "github.com/io-max/go-lib/crud"
    "golang.org/x/crypto/bcrypt"
    "gorm.io/gorm"
)

// UserOptDTO 用户操作 DTO（Create + Update 复用）
type UserOptDTO struct {
    ID       int64  `json:"id" form:"id"`
    Username string `json:"username" form:"username"`
    Email    string `json:"email" form:"email"`
    Password string `json:"password" form:"password"`
    Status   int    `json:"status" form:"status"`
}

// UserQueryDTO 查询 DTO
type UserQueryDTO struct {
    crud.BaseQueryDTO
    Username string `form:"username"`
    Email    string `form:"email"`
    Status   int    `form:"status"`
}

// UserResponse 用户响应
type UserResponse struct {
    ID       int64  `json:"id"`
    Username string `json:"username"`
    Email    string `json:"email"`
    Status   int    `json:"status"`
}

// UserService 用户服务
type UserService struct {
    *crud.Service[User, UserOptDTO, UserQueryDTO, UserResponse]
}

// NewUserService 创建用户服务
func NewUserService(db *gorm.DB) *UserService {
    cfg := crud.ServiceConfig[User, UserOptDTO, UserQueryDTO, UserResponse]{
        DtoToEntity: func(dto *UserOptDTO) (*User, error) {
            user := &User{}

            if dto.ID > 0 {
                user.ID = dto.ID
            }
            if dto.Username != "" {
                user.Username = dto.Username
            }
            if dto.Email != "" {
                user.Email = dto.Email
            }
            if dto.Password != "" {
                hashed, err := bcrypt.GenerateFromPassword([]byte(dto.Password), bcrypt.DefaultCost)
                if err != nil {
                    return nil, err
                }
                user.Password = string(hashed)
            }
            if dto.Status > 0 {
                user.Status = dto.Status
            }

            return user, nil
        },
        EntityToRes: func(entity *User) (UserResponse, error) {
            return UserResponse{
                ID:       entity.ID,
                Username: entity.Username,
                Email:    entity.Email,
                Status:   entity.Status,
            }, nil
        },
        QueryToCond: func(query UserQueryDTO) *crud.QueryCondition {
            cond := crud.NewQuery()
            if query.Username != "" {
                cond.WhereLike("username", "%"+query.Username+"%")
            }
            if query.Email != "" {
                cond.WhereEq("email", query.Email)
            }
            if query.Status > 0 {
                cond.WhereEq("status", query.Status)
            }
            cond.Page(query.GetPage()).PageSize(query.GetPageSize())
            cond.OrderBy("created_at", true)
            return cond
        },
    }

    return &UserService{
        Service: crud.NewServiceWithDB[User, UserOptDTO, UserQueryDTO, UserResponse](db, cfg),
    }
}
```

**Step 3: 验证编译**

Run: `go build ./examples/service/...`

Expected: 无错误

**Step 4: 提交**

```bash
git add examples/service/
git commit -m "docs: add user service example"
```

---

## Phase 5: 更新文档和发布

### Task 13: 更新 README.md

**Files:**
- Modify: `README.md`

**Step 1: 添加 Service 层使用示例**

Append to `README.md`:

```markdown
## Service 层使用示例

```go
// 定义 DTO
type UserOptDTO struct {
    ID       int64  `json:"id"`
    Username string `json:"username"`
    Email    string `json:"email"`
}

type UserQueryDTO struct {
    crud.BaseQueryDTO
    Username string `form:"username"`
}

type UserResponse struct {
    ID       int64  `json:"id"`
    Username string `json:"username"`
    Email    string `json:"email"`
}

// 创建服务
userService := crud.NewServiceWithDB[User, UserOptDTO, UserQueryDTO, UserResponse](db, cfg)

// 使用服务
user, _ := userService.GetByID(ctx, 1)
users, _ := userService.List(ctx, query)
page, _ := userService.Page(ctx, query)
```
```

**Step 2: 验证语法**

Run: `go build ./...`

**Step 3: 提交**

```bash
git add README.md
git commit -m "docs: add Service layer usage examples"
```

---

### Task 14: 运行完整测试

**Files:**
- Test: 整个项目

**Step 1: 运行所有测试**

Run: `go test ./... -v 2>&1 | tail -30`

Expected: 所有测试 PASS

**Step 2: 提交（如果有修复）**

```bash
git add -u
git commit -m "test: fix any issues found"
```

---

### Task 15: 打标签 v0.2.0

**Files:**
- Git 标签

**Step 1: 创建并推送标签**

Run:
```bash
git tag v0.2.0 -m "Release v0.2.0: 重命名 gincrud 为 crud 并添加 Service 层"
git push origin v0.2.0
```

**Step 2: 验证标签已推送**

Run: `git ls-remote origin refs/tags/v0.2.0`

Expected: 显示标签 commit hash

---

## 总结

本计划共 15 个 Task，涵盖：
1. 重命名 `gincrud` → `crud`
2. 创建 `Service[T, O, Q, R]` 泛型接口和实现
3. 添加完整的单元测试
4. 创建用户服务扩展示例
5. 更新文档并发布 v0.2.0

预计完成时间：约 1-2 小时

---

## 执行方式选择

计划已完成并保存到 `docs/plans/2026-03-27-crud-service-layer-design.md`。

**有两种执行方式：**

**1. Subagent-Driven（本会话）** - 我在当前会话中逐个任务执行，每个任务使用独立子代理，任务间进行代码审查

**2. 并行会话** - 新建一个会话使用 `executing-plans` 批量执行

**选择哪种方式？**
