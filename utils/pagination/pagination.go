package pagination

import "strconv"

// PageParams 分页参数结构
type PageParams struct {
	Page     int
	PageSize int
}

// PaginatedResponse 分页响应结构
// @Schema(description="分页响应结构")
type PaginatedResponse struct {
	Data       interface{} `json:"data"`        // 数据列表
	Total      int         `json:"total"`       // 总记录数
	Page       int         `json:"page"`        // 当前页码
	PageSize   int         `json:"page_size"`   // 每页数量
	TotalPages int         `json:"total_pages"` // 总页数
}

// ParsePaginationParams 解析分页参数
func ParsePaginationParams(pageStr, pageSizeStr string) PageParams {
	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	if page < 1 {
		page = 1
	}

	switch {
	case pageSize > 100:
		pageSize = 100
	case pageSize < 1:
		pageSize = 10
	}

	return PageParams{
		Page:     page,
		PageSize: pageSize,
	}
}

// NewPaginatedResponse 创建分页响应
func NewPaginatedResponse(data interface{}, total, page, pageSize int) PaginatedResponse {
	totalPages := 0
	if total > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}

	return PaginatedResponse{
		Data:       data,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}
}
