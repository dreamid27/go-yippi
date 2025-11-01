# Pagination, Filtering, and Sorting Implementation Plan

## Overview
This document outlines the plan to add cursor-based pagination, filtering, and sorting capabilities to the Product API endpoints.

## Goals
- Implement efficient cursor-based pagination
- Add flexible filtering options for product queries
- Support multi-field sorting
- Maintain hexagonal architecture principles
- Ensure backward compatibility where possible

---

## 1. Cursor-Based Pagination

### Why Cursor-Based?
- **Performance**: More efficient for large datasets (no OFFSET scan)
- **Consistency**: No duplicate/missing items when data changes during pagination
- **Scalability**: Better for infinite scroll patterns

### Cursor Design
```go
// Cursor contains pagination metadata
type Cursor struct {
    ID        int       // Primary cursor field (unique identifier)
    CreatedAt time.Time // Secondary cursor field (for time-based ordering)
}

// Encoded as base64 string in API
// Example: "eyJpZCI6MTIzLCJjcmVhdGVkX2F0IjoiMjAyNC0wMS0wMVQxMjowMDowMFoifQ=="
```

### Pagination Parameters
```go
type PaginationParams struct {
    Cursor    *string // Cursor from previous response (optional)
    Limit     int     // Items per page (default: 20, max: 100)
    Direction string  // "forward" or "backward" (default: "forward")
}
```

### Response Format
```go
type PaginatedResponse struct {
    Data     []ProductListItem `json:"data"`
    PageInfo PageInfo          `json:"page_info"`
}

type PageInfo struct {
    HasNextPage     bool    `json:"has_next_page"`
    HasPreviousPage bool    `json:"has_previous_page"`
    PreviousCursor     *string `json:"previous_cursor"` // Cursor of first item
    NextCursor       *string `json:"next_cursor"`   // Cursor of last item
    TotalCount      *int    `json:"total_count"`  // Optional, expensive to compute
}
```

---

## 2. Filtering (Flexible Approach)

### Filter Operator System

Instead of hardcoded fields, we use a flexible filter array that supports multiple operators:

```go
type FilterOperator string

const (
    OpEqual              FilterOperator = "eq"      // Equal (=)
    OpNotEqual           FilterOperator = "ne"      // Not Equal (!=)
    OpGreaterThan        FilterOperator = "gt"      // Greater Than (>)
    OpGreaterThanOrEqual FilterOperator = "gte"     // Greater Than or Equal (>=)
    OpLessThan           FilterOperator = "lt"      // Less Than (<)
    OpLessThanOrEqual    FilterOperator = "lte"     // Less Than or Equal (<=)
    OpLike               FilterOperator = "like"    // LIKE (case-sensitive partial match)
    OpILike              FilterOperator = "ilike"   // ILIKE (case-insensitive partial match)
    OpIn                 FilterOperator = "in"      // IN (value in array)
    OpNotIn              FilterOperator = "not_in"  // NOT IN (value not in array)
    OpIsNull             FilterOperator = "is_null" // IS NULL
    OpIsNotNull          FilterOperator = "not_null"// IS NOT NULL
    OpContains           FilterOperator = "contains"// Array/JSONB contains
    OpStartsWith         FilterOperator = "starts"  // String starts with
    OpEndsWith           FilterOperator = "ends"    // String ends with
)

type Filter struct {
    Field    string         `json:"field"`    // Field name (e.g., "name", "price", "status")
    Operator FilterOperator `json:"operator"` // Comparison operator
    Value    interface{}    `json:"value"`    // Value to compare (can be string, number, array, etc.)
}

// Logical grouping (for complex queries)
type FilterGroup struct {
    Logic   string   `json:"logic"`   // "and" or "or"
    Filters []Filter `json:"filters"` // Array of filters
}
```

### Supported Product Filter Fields

