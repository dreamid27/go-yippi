package services

import (
	"context"
	"strings"

	"example.com/go-yippi/internal/domain/entities"
	domainErrors "example.com/go-yippi/internal/domain/errors"
	"example.com/go-yippi/internal/domain/ports"
)

// ProductService handles business logic for products
type ProductService struct {
	repo ports.ProductRepository
}

func NewProductService(repo ports.ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

func (s *ProductService) CreateProduct(ctx context.Context, product *entities.Product) error {
	// Validate required fields
	if strings.TrimSpace(product.SKU) == "" {
		return domainErrors.NewValidationError("sku", "SKU is required")
	}
	if strings.TrimSpace(product.Name) == "" {
		return domainErrors.NewValidationError("name", "Name is required")
	}
	if product.Price <= 0 {
		return domainErrors.NewValidationError("price", "Price must be greater than 0")
	}

	// Auto-generate slug from name if not provided
	if strings.TrimSpace(product.Slug) == "" {
		product.Slug = entities.GenerateSlug(product.Name)
	}

	// Set default status to draft if not provided or empty
	if product.Status == "" {
		product.Status = entities.ProductStatusDraft
	}

	// Validate status
	if !product.IsValid() {
		return domainErrors.NewValidationError("status", "Invalid product status")
	}

	// Validate dimensions for courier calculation (if provided)
	if product.Weight < 0 {
		return domainErrors.NewValidationError("weight", "Weight cannot be negative")
	}
	if product.Length < 0 {
		return domainErrors.NewValidationError("length", "Length cannot be negative")
	}
	if product.Width < 0 {
		return domainErrors.NewValidationError("width", "Width cannot be negative")
	}
	if product.Height < 0 {
		return domainErrors.NewValidationError("height", "Height cannot be negative")
	}

	return s.repo.Create(ctx, product)
}

func (s *ProductService) GetProduct(ctx context.Context, id int) (*entities.Product, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ProductService) GetProductBySKU(ctx context.Context, sku string) (*entities.Product, error) {
	if strings.TrimSpace(sku) == "" {
		return nil, domainErrors.NewValidationError("sku", "SKU is required")
	}
	return s.repo.GetBySKU(ctx, sku)
}

func (s *ProductService) GetProductBySlug(ctx context.Context, slug string) (*entities.Product, error) {
	if strings.TrimSpace(slug) == "" {
		return nil, domainErrors.NewValidationError("slug", "Slug is required")
	}
	return s.repo.GetBySlug(ctx, slug)
}

func (s *ProductService) ListProducts(ctx context.Context) ([]*entities.Product, error) {
	return s.repo.List(ctx)
}

func (s *ProductService) ListPublishedProducts(ctx context.Context) ([]*entities.Product, error) {
	return s.repo.ListByStatus(ctx, entities.ProductStatusPublished)
}

func (s *ProductService) ListProductsByStatus(ctx context.Context, status entities.ProductStatus) ([]*entities.Product, error) {
	return s.repo.ListByStatus(ctx, status)
}

func (s *ProductService) UpdateProduct(ctx context.Context, product *entities.Product) error {
	// Validate required fields
	if strings.TrimSpace(product.SKU) == "" {
		return domainErrors.NewValidationError("sku", "SKU is required")
	}
	if strings.TrimSpace(product.Name) == "" {
		return domainErrors.NewValidationError("name", "Name is required")
	}
	if product.Price <= 0 {
		return domainErrors.NewValidationError("price", "Price must be greater than 0")
	}

	// Auto-generate slug from name if not provided
	if strings.TrimSpace(product.Slug) == "" {
		product.Slug = entities.GenerateSlug(product.Name)
	}

	// Set default status to draft if not provided or empty
	if product.Status == "" {
		product.Status = entities.ProductStatusDraft
	}

	// Validate status
	if !product.IsValid() {
		return domainErrors.NewValidationError("status", "Invalid product status")
	}

	// Validate dimensions for courier calculation (if provided)
	if product.Weight < 0 {
		return domainErrors.NewValidationError("weight", "Weight cannot be negative")
	}
	if product.Length < 0 {
		return domainErrors.NewValidationError("length", "Length cannot be negative")
	}
	if product.Width < 0 {
		return domainErrors.NewValidationError("width", "Width cannot be negative")
	}
	if product.Height < 0 {
		return domainErrors.NewValidationError("height", "Height cannot be negative")
	}

	return s.repo.Update(ctx, product)
}

func (s *ProductService) DeleteProduct(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

func (s *ProductService) PublishProduct(ctx context.Context, id int) error {
	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Business rule: can only publish draft products
	if product.Status != entities.ProductStatusDraft {
		return domainErrors.NewValidationError("status", "Only draft products can be published")
	}

	product.Status = entities.ProductStatusPublished
	return s.repo.Update(ctx, product)
}

func (s *ProductService) ArchiveProduct(ctx context.Context, id int) error {
	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	product.Status = entities.ProductStatusArchived
	return s.repo.Update(ctx, product)
}

// QueryProducts performs a flexible query with validation
func (s *ProductService) QueryProducts(ctx context.Context, params *entities.QueryParams) (*entities.QueryResult, error) {
	// Validate and set defaults for pagination
	if params.Pagination == nil {
		params.Pagination = &entities.PaginationParams{
			Limit:     20,
			Direction: "forward",
		}
	} else {
		if params.Pagination.Limit <= 0 {
			params.Pagination.Limit = 20
		}
		if params.Pagination.Limit > 100 {
			params.Pagination.Limit = 100
		}
		if params.Pagination.Direction == "" {
			params.Pagination.Direction = "forward"
		}
		if params.Pagination.Direction != "forward" && params.Pagination.Direction != "backward" {
			return nil, domainErrors.NewValidationError("direction", "Direction must be 'forward' or 'backward'")
		}
	}

	// Validate filters (max 10 filters)
	if len(params.Filters) > 10 {
		return nil, domainErrors.NewValidationError("filters", "Maximum 10 filters allowed")
	}

	for _, filter := range params.Filters {
		if err := s.validateFilter(filter); err != nil {
			return nil, err
		}
	}

	// Validate sort params (max 3 sorts)
	if len(params.Sort) > 3 {
		return nil, domainErrors.NewValidationError("sort", "Maximum 3 sort fields allowed")
	}

	for _, sort := range params.Sort {
		if err := s.validateSort(sort); err != nil {
			return nil, err
		}
	}

	return s.repo.Query(ctx, params)
}

// validateFilter validates a single filter
func (s *ProductService) validateFilter(filter entities.Filter) error {
	// Validate field name
	if !s.isValidFilterField(filter.Field) {
		return domainErrors.NewValidationError("filter.field", "Invalid filter field: "+filter.Field)
	}

	// Validate operator for field type
	if !s.isValidOperatorForField(filter.Field, filter.Operator) {
		return domainErrors.NewValidationError("filter.operator", "Invalid operator "+string(filter.Operator)+" for field "+filter.Field)
	}

	return nil
}

// validateSort validates a single sort parameter
func (s *ProductService) validateSort(sort entities.SortParam) error {
	if !s.isValidSortField(sort.Field) {
		return domainErrors.NewValidationError("sort.field", "Invalid sort field: "+sort.Field)
	}

	if sort.Order != entities.SortAsc && sort.Order != entities.SortDesc {
		return domainErrors.NewValidationError("sort.order", "Sort order must be 'asc' or 'desc'")
	}

	return nil
}

// isValidFilterField checks if a field name is valid for filtering
func (s *ProductService) isValidFilterField(field string) bool {
	validFields := map[string]bool{
		"id": true, "sku": true, "slug": true, "name": true,
		"description": true, "price": true, "weight": true,
		"length": true, "width": true, "height": true,
		"status": true, "created_at": true, "updated_at": true,
	}
	return validFields[field]
}

// isValidSortField checks if a field name is valid for sorting
func (s *ProductService) isValidSortField(field string) bool {
	validFields := map[string]bool{
		"id": true, "sku": true, "slug": true, "name": true,
		"price": true, "weight": true, "length": true,
		"width": true, "height": true, "status": true,
		"created_at": true, "updated_at": true,
	}
	return validFields[field]
}

// isValidOperatorForField checks if an operator is valid for a given field
func (s *ProductService) isValidOperatorForField(field string, op entities.FilterOperator) bool {
	stringFields := map[string]bool{
		"sku": true, "slug": true, "name": true, "description": true, "status": true,
	}
	numericFields := map[string]bool{
		"id": true, "price": true, "weight": true, "length": true, "width": true, "height": true,
	}
	timeFields := map[string]bool{
		"created_at": true, "updated_at": true,
	}

	// Universal operators
	if op == entities.OpIsNull || op == entities.OpIsNotNull {
		return true
	}

	// String field operators
	if stringFields[field] {
		switch op {
		case entities.OpEqual, entities.OpNotEqual, entities.OpLike, entities.OpILike,
			entities.OpIn, entities.OpNotIn, entities.OpStartsWith, entities.OpEndsWith, entities.OpContains:
			return true
		}
		return false
	}

	// Numeric field operators
	if numericFields[field] {
		switch op {
		case entities.OpEqual, entities.OpNotEqual, entities.OpGreaterThan,
			entities.OpGreaterThanOrEqual, entities.OpLessThan, entities.OpLessThanOrEqual, entities.OpIn:
			return true
		}
		return false
	}

	// Time field operators
	if timeFields[field] {
		switch op {
		case entities.OpEqual, entities.OpNotEqual, entities.OpGreaterThan,
			entities.OpGreaterThanOrEqual, entities.OpLessThan, entities.OpLessThanOrEqual:
			return true
		}
		return false
	}

	return false
}
