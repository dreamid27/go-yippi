package entities

// FilterOperator defines the comparison operator for filtering
type FilterOperator string

const (
	OpEqual              FilterOperator = "eq"       // Equal (=)
	OpNotEqual           FilterOperator = "ne"       // Not Equal (!=)
	OpGreaterThan        FilterOperator = "gt"       // Greater Than (>)
	OpGreaterThanOrEqual FilterOperator = "gte"      // Greater Than or Equal (>=)
	OpLessThan           FilterOperator = "lt"       // Less Than (<)
	OpLessThanOrEqual    FilterOperator = "lte"      // Less Than or Equal (<=)
	OpLike               FilterOperator = "like"     // LIKE (case-sensitive partial match)
	OpILike              FilterOperator = "ilike"    // ILIKE (case-insensitive partial match)
	OpIn                 FilterOperator = "in"       // IN (value in array)
	OpNotIn              FilterOperator = "not_in"   // NOT IN (value not in array)
	OpIsNull             FilterOperator = "is_null"  // IS NULL
	OpIsNotNull          FilterOperator = "not_null" // IS NOT NULL
	OpContains           FilterOperator = "contains" // Array/JSONB contains
	OpStartsWith         FilterOperator = "starts"   // String starts with
	OpEndsWith           FilterOperator = "ends"     // String ends with
)

// Filter represents a single filter condition
type Filter struct {
	Field    string         // Field name (e.g., "name", "price", "status")
	Operator FilterOperator // Comparison operator
	Value    interface{}    // Value to compare (can be string, number, array, etc.)
}

// SortOrder defines the sort direction
type SortOrder string

const (
	SortAsc  SortOrder = "asc"  // Ascending
	SortDesc SortOrder = "desc" // Descending
)

// SortParam represents a single sort parameter
type SortParam struct {
	Field string    // Field name to sort by
	Order SortOrder // Sort order (asc or desc)
}

// QueryParams contains all parameters for a query
type QueryParams struct {
	Filters    []Filter
	Sort       []SortParam
	Pagination *PaginationParams
}

// QueryResult contains the result of a query with pagination metadata
type QueryResult struct {
	Products []*Product
	PageInfo PageInfo
}