| Field | Type | Supported Operators | Example |
|-------|------|---------------------|---------|
| `id` | int | `eq`, `ne`, `gt`, `gte`, `lt`, `lte`, `in`, `not_in` | `{"field": "id", "operator": "gt", "value": 100}` |
| `sku` | string | `eq`, `ne`, `like`, `ilike`, `in`, `not_in`, `starts`, `ends` | `{"field": "sku", "operator": "eq", "value": "SKU-123"}` |
| `slug` | string | `eq`, `ne`, `like`, `ilike`, `starts`, `ends` | `{"field": "slug", "operator": "ilike", "value": "laptop"}` |
| `name` | string | `eq`, `ne`, `like`, `ilike`, `starts`, `ends` | `{"field": "name", "operator": "ilike", "value": "%laptop%"}` |
| `description` | string | `like`, `ilike`, `contains` | `{"field": "description", "operator": "ilike", "value": "%gaming%"}` |
| `price` | float64 | `eq`, `ne`, `gt`, `gte`, `lt`, `lte` | `{"field": "price", "operator": "gte", "value": 100.0}` |
| `weight` | float64 | `eq`, `ne`, `gt`, `gte`, `lt`, `lte` | `{"field": "weight", "operator": "lt", "value": 5.0}` |
| `length` | float64 | `eq`, `ne`, `gt`, `gte`, `lt`, `lte` | `{"field": "length", "operator": "gt", "value": 10.0}` |
| `width` | float64 | `eq`, `ne`, `gt`, `gte`, `lt`, `lte` | `{"field": "width", "operator": "lte", "value": 20.0}` |
| `height` | float64 | `eq`, `ne`, `gt`, `gte`, `lt`, `lte` | `{"field": "height", "operator": "eq", "value": 15.0}` |
| `status` | string | `eq`, `ne`, `in`, `not_in` | `{"field": "status", "operator": "in", "value": ["published", "draft"]}` |
| `created_at` | timestamp | `eq`, `ne`, `gt`, `gte`, `lt`, `lte` | `{"field": "created_at", "operator": "gte", "value": "2024-01-01T00:00:00Z"}` |
| `updated_at` | timestamp | `eq`, `ne`, `gt`, `gte`, `lt`, `lte` | `{"field": "updated_at", "operator": "lt", "value": "2024-12-31T23:59:59Z"}` |

### Query String Format (GET with URL Parameters)

**Using indexed array notation:**
```
GET /products/query?filter[0][field]=status&filter[0][operator]=eq&filter[0][value]=published&filter[1][field]=price&filter[1][operator]=gte&filter[1][value]=100&sort[0][field]=price&sort[0][order]=desc&limit=20
```

**URL decoded for readability:**
```
GET /products/query
  ?filter[0][field]=status
  &filter[0][operator]=eq
  &filter[0][value]=published
  &filter[1][field]=price
  &filter[1][operator]=gte
  &filter[1][value]=100
  &sort[0][field]=price
  &sort[0][order]=desc
  &sort[1][field]=created_at
  &sort[1][order]=desc
  &limit=20
  &cursor=eyJpZCI6MTIzfQ==
```

**How it works:**
- `filter[N][field]`, `filter[N][operator]`, `filter[N][value]` - Array of filters
- `sort[N][field]`, `sort[N][order]` - Array of sort parameters
- `limit` - Pagination limit
- `cursor` - Pagination cursor
- `direction` - Pagination direction (forward/backward)

### Filter Logic

**Phase 1 (MVP):** All filters use AND logic
```
WHERE filter[0] AND filter[1] AND filter[2] ...
```

**Phase 2 (Future):** Add support for OR logic groups via query parameter:
```
GET /products/query?filter_logic=or&filter[0][field]=price&filter[0][operator]=lt&filter[0][value]=100&filter[1][field]=price&filter[1][operator]=gt&filter[1][value]=1000
```

This would translate to:
```sql
WHERE price < 100 OR price > 1000
```

For complex nested logic (Phase 3), we may need a different approach or accept that GET has limitations for very complex queries.

---

## 3. Sorting (Array-Based)

### Sort Parameter Structure

Instead of a comma-separated string, we use an array of sort objects:

```go
type SortOrder string

const (
    SortAsc  SortOrder = "asc"
    SortDesc SortOrder = "desc"
)

type SortParam struct {
    Field string    `json:"field"` // Field name (e.g., "price", "name", "created_at")
    Order SortOrder `json:"order"` // Sort order: "asc" or "desc"
}
```

### Supported Sort Fields

