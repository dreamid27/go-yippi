package persistence

import (
	"context"
	"fmt"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	"example.com/go-yippi/internal/adapters/persistence/db/ent"
	"example.com/go-yippi/internal/adapters/persistence/db/ent/brand"
	"example.com/go-yippi/internal/adapters/persistence/db/ent/category"
	"example.com/go-yippi/internal/adapters/persistence/db/ent/predicate"
	"example.com/go-yippi/internal/adapters/persistence/db/ent/product"
	"example.com/go-yippi/internal/domain/entities"
	"github.com/google/uuid"
)

// Query performs a flexible query with filters, sorting, and pagination
func (r *ProductRepositoryImpl) Query(ctx context.Context, params *entities.QueryParams) (*entities.QueryResult, error) {
	query := r.client.Product.Query()

	// Apply filters
	if len(params.Filters) > 0 {
		predicates, err := r.buildFilterPredicates(params.Filters)
		if err != nil {
			return nil, fmt.Errorf("failed to build filter predicates: %w", err)
		}

		query = query.Where(predicates...)
	}

	// Apply sorting (default: created_at desc, id desc)
	query = r.applySorting(query, params.Sort)

	// Apply pagination
	var limit int
	var cursor *entities.Cursor
	var err error

	if params.Pagination != nil {
		limit = params.Pagination.Limit
		if params.Pagination.Cursor != nil {
			cursor, err = DecodeCursor(*params.Pagination.Cursor)
			if err != nil {
				return nil, fmt.Errorf("invalid cursor: %w", err)
			}
		}
	} else {
		limit = 20 // default
	}

	// Fetch limit + 1 to determine if there's a next page
	query = query.Limit(limit + 1)


	fmt.Println("cursor")
	fmt.Println(cursor)

	// Apply cursor pagination
	if cursor != nil {
		fmt.Println("masuk donks")
		query = r.applyCursor(query, cursor, params.Pagination.Direction, params.Sort)
	}

	// Execute query
	products, err := query.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	// Build result
	hasNextPage := len(products) > limit
	if hasNextPage {
		products = products[:limit] // Trim to actual limit
	}

	// Convert to domain entities
	domainProducts := make([]*entities.Product, len(products))
	for i, p := range products {
		domainProducts[i] = r.toEntity(p)
	}

	// Build page info
	pageInfo := r.buildPageInfo(products, hasNextPage, params.Pagination)

	return &entities.QueryResult{
		Products: domainProducts,
		PageInfo: pageInfo,
	}, nil
}

// buildFilterPredicates builds Ent predicates from filter parameters
func (r *ProductRepositoryImpl) buildFilterPredicates(filters []entities.Filter) ([]predicate.Product, error) {
	predicates := make([]predicate.Product, 0, len(filters))

	for _, filter := range filters {
		pred, err := r.buildSingleFilterPredicate(filter)
		if err != nil {
			return nil, err
		}
		predicates = append(predicates, pred)
	}

	return predicates, nil
}

// buildSingleFilterPredicate builds a single Ent predicate from a filter
func (r *ProductRepositoryImpl) buildSingleFilterPredicate(filter entities.Filter) (predicate.Product, error) {
	switch filter.Field {
	case "id":
		return r.buildIntFilter(filter, product.ID)
	case "sku":
		return r.buildStringFilter(filter, product.Sku)
	case "slug":
		return r.buildStringFilter(filter, product.Slug)
	case "name":
		return r.buildStringFilter(filter, product.Name)
	case "description":
		return r.buildStringFilter(filter, product.Description)
	case "price":
		return r.buildFloatFilter(filter, product.Price)
	case "weight":
		return r.buildIntFilter(filter, product.Weight)
	case "length":
		return r.buildIntFilter(filter, product.Length)
	case "width":
		return r.buildIntFilter(filter, product.Width)
	case "height":
		return r.buildIntFilter(filter, product.Height)
	case "status":
		return r.buildStatusFilter(filter)
	case "category_id":
		return r.buildCategoryFilter(filter)
	case "brand_id":
		return r.buildBrandIDFilter(filter)
	case "created_at":
		return r.buildTimeFilter(filter, product.CreatedAt)
	case "updated_at":
		return r.buildTimeFilter(filter, product.UpdatedAt)
	default:
		return nil, fmt.Errorf("unsupported filter field: %s", filter.Field)
	}
}

