package crud

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPageResult_Struct(t *testing.T) {
	type User struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	list := []User{
		{ID: 1, Name: "Alice"},
		{ID: 2, Name: "Bob"},
	}

	pr := &PageResult[User]{
		List:     list,
		Total:    100,
		Page:     1,
		PageSize: 10,
	}

	assert.Equal(t, 2, len(pr.List))
	assert.Equal(t, int64(100), pr.Total)
	assert.Equal(t, 1, pr.Page)
	assert.Equal(t, 10, pr.PageSize)
}

func TestNewPageResult(t *testing.T) {
	type Item struct {
		ID int `json:"id"`
	}

	list := []Item{{ID: 1}, {ID: 2}, {ID: 3}}
	query := &BaseQueryDTO{Page: 2, PageSize: 20}

	pr := NewPageResult(list, 50, query)

	assert.Equal(t, 3, len(pr.List))
	assert.Equal(t, int64(50), pr.Total)
	assert.Equal(t, 2, pr.Page)
	assert.Equal(t, 20, pr.PageSize)
}

func TestNewPageResult_DefaultQuery(t *testing.T) {
	type Data struct {
		Value string `json:"value"`
	}

	list := []Data{{Value: "test"}}
	query := &BaseQueryDTO{}
	query.Normalize()

	pr := NewPageResult(list, 1, query)

	assert.Equal(t, 1, len(pr.List))
	assert.Equal(t, int64(1), pr.Total)
	assert.Equal(t, 1, pr.Page)
	assert.Equal(t, 10, pr.PageSize)
}

func TestPageResult_EmptyList(t *testing.T) {
	type Empty struct{}

	pr := NewPageResult[Empty]([]Empty{}, 0, &BaseQueryDTO{Page: 1, PageSize: 10})

	assert.Equal(t, 0, len(pr.List))
	assert.Equal(t, int64(0), pr.Total)
	assert.Equal(t, 1, pr.Page)
	assert.Equal(t, 10, pr.PageSize)
}