| Field | Type | Default Order | Notes |
|-------|------|---------------|-------|
| `id` | int | `desc` | Primary key, stable ordering |
| `sku` | string | `asc` | Alphanumeric SKU |
| `slug` | string | `asc` | URL-friendly identifier |
| `name` | string | `asc` | Product name |
| `price` | float64 | `desc` | Product price |
| `weight` | float64 | `asc` | Product weight |
| `length` | float64 | `asc` | Product length |
| `width` | float64 | `asc` | Product width |
| `height` | float64 | `asc` | Product height |
| `status` | string | `asc` | Product status (draft, published, archived) |
| `created_at` | timestamp | `desc` | Creation timestamp |
| `updated_at` | timestamp | `desc` | Last update timestamp |

### Query Format

**URL Query Params (GET):**
```
GET /products/query?sort[0][field]=price&sort[0][order]=desc&sort[1][field]=created_at&sort[1][order]=desc
```

**JSON Body (POST):**
```json
{
  "sort": [
    {
      "field": "price",
      "order": "desc"
    },
    {
      "field": "created_at",
      "order": "desc"
    },
    {
      "field": "id",
      "order": "desc"
    }
  ]
}
```

### Default Sort Behavior
If no sort is specified:
- Primary: `created_at DESC` (newest first)
- Secondary: `id DESC` (for stable ordering)

### Multi-Field Sorting Rules
1. Sorts are applied in array order (first to last)
2. First sort is primary, subsequent sorts are tie-breakers
3. Always append `id DESC` as final sort for stable pagination
4. Maximum of 3 custom sort fields (to prevent query performance issues)

### Examples

**Sort by price (high to low):**
```json
{
  "sort": [
    {"field": "price", "order": "desc"}
  ]
}
```

**Sort by status, then price, then name:**
```json
{
  "sort": [
    {"field": "status", "order": "asc"},
    {"field": "price", "order": "desc"},
    {"field": "name", "order": "asc"}
  ]
}
```

**Sort by newest first (explicit):**
```json
{
  "sort": [
    {"field": "created_at", "order": "desc"},
    {"field": "id", "order": "desc"}
  ]
}
```

---

## 4. Architecture Impact

### 4.1 Domain Layer Changes

**New file**: `internal/domain/entities/pagination.go`
```go
// Common pagination types used across all entities
type Cursor struct {
    ID        int
    CreatedAt time.Time
}

type PageInfo struct {
    HasNextPage     bool
    HasPreviousPage bool
    PreviousCursor     *string
    NextCursor       *string
    TotalCount      *int
}

type PaginationParams struct {
    Cursor    *string
    Limit     int
    Direction string
}
```

**New file**: `internal/domain/entities/query.go`
```go
// Filter operator types
type FilterOperator string

const (
    OpEqual              FilterOperator = "eq"
    OpNotEqual           FilterOperator = "ne"
    OpGreaterThan        FilterOperator = "gt"
    OpGreaterThanOrEqual FilterOperator = "gte"
    OpLessThan           FilterOperator = "lt"
    OpLessThanOrEqual    FilterOperator = "lte"
    OpLike               FilterOperator = "like"
    OpILike              FilterOperator = "ilike"
    OpIn                 FilterOperator = "in"
    OpNotIn              FilterOperator = "not_in"
    OpIsNull             FilterOperator = "is_null"
    OpIsNotNull          FilterOperator = "not_null"
    OpContains           FilterOperator = "contains"
    OpStartsWith         FilterOperator = "starts"
    OpEndsWith           FilterOperator = "ends"
)

// Generic filter for any field
type Filter struct {
    Field    string
    Operator FilterOperator
    Value    interface{} // Can be string, number, array, etc.
}

// Sort order types
type SortOrder string

const (
    SortAsc  SortOrder = "asc"
    SortDesc SortOrder = "desc"
)

// Generic sort parameter
type SortParam struct {
    Field string
    Order SortOrder
}

// Query filter with optional logic grouping
type QueryFilter struct {
    Logic   string   // "and" or "or" (default: "and")
    Filters []Filter
}
```

### 4.2 Ports Layer Changes