// buildIntFilter builds predicates for integer fields
func (r *ProductRepositoryImpl) buildIntFilter(filter entities.Filter, fieldFunc func(int) predicate.Product) (predicate.Product, error) {
	val, ok := filter.Value.(float64) // JSON numbers are float64
	if !ok {
		return nil, fmt.Errorf("invalid value type for int filter: %T", filter.Value)
	}
	intVal := int(val)

	switch filter.Operator {
	case entities.OpEqual:
		return fieldFunc(intVal), nil
	case entities.OpNotEqual:
		return product.Not(fieldFunc(intVal)), nil
	case entities.OpGreaterThan:
		return func(s *sql.Selector) {
			s.Where(sql.GT(filter.Field, intVal))
		}, nil
	case entities.OpGreaterThanOrEqual:
		return func(s *sql.Selector) {
			s.Where(sql.GTE(filter.Field, intVal))
		}, nil
	case entities.OpLessThan:
		return func(s *sql.Selector) {
			s.Where(sql.LT(filter.Field, intVal))
		}, nil
	case entities.OpLessThanOrEqual:
		return func(s *sql.Selector) {
			s.Where(sql.LTE(filter.Field, intVal))
		}, nil
	case entities.OpIn:
		vals, ok := filter.Value.([]interface{})
		if !ok {
			return nil, fmt.Errorf("in operator requires array value")
		}
		intVals := make([]int, len(vals))
		for i, v := range vals {
			f, ok := v.(float64)
			if !ok {
				return nil, fmt.Errorf("invalid value in array: %T", v)
			}
			intVals[i] = int(f)
		}
		return product.IDIn(intVals...), nil
	default:
		return nil, fmt.Errorf("unsupported operator %s for int field", filter.Operator)
	}
}

// buildFloatFilter builds predicates for float fields
func (r *ProductRepositoryImpl) buildFloatFilter(filter entities.Filter, fieldFunc func(float64) predicate.Product) (predicate.Product, error) {
	val, ok := filter.Value.(float64)
	if !ok {
		return nil, fmt.Errorf("invalid value type for float filter: %T", filter.Value)
	}

	switch filter.Operator {
	case entities.OpEqual:
		return fieldFunc(val), nil
	case entities.OpNotEqual:
		return product.Not(fieldFunc(val)), nil
	case entities.OpGreaterThan:
		return product.PriceGT(val), nil
	case entities.OpGreaterThanOrEqual:
		return product.PriceGTE(val), nil
	case entities.OpLessThan:
		return product.PriceLT(val), nil
	case entities.OpLessThanOrEqual:
		return product.PriceLTE(val), nil
	default:
		return nil, fmt.Errorf("unsupported operator %s for float field", filter.Operator)
	}
}

