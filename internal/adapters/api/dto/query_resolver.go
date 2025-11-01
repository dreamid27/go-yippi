package dto

import (
	"regexp"
	"strconv"

	"github.com/danielgtaylor/huma/v2"
)

// Resolver implementation for QueryProductsRequest to handle complex query parameters
func (q *QueryProductsRequest) Resolve(ctx huma.Context) []error {
	// Get URL and parse query parameters
	url := ctx.URL()
	queryParams := url.Query()

	// Parse filter parameters
	// Expected format: filter[0][field]=status&filter[0][operator]=eq&filter[0][value]=published
	filterPattern := regexp.MustCompile(`^filter\[(\d+)\]\[(\w+)\]$`)
	filterMap := make(map[int]map[string]interface{})

	// Collect maximum index to ensure we process all filters
	maxFilterIndex := -1

	for key, values := range queryParams {
		if matches := filterPattern.FindStringSubmatch(key); matches != nil {
			index, _ := strconv.Atoi(matches[1])
			field := matches[2]

			if filterMap[index] == nil {
				filterMap[index] = make(map[string]interface{})
			}

			// Take the first value for each parameter
			if len(values) > 0 {
				filterMap[index][field] = values[0]
			}

			if index > maxFilterIndex {
				maxFilterIndex = index
			}
		}
	}

	// Convert map to FilterDTO slice
	for i := 0; i <= maxFilterIndex; i++ {
		if fm, ok := filterMap[i]; ok {
			filter := FilterDTO{}
			if field, ok := fm["field"].(string); ok {
				filter.Field = field
			}
			if operator, ok := fm["operator"].(string); ok {
				filter.Operator = operator
			}
			if value, ok := fm["value"]; ok {
				filter.Value = value
			}
			q.Filters = append(q.Filters, filter)
		}
	}

	// Parse sort parameters
	// Expected format: sort[0][field]=price&sort[0][order]=desc
	sortPattern := regexp.MustCompile(`^sort\[(\d+)\]\[(\w+)\]$`)
	sortMap := make(map[int]map[string]string)

	// Collect maximum index to ensure we process all sort params
	maxSortIndex := -1

	for key, values := range queryParams {
		if matches := sortPattern.FindStringSubmatch(key); matches != nil {
			index, _ := strconv.Atoi(matches[1])
			field := matches[2]

			if sortMap[index] == nil {
				sortMap[index] = make(map[string]string)
			}

			// Take the first value for each parameter
			if len(values) > 0 {
				sortMap[index][field] = values[0]
			}

			if index > maxSortIndex {
				maxSortIndex = index
			}
		}
	}

	// Convert map to SortDTO slice
	for i := 0; i <= maxSortIndex; i++ {
		if sm, ok := sortMap[i]; ok {
			sort := SortDTO{}
			if field, ok := sm["field"]; ok {
				sort.Field = field
			}
			if order, ok := sm["order"]; ok {
				sort.Order = order
			}
			q.Sort = append(q.Sort, sort)
		}
	}

	return nil
}

// Ensure QueryProductsRequest implements huma.Resolver
var _ huma.Resolver = (*QueryProductsRequest)(nil)
