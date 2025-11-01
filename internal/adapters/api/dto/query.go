package dto

// FilterDTO represents a filter condition in the API layer
type FilterDTO struct {
	Field    string      `query:"field" json:"field" doc:"Field name to filter on"`
	Operator string      `query:"operator" json:"operator" doc:"Comparison operator (eq, ne, gt, gte, lt, lte, like, ilike, in, not_in, is_null, not_null, starts, ends)"`
	Value    interface{} `query:"value" json:"value,omitempty" doc:"Value to compare against (type depends on field and operator)"`
}

// SortDTO represents a sort parameter in the API layer
type SortDTO struct {
	Field string `query:"field" json:"field" doc:"Field name to sort by"`
	Order string `query:"order" json:"order" doc:"Sort order (asc or desc)"`
}

// PageInfoDTO represents pagination metadata in the response
type PageInfoDTO struct {
	HasNextPage     bool    `json:"has_next_page" doc:"Indicates if there are more items"`
	HasPreviousPage bool    `json:"has_previous_page" doc:"Indicates if there are previous items"`
	StartCursor     *string `json:"start_cursor,omitempty" doc:"Cursor of the first item"`
	EndCursor       *string `json:"end_cursor,omitempty" doc:"Cursor of the last item"`
	TotalCount      *int    `json:"total_count,omitempty" doc:"Total count (optional, expensive to compute)"`
}