// buildStringFilter builds predicates for string fields
func (r *ProductRepositoryImpl) buildStringFilter(filter entities.Filter, fieldFunc func(string) predicate.Product) (predicate.Product, error) {
	val, ok := filter.Value.(string)
	if !ok {
		return nil, fmt.Errorf("invalid value type for string filter: %T", filter.Value)
	}

	switch filter.Operator {
	case entities.OpEqual:
		return fieldFunc(val), nil
	case entities.OpNotEqual:
		return product.Not(fieldFunc(val)), nil
	case entities.OpLike:
		return func(s *sql.Selector) {
			s.Where(sql.Like(filter.Field, val))
		}, nil
	case entities.OpILike:
		return func(s *sql.Selector) {
			s.Where(sql.Like(sql.Lower(filter.Field), strings.ToLower(val)))
		}, nil
	case entities.OpStartsWith:
		return func(s *sql.Selector) {
			s.Where(sql.HasPrefix(filter.Field, val))
		}, nil
	case entities.OpEndsWith:
		return func(s *sql.Selector) {
			s.Where(sql.HasSuffix(filter.Field, val))
		}, nil
	case entities.OpIn:
		vals, ok := filter.Value.([]interface{})
		if !ok {
			// Try comma-separated string
			strVals := strings.Split(val, ",")
			anyVals := make([]any, len(strVals))
			for i, v := range strVals {
				anyVals[i] = strings.TrimSpace(v)
			}
			return func(s *sql.Selector) {
				s.Where(sql.In(filter.Field, anyVals...))
			}, nil
		}
		anyVals := make([]any, len(vals))
		for i, v := range vals {
			str, ok := v.(string)
			if !ok {
				return nil, fmt.Errorf("invalid value in array: %T", v)
			}
			anyVals[i] = str
		}
		return func(s *sql.Selector) {
			s.Where(sql.In(filter.Field, anyVals...))
		}, nil
	default:
		return nil, fmt.Errorf("unsupported operator %s for string field", filter.Operator)
	}
}

// buildStatusFilter builds predicates for status field
func (r *ProductRepositoryImpl) buildStatusFilter(filter entities.Filter) (predicate.Product, error) {
	switch filter.Operator {
	case entities.OpEqual:
		val, ok := filter.Value.(string)
		if !ok {
			return nil, fmt.Errorf("invalid value type for status filter: %T", filter.Value)
		}
		return product.StatusEQ(product.Status(val)), nil
	case entities.OpNotEqual:
		val, ok := filter.Value.(string)
		if !ok {
			return nil, fmt.Errorf("invalid value type for status filter: %T", filter.Value)
		}
		return product.StatusNEQ(product.Status(val)), nil
	case entities.OpIn:
		vals, ok := filter.Value.([]interface{})
		if !ok {
			// Try comma-separated string
			val, ok := filter.Value.(string)
			if !ok {
				return nil, fmt.Errorf("invalid value type for status in filter: %T", filter.Value)
			}
			strVals := strings.Split(val, ",")
			statuses := make([]product.Status, len(strVals))
			for i, s := range strVals {
				statuses[i] = product.Status(strings.TrimSpace(s))
			}
			return product.StatusIn(statuses...), nil
		}
		statuses := make([]product.Status, len(vals))
		for i, v := range vals {
			str, ok := v.(string)
			if !ok {
				return nil, fmt.Errorf("invalid value in array: %T", v)
			}
			statuses[i] = product.Status(str)
		}
		return product.StatusIn(statuses...), nil
	default:
		return nil, fmt.Errorf("unsupported operator %s for status field", filter.Operator)
	}
}

