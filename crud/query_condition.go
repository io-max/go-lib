package crud

// EqCond 等于条件
type EqCond struct {
	Field string
	Value any
}

// NeCond 不等于条件
type NeCond struct {
	Field string
	Value any
}

// GtCond 大于条件
type GtCond struct {
	Field string
	Value any
}

// LtCond 小于条件
type LtCond struct {
	Field string
	Value any
}

// GeCond 大于等于条件
type GeCond struct {
	Field string
	Value any
}

// LeCond 小于等于条件
type LeCond struct {
	Field string
	Value any
}

// BetweenCond 区间条件
type BetweenCond struct {
	Field string
	Min   any
	Max   any
}

// InCond IN 条件
type InCond struct {
	Field  string
	Values []any
}

// LikeCond LIKE 条件
type LikeCond struct {
	Field   string
	Pattern string
}

// OrderByField 排序字段
type OrderByField struct {
	Field string
	Desc  bool
}

// QueryCondition 查询条件构建器
type QueryCondition struct {
	whereEq      []EqCond
	whereNe      []NeCond
	whereGt      []GtCond
	whereLt      []LtCond
	whereGe      []GeCond
	whereLe      []LeCond
	whereBetween []BetweenCond
	whereIn      []InCond
	whereLike    []LikeCond
	whereNull    []string
	whereNotNull []string
	preloads     []string
	orderBy      []OrderByField

	// 分页参数
	page     int
	pageSize int
}

// NewQuery 创建查询条件
func NewQuery() *QueryCondition {
	return &QueryCondition{
		whereEq:      make([]EqCond, 0),
		whereNe:      make([]NeCond, 0),
		whereGt:      make([]GtCond, 0),
		whereLt:      make([]LtCond, 0),
		whereGe:      make([]GeCond, 0),
		whereLe:      make([]LeCond, 0),
		whereBetween: make([]BetweenCond, 0),
		whereIn:      make([]InCond, 0),
		whereLike:    make([]LikeCond, 0),
		whereNull:    make([]string, 0),
		whereNotNull: make([]string, 0),
		preloads:     make([]string, 0),
		orderBy:      make([]OrderByField, 0),
	}
}

// WhereEq 添加等于条件
func (q *QueryCondition) WhereEq(field string, value any) *QueryCondition {
	q.whereEq = append(q.whereEq, EqCond{Field: field, Value: value})
	return q
}

// WhereNe 添加不等于条件
func (q *QueryCondition) WhereNe(field string, value any) *QueryCondition {
	q.whereNe = append(q.whereNe, NeCond{Field: field, Value: value})
	return q
}

// WhereGt 添加大于条件
func (q *QueryCondition) WhereGt(field string, value any) *QueryCondition {
	q.whereGt = append(q.whereGt, GtCond{Field: field, Value: value})
	return q
}

// WhereLt 添加小于条件
func (q *QueryCondition) WhereLt(field string, value any) *QueryCondition {
	q.whereLt = append(q.whereLt, LtCond{Field: field, Value: value})
	return q
}

// WhereGe 添加大于等于条件
func (q *QueryCondition) WhereGe(field string, value any) *QueryCondition {
	q.whereGe = append(q.whereGe, GeCond{Field: field, Value: value})
	return q
}

// WhereLe 添加小于等于条件
func (q *QueryCondition) WhereLe(field string, value any) *QueryCondition {
	q.whereLe = append(q.whereLe, LeCond{Field: field, Value: value})
	return q
}

// WhereBetween 添加区间条件
func (q *QueryCondition) WhereBetween(field string, min, max any) *QueryCondition {
	q.whereBetween = append(q.whereBetween, BetweenCond{Field: field, Min: min, Max: max})
	return q
}

// WhereIn 添加 IN 条件
func (q *QueryCondition) WhereIn(field string, values ...any) *QueryCondition {
	q.whereIn = append(q.whereIn, InCond{Field: field, Values: values})
	return q
}

// WhereLike 添加 LIKE 条件
func (q *QueryCondition) WhereLike(field string, pattern string) *QueryCondition {
	q.whereLike = append(q.whereLike, LikeCond{Field: field, Pattern: pattern})
	return q
}

// WhereNull 添加 IS NULL 条件
func (q *QueryCondition) WhereNull(field string) *QueryCondition {
	q.whereNull = append(q.whereNull, field)
	return q
}

// WhereNotNull 添加 IS NOT NULL 条件
func (q *QueryCondition) WhereNotNull(field string) *QueryCondition {
	q.whereNotNull = append(q.whereNotNull, field)
	return q
}

// Preload 添加预加载关联
func (q *QueryCondition) Preload(relations ...string) *QueryCondition {
	q.preloads = append(q.preloads, relations...)
	return q
}

// OrderBy 添加排序字段
func (q *QueryCondition) OrderBy(field string, desc ...bool) *QueryCondition {
	isDesc := false
	if len(desc) > 0 && desc[0] {
		isDesc = true
	}
	q.orderBy = append(q.orderBy, OrderByField{Field: field, Desc: isDesc})
	return q
}

// Page 设置页码
func (q *QueryCondition) Page(page int) *QueryCondition {
	q.page = page
	return q
}

// PageSize 设置每页数量
func (q *QueryCondition) PageSize(size int) *QueryCondition {
	q.pageSize = size
	return q
}

// GetPage 获取页码
func (q *QueryCondition) GetPage() int {
	if q.page < 1 {
		return 1
	}
	return q.page
}

// GetPageSize 获取每页数量
func (q *QueryCondition) GetPageSize() int {
	if q.pageSize < 1 {
		return 10
	}
	if q.pageSize > 100 {
		return 100
	}
	return q.pageSize
}

// Offset 计算偏移量
func (q *QueryCondition) Offset() int {
	return (q.GetPage() - 1) * q.GetPageSize()
}

// Limit 获取限制数量
func (q *QueryCondition) Limit() int {
	return q.GetPageSize()
}

// Getters
func (q *QueryCondition) GetWhereEq() []EqCond           { return q.whereEq }
func (q *QueryCondition) GetWhereNe() []NeCond           { return q.whereNe }
func (q *QueryCondition) GetWhereGt() []GtCond           { return q.whereGt }
func (q *QueryCondition) GetWhereLt() []LtCond           { return q.whereLt }
func (q *QueryCondition) GetWhereGe() []GeCond           { return q.whereGe }
func (q *QueryCondition) GetWhereLe() []LeCond           { return q.whereLe }
func (q *QueryCondition) GetWhereBetween() []BetweenCond { return q.whereBetween }
func (q *QueryCondition) GetWhereIn() []InCond           { return q.whereIn }
func (q *QueryCondition) GetWhereLike() []LikeCond       { return q.whereLike }
func (q *QueryCondition) GetWhereNull() []string         { return q.whereNull }
func (q *QueryCondition) GetWhereNotNull() []string      { return q.whereNotNull }
func (q *QueryCondition) GetPreloads() []string          { return q.preloads }
func (q *QueryCondition) GetOrderBy() []OrderByField     { return q.orderBy }
