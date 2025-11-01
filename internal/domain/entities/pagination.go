package entities

// Cursor contains pagination metadata for cursor-based pagination
type Cursor struct {
	ID        int    `json:"id"`
	CreatedAt string `json:"created_at"` // RFC3339 format
}

// PageInfo contains pagination metadata in the response
type PageInfo struct {
	HasNextPage     bool    `json:"has_next_page"`
	HasPreviousPage bool    `json:"has_previous_page"`
	StartCursor     *string `json:"start_cursor,omitempty"`
	EndCursor       *string `json:"end_cursor,omitempty"`
	TotalCount      *int    `json:"total_count,omitempty"`
}

// PaginationParams contains parameters for pagination
type PaginationParams struct {
	Cursor    *string
	Limit     int
	Direction string // "forward" or "backward"
}
