package gincrud

// PageResult 分页响应
type PageResult[T any] struct {
	List     []T   `json:"list"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
}

// NewPageResult 创建分页响应
func NewPageResult[T any](list []T, total int64, query QueryDTO) *PageResult[T] {
	return &PageResult[T]{
		List:     list,
		Total:    total,
		Page:     query.GetPage(),
		PageSize: query.GetPageSize(),
	}
}
