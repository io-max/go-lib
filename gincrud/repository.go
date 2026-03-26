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
