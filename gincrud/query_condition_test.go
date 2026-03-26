package gincrud

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewQuery(t *testing.T) {
	q := NewQuery()
	assert.NotNil(t, q)
	assert.Empty(t, q.GetWhereEq())
}

func TestQueryCondition_WhereEq(t *testing.T) {
	q := NewQuery()
	q.WhereEq("status", 1)
	assert.Len(t, q.GetWhereEq(), 1)
	assert.Equal(t, "status", q.GetWhereEq()[0].Field)
	assert.Equal(t, 1, q.GetWhereEq()[0].Value)
}

func TestQueryCondition_WhereNe(t *testing.T) {
	q := NewQuery()
	q.WhereNe("status", 0)
	assert.Len(t, q.GetWhereNe(), 1)
	assert.Equal(t, "status", q.GetWhereNe()[0].Field)
}

func TestQueryCondition_WhereGt(t *testing.T) {
	q := NewQuery()
	q.WhereGt("age", 18)
	assert.Len(t, q.GetWhereGt(), 1)
	assert.Equal(t, "age", q.GetWhereGt()[0].Field)
}

func TestQueryCondition_WhereLt(t *testing.T) {
	q := NewQuery()
	q.WhereLt("age", 60)
	assert.Len(t, q.GetWhereLt(), 1)
	assert.Equal(t, "age", q.GetWhereLt()[0].Field)
}

func TestQueryCondition_WhereGe(t *testing.T) {
	q := NewQuery()
	q.WhereGe("score", 60)
	assert.Len(t, q.GetWhereGe(), 1)
	assert.Equal(t, "score", q.GetWhereGe()[0].Field)
}

func TestQueryCondition_WhereLe(t *testing.T) {
	q := NewQuery()
	q.WhereLe("score", 100)
	assert.Len(t, q.GetWhereLe(), 1)
	assert.Equal(t, "score", q.GetWhereLe()[0].Field)
}

func TestQueryCondition_WhereBetween(t *testing.T) {
	q := NewQuery()
	q.WhereBetween("age", 18, 60)
	assert.Len(t, q.GetWhereBetween(), 1)
	assert.Equal(t, "age", q.GetWhereBetween()[0].Field)
	assert.Equal(t, 18, q.GetWhereBetween()[0].Min)
	assert.Equal(t, 60, q.GetWhereBetween()[0].Max)
}

func TestQueryCondition_WhereIn(t *testing.T) {
	q := NewQuery()
	q.WhereIn("role", "admin", "user")
	assert.Len(t, q.GetWhereIn(), 1)
	assert.Len(t, q.GetWhereIn()[0].Values, 2)
	assert.Equal(t, "role", q.GetWhereIn()[0].Field)
}

func TestQueryCondition_WhereLike(t *testing.T) {
	q := NewQuery()
	q.WhereLike("name", "%test%")
	assert.Len(t, q.GetWhereLike(), 1)
	assert.Equal(t, "name", q.GetWhereLike()[0].Field)
	assert.Equal(t, "%test%", q.GetWhereLike()[0].Pattern)
}

func TestQueryCondition_WhereNull(t *testing.T) {
	q := NewQuery()
	q.WhereNull("deleted_at")
	assert.Len(t, q.GetWhereNull(), 1)
	assert.Equal(t, "deleted_at", q.GetWhereNull()[0])
}

func TestQueryCondition_WhereNotNull(t *testing.T) {
	q := NewQuery()
	q.WhereNotNull("email")
	assert.Len(t, q.GetWhereNotNull(), 1)
	assert.Equal(t, "email", q.GetWhereNotNull()[0])
}

func TestQueryCondition_Preload(t *testing.T) {
	q := NewQuery()
	q.Preload("Posts", "Comments")

	preloads := q.GetPreloads()
	assert.Len(t, preloads, 2)
	assert.Equal(t, "Posts", preloads[0])
	assert.Equal(t, "Comments", preloads[1])
}

func TestQueryCondition_OrderBy(t *testing.T) {
	q := NewQuery()
	q.OrderBy("created_at").OrderBy("id", true)

	orderBy := q.GetOrderBy()
	assert.Len(t, orderBy, 2)
	assert.False(t, orderBy[0].Desc)
	assert.True(t, orderBy[1].Desc)
	assert.Equal(t, "created_at", orderBy[0].Field)
	assert.Equal(t, "id", orderBy[1].Field)
}

func TestQueryCondition_Limit(t *testing.T) {
	q := NewQuery()
	q.Limit(10)
	assert.Equal(t, 10, q.GetLimit())
}

func TestQueryCondition_Offset(t *testing.T) {
	q := NewQuery()
	q.Offset(20)
	assert.Equal(t, 20, q.GetOffset())
}

func TestQueryCondition_Chain(t *testing.T) {
	q := NewQuery()
	q.WhereEq("status", 1).
		WhereLike("name", "%test%").
		OrderBy("id", true).
		Limit(10).
		Offset(0)

	assert.Len(t, q.GetWhereEq(), 1)
	assert.Len(t, q.GetWhereLike(), 1)
	assert.Len(t, q.GetOrderBy(), 1)
	assert.Equal(t, 10, q.GetLimit())
	assert.Equal(t, 0, q.GetOffset())
}

func TestQueryCondition_AllConditions(t *testing.T) {
	q := NewQuery()
	q.WhereEq("status", 1).
		WhereNe("deleted", 1).
		WhereGt("age", 18).
		WhereLt("age", 60).
		WhereGe("score", 60).
		WhereLe("score", 100).
		WhereBetween("created_at", "2024-01-01", "2024-12-31").
		WhereIn("role", "admin", "user").
		WhereLike("name", "%test%").
		WhereNull("deleted_at").
		WhereNotNull("email").
		Preload("Posts").
		OrderBy("created_at", true).
		Limit(10).
		Offset(0)

	assert.Len(t, q.GetWhereEq(), 1)
	assert.Len(t, q.GetWhereNe(), 1)
	assert.Len(t, q.GetWhereGt(), 1)
	assert.Len(t, q.GetWhereLt(), 1)
	assert.Len(t, q.GetWhereGe(), 1)
	assert.Len(t, q.GetWhereLe(), 1)
	assert.Len(t, q.GetWhereBetween(), 1)
	assert.Len(t, q.GetWhereIn(), 1)
	assert.Len(t, q.GetWhereLike(), 1)
	assert.Len(t, q.GetWhereNull(), 1)
	assert.Len(t, q.GetWhereNotNull(), 1)
	assert.Len(t, q.GetPreloads(), 1)
	assert.Len(t, q.GetOrderBy(), 1)
	assert.Equal(t, 10, q.GetLimit())
	assert.Equal(t, 0, q.GetOffset())
}
