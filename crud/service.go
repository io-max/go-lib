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
