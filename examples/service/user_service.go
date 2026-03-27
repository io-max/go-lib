package service

import (
	"context"
	"errors"

	"github.com/io-max/go-lib/crud"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserOptDTO 用户操作 DTO（Create + Update 复用）
type UserOptDTO struct {
	ID       int64  `json:"id" form:"id"`
	Username string `json:"username" form:"username" binding:"omitempty,min=3,max=50"`
	Email    string `json:"email" form:"email" binding:"omitempty,email"`
	Password string `json:"password" form:"password" binding:"omitempty,min=6"`
	Status   int    `json:"status" form:"status" binding:"omitempty,oneof=1 2 3"`
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

// UserRepository User Repository 类型
type UserRepository = *crud.Repository[User]

// UserService 用户服务（基于 Repository 组合，不使用泛型 Service 基类）
type UserService struct {
	repo UserRepository
	db   *gorm.DB
}

// NewUserService 创建用户服务
func NewUserService(db *gorm.DB) *UserService {
	return &UserService{
		repo: crud.NewRepository[User](db),
		db:   db,
	}
}

// dtoToEntity DTO 转 Entity
func dtoToEntity(dto *UserOptDTO) (User, error) {
	user := User{}

	// Update 场景
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
			return User{}, err
		}
		user.Password = string(hashed)
	}
	if dto.Status > 0 {
		user.Status = dto.Status
	}

	return user, nil
}

// entityToResponse Entity 转 Response
func entityToResponse(user User) UserResponse {
	return UserResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Status:   user.Status,
	}
}

// queryToCond Query DTO 转 QueryCondition
func queryToCond(query UserQueryDTO) *crud.QueryCondition {
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
}

// Create 创建用户
func (s *UserService) Create(ctx context.Context, dto *UserOptDTO) (UserResponse, error) {
	user, err := dtoToEntity(dto)
	if err != nil {
		return UserResponse{}, err
	}

	if err := s.repo.Create(ctx, &user); err != nil {
		return UserResponse{}, err
	}

	return entityToResponse(user), nil
}

// GetByID 根据 ID 获取
func (s *UserService) GetByID(ctx context.Context, id int64) (UserResponse, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return UserResponse{}, err
	}

	return entityToResponse(*user), nil
}

// Update 更新用户
func (s *UserService) Update(ctx context.Context, dto *UserOptDTO) (UserResponse, error) {
	user, err := dtoToEntity(dto)
	if err != nil {
		return UserResponse{}, err
	}

	if err := s.repo.Update(ctx, &user); err != nil {
		return UserResponse{}, err
	}

	return entityToResponse(user), nil
}

// Delete 删除用户
func (s *UserService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

// List 列表查询
func (s *UserService) List(ctx context.Context, query UserQueryDTO) ([]UserResponse, error) {
	cond := queryToCond(query)
	entities, err := s.repo.Find(ctx, cond)
	if err != nil {
		return nil, err
	}

	var results []UserResponse
	for _, e := range entities {
		results = append(results, entityToResponse(*e))
	}

	return results, nil
}

// Page 分页查询
func (s *UserService) Page(ctx context.Context, query UserQueryDTO) (*crud.PageResult[UserResponse], error) {
	cond := queryToCond(query)
	entities, total, err := s.repo.FindPage(ctx, cond)
	if err != nil {
		return nil, err
	}

	var results []UserResponse
	for _, e := range entities {
		results = append(results, entityToResponse(*e))
	}

	return &crud.PageResult[UserResponse]{
		List:  results,
		Total: total,
	}, nil
}

// ChangePassword 修改密码（扩展方法）
func (s *UserService) ChangePassword(ctx context.Context, userID int64, newPassword string) error {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return crud.ErrRecordNotFound
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.Password = string(hashed)
	return s.repo.Update(ctx, user)
}

// GetByUsername 根据用户名查询（扩展方法）
func (s *UserService) GetByUsername(ctx context.Context, username string) (UserResponse, error) {
	cond := crud.NewQuery().WhereEq("username", username)
	user, err := s.repo.FindFirst(ctx, cond)
	if err != nil {
		return UserResponse{}, err
	}
	if user == nil {
		return UserResponse{}, errors.New("user not found")
	}
	return entityToResponse(*user), nil
}
