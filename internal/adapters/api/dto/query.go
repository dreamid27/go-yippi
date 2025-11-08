package dto

// FilterDTO represents a filter condition in the API layer
type FilterDTO struct {
	Field    string      `json:"field" doc:"Field name to filter on (e.g., status, price, sku, name, category_id, brand_id)"`
	Operator string      `json:"operator" enum:"eq,ne,gt,gte,lt,lte,like,ilike,in,not_in,is_null,not_null,starts,ends" doc:"Comparison operator"`
	Value    interface{} `json:"value,omitempty" doc:"Value to compare against (type depends on field and operator). For 'in' and 'not_in' operators, use array format: [value1,value2]"`
}

// SortDTO represents a sort parameter in the API layer
type SortDTO struct {
	Field string `json:"field" doc:"Field name to sort by (e.g., id, sku, name, price, created_at, updated_at)"`
	Order string `json:"order" enum:"asc,desc" doc:"Sort order: ascending or descending"`
}

// PageInfoDTO represents pagination metadata in the response
type PageInfoDTO struct {
	HasNextPage     bool   `json:"has_next_page" doc:"Indicates if there are more items"`
	HasPreviousPage bool   `json:"has_previous_page" doc:"Indicates if there are previous items"`
	PreviousCursor     string `json:"previous_cursor" doc:"Cursor of the first item (empty string if no previous page)"`
	NextCursor       string `json:"next_cursor" doc:"Cursor of the last item (empty string if no next page)"`
	TotalCount      *int   `json:"total_count,omitempty" doc:"Total count (optional, expensive to compute)"`
}