// buildCategoryFilter builds predicates for category_id field (UUID)
func (r *ProductRepositoryImpl) buildCategoryFilter(filter entities.Filter) (predicate.Product, error) {
	switch filter.Operator {
	case entities.OpEqual:
		val, ok := filter.Value.(string)
		if !ok {
			return nil, fmt.Errorf("invalid value type for category_id filter: %T", filter.Value)
		}
		categoryID, err := uuid.Parse(val)
		if err != nil {
			return nil, fmt.Errorf("invalid UUID for category_id: %w", err)
		}
		return product.HasCategoryWith(category.IDEQ(categoryID)), nil
	case entities.OpNotEqual:
		val, ok := filter.Value.(string)
		if !ok {
			return nil, fmt.Errorf("invalid value type for category_id filter: %T", filter.Value)
		}
		categoryID, err := uuid.Parse(val)
		if err != nil {
			return nil, fmt.Errorf("invalid UUID for category_id: %w", err)
		}
		return product.HasCategoryWith(category.IDNEQ(categoryID)), nil
	case entities.OpIn:
		vals, ok := filter.Value.([]interface{})
		if !ok {
			// Try comma-separated string
			val, ok := filter.Value.(string)
			if !ok {
				return nil, fmt.Errorf("invalid value type for category_id in filter: %T", filter.Value)
			}
			strVals := strings.Split(val, ",")
			categoryIDs := make([]uuid.UUID, 0, len(strVals))
			for _, s := range strVals {
				categoryID, err := uuid.Parse(strings.TrimSpace(s))
				if err != nil {
					// Skip invalid UUIDs
					continue
				}
				categoryIDs = append(categoryIDs, categoryID)
			}
			if len(categoryIDs) == 0 {
				return nil, fmt.Errorf("no valid UUIDs found in category_ids filter")
			}
			return product.HasCategoryWith(category.IDIn(categoryIDs...)), nil
		}
		categoryIDs := make([]uuid.UUID, 0, len(vals))
		for _, v := range vals {
			str, ok := v.(string)
			if !ok {
				return nil, fmt.Errorf("invalid value in array: %T", v)
			}
			categoryID, err := uuid.Parse(str)
			if err != nil {
				continue // Skip invalid UUIDs
			}
			categoryIDs = append(categoryIDs, categoryID)
		}
		if len(categoryIDs) == 0 {
			return nil, fmt.Errorf("no valid UUIDs found in category_ids filter")
		}
		return product.HasCategoryWith(category.IDIn(categoryIDs...)), nil
	default:
		return nil, fmt.Errorf("unsupported operator %s for category_id field", filter.Operator)
	}
}

// buildBrandIDFilter builds predicates for brand_id field (UUID)
func (r *ProductRepositoryImpl) buildBrandIDFilter(filter entities.Filter) (predicate.Product, error) {
	switch filter.Operator {
	case entities.OpEqual:
		val, ok := filter.Value.(string)
		if !ok {
			return nil, fmt.Errorf("invalid value type for brand_id filter: %T", filter.Value)
		}
		brandID, err := uuid.Parse(val)
		if err != nil {
			return nil, fmt.Errorf("invalid UUID for brand_id: %w", err)
		}
		return product.HasBrandWith(brand.IDEQ(brandID)), nil
	case entities.OpNotEqual:
		val, ok := filter.Value.(string)
		if !ok {
			return nil, fmt.Errorf("invalid value type for brand_id filter: %T", filter.Value)
		}
		brandID, err := uuid.Parse(val)
		if err != nil {
			return nil, fmt.Errorf("invalid UUID for brand_id: %w", err)
		}
		return product.HasBrandWith(brand.IDNEQ(brandID)), nil
	case entities.OpIn:
		vals, ok := filter.Value.([]interface{})
		if !ok {
			// Try comma-separated string
			val, ok := filter.Value.(string)
			if !ok {
				return nil, fmt.Errorf("invalid value type for brand_id in filter: %T", filter.Value)
			}
			strVals := strings.Split(val, ",")
			brandIDs := make([]uuid.UUID, 0, len(strVals))
			for _, s := range strVals {
				brandID, err := uuid.Parse(strings.TrimSpace(s))
				if err != nil {
					// Skip invalid UUIDs
					continue
				}
				brandIDs = append(brandIDs, brandID)
			}
			if len(brandIDs) == 0 {
				return nil, fmt.Errorf("no valid UUIDs found in brand_ids filter")
			}
			return product.HasBrandWith(brand.IDIn(brandIDs...)), nil
		}
		brandIDs := make([]uuid.UUID, 0, len(vals))
		for _, v := range vals {
			str, ok := v.(string)
			if !ok {
				return nil, fmt.Errorf("invalid value in array: %T", v)
			}
			brandID, err := uuid.Parse(str)
			if err != nil {
				continue // Skip invalid UUIDs
			}
			brandIDs = append(brandIDs, brandID)
		}
		if len(brandIDs) == 0 {
			return nil, fmt.Errorf("no valid UUIDs found in brand_ids filter")
		}
		return product.HasBrandWith(brand.IDIn(brandIDs...)), nil
	default:
		return nil, fmt.Errorf("unsupported operator %s for brand_id field", filter.Operator)
	}
}

