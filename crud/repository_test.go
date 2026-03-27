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

func TestRepository_FindByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository[TestEntity](db)
	ctx := context.Background()

	t.Run("成功查询存在的记录", func(t *testing.T) {
		entity := TestEntity{Name: "test1"}
		if err := repo.Create(ctx, &entity); err != nil {
			t.Fatalf("Failed to create entity: %v", err)
		}

		result, err := repo.FindByID(ctx, entity.ID)
		if err != nil {
			t.Fatalf("FindByID failed: %v", err)
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
		_, err := repo.FindByID(ctx, 99999)
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

		_, err := repo.FindByID(ctx, entity.ID)
		if err != ErrRecordNotFound {
			t.Errorf("Expected ErrRecordNotFound for deleted record, got %v", err)
		}
	})
}

func TestRepository_FindPage(t *testing.T) {
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
		cond := NewQuery().Page(1).PageSize(5)
		list, total, err := repo.FindPage(ctx, cond)
		if err != nil {
			t.Fatalf("FindPage failed: %v", err)
		}

		if total != 15 {
			t.Errorf("Expected total 15, got %d", total)
		}
		if len(list) != 5 {
			t.Errorf("Expected 5 items, got %d", len(list))
		}
	})

	t.Run("分页查询第二页", func(t *testing.T) {
		cond := NewQuery().Page(2).PageSize(5)
		list, total, err := repo.FindPage(ctx, cond)
		if err != nil {
			t.Fatalf("FindPage failed: %v", err)
		}

		if total != 15 {
			t.Errorf("Expected total 15, got %d", total)
		}
		if len(list) != 5 {
			t.Errorf("Expected 5 items, got %d", len(list))
		}
	})

	t.Run("查询所有数据", func(t *testing.T) {
		cond := NewQuery().Page(1).PageSize(100)
		list, total, err := repo.FindPage(ctx, cond)
		if err != nil {
			t.Fatalf("FindPage failed: %v", err)
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

		cond := NewQuery().WhereEq("name", "special").Page(1).PageSize(10)
		list, total, err := repo.FindPage(ctx, cond)
		if err != nil {
			t.Fatalf("FindPage with condition failed: %v", err)
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
		cond := NewQuery().Page(1).PageSize(100)
		_, totalBefore, err := repo.FindPage(ctx, cond)
		if err != nil {
			t.Fatalf("FindPage failed: %v", err)
		}

		all, _, _ := repo.FindPage(ctx, cond)
		firstID := all[0].ID
		if err := repo.Delete(ctx, firstID); err != nil {
			t.Fatalf("Failed to delete entity: %v", err)
		}

		_, totalAfter, err := repo.FindPage(ctx, cond)
		if err != nil {
			t.Fatalf("FindPage after delete failed: %v", err)
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
		result, err := repo.FindByID(ctx, entity.ID)
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
		result, err := repo.FindByID(ctx, entity.ID)
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

func TestRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository[TestEntity](db)
	ctx := context.Background()

	t.Run("成功软删除记录", func(t *testing.T) {
		entity := TestEntity{Name: "to_delete"}
		if err := repo.Create(ctx, &entity); err != nil {
			t.Fatalf("Failed to create entity: %v", err)
		}

		beforeDelete := time.Now().Unix()
		err := repo.Delete(ctx, entity.ID)
		afterDelete := time.Now().Unix()

		if err != nil {
			t.Fatalf("Delete failed: %v", err)
		}

		// 验证软删除后无法通过 FindByID 查询到
		_, err = repo.FindByID(ctx, entity.ID)
		if err != ErrRecordNotFound {
			t.Errorf("Expected ErrRecordNotFound after soft delete, got %v", err)
		}

		// 验证数据库中记录仍存在但 deleted 字段已设置
		var dbEntity TestEntity
		if err := db.Where("id = ?", entity.ID).First(&dbEntity).Error; err != nil {
			t.Fatalf("Failed to find entity in DB: %v", err)
		}
		if dbEntity.Deleted == 0 {
			t.Error("Expected Deleted to be set after soft delete")
		}
		if dbEntity.Deleted < beforeDelete || dbEntity.Deleted > afterDelete {
			t.Errorf("Expected Deleted timestamp %d to be within [%d, %d]", dbEntity.Deleted, beforeDelete, afterDelete)
		}
	})

	t.Run("删除不存在的记录返回错误", func(t *testing.T) {
		err := repo.Delete(ctx, 99999)
		if err == nil {
			t.Fatal("Expected error for non-existent record")
		}
	})

	t.Run("删除已删除的记录返回错误", func(t *testing.T) {
		entity := TestEntity{Name: "delete_twice"}
		if err := repo.Create(ctx, &entity); err != nil {
			t.Fatalf("Failed to create entity: %v", err)
		}
		if err := repo.Delete(ctx, entity.ID); err != nil {
			t.Fatalf("Failed to delete entity: %v", err)
		}

		// 再次删除应返回错误（因为记录已被软删除过滤）
		err := repo.Delete(ctx, entity.ID)
		if err == nil {
			t.Fatal("Expected error for already deleted record")
		}
	})
}

func TestRepository_TrulyDelete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository[TestEntity](db)
	ctx := context.Background()

	t.Run("成功硬删除记录", func(t *testing.T) {
		entity := TestEntity{Name: "to_truly_delete"}
		if err := repo.Create(ctx, &entity); err != nil {
			t.Fatalf("Failed to create entity: %v", err)
		}

		entityID := entity.ID

		err := repo.TrulyDelete(ctx, entityID)
		if err != nil {
			t.Fatalf("TrulyDelete failed: %v", err)
		}

		// 验证记录在数据库中完全消失
		var dbEntity TestEntity
		err = db.Where("id = ?", entityID).First(&dbEntity).Error
		if err != gorm.ErrRecordNotFound {
			t.Errorf("Expected gorm.ErrRecordNotFound, got %v", err)
		}

		// 验证 GetByID 也返回 NotFound
		_, err = repo.FindByID(ctx, entityID)
		if err != ErrRecordNotFound {
			t.Errorf("Expected ErrRecordNotFound, got %v", err)
		}
	})

	t.Run("删除不存在的记录不返回错误", func(t *testing.T) {
		// GORM 的 Delete 操作在记录不存在时不会返回错误
		err := repo.TrulyDelete(ctx, 99999)
		if err != nil {
			t.Errorf("Expected no error for non-existent record, got %v", err)
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
		cond := NewQuery().WhereEq("name", "Bob").Page(1).PageSize(10)
		list, total, err := repo.FindPage(ctx, cond)
		if err != nil {
			t.Fatalf("FindPage failed: %v", err)
		}

		if total != 1 {
			t.Errorf("Expected total 1, got %d", total)
		}
		if len(list) != 1 || list[0].Name != "Bob" {
			t.Errorf("Expected Bob, got %v", list)
		}
	})

	t.Run("排序条件", func(t *testing.T) {
		cond := NewQuery().OrderBy("id", true).Page(1).PageSize(2)
		list, total, err := repo.FindPage(ctx, cond)
		if err != nil {
			t.Fatalf("FindPage failed: %v", err)
		}

		if total != 4 {
			t.Errorf("Expected total 4, got %d", total)
		}
		if list[0].ID < list[1].ID {
			t.Error("Expected descending order by ID")
		}
	})
}

func TestRepository_BatchCreate(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository[TestEntity](db)
	ctx := context.Background()

	t.Run("成功批量创建记录", func(t *testing.T) {
		entities := []*TestEntity{
			{Name: "batch1"},
			{Name: "batch2"},
			{Name: "batch3"},
		}

		err := repo.BatchCreate(ctx, entities)
		if err != nil {
			t.Fatalf("BatchCreate failed: %v", err)
		}

		// 验证所有记录都有 ID
		for i, e := range entities {
			if e.ID == 0 {
				t.Errorf("Entity %d: Expected ID to be set", i)
			}
		}

		// 验证时间戳被设置
		for i, e := range entities {
			if e.CreatedAt.IsZero() {
				t.Errorf("Entity %d: Expected CreatedAt to be set", i)
			}
			if e.UpdatedAt.IsZero() {
				t.Errorf("Entity %d: Expected UpdatedAt to be set", i)
			}
		}

		// 验证 deleted 被设置为 0
		for i, e := range entities {
			if e.Deleted != 0 {
				t.Errorf("Entity %d: Expected Deleted to be 0, got %d", i, e.Deleted)
			}
		}

		// 验证所有记录可以在数据库中查询到
		for _, e := range entities {
			result, err := repo.FindByID(ctx, e.ID)
			if err != nil {
				t.Fatalf("Failed to get created entity %d: %v", e.ID, err)
			}
			if result.Name == "" {
				t.Errorf("Entity %d: Expected non-empty name", e.ID)
			}
		}
	})

	t.Run("批量创建时自动设置相同的时间戳", func(t *testing.T) {
		entities := []*TestEntity{
			{Name: "time1"},
			{Name: "time2"},
		}

		beforeCreate := time.Now()
		err := repo.BatchCreate(ctx, entities)
		afterCreate := time.Now()

		if err != nil {
			t.Fatalf("BatchCreate failed: %v", err)
		}

		// 验证所有实体的时间戳在合理范围内
		for i, e := range entities {
			if e.CreatedAt.Before(beforeCreate) || e.CreatedAt.After(afterCreate) {
				t.Errorf("Entity %d: CreatedAt %v is not within expected range", i, e.CreatedAt)
			}
			// 验证 CreatedAt 和 UpdatedAt 相等
			if e.CreatedAt.Unix() != e.UpdatedAt.Unix() {
				t.Errorf("Entity %d: Expected CreatedAt and UpdatedAt to be equal", i)
			}
		}
	})
}

func TestRepository_BatchUpdate(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository[TestEntity](db)
	ctx := context.Background()

	t.Run("成功批量更新记录", func(t *testing.T) {
		// 先创建记录
		entities := []*TestEntity{
			{Name: "original1"},
			{Name: "original2"},
			{Name: "original3"},
		}
		if err := repo.BatchCreate(ctx, entities); err != nil {
			t.Fatalf("BatchCreate failed: %v", err)
		}

		// 收集 ID
		ids := []int64{entities[0].ID, entities[1].ID, entities[2].ID}

		// 批量更新
		updates := map[string]any{"name": "updated"}
		err := repo.BatchUpdate(ctx, ids, updates)
		if err != nil {
			t.Fatalf("BatchUpdate failed: %v", err)
		}

		// 验证所有记录已更新
		for _, e := range entities {
			result, err := repo.FindByID(ctx, e.ID)
			if err != nil {
				t.Fatalf("Failed to get updated entity %d: %v", e.ID, err)
			}
			if result.Name != "updated" {
				t.Errorf("Entity %d: Expected name 'updated', got '%s'", e.ID, result.Name)
			}
		}
	})

	t.Run("批量更新部分记录", func(t *testing.T) {
		// 先创建记录
		entities := []*TestEntity{
			{Name: "keep1"},
			{Name: "keep2"},
			{Name: "update_me"},
		}
		if err := repo.BatchCreate(ctx, entities); err != nil {
			t.Fatalf("BatchCreate failed: %v", err)
		}

		// 只更新第三条记录
		ids := []int64{entities[2].ID}
		updates := map[string]any{"name": "updated_partial"}
		err := repo.BatchUpdate(ctx, ids, updates)
		if err != nil {
			t.Fatalf("BatchUpdate failed: %v", err)
		}

		// 验证第一条记录未变
		result1, _ := repo.FindByID(ctx, entities[0].ID)
		if result1.Name != "keep1" {
			t.Errorf("Entity 1: Expected 'keep1', got '%s'", result1.Name)
		}

		// 验证第三条记录已更新
		result3, _ := repo.FindByID(ctx, entities[2].ID)
		if result3.Name != "updated_partial" {
			t.Errorf("Entity 3: Expected 'updated_partial', got '%s'", result3.Name)
		}
	})
}

func TestRepository_BatchDelete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository[TestEntity](db)
	ctx := context.Background()

	t.Run("成功批量软删除记录", func(t *testing.T) {
		// 先创建记录
		entities := []*TestEntity{
			{Name: "delete1"},
			{Name: "delete2"},
			{Name: "delete3"},
		}
		if err := repo.BatchCreate(ctx, entities); err != nil {
			t.Fatalf("BatchCreate failed: %v", err)
		}

		// 收集 ID
		ids := []int64{entities[0].ID, entities[1].ID, entities[2].ID}

		beforeDelete := time.Now().Unix()
		err := repo.BatchDelete(ctx, ids)
		afterDelete := time.Now().Unix()

		if err != nil {
			t.Fatalf("BatchDelete failed: %v", err)
		}

		// 验证软删除后无法通过 FindByID 查询到
		for _, e := range entities {
			_, err = repo.FindByID(ctx, e.ID)
			if err != ErrRecordNotFound {
				t.Errorf("Expected ErrRecordNotFound after soft delete for entity %d", e.ID)
			}
		}

		// 验证数据库中记录仍存在但 deleted 字段已设置
		for _, e := range entities {
			var dbEntity TestEntity
			if err := db.Where("id = ?", e.ID).First(&dbEntity).Error; err != nil {
				t.Fatalf("Failed to find entity %d in DB: %v", e.ID, err)
			}
			if dbEntity.Deleted == 0 {
				t.Errorf("Entity %d: Expected Deleted to be set after soft delete", e.ID)
			}
			if dbEntity.Deleted < beforeDelete || dbEntity.Deleted > afterDelete {
				t.Errorf("Entity %d: Expected Deleted timestamp %d to be within [%d, %d]", e.ID, dbEntity.Deleted, beforeDelete, afterDelete)
			}
		}
	})

	t.Run("批量删除部分记录", func(t *testing.T) {
		// 先创建记录
		entities := []*TestEntity{
			{Name: "keep1"},
			{Name: "delete_me"},
			{Name: "keep2"},
		}
		if err := repo.BatchCreate(ctx, entities); err != nil {
			t.Fatalf("BatchCreate failed: %v", err)
		}

		// 只删除第二条记录
		ids := []int64{entities[1].ID}
		err := repo.BatchDelete(ctx, ids)
		if err != nil {
			t.Fatalf("BatchDelete failed: %v", err)
		}

		// 验证第一条记录仍存在
		result1, err := repo.FindByID(ctx, entities[0].ID)
		if err != nil {
			t.Errorf("Entity 1 should still exist: %v", err)
		}
		if result1.Name != "keep1" {
			t.Errorf("Entity 1: Expected 'keep1', got '%s'", result1.Name)
		}

		// 验证第二条记录已被删除
		_, err = repo.FindByID(ctx, entities[1].ID)
		if err != ErrRecordNotFound {
			t.Error("Entity 2 should be deleted")
		}
	})
}

func TestRepository_Find(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository[TestEntity](db)
	ctx := context.Background()

	// 创建测试数据
	for i := 1; i <= 5; i++ {
		entity := TestEntity{Name: "test"}
		if err := repo.Create(ctx, &entity); err != nil {
			t.Fatalf("Failed to create entity %d: %v", i, err)
		}
	}

	t.Run("查询所有记录", func(t *testing.T) {
		list, err := repo.Find(ctx, nil)
		if err != nil {
			t.Fatalf("Find failed: %v", err)
		}
		if len(list) != 5 {
			t.Errorf("Expected 5 records, got %d", len(list))
		}
	})

	t.Run("带条件查询", func(t *testing.T) {
		// 创建一个特殊记录
		special := TestEntity{Name: "special"}
		if err := repo.Create(ctx, &special); err != nil {
			t.Fatalf("Failed to create special entity: %v", err)
		}

		cond := NewQuery().WhereEq("name", "special")
		list, err := repo.Find(ctx, cond)
		if err != nil {
			t.Fatalf("Find with condition failed: %v", err)
		}
		if len(list) != 1 {
			t.Errorf("Expected 1 record, got %d", len(list))
		}
		if list[0].Name != "special" {
			t.Errorf("Expected name 'special', got '%s'", list[0].Name)
		}
	})

	t.Run("不包含已删除的记录", func(t *testing.T) {
		// 先查询所有记录
		all, err := repo.Find(ctx, nil)
		if err != nil {
			t.Fatalf("Find failed: %v", err)
		}
		totalBefore := len(all)

		// 删除第一条记录
		if err := repo.Delete(ctx, all[0].ID); err != nil {
			t.Fatalf("Failed to delete entity: %v", err)
		}

		// 再次查询
		after, err := repo.Find(ctx, nil)
		if err != nil {
			t.Fatalf("Find after delete failed: %v", err)
		}

		if len(after) != totalBefore-1 {
			t.Errorf("Expected %d records after delete, got %d", totalBefore-1, len(after))
		}
	})

	t.Run("带排序条件", func(t *testing.T) {
		cond := NewQuery().OrderBy("id", true) // DESC
		list, err := repo.Find(ctx, cond)
		if err != nil {
			t.Fatalf("Find with order failed: %v", err)
		}
		// 验证是降序
		for i := 0; i < len(list)-1; i++ {
			if list[i].ID < list[i+1].ID {
				t.Error("Expected descending order")
				break
			}
		}
	})
}

func TestRepository_FindFirst(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository[TestEntity](db)
	ctx := context.Background()

	t.Run("查询第一条记录", func(t *testing.T) {
		// 创建测试数据
		for i := 1; i <= 3; i++ {
			entity := TestEntity{Name: "test"}
			if err := repo.Create(ctx, &entity); err != nil {
				t.Fatalf("Failed to create entity %d: %v", i, err)
			}
		}

		result, err := repo.FindFirst(ctx, nil)
		if err != nil {
			t.Fatalf("FindFirst failed: %v", err)
		}
		if result == nil {
			t.Fatal("Expected non-nil result")
		}
	})

	t.Run("带条件查询第一条", func(t *testing.T) {
		special := TestEntity{Name: "first_special"}
		if err := repo.Create(ctx, &special); err != nil {
			t.Fatalf("Failed to create special entity: %v", err)
		}

		cond := NewQuery().WhereEq("name", "first_special")
		result, err := repo.FindFirst(ctx, cond)
		if err != nil {
			t.Fatalf("FindFirst with condition failed: %v", err)
		}
		if result == nil {
			t.Fatal("Expected non-nil result")
		}
		if result.Name != "first_special" {
			t.Errorf("Expected name 'first_special', got '%s'", result.Name)
		}
	})

	t.Run("查询不存在的记录返回 ErrRecordNotFound", func(t *testing.T) {
		cond := NewQuery().WhereEq("name", "nonexistent")
		result, err := repo.FindFirst(ctx, cond)
		if err == nil {
			t.Fatal("Expected error for non-existent record")
		}
		if result != nil {
			t.Error("Expected nil result")
		}
		// FindFirst 在记录不存在时返回 ErrRecordNotFound
	})

	t.Run("不包含已删除的记录", func(t *testing.T) {
		// 查询所有记录
		all, _ := repo.Find(ctx, nil)
		if len(all) == 0 {
			t.Fatal("Need at least one record for this test")
		}

		// 删除第一条
		if err := repo.Delete(ctx, all[0].ID); err != nil {
			t.Fatalf("Failed to delete entity: %v", err)
		}

		// 尝试用 ID 条件查询已删除的记录
		cond := NewQuery().WhereEq("id", all[0].ID)
		result, err := repo.FindFirst(ctx, cond)
		if err == nil {
			t.Error("Expected error for deleted record")
		}
		if result != nil {
			t.Error("Expected nil result for deleted record")
		}
	})
}

func TestRepository_Count(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository[TestEntity](db)
	ctx := context.Background()

	// 创建测试数据
	for i := 1; i <= 5; i++ {
		entity := TestEntity{Name: "count_test"}
		if err := repo.Create(ctx, &entity); err != nil {
			t.Fatalf("Failed to create entity %d: %v", i, err)
		}
	}

	t.Run("统计所有记录", func(t *testing.T) {
		count, err := repo.Count(ctx, nil)
		if err != nil {
			t.Fatalf("Count failed: %v", err)
		}
		if count != 5 {
			t.Errorf("Expected count 5, got %d", count)
		}
	})

	t.Run("带条件统计", func(t *testing.T) {
		special := TestEntity{Name: "count_special"}
		if err := repo.Create(ctx, &special); err != nil {
			t.Fatalf("Failed to create special entity: %v", err)
		}

		cond := NewQuery().WhereEq("name", "count_special")
		count, err := repo.Count(ctx, cond)
		if err != nil {
			t.Fatalf("Count with condition failed: %v", err)
		}
		if count != 1 {
			t.Errorf("Expected count 1, got %d", count)
		}
	})

	t.Run("不包含已删除的记录", func(t *testing.T) {
		// 统计删除前的数量
		countBefore, err := repo.Count(ctx, nil)
		if err != nil {
			t.Fatalf("Count failed: %v", err)
		}

		// 查询并删除一条记录
		all, _ := repo.Find(ctx, nil)
		if err := repo.Delete(ctx, all[0].ID); err != nil {
			t.Fatalf("Failed to delete entity: %v", err)
		}

		// 统计删除后的数量
		countAfter, err := repo.Count(ctx, nil)
		if err != nil {
			t.Fatalf("Count after delete failed: %v", err)
		}

		if countAfter != countBefore-1 {
			t.Errorf("Expected count %d after delete, got %d", countBefore-1, countAfter)
		}
	})

	t.Run("空表统计返回 0", func(t *testing.T) {
		// 创建一个新表
		newDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		if err != nil {
			t.Fatalf("Failed to open new database: %v", err)
		}
		if err := newDB.AutoMigrate(&TestEntity{}); err != nil {
			t.Fatalf("Failed to migrate: %v", err)
		}
		newRepo := NewRepository[TestEntity](newDB)

		count, err := newRepo.Count(ctx, nil)
		if err != nil {
			t.Fatalf("Count on empty table failed: %v", err)
		}
		if count != 0 {
			t.Errorf("Expected count 0, got %d", count)
		}
	})
}

func TestRepository_Exists(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository[TestEntity](db)
	ctx := context.Background()

	t.Run("存在的记录返回 true", func(t *testing.T) {
		entity := TestEntity{Name: "exists_test"}
		if err := repo.Create(ctx, &entity); err != nil {
			t.Fatalf("Failed to create entity: %v", err)
		}

		exists, err := repo.Exists(ctx, entity.ID)
		if err != nil {
			t.Fatalf("Exists failed: %v", err)
		}
		if !exists {
			t.Error("Expected true for existing record")
		}
	})

	t.Run("不存在的记录返回 false", func(t *testing.T) {
		exists, err := repo.Exists(ctx, 99999)
		if err != nil {
			t.Fatalf("Exists failed: %v", err)
		}
		if exists {
			t.Error("Expected false for non-existent record")
		}
	})

	t.Run("已删除的记录返回 false", func(t *testing.T) {
		entity := TestEntity{Name: "to_delete"}
		if err := repo.Create(ctx, &entity); err != nil {
			t.Fatalf("Failed to create entity: %v", err)
		}

		// 删除前存在
		existsBefore, err := repo.Exists(ctx, entity.ID)
		if err != nil {
			t.Fatalf("Exists before delete failed: %v", err)
		}
		if !existsBefore {
			t.Error("Expected true before delete")
		}

		// 删除后不存在
		if err := repo.Delete(ctx, entity.ID); err != nil {
			t.Fatalf("Failed to delete entity: %v", err)
		}

		existsAfter, err := repo.Exists(ctx, entity.ID)
		if err != nil {
			t.Fatalf("Exists after delete failed: %v", err)
		}
		if existsAfter {
			t.Error("Expected false for deleted record")
		}
	})
}

func TestRepository_FindByIDs(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository[TestEntity](db)
	ctx := context.Background()

	// 创建测试数据
	entities := []*TestEntity{
		{Name: "entity1"},
		{Name: "entity2"},
		{Name: "entity3"},
	}
	if err := repo.BatchCreate(ctx, entities); err != nil {
		t.Fatalf("BatchCreate failed: %v", err)
	}

	t.Run("成功批量查询存在的记录", func(t *testing.T) {
		ids := []int64{entities[0].ID, entities[1].ID, entities[2].ID}
		results, err := repo.FindByIDs(ctx, ids)
		if err != nil {
			t.Fatalf("FindByIDs failed: %v", err)
		}

		if len(results) != 3 {
			t.Errorf("Expected 3 results, got %d", len(results))
		}

		// 验证返回的结果
		names := make(map[string]bool)
		for _, r := range results {
			names[r.Name] = true
		}
		if !names["entity1"] || !names["entity2"] || !names["entity3"] {
			t.Errorf("Expected all entities to be returned")
		}
	})

	t.Run("查询部分存在的记录", func(t *testing.T) {
		ids := []int64{entities[0].ID, 99999}
		results, err := repo.FindByIDs(ctx, ids)
		if err != nil {
			t.Fatalf("FindByIDs failed: %v", err)
		}

		if len(results) != 1 {
			t.Errorf("Expected 1 result, got %d", len(results))
		}
		if results[0].ID != entities[0].ID {
			t.Errorf("Expected entity ID %d, got %d", entities[0].ID, results[0].ID)
		}
	})

	t.Run("查询空 ID 列表", func(t *testing.T) {
		ids := []int64{}
		results, err := repo.FindByIDs(ctx, ids)
		if err != nil {
			t.Fatalf("FindByIDs failed: %v", err)
		}

		if len(results) != 0 {
			t.Errorf("Expected 0 results for empty ID list, got %d", len(results))
		}
	})

	t.Run("不包含已删除的记录", func(t *testing.T) {
		// 删除第一条记录
		if err := repo.Delete(ctx, entities[0].ID); err != nil {
			t.Fatalf("Delete failed: %v", err)
		}

		// 查询所有 ID
		ids := []int64{entities[0].ID, entities[1].ID}
		results, err := repo.FindByIDs(ctx, ids)
		if err != nil {
			t.Fatalf("FindByIDs failed: %v", err)
		}

		// 只应返回未删除的记录
		if len(results) != 1 {
			t.Errorf("Expected 1 result (excluding deleted), got %d", len(results))
		}
		if results[0].ID != entities[1].ID {
			t.Errorf("Expected entity ID %d, got %d", entities[1].ID, results[0].ID)
		}
	})
}

func TestRepository_DeleteByIDs(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository[TestEntity](db)
	ctx := context.Background()

	t.Run("成功批量软删除记录", func(t *testing.T) {
		// 创建测试数据
		entities := []*TestEntity{
			{Name: "to_delete1"},
			{Name: "to_delete2"},
			{Name: "to_delete3"},
		}
		if err := repo.BatchCreate(ctx, entities); err != nil {
			t.Fatalf("BatchCreate failed: %v", err)
		}

		ids := []int64{entities[0].ID, entities[1].ID, entities[2].ID}

		beforeDelete := time.Now().Unix()
		err := repo.DeleteByIDs(ctx, ids)
		afterDelete := time.Now().Unix()

		if err != nil {
			t.Fatalf("DeleteByIDs failed: %v", err)
		}

		// 验证软删除后无法通过 FindByID 查询到
		for _, e := range entities {
			_, err = repo.FindByID(ctx, e.ID)
			if err != ErrRecordNotFound {
				t.Errorf("Expected ErrRecordNotFound after soft delete for entity %d", e.ID)
			}
		}

		// 验证数据库中记录仍存在但 deleted 字段已设置
		for _, e := range entities {
			var dbEntity TestEntity
			if err := db.Where("id = ?", e.ID).First(&dbEntity).Error; err != nil {
				t.Fatalf("Failed to find entity %d in DB: %v", e.ID, err)
			}
			if dbEntity.Deleted == 0 {
				t.Errorf("Entity %d: Expected Deleted to be set after soft delete", e.ID)
			}
			if dbEntity.Deleted < beforeDelete || dbEntity.Deleted > afterDelete {
				t.Errorf("Entity %d: Expected Deleted timestamp %d to be within [%d, %d]", e.ID, dbEntity.Deleted, beforeDelete, afterDelete)
			}
		}
	})

	t.Run("批量删除部分记录", func(t *testing.T) {
		// 创建测试数据
		entities := []*TestEntity{
			{Name: "keep1"},
			{Name: "delete_me"},
			{Name: "keep2"},
		}
		if err := repo.BatchCreate(ctx, entities); err != nil {
			t.Fatalf("BatchCreate failed: %v", err)
		}

		// 只删除第二条记录
		ids := []int64{entities[1].ID}
		err := repo.DeleteByIDs(ctx, ids)
		if err != nil {
			t.Fatalf("DeleteByIDs failed: %v", err)
		}

		// 验证第一条记录仍存在
		result1, err := repo.FindByID(ctx, entities[0].ID)
		if err != nil {
			t.Errorf("Entity 1 should still exist: %v", err)
		}
		if result1.Name != "keep1" {
			t.Errorf("Entity 1: Expected 'keep1', got '%s'", result1.Name)
		}

		// 验证第二条记录已被删除
		_, err = repo.FindByID(ctx, entities[1].ID)
		if err != ErrRecordNotFound {
			t.Error("Entity 2 should be deleted")
		}
	})

	t.Run("删除空 ID 列表不报错", func(t *testing.T) {
		ids := []int64{}
		err := repo.DeleteByIDs(ctx, ids)
		if err != nil {
			t.Errorf("DeleteByIDs with empty ID list should not return error: %v", err)
		}
	})
}
