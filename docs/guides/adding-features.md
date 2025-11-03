# Adding New Features

This guide walks you through adding a new entity/feature to the application following hexagonal architecture principles.

## Overview

Adding a new feature involves creating components in each layer, from the inside out:
1. Domain entities and ports
2. Database schema
3. Repository implementation
4. Service implementation
5. DTOs and handlers
6. Wire dependencies

## Step-by-Step Guide

Let's add a new `Category` entity as an example.

### Step 1: Create Domain Entity

Create the pure business entity in the domain layer.

**File**: `internal/domain/entities/category.go`

```go
package entities

import "time"

type Category struct {
    ID          int
    Name        string
    Description string
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

**Rules**:
- No external dependencies
- Pure Go types
- Business-focused fields only

### Step 2: Define Repository Port

Add the repository interface to the domain ports.

**File**: `internal/domain/ports/repository.go`

```go
package ports

import (
    "context"
    "myapp/internal/domain/entities"
)

// Add to existing file or create new interface
type CategoryRepository interface {
    Create(ctx context.Context, category *entities.Category) (*entities.Category, error)
    FindByID(ctx context.Context, id int) (*entities.Category, error)
    List(ctx context.Context) ([]*entities.Category, error)
    Update(ctx context.Context, category *entities.Category) (*entities.Category, error)
    Delete(ctx context.Context, id int) error
}
```

### Step 3: Create Ent Schema

Create the database schema using Ent.

**Generate schema file**:
```bash
go run -mod=mod entgo.io/ent/cmd/ent new --target internal/adapters/persistence/db/schema Category
```

**Edit**: `internal/adapters/persistence/db/schema/category.go`

```go
package schema

import (
    "time"
    "entgo.io/ent"
    "entgo.io/ent/schema/field"
)

type Category struct {
    ent.Schema
}

func (Category) Fields() []ent.Field {
    return []ent.Field{
        field.Int("id").Unique().Immutable(),
        field.String("name").NotEmpty(),
        field.String("description").Optional(),
        field.Time("created_at").Default(time.Now).Immutable(),
        field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
    }
}
```

**Generate Ent code**:
```bash
make generate
```

This generates ORM code in `internal/adapters/persistence/db/ent/`.

### Step 4: Implement Repository

Create the repository implementation that satisfies the port.

**File**: `internal/adapters/persistence/category_repository.go`

```go
package persistence

import (
    "context"
    "myapp/internal/adapters/persistence/db/ent"
    "myapp/internal/adapters/persistence/db/ent/category"
    "myapp/internal/domain/entities"
    "myapp/internal/domain/errors"
)

type CategoryRepository struct {
    client *ent.Client
}

func NewCategoryRepository(client *ent.Client) *CategoryRepository {
    return &CategoryRepository{client: client}
}

func (r *CategoryRepository) Create(ctx context.Context, cat *entities.Category) (*entities.Category, error) {
    entCat, err := r.client.Category.Create().
        SetName(cat.Name).
        SetDescription(cat.Description).
        Save(ctx)
    if err != nil {
        return nil, err
    }
    return r.mapToEntity(entCat), nil
}

func (r *CategoryRepository) FindByID(ctx context.Context, id int) (*entities.Category, error) {
    entCat, err := r.client.Category.Get(ctx, id)
    if err != nil {
        if ent.IsNotFound(err) {
            return nil, errors.NewNotFoundError("Category", id)
        }
        return nil, err
    }
    return r.mapToEntity(entCat), nil
}

func (r *CategoryRepository) List(ctx context.Context) ([]*entities.Category, error) {
    entCats, err := r.client.Category.Query().All(ctx)
    if err != nil {
        return nil, err
    }

    categories := make([]*entities.Category, len(entCats))
    for i, ec := range entCats {
        categories[i] = r.mapToEntity(ec)
    }
    return categories, nil
}

func (r *CategoryRepository) Update(ctx context.Context, cat *entities.Category) (*entities.Category, error) {
    entCat, err := r.client.Category.UpdateOneID(cat.ID).
        SetName(cat.Name).
        SetDescription(cat.Description).
        Save(ctx)
    if err != nil {
        if ent.IsNotFound(err) {
            return nil, errors.NewNotFoundError("Category", cat.ID)
        }
        return nil, err
    }
    return r.mapToEntity(entCat), nil
}

