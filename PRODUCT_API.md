# Product API Documentation

## Overview

The Product API provides endpoints for managing e-commerce products with complete hexagonal architecture implementation.

## Product Fields

- **SKU** (string, required, unique): Stock Keeping Unit
- **Slug** (string, required, unique): URL-friendly identifier
- **Name** (string, required): Product name
- **Price** (float, required): Product price (must be > 0)
- **Description** (text, optional): Product description
- **Weight** (int): Weight in grams for courier calculation
- **Length** (int): Length in cm for courier calculation
- **Width** (int): Width in cm for courier calculation
- **Height** (int): Height in cm for courier calculation (ukuran/dimensions)
- **Status** (enum): Product status (draft, published, archived)
- **CreatedAt** (timestamp): Creation timestamp
- **UpdatedAt** (timestamp): Last update timestamp

## Available Endpoints

### 1. Create Product
**POST** `/products`

Creates a new product with draft status by default.

**Request Body:**
```json
{
  "sku": "PROD-001",
  "slug": "amazing-product",
  "name": "Amazing Product",
  "price": 99.99,
  "description": "This is an amazing product",
  "weight": 500,
  "length": 20,
  "width": 15,
  "height": 10,
  "status": "draft"
}
```

**Response:** `200 OK`
```json
{
  "id": 1,
  "sku": "PROD-001",
  "slug": "amazing-product",
  "name": "Amazing Product",
  "price": 99.99,
  "description": "This is an amazing product",
  "weight": 500,
  "length": 20,
  "width": 15,
  "height": 10,
  "status": "draft",
  "created_at": "2025-11-01T10:00:00Z",
  "updated_at": "2025-11-01T10:00:00Z"
}
```

### 2. List All Products
**GET** `/products`

Returns all products regardless of status.

**Response:** `200 OK`

### 3. Get Product by ID
**GET** `/products/{id}`

Retrieves a specific product by its ID.

**Response:** `200 OK` or `404 Not Found`

### 4. Get Product by SKU
**GET** `/products/sku/{sku}`

Retrieves a product by its Stock Keeping Unit.

**Response:** `200 OK` or `404 Not Found`

### 5. Get Product by Slug
**GET** `/products/slug/{slug}`

Retrieves a product by its URL-friendly slug.

**Response:** `200 OK` or `404 Not Found`

### 6. List Products by Status
**GET** `/products/status/{status}`

Retrieves products filtered by status (draft, published, or archived).

**Response:** `200 OK`

### 7. Update Product
**PUT** `/products/{id}`

Updates an existing product. All fields must be provided.

**Request Body:** Same as Create Product

**Response:** `200 OK`, `404 Not Found`, or `409 Conflict` (if SKU/slug already exists)

### 8. Publish Product
**POST** `/products/{id}/publish`

Changes product status from draft to published. Only draft products can be published.

**Response:** `200 OK` or `400 Bad Request` (if not in draft status)

### 9. Archive Product
**POST** `/products/{id}/archive`

Changes product status to archived. Can be called from any status.

**Response:** `200 OK` or `404 Not Found`

### 10. Delete Product
**DELETE** `/products/{id}`

Permanently deletes a product from the system.

**Response:** `204 No Content` or `404 Not Found`

## Business Rules

1. **SKU Uniqueness**: Each product must have a unique SKU
2. **Slug Uniqueness**: Each product must have a unique slug
3. **Price Validation**: Price must be greater than 0
4. **Publishing Rule**: Only products with "draft" status can be published
5. **Weight/Dimensions**: Used for courier/shipping calculations (Indonesian e-commerce standard)

## Error Handling

The API uses domain-specific errors that are properly converted to HTTP status codes:

- `400 Bad Request`: Invalid input or validation errors
- `404 Not Found`: Product not found
- `409 Conflict`: Duplicate SKU or slug
- `500 Internal Server Error`: Unexpected server errors

## Architecture Implementation

The Product API follows hexagonal architecture:

1. **Domain Layer**: `internal/domain/entities/product.go`, `internal/domain/ports/repository.go`
2. **Application Layer**: `internal/application/services/product_service.go`
3. **Infrastructure Layer**: `internal/infrastructure/adapters/persistence/product_repository.go`
4. **API Layer**: `internal/api/handlers/product_handler.go`, `internal/api/dto/product_dto.go`

## Testing the API

Once the server is running, access the interactive API documentation at:
```
http://localhost:8080/docs
```

You can test all endpoints directly from the Swagger UI.

## Example Workflow

1. **Create a product** with status "draft"
2. **Update product details** as needed
3. **Publish the product** when ready to go live
4. **Customers can view** published products (filter by status=published)
5. **Archive old products** instead of deleting them
