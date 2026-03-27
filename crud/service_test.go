package crud

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// ServiceTestEntity 测试实体（Service 层专用）
type ServiceTestEntity struct {
	BaseEntity
	Name string `gorm:"size:255" json:"name"`
}

func (ServiceTestEntity) TableName() string { return "service_test_entities" }

// 重写 Entity 接口方法为值接收者
func (e ServiceTestEntity) GetID() int64            { return e.ID }
func (e ServiceTestEntity) SetID(id int64)          { e.ID = id }
func (e ServiceTestEntity) GetDeleted() int64       { return e.Deleted }
func (e ServiceTestEntity) SetDeleted(ts int64)     { e.Deleted = ts }
func (e ServiceTestEntity) GetCreatedAt() time.Time { return e.CreatedAt }
func (e ServiceTestEntity) SetCreatedAt(t time.Time) { e.CreatedAt = t }
func (e ServiceTestEntity) GetUpdatedAt() time.Time  { return e.UpdatedAt }
func (e ServiceTestEntity) SetUpdatedAt(t time.Time) { e.UpdatedAt = t }

// ServiceTestDTO 测试操作 DTO
type ServiceTestDTO struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// ServiceTestQueryDTO 测试查询 DTO
type ServiceTestQueryDTO struct {
	BaseQueryDTO
	Name string `form:"name"`
}

// ServiceTestResponse 测试响应
type ServiceTestResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func setupTestService(t *testing.T) (*Service[ServiceTestEntity, ServiceTestDTO, ServiceTestQueryDTO, ServiceTestResponse], *gorm.DB) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	assert.NoError(t, err)

	err = db.AutoMigrate(&ServiceTestEntity{})
	assert.NoError(t, err)

	cfg := ServiceConfig[ServiceTestEntity, ServiceTestDTO, ServiceTestQueryDTO, ServiceTestResponse]{
		DtoToEntity: func(dto *ServiceTestDTO) (*ServiceTestEntity, error) {
			entity := &ServiceTestEntity{}
			if dto.ID > 0 {
				entity.ID = dto.ID
			}
			if dto.Name != "" {
				entity.Name = dto.Name
			}
			return entity, nil
		},
		EntityToRes: func(entity *ServiceTestEntity) (ServiceTestResponse, error) {
			return ServiceTestResponse{
				ID:   entity.ID,
				Name: entity.Name,
			}, nil
		},
		QueryToCond: func(query ServiceTestQueryDTO) *QueryCondition {
			cond := NewQuery()
			if query.Name != "" {
				cond.WhereEq("name", query.Name)
			}
			cond.Page(query.GetPage()).PageSize(query.GetPageSize())
			return cond
		},
	}

	return NewService(NewRepository[ServiceTestEntity](db), cfg), db
}

func TestService_Create(t *testing.T) {
	svc, _ := setupTestService(t)
	ctx := context.Background()

	dto := &ServiceTestDTO{
		Name: "Test User",
	}

	result, err := svc.Create(ctx, dto)

	assert.NoError(t, err)
	assert.NotZero(t, result.ID)
	assert.Equal(t, "Test User", result.Name)
}

func TestService_GetByID(t *testing.T) {
	svc, _ := setupTestService(t)
	ctx := context.Background()

	// 先创建
	dto := &ServiceTestDTO{Name: "Test"}
	created, _ := svc.Create(ctx, dto)

	// 再查询
	result, err := svc.GetByID(ctx, created.ID)

	assert.NoError(t, err)
	assert.Equal(t, created.ID, result.ID)
	assert.Equal(t, "Test", result.Name)
}

func TestService_List(t *testing.T) {
	svc, _ := setupTestService(t)
	ctx := context.Background()

	// 创建测试数据
	svc.Create(ctx, &ServiceTestDTO{Name: "User1"})
	svc.Create(ctx, &ServiceTestDTO{Name: "User2"})

	query := ServiceTestQueryDTO{}
	results, err := svc.List(ctx, query)

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(results), 2)
}

func TestService_Page(t *testing.T) {
	svc, db := setupTestService(t)
	ctx := context.Background()

	// 清空可能存在的测试数据
	db.Exec("DELETE FROM service_test_entities")

	// 创建测试数据
	for i := 0; i < 15; i++ {
		svc.Create(ctx, &ServiceTestDTO{Name: "User"})
	}

	query := ServiceTestQueryDTO{
		BaseQueryDTO: BaseQueryDTO{Page: 1, PageSize: 10},
	}
	result, err := svc.Page(ctx, query)

	assert.NoError(t, err)
	assert.Equal(t, int64(15), result.Total)
	assert.Len(t, result.List, 10)
}
