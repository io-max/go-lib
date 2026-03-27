# gincrud Repository 层设计

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development to implement this plan task-by-task.

**Goal:** 实现 gincrud 模块的 Repository 层，提供泛型数据访问抽象，支持 CRUD、条件查询、分页、事务等功能。

**Architecture:** 采用接口 + 实现分离的设计。`IRepository[T Entity]` 接口定义数据访问契约，`Repository[T Entity]` 结构体提供通用实现。用户可通过嵌入基础 Repository 轻松扩展自定义数据访问逻辑。

**Tech Stack:** Go 1.21+ 泛型，GORM v1.25+，Context 用于超时和取消控制

---

## 设计决策

### 泛型约束
- 使用 `T Entity` 而非 `T any`，确保实体有 `ID`、`Deleted`、`CreatedAt`、`UpdatedAt` 字段

### 返回值类型
- 单个实体返回 `*T`（指针），避免大结构体拷贝
- 列表返回 `[]*T`（指针切片），保持一致性

### 软删除处理
- 默认自动过滤软删除（`WHERE deleted = 0`）
- 提供 `TrulyDelete` 方法执行硬删除
- 用户可通过重写方法自定义软删除逻辑

### 查询条件
- `QueryCondition` 独立于 `QueryDTO`（分页参数）
- 支持链式调用：`NewQuery().WhereEq("status", 1).OrderBy("id", true)`
- 支持预加载：`Preload("Posts")`

### 事务支持
- `WithTx(tx *gorm.DB) IRepository[T]` 返回带事务的 Repository
- 用户可组合使用：`txRepo := repo.WithTx(tx)`

---

## Phase 1: 基础结构定义

### Task 1: 定义 QueryCondition 查询条件构建器

**Files:**
- Create: `gincrud/query_condition.go`
- Test: `gincrud/query_condition_test.go`

**Step 1: 创建查询条件结构**

Create `gincrud/query_condition.go`:

```go
package gincrud

// EqCond 等于条件
type EqCond struct {
	Field string
	Value any
}

// NeCond 不等于条件
type NeCond struct {
	Field string
	Value any
}

// GtCond 大于条件
type GtCond struct {
	Field string
	Value any
}

// LtCond 小于条件
type LtCond struct {
	Field string
	Value any
}

// GeCond 大于等于条件
type GeCond struct {
	Field string
	Value any
}

// LeCond 小于等于条件
type LeCond struct {
	Field string
	Value any
}

// BetweenCond 区间条件
type BetweenCond struct {
	Field string
	Min   any
	Max   any
}

// InCond IN 条件
type InCond struct {
	Field  string
	Values []any
}

// LikeCond LIKE 条件
type LikeCond struct {
	Field   string
	Pattern string
}

// OrderByField 排序字段
type OrderByField struct {
	Field string
	Desc  bool
}

// QueryCondition 查询条件构建器
type QueryCondition struct {
	whereEq      []EqCond
	whereNe      []NeCond
	whereGt      []GtCond
	whereLt      []LtCond
	whereGe      []GeCond
	whereLe      []LeCond
	whereBetween []BetweenCond
	whereIn      []InCond
	whereLike    []LikeCond
	whereNull    []string
	whereNotNull []string
	preloads     []string
	orderBy      []OrderByField
	limit        int
	offset       int
}

// NewQuery 创建查询条件
func NewQuery() *QueryCondition {
	return &QueryCondition{
		whereEq:      make([]EqCond, 0),
		whereNe:      make([]NeCond, 0),
		whereGt:      make([]GtCond, 0),
		whereLt:      make([]LtCond, 0),
		whereGe:      make([]GeCond, 0),
		whereLe:      make([]LeCond, 0),
		whereBetween: make([]BetweenCond, 0),
		whereIn:      make([]InCond, 0),
		whereLike:    make([]LikeCond, 0),
		whereNull:    make([]string, 0),
		whereNotNull: make([]string, 0),
		preloads:     make([]string, 0),
		orderBy:      make([]OrderByField, 0),
	}
}

// WhereEq 添加等于条件
func (q *QueryCondition) WhereEq(field string, value any) *QueryCondition {
	q.whereEq = append(q.whereEq, EqCond{Field: field, Value: value})
	return q
}

// WhereNe 添加不等于条件
func (q *QueryCondition) WhereNe(field string, value any) *QueryCondition {
	q.whereNe = append(q.whereNe, NeCond{Field: field, Value: value})
	return q
}

// WhereGt 添加大于条件
func (q *QueryCondition) WhereGt(field string, value any) *QueryCondition {
	q.whereGt = append(q.whereGt, GtCond{Field: field, Value: value})
	return q
}

// WhereLt 添加小于条件
func (q *QueryCondition) WhereLt(field string, value any) *QueryCondition {
	q.whereLt = append(q.whereLt, LtCond{Field: field, Value: value})
	return q
}

// WhereGe 添加大于等于条件
func (q *QueryCondition) WhereGe(field string, value any) *QueryCondition {
	q.whereGe = append(q.whereGe, GeCond{Field: field, Value: value})
	return q
}

// WhereLe 添加小于等于条件
func (q *QueryCondition) WhereLe(field string, value any) *QueryCondition {
	q.whereLe = append(q.whereLe, LeCond{Field: field, Value: value})
	return q
}

// WhereBetween 添加区间条件
func (q *QueryCondition) WhereBetween(field string, min, max any) *QueryCondition {
	q.whereBetween = append(q.whereBetween, BetweenCond{Field: field, Min: min, Max: max})
	return q
}

// WhereIn 添加 IN 条件
func (q *QueryCondition) WhereIn(field string, values ...any) *QueryCondition {
	q.whereIn = append(q.whereIn, InCond{Field: field, Values: values})
	return q
}

// WhereLike 添加 LIKE 条件
func (q *QueryCondition) WhereLike(field string, pattern string) *QueryCondition {
	q.whereLike = append(q.whereLike, LikeCond{Field: field, Pattern: pattern})
	return q
}

// WhereNull 添加 IS NULL 条件
func (q *QueryCondition) WhereNull(field string) *QueryCondition {
	q.whereNull = append(q.whereNull, field)
	return q
}

// WhereNotNull 添加 IS NOT NULL 条件
func (q *QueryCondition) WhereNotNull(field string) *QueryCondition {
	q.whereNotNull = append(q.whereNotNull, field)
	return q
}

// Preload 添加预加载关联
func (q *QueryCondition) Preload(relations ...string) *QueryCondition {
	q.preloads = append(q.preloads, relations...)
	return q
}

// OrderBy 添加排序字段
func (q *QueryCondition) OrderBy(field string, desc ...bool) *QueryCondition {
	isDesc := false
	if len(desc) > 0 && desc[0] {
		isDesc = true
	}
	q.orderBy = append(q.orderBy, OrderByField{Field: field, Desc: isDesc})
	return q
}

// Limit 设置限制
func (q *QueryCondition) Limit(n int) *QueryCondition {
	q.limit = n
	return q
}

// Offset 设置偏移
func (q *QueryCondition) Offset(n int) *QueryCondition {
	q.offset = n
	return q
}

// Getters
func (q *QueryCondition) GetWhereEq() []EqCond           { return q.whereEq }
func (q *QueryCondition) GetWhereNe() []NeCond           { return q.whereNe }
func (q *QueryCondition) GetWhereGt() []GtCond           { return q.whereGt }
func (q *QueryCondition) GetWhereLt() []LtCond           { return q.whereLt }
func (q *QueryCondition) GetWhereGe() []GeCond           { return q.whereGe }
func (q *QueryCondition) GetWhereLe() []LeCond           { return q.whereLe }
func (q *QueryCondition) GetWhereBetween() []BetweenCond { return q.whereBetween }
func (q *QueryCondition) GetWhereIn() []InCond           { return q.whereIn }
func (q *QueryCondition) GetWhereLike() []LikeCond       { return q.whereLike }
func (q *QueryCondition) GetWhereNull() []string         { return q.whereNull }
func (q *QueryCondition) GetWhereNotNull() []string      { return q.whereNotNull }
func (q *QueryCondition) GetPreloads() []string          { return q.preloads }
func (q *QueryCondition) GetOrderBy() []OrderByField     { return q.orderBy }
func (q *QueryCondition) GetLimit() int                  { return q.limit }
func (q *QueryCondition) GetOffset() int                 { return q.offset }
```

**Step 2: 创建测试**

Create `gincrud/query_condition_test.go`:

```go
package gincrud

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestNewQuery(t *testing.T) {
	q := NewQuery()
	assert.NotNil(t, q)
	assert.Empty(t, q.GetWhereEq())
}

func TestQueryCondition_WhereEq(t *testing.T) {
	q := NewQuery()
	q.WhereEq("status", 1)
	assert.Len(t, q.GetWhereEq(), 1)
	assert.Equal(t, "status", q.GetWhereEq()[0].Field)
}

func TestQueryCondition_WhereIn(t *testing.T) {
	q := NewQuery()
	q.WhereIn("role", "admin", "user")
	assert.Len(t, q.GetWhereIn(), 1)
	assert.Len(t, q.GetWhereIn()[0].Values, 2)
}

func TestQueryCondition_OrderBy(t *testing.T) {
	q := NewQuery()
	q.OrderBy("created_at").OrderBy("id", true)

	orderBy := q.GetOrderBy()
	assert.Len(t, orderBy, 2)
	assert.False(t, orderBy[0].Desc)
	assert.True(t, orderBy[1].Desc)
}

func TestQueryCondition_Chain(t *testing.T) {
	q := NewQuery()
	q.WhereEq("status", 1).
		WhereLike("name", "%test%").
		OrderBy("id", true).
		Limit(10).
		Offset(0)

	assert.Len(t, q.GetWhereEq(), 1)
	assert.Len(t, q.GetWhereLike(), 1)
	assert.Len(t, q.GetOrderBy(), 1)
	assert.Equal(t, 10, q.GetLimit())
	assert.Equal(t, 0, q.GetOffset())
}

func TestQueryCondition_Preload(t *testing.T) {
	q := NewQuery()
	q.Preload("Posts", "Comments")

	preloads := q.GetPreloads()
	assert.Len(t, preloads, 2)
	assert.Equal(t, "Posts", preloads[0])
	assert.Equal(t, "Comments", preloads[1])
}
```

**Step 3: 运行测试**

Run: `go test ./gincrud -run TestQueryCondition -v`

Expected: 所有测试通过

**Step 4: 提交**

```bash
git add gincrud/query_condition.go gincrud/query_condition_test.go
git commit -m "feat(gincrud): 实现 QueryCondition 查询条件构建器"
```

---

### Task 2: 定义错误码

**Files:**
- Create: `gincrud/errors.go`
- Test: 无需测试

**Step 1: 创建错误码**

Create `gincrud/errors.go`:

```go
package gincrud

import "errors"

var (
	ErrRecordNotFound          = errors.New("record not found")
	ErrCannotDeleteHasChildren = errors.New("cannot delete record with children")
	ErrDuplicateEntry          = errors.New("duplicate entry")
	ErrEntityIDRequired        = errors.New("entity ID is required")
)
```

**Step 2: 提交**

```bash
git add gincrud/errors.go
git commit -m "feat(gincrud): 定义错误码"
```

---

## Phase 2: IRepository 接口和 Repository 实现

### Task 3: 定义 IRepository 接口

**Files:**
- Create: `gincrud/repository.go`（部分）

**Step 1: 定义接口**

Append to `gincrud/repository.go`:

```go
package gincrud

import (
	"context"
	"gorm.io/gorm"
)

// IRepository Repository 接口
type IRepository[T Entity] interface {
	// 基础 CRUD
	GetByID(ctx context.Context, id int64) (*T, error)
	List(ctx context.Context, cond *QueryCondition, dto QueryDTO) ([]*T, int64, error)
	Create(ctx context.Context, entity *T) error
	Update(ctx context.Context, entity *T) error
	Delete(ctx context.Context, id int64) error
	TrulyDelete(ctx context.Context, id int64) error

	// 批量操作
	BatchCreate(ctx context.Context, entities []*T) error
	BatchUpdate(ctx context.Context, ids []int64, updates map[string]any) error
	BatchDelete(ctx context.Context, ids []int64) error

	// 查询
	Find(ctx context.Context, cond *QueryCondition) ([]*T, error)
	FindFirst(ctx context.Context, cond *QueryCondition) (*T, error)
	Count(ctx context.Context, cond *QueryCondition) (int64, error)
	Exists(ctx context.Context, id int64) (bool, error)

	// 事务
	WithTx(tx *gorm.DB) IRepository[T]

	// DB 访问
	DB() *gorm.DB
}
```

**Step 2: 提交**

```bash
git add gincrud/repository.go
git commit -m "feat(gincrud): 定义 IRepository 接口"
```

---

