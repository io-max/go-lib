package gincrud

import (
	"testing"
	"time"
)

func TestBaseEntity_GetID(t *testing.T) {
	e := &BaseEntity{}
	e.SetID(123)
	if e.GetID() != 123 {
		t.Errorf("Expected ID 123, got %d", e.GetID())
	}
}

func TestBaseEntity_SetID(t *testing.T) {
	e := &BaseEntity{}
	e.SetID(456)
	if e.ID != 456 {
		t.Errorf("Expected ID 456, got %d", e.ID)
	}
}

func TestBaseEntity_GetDeleted(t *testing.T) {
	e := &BaseEntity{}
	e.SetDeleted(1234567890)
	if e.GetDeleted() != 1234567890 {
		t.Errorf("Expected Deleted 1234567890, got %d", e.GetDeleted())
	}
}

func TestBaseEntity_SetDeleted(t *testing.T) {
	e := &BaseEntity{}
	e.SetDeleted(9876543210)
	if e.Deleted != 9876543210 {
		t.Errorf("Expected Deleted 9876543210, got %d", e.Deleted)
	}
}

func TestBaseEntity_GetCreatedAt(t *testing.T) {
	e := &BaseEntity{}
	expected := time.Now()
	e.SetCreatedAt(expected)
	if e.GetCreatedAt().Unix() != expected.Unix() {
		t.Errorf("Expected CreatedAt %v, got %v", expected, e.GetCreatedAt())
	}
}

func TestBaseEntity_SetCreatedAt(t *testing.T) {
	e := &BaseEntity{}
	expected := time.Now()
	e.SetCreatedAt(expected)
	if e.CreatedAt.Unix() != expected.Unix() {
		t.Errorf("Expected CreatedAt %v, got %v", expected, e.CreatedAt)
	}
}

func TestBaseEntity_GetUpdatedAt(t *testing.T) {
	e := &BaseEntity{}
	expected := time.Now()
	e.SetUpdatedAt(expected)
	if e.GetUpdatedAt().Unix() != expected.Unix() {
		t.Errorf("Expected UpdatedAt %v, got %v", expected, e.GetUpdatedAt())
	}
}

func TestBaseEntity_SetUpdatedAt(t *testing.T) {
	e := &BaseEntity{}
	expected := time.Now()
	e.SetUpdatedAt(expected)
	if e.UpdatedAt.Unix() != expected.Unix() {
		t.Errorf("Expected UpdatedAt %v, got %v", expected, e.UpdatedAt)
	}
}

func TestBaseEntity_IsDeleted(t *testing.T) {
	e := &BaseEntity{}

	// Test not deleted
	if e.IsDeleted() {
		t.Error("Expected IsDeleted to be false")
	}

	// Test deleted
	e.SetDeleted(time.Now().Unix())
	if !e.IsDeleted() {
		t.Error("Expected IsDeleted to be true")
	}
}

func TestBaseEntity_MarkDeleted(t *testing.T) {
	e := &BaseEntity{}
	before := time.Now().Unix()
	e.MarkDeleted()
	after := time.Now().Unix()

	if !e.IsDeleted() {
		t.Error("Expected entity to be marked as deleted")
	}

	deletedTime := e.GetDeleted()
	if deletedTime < before || deletedTime > after {
		t.Errorf("Expected deleted time between %d and %d, got %d", before, after, deletedTime)
	}
}

func TestAuditEntity_GetCreatedBy(t *testing.T) {
	e := &AuditEntity{}
	e.SetCreatedBy(100)
	if e.GetCreatedBy() != 100 {
		t.Errorf("Expected CreatedBy 100, got %d", e.GetCreatedBy())
	}
}

func TestAuditEntity_SetCreatedBy(t *testing.T) {
	e := &AuditEntity{}
	e.SetCreatedBy(200)
	if e.CreatedBy != 200 {
		t.Errorf("Expected CreatedBy 200, got %d", e.CreatedBy)
	}
}

func TestAuditEntity_GetUpdatedBy(t *testing.T) {
	e := &AuditEntity{}
	e.SetUpdatedBy(300)
	if e.GetUpdatedBy() != 300 {
		t.Errorf("Expected UpdatedBy 300, got %d", e.GetUpdatedBy())
	}
}

func TestAuditEntity_SetUpdatedBy(t *testing.T) {
	e := &AuditEntity{}
	e.SetUpdatedBy(400)
	if e.UpdatedBy != 400 {
		t.Errorf("Expected UpdatedBy 400, got %d", e.UpdatedBy)
	}
}

func TestAuditEntity_InheritsBaseEntity(t *testing.T) {
	e := &AuditEntity{}

	// Test inherited methods
	e.SetID(999)
	if e.GetID() != 999 {
		t.Errorf("Expected ID 999, got %d", e.GetID())
	}

	e.SetDeleted(111)
	if !e.IsDeleted() {
		t.Error("Expected AuditEntity to inherit IsDeleted")
	}

	expected := time.Now()
	e.SetCreatedAt(expected)
	if e.GetCreatedAt().Unix() != expected.Unix() {
		t.Error("Expected AuditEntity to inherit CreatedAt methods")
	}
}

func TestEntity_InterfaceImplementation(t *testing.T) {
	// Verify BaseEntity implements Entity interface
	var _ Entity = (*BaseEntity)(nil)

	// Verify AuditEntity implements Entity interface through embedding
	var _ Entity = (*AuditEntity)(nil)
}
