package repository

import "github.com/your-org/go-lib/gincrud"

// User 用户模型
type User struct {
	*gincrud.BaseEntity
	Username string `gorm:"uniqueIndex;size:50" json:"username"`
	Email    string `gorm:"uniqueIndex;size:100" json:"email"`
	Password string `gorm:"size:255" json:"-"`
	Role     string `gorm:"size:20" json:"role"`
	Status   int    `gorm:"default:1" json:"status"`
}

func NewUser() *User {
	return &User{BaseEntity: &gincrud.BaseEntity{}}
}

func (User) TableName() string { return "users" }

// Post 文章模型
type Post struct {
	*gincrud.BaseEntity
	Title   string `json:"title"`
	Content string `json:"content"`
	UserID  int64  `json:"user_id"`
}

func NewPost() *Post {
	return &Post{BaseEntity: &gincrud.BaseEntity{}}
}

func (Post) TableName() string { return "posts" }
