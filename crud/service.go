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