**Update**: `internal/domain/ports/repository.go`
```go
type ProductRepository interface {
    // Existing methods...
    Create(ctx context.Context, product *entities.Product) error
    GetByID(ctx context.Context, id int) (*entities.Product, error)
    GetBySKU(ctx context.Context, sku string) (*entities.Product, error)
    GetBySlug(ctx context.Context, slug string) (*entities.Product, error)
    List(ctx context.Context) ([]*entities.Product, error)
    ListByStatus(ctx context.Context, status entities.ProductStatus) ([]*entities.Product, error)
    Update(ctx context.Context, product *entities.Product) error
    Delete(ctx context.Context, id int) error

    // NEW: Flexible query method with filters, sorting, and pagination
    Query(ctx context.Context, params *entities.QueryParams) (*entities.QueryResult, error)
}

// Generic query parameters (reusable across entities)
type QueryParams struct {
    Filters    []entities.Filter      // Array of filters
    Sort       []entities.SortParam   // Array of sort parameters
    Pagination *entities.PaginationParams
}

// Generic query result (reusable across entities)
type QueryResult struct {
    Products []*entities.Product
    PageInfo entities.PageInfo
}
```

### 4.3 Application Layer Changes

**Update**: `internal/application/services/product_service.go`
```go
type ProductService struct {
    repo ports.ProductRepository
}

// NEW: Query method with validation
func (s *ProductService) QueryProducts(
    ctx context.Context,
    params *entities.QueryParams,
) (*entities.QueryResult, error) {
    // Validate pagination params
    if params.Pagination != nil {
        if params.Pagination.Limit <= 0 {
            params.Pagination.Limit = 20 // Default
        }
        if params.Pagination.Limit > 100 {
            params.Pagination.Limit = 100 // Max
        }
        if params.Pagination.Direction == "" {
            params.Pagination.Direction = "forward"
        }
    }

    // Validate sort params
    if len(params.Sort) > 3 {
        return nil, domainErrors.ErrInvalidInput // Max 3 sort fields
    }
    for _, sort := range params.Sort {
        if !s.isValidSortField(sort.Field) {
            return nil, domainErrors.NewInvalidInputError(
                fmt.Sprintf("invalid sort field: %s", sort.Field),
            )
        }
        if sort.Order != entities.SortAsc && sort.Order != entities.SortDesc {
            return nil, domainErrors.NewInvalidInputError(
                fmt.Sprintf("invalid sort order: %s", sort.Order),
            )
        }
    }

    // Validate filter params
    for _, filter := range params.Filters {
        if !s.isValidFilterField(filter.Field) {
            return nil, domainErrors.NewInvalidInputError(
                fmt.Sprintf("invalid filter field: %s", filter.Field),
            )
        }
        if !s.isValidOperatorForField(filter.Field, filter.Operator) {
            return nil, domainErrors.NewInvalidInputError(
                fmt.Sprintf("invalid operator %s for field %s", filter.Operator, filter.Field),
            )
        }
    }

    return s.repo.Query(ctx, params)
}

// Validation helpers
func (s *ProductService) isValidSortField(field string) bool {
    validFields := map[string]bool{
        "id": true, "sku": true, "slug": true, "name": true,
        "price": true, "weight": true, "length": true, "width": true,
        "height": true, "status": true, "created_at": true, "updated_at": true,
    }
    return validFields[field]
}

func (s *ProductService) isValidFilterField(field string) bool {
    // Same as sort fields
    return s.isValidSortField(field) || field == "description"
}

func (s *ProductService) isValidOperatorForField(field string, op entities.FilterOperator) bool {
    // Field-specific operator validation
    // String fields: eq, ne, like, ilike, in, not_in, starts, ends
    // Numeric fields: eq, ne, gt, gte, lt, lte
    // All fields: is_null, not_null

    stringFields := map[string]bool{"sku": true, "slug": true, "name": true, "description": true, "status": true}
    numericFields := map[string]bool{"id": true, "price": true, "weight": true, "length": true, "width": true, "height": true}

    // Universal operators
    if op == entities.OpIsNull || op == entities.OpIsNotNull {
        return true
    }

    if stringFields[field] {
        switch op {
        case entities.OpEqual, entities.OpNotEqual, entities.OpLike, entities.OpILike,
             entities.OpIn, entities.OpNotIn, entities.OpStartsWith, entities.OpEndsWith:
            return true
        }
    }

    if numericFields[field] {
        switch op {
        case entities.OpEqual, entities.OpNotEqual, entities.OpGreaterThan,
             entities.OpGreaterThanOrEqual, entities.OpLessThan, entities.OpLessThanOrEqual:
            return true
        }
    }

    return false
}
```