func (r *CategoryRepository) Delete(ctx context.Context, id int) error {
    err := r.client.Category.DeleteOneID(id).Exec(ctx)
    if err != nil && ent.IsNotFound(err) {
        return errors.NewNotFoundError("Category", id)
    }
    return err
}

// Helper: Map Ent model to domain entity
func (r *CategoryRepository) mapToEntity(entCat *ent.Category) *entities.Category {
    return &entities.Category{
        ID:          entCat.ID,
        Name:        entCat.Name,
        Description: entCat.Description,
        CreatedAt:   entCat.CreatedAt,
        UpdatedAt:   entCat.UpdatedAt,
    }
}
```

**Key Points**:
- Implements `ports.CategoryRepository` interface
- Converts Ent errors to domain errors
- Maps between Ent models and domain entities

### Step 5: Create Service

Implement business logic in the application layer.

**File**: `internal/application/services/category_service.go`

```go
package services

import (
    "context"
    "myapp/internal/domain/entities"
    "myapp/internal/domain/errors"
    "myapp/internal/domain/ports"
)

type CategoryService struct {
    repo ports.CategoryRepository
}

func NewCategoryService(repo ports.CategoryRepository) *CategoryService {
    return &CategoryService{repo: repo}
}

type CreateCategoryInput struct {
    Name        string
    Description string
}

func (s *CategoryService) CreateCategory(ctx context.Context, input CreateCategoryInput) (*entities.Category, error) {
    // Business validation
    if input.Name == "" {
        return nil, errors.NewValidationError("category name is required")
    }

    category := &entities.Category{
        Name:        input.Name,
        Description: input.Description,
    }

    return s.repo.Create(ctx, category)
}

func (s *CategoryService) GetCategory(ctx context.Context, id int) (*entities.Category, error) {
    if id <= 0 {
        return nil, errors.NewValidationError("invalid category ID")
    }
    return s.repo.FindByID(ctx, id)
}

func (s *CategoryService) ListCategories(ctx context.Context) ([]*entities.Category, error) {
    return s.repo.List(ctx)
}

func (s *CategoryService) UpdateCategory(ctx context.Context, id int, input CreateCategoryInput) (*entities.Category, error) {
    if id <= 0 {
        return nil, errors.NewValidationError("invalid category ID")
    }
    if input.Name == "" {
        return nil, errors.NewValidationError("category name is required")
    }

    category := &entities.Category{
        ID:          id,
        Name:        input.Name,
        Description: input.Description,
    }

    return s.repo.Update(ctx, category)
}

func (s *CategoryService) DeleteCategory(ctx context.Context, id int) error {
    if id <= 0 {
        return errors.NewValidationError("invalid category ID")
    }
    return s.repo.Delete(ctx, id)
}
```

**Key Points**:
- Depends only on `ports.CategoryRepository` (interface)
- Contains business validation logic
- Returns domain entities

### Step 6: Create DTOs

Define request/response structures for the API.

**File**: `internal/adapters/api/dto/category_dto.go`

```go
package dto

import "time"

// Request DTOs
type CreateCategoryRequest struct {
    Body struct {
        Name        string `json:"name" minLength:"1" maxLength:"100" doc:"Category name"`
        Description string `json:"description" maxLength:"500" doc:"Category description"`
    }
}

type UpdateCategoryRequest struct {
    ID   int `path:"id" doc:"Category ID"`
    Body struct {
        Name        string `json:"name" minLength:"1" maxLength:"100" doc:"Category name"`
        Description string `json:"description" maxLength:"500" doc:"Category description"`
    }
}

type GetCategoryRequest struct {
    ID int `path:"id" doc:"Category ID"`
}

type DeleteCategoryRequest struct {
    ID int `path:"id" doc:"Category ID"`
}

// Response DTOs
type CategoryItem struct {
    ID          int       `json:"id" doc:"Category ID"`
    Name        string    `json:"name" doc:"Category name"`
    Description string    `json:"description" doc:"Category description"`
    CreatedAt   time.Time `json:"created_at" doc:"Creation timestamp"`
    UpdatedAt   time.Time `json:"updated_at" doc:"Last update timestamp"`
}

