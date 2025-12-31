package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"example.com/go-yippi/internal/domain/entities"
	domainErrors "example.com/go-yippi/internal/domain/errors"
	"example.com/go-yippi/internal/domain/ports"
	"example.com/go-yippi/internal/infrastructure/observability"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ProductService handles business logic for products
type ProductService struct {
	repo         ports.ProductRepository
	categoryRepo ports.CategoryRepository
	observer     *observability.ServiceObserver
}

func NewProductService(repo ports.ProductRepository, categoryRepo ports.CategoryRepository) *ProductService {
	return &ProductService{
		repo:         repo,
		categoryRepo: categoryRepo,
		observer:     observability.NewServiceObserver("ProductService"),
	}
}

func (s *ProductService) CreateProduct(ctx context.Context, product *entities.Product) error {
	// Start observable operation
	op := s.observer.StartOperation(ctx, "CreateProduct")
	defer op.End(nil)

	// Add input attributes for tracing
	op.AddAttribute("name", product.Name)
	op.AddAttribute("base_price", fmt.Sprintf("%.2f", product.BasePrice))

	// Validate required fields
	if strings.TrimSpace(product.Name) == "" {
		err := domainErrors.NewValidationError("name", "Name is required")
		observability.LogValidationError(op.Context(), "name", "Name is required")
		op.End(err)
		return err
	}
	if product.BasePrice <= 0 {
		err := domainErrors.NewValidationError("base_price", "Base price must be greater than 0")
		observability.LogValidationError(op.Context(), "base_price", "Base price must be greater than 0")
		op.End(err)
		return err
	}

	// Auto-generate slug from name if not provided
	if strings.TrimSpace(product.Slug) == "" {
		product.Slug = entities.GenerateSlug(product.Name)
		op.LogInfo("Auto-generated slug from product name",
			zap.String("slug", product.Slug))
	}

	// Set default status to draft if not provided or empty
	if product.Status == "" {
		product.Status = entities.ProductStatusDraft
		op.LogInfo("Set default product status to draft")
	}

	// Validate status
	if !product.IsValid() {
		err := domainErrors.NewValidationError("status", "Invalid product status")
		observability.LogValidationError(op.Context(), "status", "Invalid product status")
		op.End(err)
		return err
	}

	// Validate dimensions for courier calculation (if provided)
	if product.Weight < 0 {
		err := domainErrors.NewValidationError("weight", "Weight cannot be negative")
		observability.LogValidationError(op.Context(), "weight", "Weight cannot be negative")
		op.End(err)
		return err
	}
	if product.Length < 0 {
		err := domainErrors.NewValidationError("length", "Length cannot be negative")
		observability.LogValidationError(op.Context(), "length", "Length cannot be negative")
		op.End(err)
		return err
	}
	if product.Width < 0 {
		err := domainErrors.NewValidationError("width", "Width cannot be negative")
		observability.LogValidationError(op.Context(), "width", "Width cannot be negative")
		op.End(err)
		return err
	}
	if product.Height < 0 {
		err := domainErrors.NewValidationError("height", "Height cannot be negative")
		observability.LogValidationError(op.Context(), "height", "Height cannot be negative")
		op.End(err)
		return err
	}

	// Create product in repository
	startTime := time.Now()
	err := s.repo.Create(op.Context(), product)

	// Log slow operation if it takes too long
	observability.LogSlowOperation(op.Context(), "CreateProduct", time.Since(startTime), 100*time.Millisecond)

	if err != nil {
		op.End(err)
		return err
	}

	// Record business event
	op.RecordBusinessEvent("product", "product_created",
		zap.String("product_id", product.ID.String()),
		zap.String("status", string(product.Status)),
	)

	op.End(nil)
	return nil
}