### 4.4 Adapter Layer Changes

**New file**: `internal/adapters/api/dto/query.go`
```go
// Generic filter DTO
type FilterDTO struct {
    Field    string      `json:"field" doc:"Field name to filter on"`
    Operator string      `json:"operator" doc:"Comparison operator (eq, ne, gt, gte, lt, lte, like, ilike, in, not_in, is_null, not_null, starts, ends)"`
    Value    interface{} `json:"value,omitempty" doc:"Value to compare against (type depends on field and operator)"`
}

// Generic sort DTO
type SortDTO struct {
    Field string `json:"field" doc:"Field name to sort by"`
    Order string `json:"order" doc:"Sort order (asc or desc)"`
}

// Pagination DTO
type PaginationDTO struct {
    Cursor    *string `json:"cursor,omitempty" doc:"Pagination cursor from previous response"`
    Limit     *int    `json:"limit,omitempty" doc:"Items per page (default: 20, max: 100)"`
    Direction *string `json:"direction,omitempty" doc:"Pagination direction: forward or backward (default: forward)"`
}

// Page info response DTO
type PageInfoDTO struct {
    HasNextPage     bool    `json:"has_next_page" doc:"Indicates if there are more items"`
    HasPreviousPage bool    `json:"has_previous_page" doc:"Indicates if there are previous items"`
    PreviousCursor     *string `json:"previous_cursor,omitempty" doc:"Cursor of the first item"`
    NextCursor       *string `json:"next_cursor,omitempty" doc:"Cursor of the last item"`
    TotalCount      *int    `json:"total_count,omitempty" doc:"Total count (optional, expensive to compute)"`
}
```

**Update**: `internal/adapters/api/dto/product.go`
```go
// Query request using URL query parameters (GET)
type QueryProductsRequest struct {
    // Filters - array of filter conditions
    // Usage: ?filter[0][field]=status&filter[0][operator]=eq&filter[0][value]=published
    Filters []FilterDTO `query:"filter" doc:"Array of filter conditions"`

    // Sort - array of sort parameters
    // Usage: ?sort[0][field]=price&sort[0][order]=desc
    Sort []SortDTO `query:"sort" doc:"Array of sort parameters"`

    // Pagination parameters
    Cursor    *string `query:"cursor" doc:"Pagination cursor from previous response"`
    Limit     *int    `query:"limit" doc:"Items per page (default: 20, max: 100)"`
    Direction *string `query:"direction" doc:"Pagination direction: forward or backward (default: forward)"`
}

// Query response with pagination metadata
type QueryProductsResponse struct {
    Body struct {
        Data     []ProductListItem `json:"data" doc:"List of products"`
        PageInfo PageInfoDTO       `json:"page_info" doc:"Pagination information"`
    }
}
```

**Update**: `internal/adapters/api/handlers/product_handler.go`
```go
func (h *ProductHandler) RegisterRoutes(api huma.API) {
    // Existing routes...

    // NEW: Query products with pagination, filtering, and sorting
    huma.Register(api, huma.Operation{
        OperationID: "query-products",
        Method:      http.MethodGet,
        Path:        "/products/query",
        Summary:     "Query products with pagination, filtering, and sorting",
        Description: "Advanced product search with cursor-based pagination",
        Tags:        []string{"Products"},
        Errors:      []int{http.StatusBadRequest, http.StatusInternalServerError},
    }, h.QueryProducts)
}

func (h *ProductHandler) QueryProducts(
    ctx context.Context,
    input *dto.QueryProductsRequest,
) (*dto.QueryProductsResponse, error) {
    // Convert DTO to domain entities
    // Call service
    // Convert response back to DTO
}
```