type CategoryResponse struct {
    Body CategoryItem
}

type CategoryListResponse struct {
    Body struct {
        Categories []CategoryItem `json:"categories" doc:"List of categories"`
    }
}
```

**Important**: Always use named types (like `CategoryItem`) instead of inline anonymous structs to ensure proper OpenAPI schema generation.

### Step 7: Create Handler

Implement HTTP handlers for the API.

**File**: `internal/adapters/api/handlers/category_handler.go`

```go
package handlers

import (
    "context"
    "errors"
    "myapp/internal/adapters/api/dto"
    "myapp/internal/application/services"
    "myapp/internal/domain/entities"
    domainErrors "myapp/internal/domain/errors"

    "github.com/danielgtaylor/huma/v2"
)

type CategoryHandler struct {
    service *services.CategoryService
}

func NewCategoryHandler(service *services.CategoryService) *CategoryHandler {
    return &CategoryHandler{service: service}
}

func (h *CategoryHandler) RegisterRoutes(api huma.API) {
    huma.Register(api, huma.Operation{
        OperationID: "create-category",
        Method:      "POST",
        Path:        "/categories",
        Summary:     "Create a new category",
    }, h.CreateCategory)

    huma.Register(api, huma.Operation{
        OperationID: "get-category",
        Method:      "GET",
        Path:        "/categories/{id}",
        Summary:     "Get category by ID",
    }, h.GetCategory)

    huma.Register(api, huma.Operation{
        OperationID: "list-categories",
        Method:      "GET",
        Path:        "/categories",
        Summary:     "List all categories",
    }, h.ListCategories)

    huma.Register(api, huma.Operation{
        OperationID: "update-category",
        Method:      "PUT",
        Path:        "/categories/{id}",
        Summary:     "Update category",
    }, h.UpdateCategory)

    huma.Register(api, huma.Operation{
        OperationID: "delete-category",
        Method:      "DELETE",
        Path:        "/categories/{id}",
        Summary:     "Delete category",
    }, h.DeleteCategory)
}

func (h *CategoryHandler) CreateCategory(ctx context.Context, req *dto.CreateCategoryRequest) (*dto.CategoryResponse, error) {
    input := services.CreateCategoryInput{
        Name:        req.Body.Name,
        Description: req.Body.Description,
    }

    category, err := h.service.CreateCategory(ctx, input)
    if err != nil {
        return nil, h.handleError(err)
    }

    return &dto.CategoryResponse{Body: h.mapToDTO(category)}, nil
}

func (h *CategoryHandler) GetCategory(ctx context.Context, req *dto.GetCategoryRequest) (*dto.CategoryResponse, error) {
    category, err := h.service.GetCategory(ctx, req.ID)
    if err != nil {
        return nil, h.handleError(err)
    }

    return &dto.CategoryResponse{Body: h.mapToDTO(category)}, nil
}

func (h *CategoryHandler) ListCategories(ctx context.Context, _ *struct{}) (*dto.CategoryListResponse, error) {
    categories, err := h.service.ListCategories(ctx)
    if err != nil {
        return nil, h.handleError(err)
    }

    resp := &dto.CategoryListResponse{}
    resp.Body.Categories = make([]dto.CategoryItem, len(categories))
    for i, cat := range categories {
        resp.Body.Categories[i] = h.mapToDTO(cat)
    }

    return resp, nil
}

func (h *CategoryHandler) UpdateCategory(ctx context.Context, req *dto.UpdateCategoryRequest) (*dto.CategoryResponse, error) {
    input := services.CreateCategoryInput{
        Name:        req.Body.Name,
        Description: req.Body.Description,
    }

    category, err := h.service.UpdateCategory(ctx, req.ID, input)
    if err != nil {
        return nil, h.handleError(err)
    }

    return &dto.CategoryResponse{Body: h.mapToDTO(category)}, nil
}

