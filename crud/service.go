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
	_, err := s.dtoToEntity(dto)
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