**New file**: `internal/adapters/persistence/product_repository_query.go`
```go
// Implements the Query method using Ent
func (r *ProductRepository) Query(
    ctx context.Context,
    params *entities.ProductQueryParams,
) (*entities.ProductQueryResult, error) {
    query := r.client.Product.Query()

    // Apply filters
    if params.Filter != nil {
        query = r.applyFilters(query, params.Filter)
    }

    // Apply sorting
    if len(params.Sort) > 0 {
        query = r.applySorting(query, params.Sort)
    }

    // Apply pagination (cursor-based)
    if params.Pagination != nil {
        query = r.applyPagination(query, params.Pagination)
    }

    // Execute query
    products, err := query.All(ctx)
    if err != nil {
        return nil, err
    }

    // Build page info
    pageInfo := r.buildPageInfo(products, params.Pagination)

    return &entities.ProductQueryResult{
        Products: products,
        PageInfo: pageInfo,
    }, nil
}
```

---

## 5. Implementation Phases

### Phase 1: Foundation
1. Create domain entities for pagination, filtering, sorting
2. Update repository port interface
3. Create DTO types for API layer

### Phase 2: Repository Implementation
1. Implement cursor encoding/decoding utilities
2. Implement filter application in Ent queries
3. Implement sorting application in Ent queries
4. Implement cursor-based pagination in Ent queries
5. Implement page info calculation

### Phase 3: Service Layer
1. Add validation logic for query parameters
2. Implement service method for querying products

### Phase 4: API Layer
1. Create handler method for query endpoint
2. Add DTO conversion utilities
3. Register new route
4. Update OpenAPI documentation

### Phase 5: Testing
1. Unit tests for cursor encoding/decoding
2. Unit tests for filter/sort/pagination logic
3. Integration tests for query endpoint
4. Performance testing with large datasets

---

## 6. Alternative Approaches

### A. Offset-Based Pagination
**Pros:**
- Simpler to implement
- Allows jumping to specific pages

**Cons:**
- Performance degrades with large offsets
- Inconsistent results when data changes
- Not recommended for modern APIs

**Verdict:** ❌ Not recommended for this use case

### B. Keyset Pagination (Simplified Cursor)
**Pros:**
- Similar performance to cursor-based
- Simpler implementation

**Cons:**
- Less flexible for complex sorting
- Harder to implement bidirectional pagination

**Verdict:** ⚠️ Could work, but full cursor is better

### C. GraphQL-Style Relay Cursor
**Pros:**
- Industry standard
- Rich pagination metadata
- Good for complex queries

**Cons:**
- More complex implementation
- Overkill for REST API

**Verdict:** ⚠️ Good pattern, but simplified version is sufficient

---

## 7. Open Questions & Decisions

1. **Should we keep the existing `/products` endpoint?**
   - ✅ **Decision**: Keep both for backward compatibility
   - `/products` - Simple list (existing)
   - `/products/query` - Advanced filtering/sorting/pagination (new)

2. **Should we include total count in responses?**
   - ✅ **Decision**: Make it optional via query param `include_total=true`
   - Pro: Useful for UI (showing "Page 1 of 10")
   - Con: Expensive query (COUNT(*) on large tables)

3. **Should we support full-text search?**
   - ✅ **Decision**: Start with ILIKE, upgrade to full-text search if needed
   - Phase 1: Use `ilike` operator with wildcards
   - Phase 2: Add PostgreSQL full-text search if performance becomes an issue

4. **How to handle cursor invalidation?**
   - ✅ **Decision**: Return next valid item, document behavior in API docs
   - If cursor item is deleted, find next valid item based on sort criteria
   - Return empty results if no items exist after cursor

5. **How to handle IN operator values in URL?**
   - ✅ **Decision**: Use comma-separated values
   - Example: `filter[0][value]=published,draft,archived`
   - Parse and split on comma in handler layer

6. **Rate limiting on complex queries?**
   - ✅ **Decision**: Not in MVP, consider for future
   - Phase 1: Basic validation (max 10 filters, max 3 sorts)
   - Phase 2: Add rate limiting if abuse is detected

---

## 8. Performance Considerations

### Indexes Required
```sql
-- For cursor-based pagination
CREATE INDEX idx_products_created_at_id ON products(created_at DESC, id DESC);

-- For price filtering + sorting
CREATE INDEX idx_products_price ON products(price);

-- For status filtering
CREATE INDEX idx_products_status ON products(status);

-- For text search (if using full-text)
CREATE INDEX idx_products_search ON products USING gin(to_tsvector('english', name || ' ' || description));
```