### Task 4: 实现 Repository 基础结构

**Files:**
- Modify: `gincrud/repository.go`

**Step 1: 实现基础结构**

Append to `gincrud/repository.go`:

```go
// Repository Repository 实现
type Repository[T Entity] struct {
	db *gorm.DB
}

// NewRepository 创建 Repository
func NewRepository[T Entity](db *gorm.DB) *Repository[T] {
	return &Repository[T]{db: db}
}

// DB 获取底层 db
func (r *Repository[T]) DB() *gorm.DB {
	return r.db
}

// WithTx 创建带事务的 Repository
func (r *Repository[T]) WithTx(tx *gorm.DB) IRepository[T] {
	return &Repository[T]{db: tx}
}
```

**Step 2: 提交**

```bash
git add gincrud/repository.go
git commit -m "feat(gincrud): 实现 Repository 基础结构"
```

---

### Task 5: 实现 GetByID 和 List

**Files:**
- Modify: `gincrud/repository.go`
- Test: `gincrud/repository_test.go`

**Step 1: 实现 GetByID**

Append to `gincrud/repository.go`:

```go
import (
	"errors"
	"time"
)

// GetByID 根据 ID 查询
func (r *Repository[T]) GetByID(ctx context.Context, id int64) (*T, error) {
	var entity T
	entity.SetID(id)

	db := r.db.WithContext(ctx)
	db = r.withSoftDelete(db)

	if err := db.First(&entity).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}
	return &entity, nil
}
```

**Step 2: 实现 List**

Append to `gincrud/repository.go`:

```go
// List 列表查询（分页）
func (r *Repository[T]) List(ctx context.Context, cond *QueryCondition, dto QueryDTO) ([]*T, int64, error) {
	db := r.db.WithContext(ctx)
	db = r.withSoftDelete(db)
	db = r.applyCondition(db, cond)
	db = r.applyPagination(db, dto)

	var list []*T
	var total int64

	// 计数
	if err := db.Model(new(T)).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 查询
	if err := db.Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
```

**Step 3: 创建测试**

Create `gincrud/repository_test.go`:

```go
package gincrud

import (
	"context"
	"testing"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"github.com/stretchr/testify/assert"
)

// TestUser 测试用模型
type TestUser struct {
	BaseEntity
	Name   string `json:"name"`
	Email  string `json:"email"`
	Age    int    `json:"age"`
	Status int    `gorm:"default:1" json:"status"`
}

func (TestUser) TableName() string { return "test_users" }

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	assert.NoError(t, err)

	err = db.AutoMigrate(&TestUser{})
	assert.NoError(t, err)

	return db
}

func TestRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository[TestUser](db)
	ctx := context.Background()

	// 创建测试数据
	user := &TestUser{Name: "Test", Email: "test@example.com", Age: 25}
	err := repo.Create(ctx, user)
	assert.NoError(t, err)

	// 查询
	found, err := repo.GetByID(ctx, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, found.ID)
	assert.Equal(t, "Test", found.Name)
}

func TestRepository_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository[TestUser](db)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, 999)
	assert.Error(t, err)
	assert.Equal(t, ErrRecordNotFound, err)
}
```

**Step 4: 运行测试**

Run: `go test ./gincrud -run TestRepository_GetByID -v`

Expected: 测试通过

**Step 5: 提交**

```bash
git add gincrud/repository.go gincrud/repository_test.go
git commit -m "feat(gincrud): 实现 GetByID 和 List 方法"
```

---

### Task 6: 实现 Create 和 Update

**Files:**
- Modify: `gincrud/repository.go`

**Step 1: 实现 Create**

Append to `gincrud/repository.go`:

```go
// Create 创建
func (r *Repository[T]) Create(ctx context.Context, entity *T) error {
	entity.SetCreatedAt(time.Now())
	entity.SetUpdatedAt(time.Now())
	entity.SetDeleted(0)

	return r.db.WithContext(ctx).Create(entity).Error
}
```

**Step 2: 实现 Update**

Append to `gincrud/repository.go`:

```go
// Update 更新
func (r *Repository[T]) Update(ctx context.Context, entity *T) error {
	if entity.GetID() == 0 {
		return ErrEntityIDRequired
	}

	entity.SetUpdatedAt(time.Now())

	return r.db.WithContext(ctx).Save(entity).Error
}
```

**Step 3: 添加测试**