func (s *ProductService) GetProduct(ctx context.Context, id uuid.UUID) (*entities.Product, error) {
	// Start observable operation
	op := s.observer.StartOperation(ctx, "GetProduct")
	defer func() { op.End(nil) }()

	op.AddAttribute("product_id", id.String())

	product, err := s.repo.GetByID(op.Context(), id)
	if err != nil {
		observability.LogNotFoundError(op.Context(), "Product", id)
		op.End(err)
		return nil, err
	}

	op.End(nil)
	return product, nil
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
	if strings.TrimSpace(product.Name) == "" {
		return domainErrors.NewValidationError("name", "Name is required")
	}
	if product.BasePrice <= 0 {
		return domainErrors.NewValidationError("base_price", "Base price must be greater than 0")
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

func (s *ProductService) DeleteProduct(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *ProductService) PublishProduct(ctx context.Context, id uuid.UUID) error {
	// Start observable operation
	op := s.observer.StartOperation(ctx, "PublishProduct")
	defer func() { op.End(nil) }()

	op.AddAttribute("product_id", id)

	product, err := s.repo.GetByID(op.Context(), id)
	if err != nil {
		observability.LogNotFoundError(op.Context(), "Product", id)
		op.End(err)
		return err
	}

	op.AddAttribute("current_status", string(product.Status))

	// Business rule: can only publish draft products
	if product.Status != entities.ProductStatusDraft {
		err := domainErrors.NewValidationError("status", "Only draft products can be published")
		observability.LogBusinessRule(op.Context(), "publish_draft_only",
			fmt.Sprintf("Cannot publish product with status: %s", product.Status))
		op.End(err)
		return err
	}

	oldStatus := product.Status
	product.Status = entities.ProductStatusPublished

	err = s.repo.Update(op.Context(), product)
	if err != nil {
		op.End(err)
		return err
	}

	// Record business event
	op.RecordBusinessEvent("product", "product_published",
		zap.String("product_id", id.String()),
		zap.String("old_status", string(oldStatus)),
		zap.String("new_status", string(product.Status)),
	)

	op.LogInfo("Product status changed",
		zap.String("from_status", string(oldStatus)),
		zap.String("to_status", string(product.Status)))

	op.End(nil)
	return nil
}

func (s *ProductService) ArchiveProduct(ctx context.Context, id uuid.UUID) error{
	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	product.Status = entities.ProductStatusArchived
	return s.repo.Update(ctx, product)
}

// QueryProducts performs a flexible query with validation
func (s *ProductService) QueryProducts(ctx context.Context, params *entities.QueryParams) (*entities.QueryResult, error) {
	// Start observable operation
	op := s.observer.StartOperation(ctx, "QueryProducts")
	defer func() { op.End(nil) }()

	op.AddAttribute("filters_count", len(params.Filters))
	op.AddAttribute("sort_count", len(params.Sort))

	// Validate and set defaults for pagination
	if params.Pagination == nil {
		params.Pagination = &entities.PaginationParams{
			Limit:     20,
			Direction: "forward",
		}
		op.LogInfo("Using default pagination parameters")
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
			err := domainErrors.NewValidationError("direction", "Direction must be 'forward' or 'backward'")
			observability.LogValidationError(op.Context(), "direction", "Direction must be 'forward' or 'backward'")
			op.End(err)
			return nil, err
		}
	}

	op.AddAttribute("pagination_limit", params.Pagination.Limit)
	op.AddAttribute("pagination_direction", params.Pagination.Direction)

	// Validate filters (max 10 filters)
	if len(params.Filters) > 10 {
		err := domainErrors.NewValidationError("filters", "Maximum 10 filters allowed")
		observability.LogValidationError(op.Context(), "filters", "Maximum 10 filters allowed")
		op.End(err)
		return nil, err
	}

	for _, filter := range params.Filters {
		if err := s.validateFilter(filter); err != nil {
			observability.LogValidationError(op.Context(), "filter", err.Error())
			op.End(err)
			return nil, err
		}
	}

	// Validate sort params (max 3 sorts)
	if len(params.Sort) > 3 {
		err := domainErrors.NewValidationError("sort", "Maximum 3 sort fields allowed")
		observability.LogValidationError(op.Context(), "sort", "Maximum 3 sort fields allowed")
		op.End(err)
		return nil, err
	}

	for _, sort := range params.Sort {
		if err := s.validateSort(sort); err != nil {
			observability.LogValidationError(op.Context(), "sort", err.Error())
			op.End(err)
			return nil, err
		}
	}

	// Expand category_id filters to include descendants
	for i, filter := range params.Filters {
		if filter.Field == "category_id" && (filter.Operator == entities.OpIn || filter.Operator == entities.OpEqual) {
			// Extract category IDs from filter value
			var categoryIDs []uuid.UUID

			if filter.Operator == entities.OpEqual {
				// Single ID: parse UUID string
				switch v := filter.Value.(type) {
				case string:
					parsedID, err := uuid.Parse(v)
					if err != nil {
						return nil, domainErrors.NewValidationError("category_id", "Invalid UUID format")
					}
					categoryIDs = []uuid.UUID{parsedID}
				default:
					return nil, domainErrors.NewValidationError("category_id", "Invalid category_id value type (expected string UUID)")
				}
			} else if filter.Operator == entities.OpIn {
				// Multiple IDs: convert array to UUID slice
				switch v := filter.Value.(type) {
				case []interface{}:
					categoryIDs = make([]uuid.UUID, 0, len(v))
					for _, id := range v {
						switch idVal := id.(type) {
						case string:
							parsedID, err := uuid.Parse(idVal)
							if err != nil {
								return nil, domainErrors.NewValidationError("category_id", "Invalid UUID format in array")
							}
							categoryIDs = append(categoryIDs, parsedID)
						default:
							return nil, domainErrors.NewValidationError("category_id", "Invalid category_id array value type (expected string UUID)")
						}
					}
				default:
					return nil, domainErrors.NewValidationError("category_id", "Invalid category_id value type for 'in' operator")
				}
			}

			// Expand to include all descendants
			if len(categoryIDs) > 0 {
				expandedIDs, err := s.categoryRepo.GetDescendantIDs(op.Context(), categoryIDs)
				if err != nil {
					err := domainErrors.NewValidationError("category_id", "Failed to expand category IDs")
					op.End(err)
					return nil, err
				}

				// Log category expansion
				if len(expandedIDs) > len(categoryIDs) {
					op.LogInfo("Expanded category filter to include descendants",
						zap.Int("original_count", len(categoryIDs)),
						zap.Int("expanded_count", len(expandedIDs)))
				}

				// Update the filter with expanded IDs (convert UUID back to string)
				if len(expandedIDs) == 1 {
					params.Filters[i].Value = expandedIDs[0].String()
					params.Filters[i].Operator = entities.OpEqual
				} else {
					// Convert to []interface{} for compatibility
					expandedValues := make([]interface{}, len(expandedIDs))
					for j, id := range expandedIDs {
						expandedValues[j] = id.String()
					}
					params.Filters[i].Value = expandedValues
					params.Filters[i].Operator = entities.OpIn
				}
			}
		}
	}

	// Execute query
	startTime := time.Now()
	result, err := s.repo.Query(op.Context(), params)

	// Log slow query
	observability.LogSlowOperation(op.Context(), "QueryProducts", time.Since(startTime), 200*time.Millisecond)

	if err != nil {
		op.End(err)
		return nil, err
	}

	// Log query results
	op.AddAttribute("results_count", len(result.Products))
	op.AddAttribute("has_next", result.PageInfo.HasNextPage)
	op.AddAttribute("has_previous", result.PageInfo.HasPreviousPage)

	op.LogInfo("Query completed successfully",
		zap.Int("results_count", len(result.Products)),
		zap.Bool("has_next", result.PageInfo.HasNextPage),
		zap.Bool("has_previous", result.PageInfo.HasPreviousPage))

	op.End(nil)
	return result, nil
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
		"status": true, "category_id": true, "brand_id": true,
		"created_at": true, "updated_at": true,
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
	uuidFields := map[string]bool{
		"category_id": true, "brand_id": true,
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

	// UUID field operators (category_id, brand_id)
	if uuidFields[field] {
		switch op {
		case entities.OpEqual, entities.OpNotEqual, entities.OpIn, entities.OpNotIn:
			return true
		}
		return false
	}

	return false
}