### Query Optimization
- Fetch `limit + 1` items to determine `has_next_page`
- Avoid COUNT(*) unless explicitly requested
- Use covering indexes where possible
- Set reasonable max limit (100 items)

---

## 9. Example API Usage

All examples use `GET /products/query` with query parameters.

### Basic Pagination

**Get first page:**
```
GET /products/query?limit=20
```

**Response:**
```json
{
  "data": [
    {
      "id": 1,
      "sku": "SKU-001",
      "name": "Product 1",
      "price": 99.99,
      "status": "published",
      "created_at": "2024-01-01T12:00:00Z",
      "updated_at": "2024-01-01T12:00:00Z"
    }
  ],
  "page_info": {
    "has_next_page": true,
    "has_previous_page": false,
    "previous_cursor": "eyJpZCI6MSwi...",
    "next_cursor": "eyJpZCI6MjAsI..."
  }
}
```

**Get next page:**
```
GET /products/query?limit=20&cursor=eyJpZCI6MjAsI...
```

### Filtering Examples

**Single filter (published products):**
```
GET /products/query?filter[0][field]=status&filter[0][operator]=eq&filter[0][value]=published&limit=20
```

**Multiple filters (published AND expensive products):**
```
GET /products/query?filter[0][field]=status&filter[0][operator]=eq&filter[0][value]=published&filter[1][field]=price&filter[1][operator]=gte&filter[1][value]=1000&limit=20
```

**Range filter (price between 100 and 1000):**
```
GET /products/query?filter[0][field]=price&filter[0][operator]=gte&filter[0][value]=100&filter[1][field]=price&filter[1][operator]=lte&filter[1][value]=1000&limit=20
```

**Text search (case-insensitive name search):**
```
GET /products/query?filter[0][field]=name&filter[0][operator]=ilike&filter[0][value]=%laptop%&limit=20
```

**IN operator (multiple statuses):**
```
GET /products/query?filter[0][field]=status&filter[0][operator]=in&filter[0][value]=published,draft&limit=20
```
Note: For `in` operator, values are comma-separated in the URL.

**Starts with:**
```
GET /products/query?filter[0][field]=sku&filter[0][operator]=starts&filter[0][value]=LAPTOP&limit=20
```

### Sorting Examples

**Sort by price (high to low):**
```
GET /products/query?sort[0][field]=price&sort[0][order]=desc&limit=20
```

**Multi-field sort (status, then price, then name):**
```
GET /products/query?sort[0][field]=status&sort[0][order]=asc&sort[1][field]=price&sort[1][order]=desc&sort[2][field]=name&sort[2][order]=asc&limit=20
```

### Combined (Filter + Sort + Pagination)

**Published laptops with price >= 500, sorted by price (high to low):**
```
GET /products/query?filter[0][field]=status&filter[0][operator]=eq&filter[0][value]=published&filter[1][field]=name&filter[1][operator]=ilike&filter[1][value]=%laptop%&filter[2][field]=price&filter[2][operator]=gte&filter[2][value]=500&sort[0][field]=price&sort[0][order]=desc&sort[1][field]=created_at&sort[1][order]=desc&limit=20
```

**URL decoded for readability:**
```
GET /products/query
  ?filter[0][field]=status
  &filter[0][operator]=eq
  &filter[0][value]=published
  &filter[1][field]=name
  &filter[1][operator]=ilike
  &filter[1][value]=%laptop%
  &filter[2][field]=price
  &filter[2][operator]=gte
  &filter[2][value]=500
  &sort[0][field]=price
  &sort[0][order]=desc
  &sort[1][field]=created_at
  &sort[1][order]=desc
  &limit=20
```

---

## 10. Next Steps

1. **Review this plan** - Discuss any concerns or alternative approaches
2. **Decide on open questions** - Clarify requirements
3. **Approve architecture changes** - Ensure alignment with hexagonal principles
4. **Begin implementation** - Start with Phase 1

---

## Notes

- All changes maintain hexagonal architecture principles
- Domain layer remains pure (no infrastructure dependencies)
- Repository implements the complex query logic using Ent
- Service layer adds business validation
- Handler layer manages HTTP concerns and DTO conversion
- Backward compatible (existing endpoints remain unchanged)