Append to `gincrud/repository_test.go`:

```go
func TestRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository[TestUser](db)
	ctx := context.Background()

	user := &TestUser{Name: "New User", Email: "new@example.com"}
	err := repo.Create(ctx, user)
	assert.NoError(t, err)
	assert.NotZero(t, user.ID)
	assert.NotZero(t, user.CreatedAt)
	assert.NotZero(t, user.UpdatedAt)
	assert.Zero(t, user.Deleted)
}

func TestRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository[TestUser](db)
	ctx := context.Background()

	user := &TestUser{Name: "Original", Email: "original@example.com"}
	repo.Create(ctx, user)

	user.Name = "Updated"
	err := repo.Update(ctx, user)
	assert.NoError(t, err)

	found, _ := repo.GetByID(ctx, user.ID)
	assert.Equal(t, "Updated", found.Name)
}

func TestRepository_Update_IDRequired(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository[TestUser](db)
	ctx := context.Background()

	user := &TestUser{}
	err := repo.Update(ctx, user)
	assert.Equal(t, ErrEntityIDRequired, err)
}
```

**Step 4: 提交**

```bash
git add gincrud/repository.go gincrud/repository_test.go
git commit -m "feat(gincrud): 实现 Create 和 Update 方法"
```

---

### Task 7: 实现 Delete 和 TrulyDelete

**Files:**
- Modify: `gincrud/repository.go`

**Step 1: 实现 Delete（软删除）**

Append to `gincrud/repository.go`:

```go
// Delete 软删除
func (r *Repository[T]) Delete(ctx context.Context, id int64) error {
	var entity T
	entity.SetID(id)
	entity.SetDeleted(time.Now().Unix())

	return r.db.WithContext(ctx).Save(&entity).Error
}
```

**Step 2: 实现 TrulyDelete（硬删除）**

Append to `gincrud/repository.go`:

```go
// TrulyDelete 硬删除
func (r *Repository[T]) TrulyDelete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(new(T), id).Error
}
```

**Step 3: 添加测试**

Append to `gincrud/repository_test.go`:

```go
func TestRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository[TestUser](db)
	ctx := context.Background()

	user := &TestUser{Name: "To Delete", Email: "delete@example.com"}
	repo.Create(ctx, user)

	err := repo.Delete(ctx, user.ID)
	assert.NoError(t, err)

	// 软删除后无法通过 GetByID 查询
	_, err = repo.GetByID(ctx, user.ID)
	assert.Equal(t, ErrRecordNotFound, err)
}

func TestRepository_TrulyDelete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository[TestUser](db)
	ctx := context.Background()

	user := &TestUser{Name: "To Really Delete", Email: "really@example.com"}
	repo.Create(ctx, user)

	err := repo.TrulyDelete(ctx, user.ID)
	assert.NoError(t, err)

	// 硬删除后记录不存在
	var count int64
	db.Unscoped().Model(&TestUser{}).Where("id = ?", user.ID).Count(&count)
	assert.Zero(t, count)
}
```

**Step 4: 提交**

```bash
git add gincrud/repository.go gincrud/repository_test.go
git commit -m "feat(gincrud): 实现 Delete 和 TrulyDelete 方法"
```

---

### Task 8: 实现批量操作

**Files:**
- Modify: `gincrud/repository.go`

**Step 1: 实现 BatchCreate**

Append to `gincrud/repository.go`:

```go
// BatchCreate 批量创建
func (r *Repository[T]) BatchCreate(ctx context.Context, entities []*T) error {
	now := time.Now()
	for _, e := range entities {
		e.SetCreatedAt(now)
		e.SetUpdatedAt(now)
		e.SetDeleted(0)
	}

	return r.db.WithContext(ctx).Create(&entities).Error
}
```

**Step 2: 实现 BatchUpdate**

Append to `gincrud/repository.go`:

```go
// BatchUpdate 批量更新
func (r *Repository[T]) BatchUpdate(ctx context.Context, ids []int64, updates map[string]any) error {
	return r.db.WithContext(ctx).
		Model(new(T)).
		Where("id IN ?", ids).
		Updates(updates).Error
}
```

**Step 3: 实现 BatchDelete**

Append to `gincrud/repository.go`:

```go
// BatchDelete 批量软删除
func (r *Repository[T]) BatchDelete(ctx context.Context, ids []int64) error {
	return r.db.WithContext(ctx).
		Model(new(T)).
		Where("id IN ?", ids).
		Update("deleted", time.Now().Unix()).Error
}
```