func (h *CategoryHandler) DeleteCategory(ctx context.Context, req *dto.DeleteCategoryRequest) (*struct{}, error) {
    err := h.service.DeleteCategory(ctx, req.ID)
    if err != nil {
        return nil, h.handleError(err)
    }

    return &struct{}{}, nil
}

// Helper: Map domain entity to DTO
func (h *CategoryHandler) mapToDTO(category *entities.Category) dto.CategoryItem {
    return dto.CategoryItem{
        ID:          category.ID,
        Name:        category.Name,
        Description: category.Description,
        CreatedAt:   category.CreatedAt,
        UpdatedAt:   category.UpdatedAt,
    }
}

// Helper: Handle domain errors
func (h *CategoryHandler) handleError(err error) error {
    if errors.Is(err, domainErrors.ErrNotFound) {
        return huma.Error404NotFound("Category not found")
    }
    if errors.Is(err, domainErrors.ErrValidation) {
        return huma.Error400BadRequest(err.Error())
    }
    return huma.Error500InternalServerError("Internal server error")
}
```

### Step 8: Wire Dependencies

Connect all components in the main application entry point.

**File**: `cmd/api/main.go`

```go
func main() {
    // ... existing setup ...

    // Create repository
    categoryRepo := persistence.NewCategoryRepository(client)

    // Create service
    categoryService := services.NewCategoryService(categoryRepo)

    // Create handler
    categoryHandler := handlers.NewCategoryHandler(categoryService)

    // Register routes
    categoryHandler.RegisterRoutes(humaAPI)

    // ... existing code ...
}
```

### Step 9: Run Migrations

Generate and apply database migrations:

```bash
make generate  # Regenerate Ent code
make run       # Run app (auto-migrates on startup)
```

### Step 10: Test the API

Access the OpenAPI documentation:
```
http://localhost:8080/docs
```

Test endpoints:
```bash
# Create category
curl -X POST http://localhost:8080/categories \
  -H "Content-Type: application/json" \
  -d '{"name":"Electronics","description":"Electronic items"}'

# Get category
curl http://localhost:8080/categories/1

# List categories
curl http://localhost:8080/categories

# Update category
curl -X PUT http://localhost:8080/categories/1 \
  -H "Content-Type: application/json" \
  -d '{"name":"Updated","description":"New description"}'

# Delete category
curl -X DELETE http://localhost:8080/categories/1
```

## Summary Checklist

- [ ] Create domain entity (`internal/domain/entities/`)
- [ ] Define repository port (`internal/domain/ports/repository.go`)
- [ ] Create Ent schema (`internal/adapters/persistence/db/schema/`)
- [ ] Run `make generate` to generate Ent code
- [ ] Implement repository (`internal/adapters/persistence/`)
- [ ] Create service (`internal/application/services/`)
- [ ] Create DTOs (`internal/adapters/api/dto/`)
- [ ] Create handler (`internal/adapters/api/handlers/`)
- [ ] Wire dependencies (`cmd/api/main.go`)
- [ ] Test endpoints

## Common Patterns

### Adding Relationships

To add relationships between entities, update Ent schemas:

```go
// In Product schema
func (Product) Edges() []ent.Edge {
    return []ent.Edge{
        edge.From("category", Category.Type).
            Ref("products").
            Unique(),
    }
}

// In Category schema
func (Category) Edges() []ent.Edge {
    return []ent.Edge{
        edge.To("products", Product.Type),
    }
}
```

### Adding Pagination

See [Pagination Guide](../architecture/pagination-filter-sort.md) for cursor-based pagination implementation.

### Adding Validation

Business validation goes in services, request validation in DTOs:

```go
// DTO validation (automatic via Huma)
type CreateCategoryRequest struct {
    Body struct {
        Name string `json:"name" minLength:"1" maxLength:"100"`
    }
}

// Service validation (business rules)
func (s *CategoryService) CreateCategory(ctx context.Context, input CreateCategoryInput) (*entities.Category, error) {
    if input.Name == "" {
        return nil, errors.NewValidationError("name required")
    }
    // ...
}
```

## Related Documentation

- [Architecture Overview](../architecture/overview.md)
- [Dependency Flow](../architecture/dependency-flow.md)
- [Hexagonal Pattern](../architecture/hexagonal-pattern.md)
