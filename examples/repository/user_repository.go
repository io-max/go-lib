package repository

import (
	"context"

	"github.com/io-max/go-lib/crud"
	"gorm.io/gorm"
)

// UserRepository 用户数据访问层
type UserRepository struct {
	*crud.Repository[User]
}

// NewUserRepository 创建 UserRepository
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		Repository: crud.NewRepository[User](db),
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
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// GetByUsernameWithPosts 根据用户名查询（带文章预加载）
func (r *UserRepository) GetByUsernameWithPosts(ctx context.Context, username string) (*User, error) {
	var user User
	err := r.DB().WithContext(ctx).
		Where("username = ?", username).
		Preload("Posts").
		First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
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
