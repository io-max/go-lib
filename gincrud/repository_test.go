package gincrud

import (
	"context"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestEntity 测试实体 - 重写所有 Entity 接口方法为值接收者
type TestEntity struct {
	BaseEntity
	Name string `gorm:"size:255" json:"name"`
}

// 重写 Entity 接口方法为值接收者（覆盖 BaseEntity 的指针接收者方法）
func (e TestEntity) GetID() int64 { return e.ID }
func (e TestEntity) SetID(id int64) { e.ID = id }
func (e TestEntity) GetDeleted() int64 { return e.Deleted }
func (e TestEntity) SetDeleted(ts int64) { e.Deleted = ts }
func (e TestEntity) GetCreatedAt() time.Time { return e.CreatedAt }
func (e TestEntity) SetCreatedAt(t time.Time) { e.CreatedAt = t }
func (e TestEntity) GetUpdatedAt() time.Time { return e.UpdatedAt }
func (e TestEntity) SetUpdatedAt(t time.Time) { e.UpdatedAt = t }

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	err = db.AutoMigrate(&TestEntity{})
	if err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	return db
}

func TestRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository[TestEntity](db)
	ctx := context.Background()

	t.Run("成功查询存在的记录", func(t *testing.T) {
		entity := TestEntity{Name: "test1"}
		if err := repo.Create(ctx, &entity); err != nil {
			t.Fatalf("Failed to create entity: %v", err)
		}

		result, err := repo.GetByID(ctx, entity.ID)
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}

		if result == nil {
			t.Fatal("Expected non-nil result")
		}
		if result.ID != entity.ID {
			t.Errorf("Expected ID %d, got %d", entity.ID, result.ID)
		}
		if result.Name != "test1" {
			t.Errorf("Expected Name 'test1', got '%s'", result.Name)
		}
	})

	t.Run("查询不存在的记录返回 ErrRecordNotFound", func(t *testing.T) {
		_, err := repo.GetByID(ctx, 99999)
		if err == nil {
			t.Fatal("Expected error for non-existent record")
		}
		if err != ErrRecordNotFound {
			t.Errorf("Expected ErrRecordNotFound, got %v", err)
		}
	})

	t.Run("查询已删除的记录返回 ErrRecordNotFound", func(t *testing.T) {
		entity := TestEntity{Name: "to_delete"}
		if err := repo.Create(ctx, &entity); err != nil {
			t.Fatalf("Failed to create entity: %v", err)
		}
		if err := repo.Delete(ctx, entity.ID); err != nil {
			t.Fatalf("Failed to delete entity: %v", err)
		}

		_, err := repo.GetByID(ctx, entity.ID)
		if err != ErrRecordNotFound {
			t.Errorf("Expected ErrRecordNotFound for deleted record, got %v", err)
		}
	})
}

func TestRepository_List(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository[TestEntity](db)
	ctx := context.Background()

	for i := 1; i <= 15; i++ {
		entity := TestEntity{Name: "test"}
		if err := repo.Create(ctx, &entity); err != nil {
			t.Fatalf("Failed to create entity %d: %v", i, err)
		}
	}

	t.Run("分页查询第一页", func(t *testing.T) {
		dto := &BaseQueryDTO{Page: 1, PageSize: 5}
		dto.Normalize()

		list, total, err := repo.List(ctx, nil, dto)
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}

		if total != 15 {
			t.Errorf("Expected total 15, got %d", total)
		}
		if len(list) != 5 {
			t.Errorf("Expected 5 items, got %d", len(list))
		}
	})

	t.Run("分页查询第二页", func(t *testing.T) {
		dto := &BaseQueryDTO{Page: 2, PageSize: 5}
		dto.Normalize()

		list, total, err := repo.List(ctx, nil, dto)
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}

		if total != 15 {
			t.Errorf("Expected total 15, got %d", total)
		}
		if len(list) != 5 {
			t.Errorf("Expected 5 items, got %d", len(list))
		}
	})

	t.Run("查询所有数据", func(t *testing.T) {
		dto := &BaseQueryDTO{Page: 1, PageSize: 100}
		dto.Normalize()

		list, total, err := repo.List(ctx, nil, dto)
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}

		if total != 15 {
			t.Errorf("Expected total 15, got %d", total)
		}
		if len(list) != 15 {
			t.Errorf("Expected 15 items, got %d", len(list))
		}
	})

	t.Run("带条件查询", func(t *testing.T) {
		special := TestEntity{Name: "special"}
		if err := repo.Create(ctx, &special); err != nil {
			t.Fatalf("Failed to create special entity: %v", err)
		}

		cond := NewQuery().WhereEq("name", "special")
		dto := &BaseQueryDTO{Page: 1, PageSize: 10}
		dto.Normalize()

		list, total, err := repo.List(ctx, cond, dto)
		if err != nil {
			t.Fatalf("List with condition failed: %v", err)
		}

		if total != 1 {
			t.Errorf("Expected total 1, got %d", total)
		}
		if len(list) != 1 {
			t.Errorf("Expected 1 item, got %d", len(list))
		}
		if list[0].Name != "special" {
			t.Errorf("Expected name 'special', got '%s'", list[0].Name)
		}
	})

	t.Run("不包含已删除的记录", func(t *testing.T) {
		dto := &BaseQueryDTO{Page: 1, PageSize: 100}
		dto.Normalize()

		_, totalBefore, err := repo.List(ctx, nil, dto)
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}

		all, _, _ := repo.List(ctx, nil, dto)
		firstID := all[0].ID
		if err := repo.Delete(ctx, firstID); err != nil {
			t.Fatalf("Failed to delete entity: %v", err)
		}

		_, totalAfter, err := repo.List(ctx, nil, dto)
		if err != nil {
			t.Fatalf("List after delete failed: %v", err)
		}

		if totalAfter != totalBefore-1 {
			t.Errorf("Expected total %d after delete, got %d", totalBefore-1, totalAfter)
		}
	})
}

func TestRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository[TestEntity](db)
	ctx := context.Background()

	t.Run("成功创建记录", func(t *testing.T) {
		entity := &TestEntity{Name: "test_create"}

		err := repo.Create(ctx, entity)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		// 验证 ID 被设置
		if entity.ID == 0 {
			t.Error("Expected ID to be set")
		}

		// 验证时间戳被设置
		if entity.CreatedAt.IsZero() {
			t.Error("Expected CreatedAt to be set")
		}
		if entity.UpdatedAt.IsZero() {
			t.Error("Expected UpdatedAt to be set")
		}

		// 验证 deleted 被设置为 0
		if entity.Deleted != 0 {
			t.Errorf("Expected Deleted to be 0, got %d", entity.Deleted)
		}

		// 验证记录可以在数据库中查询到
		result, err := repo.GetByID(ctx, entity.ID)
		if err != nil {
			t.Fatalf("Failed to get created entity: %v", err)
		}
		if result.Name != "test_create" {
			t.Errorf("Expected name 'test_create', got '%s'", result.Name)
		}
	})

	t.Run("创建记录时自动设置时间戳", func(t *testing.T) {
		entity := &TestEntity{Name: "test_timestamp"}
		beforeCreate := time.Now()

		err := repo.Create(ctx, entity)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		afterCreate := time.Now()

		// 验证 CreatedAt 和 UpdatedAt 在合理范围内
		if entity.CreatedAt.Before(beforeCreate) || entity.CreatedAt.After(afterCreate) {
			t.Errorf("CreatedAt %v is not within expected range [%v, %v]", entity.CreatedAt, beforeCreate, afterCreate)
		}
		if entity.UpdatedAt.Before(beforeCreate) || entity.UpdatedAt.After(afterCreate) {
			t.Errorf("UpdatedAt %v is not within expected range [%v, %v]", entity.UpdatedAt, beforeCreate, afterCreate)
		}

		// 验证 CreatedAt 和 UpdatedAt 相等
		if entity.CreatedAt.Unix() != entity.UpdatedAt.Unix() {
			t.Error("Expected CreatedAt and UpdatedAt to be equal on create")
		}
	})
}

func TestRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository[TestEntity](db)
	ctx := context.Background()

	t.Run("成功更新记录", func(t *testing.T) {
		// 先创建记录
		entity := &TestEntity{Name: "original"}
		if err := repo.Create(ctx, entity); err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		oldUpdatedAt := entity.UpdatedAt
		time.Sleep(10 * time.Millisecond) // 确保时间戳有差异

		// 更新记录
		entity.Name = "updated"
		err := repo.Update(ctx, entity)
		if err != nil {
			t.Fatalf("Update failed: %v", err)
		}

		// 验证 UpdatedAt 被更新
		if !entity.UpdatedAt.After(oldUpdatedAt) {
			t.Errorf("Expected UpdatedAt to be updated, old: %v, new: %v", oldUpdatedAt, entity.UpdatedAt)
		}

		// 验证记录已更新
		result, err := repo.GetByID(ctx, entity.ID)
		if err != nil {
			t.Fatalf("Failed to get updated entity: %v", err)
		}
		if result.Name != "updated" {
			t.Errorf("Expected name 'updated', got '%s'", result.Name)
		}
	})

	t.Run("更新记录时 ID 必须存在", func(t *testing.T) {
		entity := &TestEntity{Name: "no_id"}
		entity.SetID(0) // 明确设置 ID 为 0

		err := repo.Update(ctx, entity)
		if err == nil {
			t.Fatal("Expected error for missing ID")
		}
		if err != ErrEntityIDRequired {
			t.Errorf("Expected ErrEntityIDRequired, got %v", err)
		}
	})

	t.Run("更新不存在的 ID 返回错误", func(t *testing.T) {
		entity := &TestEntity{Name: "non_existent"}
		entity.SetID(99999)

		err := repo.Update(ctx, entity)
		if err == nil {
			t.Fatal("Expected error for non-existent ID")
		}
	})
}

func TestRepository_ApplyCondition(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository[TestEntity](db)
	ctx := context.Background()

	names := []string{"Alice", "Bob", "Charlie", "David"}
	for _, name := range names {
		entity := TestEntity{Name: name}
		if err := repo.Create(ctx, &entity); err != nil {
			t.Fatalf("Failed to create entity: %v", err)
		}
	}

	t.Run("等于条件", func(t *testing.T) {
		cond := NewQuery().WhereEq("name", "Bob")
		dto := &BaseQueryDTO{Page: 1, PageSize: 10}
		dto.Normalize()

		list, total, err := repo.List(ctx, cond, dto)
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}

		if total != 1 {
			t.Errorf("Expected total 1, got %d", total)
		}
		if len(list) != 1 || list[0].Name != "Bob" {
			t.Errorf("Expected Bob, got %v", list)
		}
	})

	t.Run("排序条件", func(t *testing.T) {
		cond := NewQuery().OrderBy("id", true)
		dto := &BaseQueryDTO{Page: 1, PageSize: 2}
		dto.Normalize()

		list, total, err := repo.List(ctx, cond, dto)
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}

		if total != 4 {
			t.Errorf("Expected total 4, got %d", total)
		}
		if list[0].ID < list[1].ID {
			t.Error("Expected descending order by ID")
		}
	})
}
