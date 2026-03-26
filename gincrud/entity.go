package gincrud

import "time"

// Entity 实体接口
type Entity interface {
	GetID() int64
	SetID(id int64)
	GetDeleted() int64
	SetDeleted(ts int64)
	GetCreatedAt() time.Time
	SetCreatedAt(t time.Time)
	GetUpdatedAt() time.Time
	SetUpdatedAt(t time.Time)
}

// BaseEntity 基础实体
type BaseEntity struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	Deleted   int64     `gorm:"default:0;index" json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (e *BaseEntity) GetID() int64        { return e.ID }
func (e *BaseEntity) SetID(id int64)      { e.ID = id }
func (e *BaseEntity) GetDeleted() int64   { return e.Deleted }
func (e *BaseEntity) SetDeleted(ts int64) { e.Deleted = ts }
func (e *BaseEntity) GetCreatedAt() time.Time { return e.CreatedAt }
func (e *BaseEntity) SetCreatedAt(t time.Time) { e.CreatedAt = t }
func (e *BaseEntity) GetUpdatedAt() time.Time { return e.UpdatedAt }
func (e *BaseEntity) SetUpdatedAt(t time.Time) { e.UpdatedAt = t }

// IsDeleted 检查是否已删除
func (e *BaseEntity) IsDeleted() bool {
	return e.Deleted > 0
}

// MarkDeleted 标记为删除
func (e *BaseEntity) MarkDeleted() {
	e.Deleted = time.Now().Unix()
}

// AuditEntity 审计实体
type AuditEntity struct {
	BaseEntity
	CreatedBy int64 `gorm:"default:0" json:"created_by"`
	UpdatedBy int64 `gorm:"default:0" json:"updated_by"`
}

func (e *AuditEntity) GetCreatedBy() int64 { return e.CreatedBy }
func (e *AuditEntity) SetCreatedBy(id int64) { e.CreatedBy = id }
func (e *AuditEntity) GetUpdatedBy() int64 { return e.UpdatedBy }
func (e *AuditEntity) SetUpdatedBy(id int64) { e.UpdatedBy = id }
