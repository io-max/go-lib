package service

import (
	"time"
)

// User 用户模型（使用值接收者方法以满足 crud.Entity 接口）
type User struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	Deleted   int64     `gorm:"default:0;index" json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Username  string    `gorm:"uniqueIndex;size:50" json:"username"`
	Email     string    `gorm:"uniqueIndex;size:100" json:"email"`
	Password  string    `gorm:"size:255" json:"-"`
	Status    int       `gorm:"default:1" json:"status"`
}

// TableName 表名
func (User) TableName() string { return "users" }

// 实现 crud.Entity 接口（值接收者，这样 User 值类型也满足接口）
func (u User) GetID() int64              { return u.ID }
func (u User) SetID(id int64)            { u.ID = id }
func (u User) GetDeleted() int64         { return u.Deleted }
func (u User) SetDeleted(ts int64)       { u.Deleted = ts }
func (u User) GetCreatedAt() time.Time   { return u.CreatedAt }
func (u User) SetCreatedAt(t time.Time)  { u.CreatedAt = t }
func (u User) GetUpdatedAt() time.Time   { return u.UpdatedAt }
func (u User) SetUpdatedAt(t time.Time)  { u.UpdatedAt = t }

// Post 文章模型（用于演示关联）
type Post struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	Deleted   int64     `gorm:"default:0;index" json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Title     string    `gorm:"size:100" json:"title"`
	Content   string    `gorm:"type:text" json:"content"`
	UserID    int64     `gorm:"index" json:"user_id"`
}

func (Post) TableName() string { return "posts" }

// 实现 crud.Entity 接口（值接收者）
func (p Post) GetID() int64              { return p.ID }
func (p Post) SetID(id int64)            { p.ID = id }
func (p Post) GetDeleted() int64         { return p.Deleted }
func (p Post) SetDeleted(ts int64)       { p.Deleted = ts }
func (p Post) GetCreatedAt() time.Time   { return p.CreatedAt }
func (p Post) SetCreatedAt(t time.Time)  { p.CreatedAt = t }
func (p Post) GetUpdatedAt() time.Time   { return p.UpdatedAt }
func (p Post) SetUpdatedAt(t time.Time)  { p.UpdatedAt = t }