**Step 4: 添加测试**

Append to `gincrud/repository_test.go`:

```go
func TestRepository_BatchCreate(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository[TestUser](db)
	ctx := context.Background()

	users := []*TestUser{
		{Name: "User1", Email: "user1@example.com"},
		{Name: "User2", Email: "user2@example.com"},
	}

	err := repo.BatchCreate(ctx, users)
	assert.NoError(t, err)
	assert.NotZero(t, users[0].ID)
	assert.NotZero(t, users[1].ID)
}

func TestRepository_BatchUpdate(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository[TestUser](db)
	ctx := context.Background()

	users := []*TestUser{
		{Name: "User1", Email: "user1@example.com", Status: 1},
		{Name: "User2", Email: "user2@example.com", Status: 1},
	}
	repo.BatchCreate(ctx, users)

	ids := []int64{users[0].ID, users[1].ID}
	err := repo.BatchUpdate(ctx, ids, map[string]any{"status": 0})
	assert.NoError(t, err)

	// 验证更新
	found, _ := repo.GetByID(ctx, users[0].ID)
	assert.Equal(t, 0, found.Status)
}

func TestRepository_BatchDelete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository[TestUser](db)
	ctx := context.Background()

	users := []*TestUser{
		{Name: "User1", Email: "user1@example.com"},
		{Name: "User2", Email: "user2@example.com"},
	}
	repo.BatchCreate(ctx, users)

	ids := []int64{users[0].ID, users[1].ID}
	err := repo.BatchDelete(ctx, ids)
	assert.NoError(t, err)

	// 验证软删除
	_, err = repo.GetByID(ctx, users[0].ID)
	assert.Equal(t, ErrRecordNotFound, err)
}
```

**Step 5: 提交**

```bash
git add gincrud/repository.go gincrud/repository_test.go
git commit -m "feat(gincrud): 实现批量操作方法"
```

---

### Task 9: 实现 Find 系列查询方法

**Files:**
- Modify: `gincrud/repository.go`

**Step 1: 实现 Find**

Append to `gincrud/repository.go`:

```go
// Find 条件查询
func (r *Repository[T]) Find(ctx context.Context, cond *QueryCondition) ([]*T, error) {
	db := r.db.WithContext(ctx)
	db = r.withSoftDelete(db)
	db = r.applyCondition(db, cond)

	var list []*T
	if err := db.Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}
```

**Step 2: 实现 FindFirst**

Append to `gincrud/repository.go`:

```go
// FindFirst 查询第一条
func (r *Repository[T]) FindFirst(ctx context.Context, cond *QueryCondition) (*T, error) {
	db := r.db.WithContext(ctx)
	db = r.withSoftDelete(db)
	db = r.applyCondition(db, cond)

	var entity T
	if err := db.First(&entity).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &entity, nil
}
```

**Step 3: 实现 Count**

Append to `gincrud/repository.go`:

```go
// Count 计数
func (r *Repository[T]) Count(ctx context.Context, cond *QueryCondition) (int64, error) {
	db := r.db.WithContext(ctx)
	db = r.withSoftDelete(db)
	db = r.applyCondition(db, cond)

	var count int64
	if err := db.Model(new(T)).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
```

**Step 4: 实现 Exists**

Append to `gincrud/repository.go`:

```go
// Exists 检查是否存在
func (r *Repository[T]) Exists(ctx context.Context, id int64) (bool, error) {
	var count int64
	db := r.db.WithContext(ctx)
	db = r.withSoftDelete(db)

	err := db.Model(new(T)).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}
```

**Step 5: 添加测试**

Append to `gincrud/repository_test.go`:

```go
func TestRepository_Find(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository[TestUser](db)
	ctx := context.Background()

	users := []*TestUser{
		{Name: "Alice", Email: "alice@example.com", Status: 1},
		{Name: "Bob", Email: "bob@example.com", Status: 0},
	}
	repo.BatchCreate(ctx, users)

	cond := NewQuery().WhereEq("status", 1)
	found, err := repo.Find(ctx, cond)
	assert.NoError(t, err)
	assert.Len(t, found, 1)
	assert.Equal(t, "Alice", found[0].Name)
}

func TestRepository_FindFirst(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository[TestUser](db)
	ctx := context.Background()

	repo.Create(ctx, &TestUser{Name: "First", Email: "first@example.com"})

	cond := NewQuery().WhereEq("name", "First")
	found, err := repo.FindFirst(ctx, cond)
	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, "First", found.Name)
}

func TestRepository_Count(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository[TestUser](db)
	ctx := context.Background()

	repo.Create(ctx, &TestUser{Name: "User1", Status: 1})
	repo.Create(ctx, &TestUser{Name: "User2", Status: 1})
	repo.Create(ctx, &TestUser{Name: "User3", Status: 0})

	cond := NewQuery().WhereEq("status", 1)
	count, err := repo.Count(ctx, cond)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

func TestRepository_Exists(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository[TestUser](db)
	ctx := context.Background()

	user := &TestUser{Name: "Exists", Email: "exists@example.com"}
	repo.Create(ctx, user)

	exists, err := repo.Exists(ctx, user.ID)
	assert.NoError(t, err)
	assert.True(t, exists)

	exists, err = repo.Exists(ctx, 999)
	assert.NoError(t, err)
	assert.False(t, exists)
}
```

**Step 6: 提交**

```bash
git add gincrud/repository.go gincrud/repository_test.go
git commit -m "feat(gincrud): 实现 Find 系列查询方法"
```

---

## Phase 3: 辅助方法

### Task 10: 实现 withSoftDelete 和 applyCondition

**Files:**
- Modify: `gincrud/repository.go`

**Step 1: 实现 withSoftDelete**

Append to `gincrud/repository.go`:

```go
// =============================================================================
// 辅助方法
// =============================================================================

// withSoftDelete 自动过滤软删除
func (r *Repository[T]) withSoftDelete(db *gorm.DB) *gorm.DB {
	return db.Where("deleted = 0")
}
```

**Step 2: 实现 applyCondition**

Append to `gincrud/repository.go`:

```go
// applyCondition 应用查询条件
func (r *Repository[T]) applyCondition(db *gorm.DB, cond *QueryCondition) *gorm.DB {
	if cond == nil {
		return db
	}

	// 等于条件
	for _, c := range cond.GetWhereEq() {
		db = db.Where(c.Field+" = ?", c.Value)
	}

	// 不等于条件
	for _, c := range cond.GetWhereNe() {
		db = db.Where(c.Field+" != ?", c.Value)
	}

	// 大于条件
	for _, c := range cond.GetWhereGt() {
		db = db.Where(c.Field+" > ?", c.Value)
	}

	// 小于条件
	for _, c := range cond.GetWhereLt() {
		db = db.Where(c.Field+" < ?", c.Value)
	}

	// 大于等于条件
	for _, c := range cond.GetWhereGe() {
		db = db.Where(c.Field+" >= ?", c.Value)
	}

	// 小于等于条件
	for _, c := range cond.GetWhereLe() {
		db = db.Where(c.Field+" <= ?", c.Value)
	}

	// 区间条件
	for _, c := range cond.GetWhereBetween() {
		db = db.Where(c.Field+" BETWEEN ? AND ?", c.Min, c.Max)
	}

	// IN 条件
	for _, c := range cond.GetWhereIn() {
		db = db.Where(c.Field+" IN ?", c.Values)
	}

	// LIKE 条件
	for _, c := range cond.GetWhereLike() {
		db = db.Where(c.Field+" LIKE ?", c.Pattern)
	}

	// IS NULL 条件
	for _, field := range cond.GetWhereNull() {
		db = db.Where(field+" IS NULL")
	}

	// IS NOT NULL 条件
	for _, field := range cond.GetWhereNotNull() {
		db = db.Where(field+" IS NOT NULL")
	}

	// 排序
	for _, o := range cond.GetOrderBy() {
		if o.Desc {
			db = db.Order(o.Field + " DESC")
		} else {
			db = db.Order(o.Field + " ASC")
		}
	}

	// 预加载
	for _, relation := range cond.GetPreloads() {
		db = db.Preload(relation)
	}

	// Limit/Offset
	if cond.GetLimit() > 0 {
		db = db.Limit(cond.GetLimit())
	}
	if cond.GetOffset() > 0 {
		db = db.Offset(cond.GetOffset())
	}

	return db
}
```

**Step 3: 实现 applyPagination**

Append to `gincrud/repository.go`:

```go
// applyPagination 应用分页
func (r *Repository[T]) applyPagination(db *gorm.DB, dto QueryDTO) *gorm.DB {
	// 排序
	if dto.GetSortBy() != "" {
		order := dto.GetSortBy() + " " + dto.GetSortOrder()
		db = db.Order(order)
	}

	// 分页
	if dto.GetPage() > 0 && dto.GetPageSize() > 0 {
		db = db.Offset(dto.Offset()).Limit(dto.Limit())
	}

	return db
}
```

**Step 4: 运行所有测试**

Run: `go test ./gincrud -v`

Expected: 所有测试通过

**Step 5: 提交**

```bash
git add gincrud/repository.go
git commit -m "feat(gincrud): 实现辅助方法"
```

---

## Phase 4: 扩展示例和文档

### Task 11: 创建扩展示例

**Files:**
- Create: `examples/repository/user_repository.go`
- Create: `examples/repository/models.go`

**Step 1: 创建模型**

Create `examples/repository/models.go`:

```go
package repository

import "github.com/io-max/go-lib/gincrud"

// User 用户模型
type User struct {
	gincrud.BaseEntity
	Username string `gorm:"uniqueIndex;size:50" json:"username"`
	Email    string `gorm:"uniqueIndex;size:100" json:"email"`
	Password string `gorm:"size:255" json:"-"`
	Role     string `gorm:"size:20" json:"role"`
	Status   int    `gorm:"default:1" json:"status"`
}

func (User) TableName() string { return "users" }

// Post 文章模型
type Post struct {
	gincrud.BaseEntity
	Title   string `json:"title"`
	Content string `json:"content"`
	UserID  int64  `json:"user_id"`
}

func (Post) TableName() string { return "posts" }
```

**Step 2: 创建自定义 Repository**

Create `examples/repository/user_repository.go`:

```go
package repository

import (
	"context"
	"github.com/io-max/go-lib/gincrud"
	"gorm.io/gorm"
)

// UserRepository 用户数据访问层
type UserRepository struct {
	*gincrud.Repository[User]
}

// NewUserRepository 创建 UserRepository
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		Repository: gincrud.NewRepository[User](db),
	}
}

// GetByUsername 根据用户名查询
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*User, error) {
	var user User
	err := r.DB().WithContext(ctx).
		Where("username = ?", username).
		First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// GetByEmail 根据邮箱查询
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	err := r.DB().WithContext(ctx).
		Where("email = ?", email).
		First(&user).Error
	return &user, err
}

// GetByUsernameWithPosts 根据用户名查询（带文章预加载）
func (r *UserRepository) GetByUsernameWithPosts(ctx context.Context, username string) (*User, error) {
	var user User
	err := r.DB().WithContext(ctx).
		Where("username = ?", username).
		Preload("Posts").
		First(&user).Error
	return &user, err
}

// UpdateStatus 批量更新状态
func (r *UserRepository) UpdateStatus(ctx context.Context, ids []int64, status int) error {
	return r.DB().WithContext(ctx).
		Model(&User{}).
		Where("id IN ?", ids).
		Update("status", status).Error
}

// ExistsByEmail 检查邮箱是否存在
func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	err := r.DB().WithContext(ctx).
		Model(&User{}).
		Where("email = ?", email).
		Count(&count).Error
	return count > 0, err
}
```

**Step 3: 提交**

```bash
git add examples/repository/
git commit -m "docs: 添加 Repository 扩展示例"
```

---

### Task 12: 更新 README

**Files:**
- Modify: `README.md`

**Step 1: 添加 Repository 使用示例**

Append to `README.md`:

```markdown
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
        OrderBy("created_at", true)

    users, total, _ := repo.List(ctx, cond, &gincrud.BaseQueryDTO{
        Page: 1, PageSize: 10,
    })
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
```

**Step 2: 提交**

```bash
git add README.md
git commit -m "docs: 添加 Repository 使用文档"
```

---

## 总结

本计划共 12 个 Task，涵盖：
1. QueryCondition 查询条件构建器
2. 错误码定义
3. IRepository 接口
4. Repository 基础结构
5. GetByID/List 实现
6. Create/Update 实现
7. Delete/TrulyDelete 实现
8. 批量操作实现
9. Find 系列查询方法
10. 辅助方法
11. 扩展示例
12. 文档更新

预计完成时间：约 1-2 小时
