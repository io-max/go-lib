package gincrud

// QueryDTO 查询参数接口
type QueryDTO interface {
	GetPage() int
	SetPage(page int)
	GetPageSize() int
	SetPageSize(size int)
	GetSortBy() string
	SetSortBy(field string)
	GetSortOrder() string
	SetSortOrder(order string)
	Normalize()
	Offset() int
	Limit() int
}

// BaseQueryDTO 基础查询参数
type BaseQueryDTO struct {
	Page      int    `form:"page" binding:"min=1"`
	PageSize  int    `form:"page_size" binding:"min=1,max=100"`
	SortBy    string `form:"sort_by"`
	SortOrder string `form:"sort_order" binding:"oneof=asc desc"`
}

func (q *BaseQueryDTO) GetPage() int      { return q.Page }
func (q *BaseQueryDTO) SetPage(page int)  { q.Page = page }
func (q *BaseQueryDTO) GetPageSize() int  { return q.PageSize }
func (q *BaseQueryDTO) SetPageSize(size int) { q.PageSize = size }
func (q *BaseQueryDTO) GetSortBy() string { return q.SortBy }
func (q *BaseQueryDTO) SetSortBy(field string) { q.SortBy = field }
func (q *BaseQueryDTO) GetSortOrder() string { return q.SortOrder }
func (q *BaseQueryDTO) SetSortOrder(order string) { q.SortOrder = order }

func (q *BaseQueryDTO) Normalize() {
	if q.Page < 1 { q.Page = 1 }
	if q.PageSize < 1 { q.PageSize = 10 }
	if q.PageSize > 100 { q.PageSize = 100 }
	if q.SortBy == "" { q.SortBy = "id" }
	if q.SortOrder == "" { q.SortOrder = "desc" }
}

func (q *BaseQueryDTO) Offset() int { return (q.Page - 1) * q.PageSize }
func (q *BaseQueryDTO) Limit() int {
	if q.PageSize > 100 { return 100 }
	return q.PageSize
}
