# Brands Filtered by Category - Design Document

**Date:** 2026-01-02
**Feature:** Filter brands by category_ids
**Status:** Design Approved

## Overview

Add filtering capability to the existing `GET /brands` endpoint to allow filtering brands by the categories their products belong to. This enables use cases like "show me all brands that have products in the Electronics or Books categories."

## API Endpoint

### Endpoint
```
GET /brands?category_ids=<uuid1>,<uuid2>,<uuid3>
```

### Request Parameters
- `category_ids` (optional, query parameter): Comma-separated list of category UUIDs
  - Example: `/brands?category_ids=123e4567-e89b-12d3-a456-426614174000,987f6543-e21b-43d3-b456-555614174000`
  - If omitted or empty: Returns all brands (existing behavior preserved)
  - If provided: Returns only brands that have at least one product in ANY of the specified categories

### Response Format
```json
{
  "brands": [
    {
      "id": "uuid",
      "name": "Brand Name",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

Same response structure as existing `GET /brands` endpoint for consistency.

### Example Requests
- `GET /brands` → Returns all brands (current behavior, unchanged)
- `GET /brands?category_ids=uuid1` → Returns brands with products in category uuid1
- `GET /brands?category_ids=uuid1,uuid2` → Returns brands with products in category uuid1 OR uuid2

## Implementation Architecture

### Layer 1: Handler (`internal/adapters/api/handlers/brand_handler.go`)

**Changes:**
- Modify `ListBrands` handler to accept new request DTO with `category_ids` field
- Parse comma-separated UUIDs and validate format
- Call service layer with parsed category IDs
- Map domain entities to DTOs

**Validation:**
- Each UUID in `category_ids` must be valid UUID format
- Invalid UUIDs → `400 Bad Request`

### Layer 2: Service (`internal/application/services/brand_service.go`)

**Changes:**
- Update `ListBrands` method signature to accept `categoryIDs []uuid.UUID`
- Pass `nil` if no categories provided (return all brands)
- Delegate to repository layer

**Business Logic:**
- If `categoryIDs` is `nil` or empty → return all brands
- If `categoryIDs` provided → filter by categories

### Layer 3: Repository (`internal/adapters/persistence/brand_repository.go`)

**Changes:**
- Update `ListBrands` method to accept optional `categoryIDs []uuid.UUID`
- Implement filtered query using Ent

**Query Logic:**
```go
// If no category filter
client.Brand.Query().All(ctx)

// With category filter
client.Brand.Query().
    Where(
        func(s *sql.Selector) {
            // Join with products and filter by category_ids
            t := sql.Table(products.Table).
                // Get distinct brands that have products in any of the categories
                // INNER JOIN products ON products.brand_id = brands.id
                // WHERE products.category_id IN (?)
        },
    ).
    Distinct().
    Order(ent.Asc(brand.FieldName)).
    All(ctx)
```

**SQL Equivalent:**
```sql
SELECT DISTINCT brands.*
FROM brands
INNER JOIN products ON products.brand_id = brands.id
WHERE products.category_id IN (?, ?, ?)
ORDER BY brands.name
```

### Layer 4: DTO (`internal/adapters/api/dto/brand_dto.go`)

**New DTO:**
```go
type ListBrandsRequest struct {
    CategoryIDs string `query:"category_ids" doc:"Comma-separated list of category UUIDs to filter brands"`
}

type ListBrandsResponse struct {
    Body struct {
        Brands []BrandListItem `json:"brands"`
    }
}
```

## Error Handling

| Scenario | HTTP Status | Behavior |
|----------|-------------|----------|
| Invalid UUID format | `400 Bad Request` | Return error immediately (fail fast) |
| Empty/No category_ids | `200 OK` | Return all brands (backward compatible) |
| No matching products | `200 OK` | Return empty brand array |
| Non-existent category UUID | `200 OK` | Return empty brand array (not an error) |
| Duplicate category_ids | `200 OK` | Deduplicate before querying |
| Brands with no products | `200 OK` | Excluded from filtered results |

## Testing Strategy

### Service Layer Tests (`internal/application/services/brand_service_test.go`)

Test cases:
- ✅ `ListBrands_NoCategoryFilter_ReturnsAllBrands`
- ✅ `ListBrands_WithValidCategoryIDs_ReturnsFilteredBrands`
- ✅ `ListBrands_WithMultipleCategoryIDs_ReturnsBrandsFromAnyCategory`
- ✅ `ListBrands_WithNonExistentCategories_ReturnsEmptyList`

### Handler Layer Tests (`internal/adapters/api/handlers/brand_handler_test.go`)

Test cases:
- ✅ `ListBrands_NoFilter_Success`
- ✅ `ListBrands_SingleCategoryID_Success`
- ✅ `ListBrands_MultipleCategoryIDs_Success`
- ✅ `ListBrands_InvalidUUID_Returns400`
- ✅ `ListBrands_EmptyCategoryIDs_ReturnsAllBrands`

**Coverage Goals:**
- Service layer: 80%+
- Handler layer: 70%+

### No Repository Tests
- Repository tests are optional per project guidelines
- Query logic is straightforward Ent operations
- Focus on service and handler test coverage

## Files to Modify

1. `internal/adapters/api/handlers/brand_handler.go` - Update handler logic
2. `internal/adapters/api/handlers/brand_handler_test.go` - Add tests
3. `internal/adapters/api/dto/brand_dto.go` - Add request DTO
4. `internal/application/services/brand_service.go` - Update service method
5. `internal/application/services/brand_service_test.go` - Add tests
6. `internal/adapters/persistence/brand_repository.go` - Update repository query
7. `internal/domain/ports/brand_service.go` - Update service interface
8. `internal/domain/ports/brand_repository.go` - Update repository interface

## Implementation Notes

- Backward compatible: Existing `GET /brands` behavior unchanged
- OR logic: Brands appear if they have products in ANY of the specified categories
- Deduplication: Duplicate category_ids are handled gracefully
- Distinct results: Brands appear only once even if they have multiple matching products
- Sorting: Results ordered by brand name (ascending)
