package gincrud

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

// IRepository Repository 接口
type IRepository[T Entity] interface {
	// 基础 CRUD
	GetByID(ctx context.Context, id int64) (*T, error)
	FindPage(ctx context.Context, cond *QueryCondition) ([]*T, int64, error)
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

// withSoftDelete 添加软删除过滤
func (r *Repository[T]) withSoftDelete(db *gorm.DB) *gorm.DB {
	// 始终添加软删除过滤，因为 Entity 接口要求有 Deleted 字段
	return db.Where("deleted = 0")
}

// applyCondition 应用查询条件
func (r *Repository[T]) applyCondition(db *gorm.DB, cond *QueryCondition) *gorm.DB {
	if cond == nil {
		return db
	}

	for _, c := range cond.GetWhereEq() {
		db = db.Where(c.Field+" = ?", c.Value)
	}
	for _, c := range cond.GetWhereNe() {
		db = db.Where(c.Field+" != ?", c.Value)
	}
	for _, c := range cond.GetWhereGt() {
		db = db.Where(c.Field+" > ?", c.Value)
	}
	for _, c := range cond.GetWhereLt() {
		db = db.Where(c.Field+" < ?", c.Value)
	}
	for _, c := range cond.GetWhereGe() {
		db = db.Where(c.Field+" >= ?", c.Value)
	}
	for _, c := range cond.GetWhereLe() {
		db = db.Where(c.Field+" <= ?", c.Value)
	}
	for _, c := range cond.GetWhereBetween() {
		db = db.Where(c.Field+" BETWEEN ? AND ?", c.Min, c.Max)
	}
	for _, c := range cond.GetWhereIn() {
		db = db.Where(c.Field+" IN ?", c.Values)
	}
	for _, c := range cond.GetWhereLike() {
		db = db.Where(c.Field+" LIKE ?", c.Pattern)
	}
	for _, field := range cond.GetWhereNull() {
		db = db.Where(field+" IS NULL")
	}
	for _, field := range cond.GetWhereNotNull() {
		db = db.Where(field+" IS NOT NULL")
	}
	for _, order := range cond.GetOrderBy() {
		if order.Desc {
			db = db.Order(order.Field + " DESC")
		} else {
			db = db.Order(order.Field + " ASC")
		}
	}

	return db
}

// applyPagination 应用分页和排序
func (r *Repository[T]) applyPagination(db *gorm.DB, cond *QueryCondition) *gorm.DB {
	if cond == nil {
		return db
	}
	db = db.Limit(cond.Limit()).Offset(cond.Offset())

	// 应用排序
	for _, order := range cond.GetOrderBy() {
		if order.Desc {
			db = db.Order(order.Field + " DESC")
		} else {
			db = db.Order(order.Field + " ASC")
		}
	}

	return db
}

// GetByID 根据 ID 查询
func (r *Repository[T]) GetByID(ctx context.Context, id int64) (*T, error) {
	var entity T

	db := r.db.WithContext(ctx)
	db = r.withSoftDelete(db)

	if err := db.Where("id = ?", id).First(&entity).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}
	return &entity, nil
}

// FindPage 分页查询
func (r *Repository[T]) FindPage(ctx context.Context, cond *QueryCondition) ([]*T, int64, error) {
	db := r.db.WithContext(ctx)
	db = r.withSoftDelete(db)
	db = r.applyCondition(db, cond)

	var total int64

	// 计数（不应用分页）
	if err := db.Model(new(T)).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 应用分页
	db = r.applyPagination(db, cond)

	// 查询
	var list []*T
	if err := db.Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// Create 创建
func (r *Repository[T]) Create(ctx context.Context, entity *T) error {
	(*entity).SetCreatedAt(time.Now())
	(*entity).SetUpdatedAt(time.Now())
	(*entity).SetDeleted(0)

	return r.db.WithContext(ctx).Create(entity).Error
}

// Update 更新
func (r *Repository[T]) Update(ctx context.Context, entity *T) error {
	if (*entity).GetID() == 0 {
		return ErrEntityIDRequired
	}

	(*entity).SetUpdatedAt(time.Now())

	return r.db.WithContext(ctx).Save(entity).Error
}

// Delete 软删除记录
func (r *Repository[T]) Delete(ctx context.Context, id int64) error {
	db := r.db.WithContext(ctx)
	db = r.withSoftDelete(db)

	// 先查询实体
	var entity T
	if err := db.Where("id = ?", id).First(&entity).Error; err != nil {
		return err
	}

	// 使用 Update 直接更新 deleted 字段，不再次应用软删除过滤
	return r.db.Model(&entity).Where("id = ?", id).Update("deleted", time.Now().Unix()).Error
}

// TrulyDelete 硬删除记录
func (r *Repository[T]) TrulyDelete(ctx context.Context, id int64) error {
	var entity T
	entity.SetID(id)
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&entity).Error
}

// BatchCreate 批量创建
func (r *Repository[T]) BatchCreate(ctx context.Context, entities []*T) error {
	now := time.Now()
	for _, e := range entities {
		(*e).SetCreatedAt(now)
		(*e).SetUpdatedAt(now)
		(*e).SetDeleted(0)
	}

	return r.db.WithContext(ctx).Create(&entities).Error
}

// BatchUpdate 批量更新
func (r *Repository[T]) BatchUpdate(ctx context.Context, ids []int64, updates map[string]any) error {
	db := r.db.WithContext(ctx)
	var entity T
	return db.Model(&entity).Where("id IN ?", ids).Updates(updates).Error
}

// BatchDelete 批量软删除
func (r *Repository[T]) BatchDelete(ctx context.Context, ids []int64) error {
	db := r.db.WithContext(ctx)
	updates := map[string]any{"deleted": time.Now().Unix()}
	var entity T
	return db.Model(&entity).Where("id IN ?", ids).Updates(updates).Error
}

// Find 条件查询
func (r *Repository[T]) Find(ctx context.Context, cond *QueryCondition) ([]*T, error) {
	db := r.db.WithContext(ctx)
	db = r.withSoftDelete(db)
	db = r.applyCondition(db, cond)

	var list []*T
	err := db.Find(&list).Error
	return list, err
}

// FindFirst 查询第一条
func (r *Repository[T]) FindFirst(ctx context.Context, cond *QueryCondition) (*T, error) {
	db := r.db.WithContext(ctx)
	db = r.withSoftDelete(db)
	db = r.applyCondition(db, cond)

	var entity T
	err := db.First(&entity).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}
	return &entity, nil
}

// Count 条件计数
func (r *Repository[T]) Count(ctx context.Context, cond *QueryCondition) (int64, error) {
	db := r.db.WithContext(ctx)
	db = r.withSoftDelete(db)
	db = r.applyCondition(db, cond)

	var total int64
	err := db.Model(new(T)).Count(&total).Error
	return total, err
}

// Exists 检查是否存在
func (r *Repository[T]) Exists(ctx context.Context, id int64) (bool, error) {
	db := r.db.WithContext(ctx)
	db = r.withSoftDelete(db)

	var entity T
	entity.SetID(id)

	var exists int64
	err := db.Model(&entity).Where("id = ?", id).Count(&exists).Error
	return exists > 0, err
}
