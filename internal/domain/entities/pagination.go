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
	PreviousCursor     string  `json:"previous_cursor"` // Empty string if no previous page
	NextCursor       string  `json:"next_cursor"`   // Empty string if no next page
	TotalCount      *int    `json:"total_count,omitempty"`
}

// PaginationParams contains parameters for pagination
type PaginationParams struct {
	Cursor    *string
	Limit     int
	Direction string // "forward" or "backward"
}