// buildTimeFilter builds predicates for time fields
func (r *ProductRepositoryImpl) buildTimeFilter(filter entities.Filter, fieldFunc func(time.Time) predicate.Product) (predicate.Product, error) {
	val, ok := filter.Value.(string)
	if !ok {
		return nil, fmt.Errorf("invalid value type for time filter: %T", filter.Value)
	}

	t, err := time.Parse(time.RFC3339, val)
	if err != nil {
		return nil, fmt.Errorf("invalid time format: %w", err)
	}

	switch filter.Operator {
	case entities.OpEqual:
		return fieldFunc(t), nil
	case entities.OpNotEqual:
		return product.Not(fieldFunc(t)), nil
	case entities.OpGreaterThan:
		return func(s *sql.Selector) {
			s.Where(sql.GT(filter.Field, t))
		}, nil
	case entities.OpGreaterThanOrEqual:
		return func(s *sql.Selector) {
			s.Where(sql.GTE(filter.Field, t))
		}, nil
	case entities.OpLessThan:
		return func(s *sql.Selector) {
			s.Where(sql.LT(filter.Field, t))
		}, nil
	case entities.OpLessThanOrEqual:
		return func(s *sql.Selector) {
			s.Where(sql.LTE(filter.Field, t))
		}, nil
	default:
		return nil, fmt.Errorf("unsupported operator %s for time field", filter.Operator)
	}
}

// applySorting applies sorting to the query
func (r *ProductRepositoryImpl) applySorting(query *ent.ProductQuery, sortParams []entities.SortParam) *ent.ProductQuery {
	if len(sortParams) == 0 {
		// Default sorting
		return query.Order(product.ByCreatedAt(sql.OrderDesc()), product.ByID(sql.OrderDesc()))
	}

	orderFuncs := make([]product.OrderOption, 0, len(sortParams)+1)
	for _, sort := range sortParams {
		orderFunc := r.getSortOrderFunc(sort.Field, sort.Order)
		if orderFunc != nil {
			orderFuncs = append(orderFuncs, orderFunc)
		}
	}

	// Always add ID as final sort for stable ordering
	orderFuncs = append(orderFuncs, product.ByID(sql.OrderDesc()))

	return query.Order(orderFuncs...)
}

// getSortOrderFunc returns the appropriate order function for a field
func (r *ProductRepositoryImpl) getSortOrderFunc(field string, order entities.SortOrder) product.OrderOption {
	desc := order == entities.SortDesc

	switch field {
	case "id":
		if desc {
			return product.ByID(sql.OrderDesc())
		}
		return product.ByID(sql.OrderAsc())
	case "sku":
		if desc {
			return product.BySku(sql.OrderDesc())
		}
		return product.BySku(sql.OrderAsc())
	case "slug":
		if desc {
			return product.BySlug(sql.OrderDesc())
		}
		return product.BySlug(sql.OrderAsc())
	case "name":
		if desc {
			return product.ByName(sql.OrderDesc())
		}
		return product.ByName(sql.OrderAsc())
	case "price":
		if desc {
			return product.ByPrice(sql.OrderDesc())
		}
		return product.ByPrice(sql.OrderAsc())
	case "weight":
		if desc {
			return product.ByWeight(sql.OrderDesc())
		}
		return product.ByWeight(sql.OrderAsc())
	case "length":
		if desc {
			return product.ByLength(sql.OrderDesc())
		}
		return product.ByLength(sql.OrderAsc())
	case "width":
		if desc {
			return product.ByWidth(sql.OrderDesc())
		}
		return product.ByWidth(sql.OrderAsc())
	case "height":
		if desc {
			return product.ByHeight(sql.OrderDesc())
		}
		return product.ByHeight(sql.OrderAsc())
	case "status":
		if desc {
			return product.ByStatus(sql.OrderDesc())
		}
		return product.ByStatus(sql.OrderAsc())
	case "created_at":
		if desc {
			return product.ByCreatedAt(sql.OrderDesc())
		}
		return product.ByCreatedAt(sql.OrderAsc())
	case "updated_at":
		if desc {
			return product.ByUpdatedAt(sql.OrderDesc())
		}
		return product.ByUpdatedAt(sql.OrderAsc())
	default:
		return nil
	}
}

