package gincrud

import (
	"testing"
)

func TestBaseQueryDTO_GettersAndSetters(t *testing.T) {
	q := &BaseQueryDTO{}

	// Test Page
	q.SetPage(5)
	if q.GetPage() != 5 {
		t.Errorf("Expected page 5, got %d", q.GetPage())
	}

	// Test PageSize
	q.SetPageSize(20)
	if q.GetPageSize() != 20 {
		t.Errorf("Expected page size 20, got %d", q.GetPageSize())
	}

	// Test SortBy
	q.SetSortBy("name")
	if q.GetSortBy() != "name" {
		t.Errorf("Expected sort_by 'name', got '%s'", q.GetSortBy())
	}

	// Test SortOrder
	q.SetSortOrder("asc")
	if q.GetSortOrder() != "asc" {
		t.Errorf("Expected sort_order 'asc', got '%s'", q.GetSortOrder())
	}
}

func TestBaseQueryDTO_Normalize(t *testing.T) {
	tests := []struct {
		name         string
		input        *BaseQueryDTO
		expectedPage int
		expectedSize int
		expectedBy   string
		expectedOrder string
	}{
		{
			name:         "all defaults",
			input:        &BaseQueryDTO{},
			expectedPage: 1,
			expectedSize: 10,
			expectedBy:   "id",
			expectedOrder: "desc",
		},
		{
			name:         "negative page",
			input:        &BaseQueryDTO{Page: -1, PageSize: 20},
			expectedPage: 1,
			expectedSize: 20,
			expectedBy:   "id",
			expectedOrder: "desc",
		},
		{
			name:         "zero page size",
			input:        &BaseQueryDTO{Page: 2, PageSize: 0},
			expectedPage: 2,
			expectedSize: 10,
			expectedBy:   "id",
			expectedOrder: "desc",
		},
		{
			name:         "page size too large",
			input:        &BaseQueryDTO{Page: 1, PageSize: 200},
			expectedPage: 1,
			expectedSize: 100,
			expectedBy:   "id",
			expectedOrder: "desc",
		},
		{
			name:         "custom values preserved",
			input:        &BaseQueryDTO{Page: 3, PageSize: 25, SortBy: "created_at", SortOrder: "asc"},
			expectedPage: 3,
			expectedSize: 25,
			expectedBy:   "created_at",
			expectedOrder: "asc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.input.Normalize()
			if tt.input.Page != tt.expectedPage {
				t.Errorf("Expected page %d, got %d", tt.expectedPage, tt.input.Page)
			}
			if tt.input.PageSize != tt.expectedSize {
				t.Errorf("Expected page size %d, got %d", tt.expectedSize, tt.input.PageSize)
			}
			if tt.input.SortBy != tt.expectedBy {
				t.Errorf("Expected sort_by '%s', got '%s'", tt.expectedBy, tt.input.SortBy)
			}
			if tt.input.SortOrder != tt.expectedOrder {
				t.Errorf("Expected sort_order '%s', got '%s'", tt.expectedOrder, tt.input.SortOrder)
			}
		})
	}
}

func TestBaseQueryDTO_Offset(t *testing.T) {
	tests := []struct {
		page     int
		pageSize int
		expected int
	}{
		{1, 10, 0},
		{2, 10, 10},
		{3, 20, 40},
		{5, 25, 100},
		{1, 100, 0},
	}

	for _, tt := range tests {
		q := &BaseQueryDTO{Page: tt.page, PageSize: tt.pageSize}
		if q.Offset() != tt.expected {
			t.Errorf("Page %d, PageSize %d: Expected offset %d, got %d",
				tt.page, tt.pageSize, tt.expected, q.Offset())
		}
	}
}

func TestBaseQueryDTO_Limit(t *testing.T) {
	tests := []struct {
		pageSize int
		expected int
	}{
		{1, 1},
		{10, 10},
		{50, 50},
		{100, 100},
		{150, 100}, // capped at 100
		{200, 100}, // capped at 100
	}

	for _, tt := range tests {
		q := &BaseQueryDTO{PageSize: tt.pageSize}
		if q.Limit() != tt.expected {
			t.Errorf("PageSize %d: Expected limit %d, got %d", tt.pageSize, tt.expected, q.Limit())
		}
	}
}

func TestBaseQueryDTO_Interface(t *testing.T) {
	// Verify BaseQueryDTO implements QueryDTO interface
	var _ QueryDTO = (*BaseQueryDTO)(nil)
}