// applyCursor applies cursor-based pagination
func (r *ProductRepositoryImpl) applyCursor(query *ent.ProductQuery, cursor *entities.Cursor, direction string, sortParams []entities.SortParam) *ent.ProductQuery {
	// Parse cursor timestamp - try RFC3339Nano first, then RFC3339 for backward compatibility
	cursorTime, err := time.Parse(time.RFC3339Nano, cursor.CreatedAt)
	if err != nil {
		cursorTime, err = time.Parse(time.RFC3339, cursor.CreatedAt)
		if err != nil {
			// If parsing fails, fall back to ID-only comparison
			if direction == "backward" {
				return query.Where(product.IDLT(cursor.ID))
			}
			return query.Where(product.IDGT(cursor.ID))
		}
	}

	if direction == "backward" {
		// For backward pagination (prev page), get items BEFORE cursor
		// With ORDER BY created_at DESC, id DESC:
		// WHERE (created_at > cursor.created_at) OR (created_at = cursor.created_at AND id > cursor.id)
		return query.Where(
			product.Or(
				product.CreatedAtGT(cursorTime),
				product.And(
					product.CreatedAtEQ(cursorTime),
					product.IDGT(cursor.ID),
				),
			),
		)
	}

	// Forward pagination (next page), get items AFTER cursor
	// With ORDER BY created_at DESC, id DESC:
	// WHERE (created_at < cursor.created_at) OR (created_at = cursor.created_at AND id < cursor.id)
	return query.Where(
		product.Or(
			product.CreatedAtLT(cursorTime),
			product.And(
				product.CreatedAtEQ(cursorTime),
				product.IDLT(cursor.ID),
			),
		),
	)
}

// buildPageInfo builds pagination metadata
func (r *ProductRepositoryImpl) buildPageInfo(products []*ent.Product, hasNextPage bool, pagination *entities.PaginationParams) entities.PageInfo {
	// Determine if there's a previous page
	// If we received a cursor in forward direction, there's a previous page
	// If we received a cursor in backward direction, there's a next page
	hasPreviousPage := pagination != nil && pagination.Cursor != nil && pagination.Direction != "backward"

	pageInfo := entities.PageInfo{
		HasNextPage:     hasNextPage,
		HasPreviousPage: hasPreviousPage,
		PreviousCursor:     "",  // Will be set below if applicable
		NextCursor:       "",  // Will be set below if applicable
	}

	if len(products) > 0 {
		// Set start_cursor if there's a previous page
		if hasPreviousPage {
			startCursor := entities.Cursor{
				ID:        products[0].ID,
				CreatedAt: products[0].CreatedAt.Format(time.RFC3339Nano),
			}
			startCursorStr, _ := EncodeCursor(startCursor)
			pageInfo.PreviousCursor = startCursorStr
		}

		// Set end_cursor if there's a next page
		if hasNextPage {
			endCursor := entities.Cursor{
				ID:        products[len(products)-1].ID,
				CreatedAt: products[len(products)-1].CreatedAt.Format(time.RFC3339Nano),
			}
			endCursorStr, _ := EncodeCursor(endCursor)
			pageInfo.NextCursor = endCursorStr
		}
	}

	return pageInfo
}
