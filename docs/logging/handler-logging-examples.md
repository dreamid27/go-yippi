# Handler Structured Logging Examples

This document provides practical examples of how to implement structured logging in API handlers using Zap.

## Enhanced User Handler with Full Logging

```go
package handlers

import (
    "context"
    "errors"
    "time"

    "example.com/go-yippi/internal/adapters/api/dto"
    "example.com/go-yippi/internal/application/services"
    "example.com/go-yippi/internal/domain/entities"
    "example.com/go-yippi/internal/infrastructure/logging"
    "github.com/danielgtaylor/huma/v2"
    "github.com/gofiber/fiber/v2"
    "go.uber.org/zap"
)

type UserHandler struct {
    service ports.UserService
    logger  *zap.Logger
}

func NewUserHandler(service ports.UserService) *UserHandler {
    return &UserHandler{
        service: service,
        logger:  logging.GetGlobalLogger(),
    }
}

// CreateUser with comprehensive logging
func (h *UserHandler) CreateUser(ctx context.Context, input *dto.CreateUserRequest) (*dto.CreateUserResponse, error) {
    start := time.Now()

    // Log request received
    requestID := ctx.Value("request_id").(string)
    h.logger.Info("Create user request received",
        zap.String("request_id", requestID),
        zap.String("operation", "create_user"),
        zap.String("user_email", input.Body.Email),
        zap.String("user_name", input.Body.Name),
        zap.String("client_ip", getClientIP(ctx)),
    )

    // Input validation with logging
    if err := h.validateCreateUserInput(input.Body); err != nil {
        duration := time.Since(start)
        h.logger.Warn("Create user validation failed",
            zap.Error(err),
            zap.String("request_id", requestID),
            zap.String("validation_type", "input_validation"),
            zap.Float64("duration_ms", float64(duration.Nanoseconds())/1e6),
            zap.Any("input_data", input.Body),
        )
        return nil, huma.Error400BadRequest("Invalid input: "+err.Error())
    }

    // Business logic execution with tracing
    startBusiness := time.Now()
    user := &entities.User{
        Name:  input.Body.Name,
        Email: input.Body.Email,
        Password: input.Body.Password,
        Role:  "user", // Default role
    }

    err := h.service.CreateUser(ctx, user)
    businessDuration := time.Since(startBusiness)

    // Log business operation result
    duration := time.Since(start)
    if err != nil {
        h.logger.Error("Create user request failed",
            zap.Error(err),
            zap.String("request_id", requestID),
            zap.Float64("business_duration_ms", float64(businessDuration.Nanoseconds())/1e6),
            zap.Float64("total_duration_ms", float64(duration.Nanoseconds())/1e6),
            zap.String("failure_type", getErrorType(err)),
        )
        return nil, h.mapServiceErrorToHTTPError(err)
    }

    // Success logging
    h.logger.Info("Create user request completed successfully",
        zap.String("request_id", requestID),
        zap.String("user_id", fmt.Sprintf("%d", user.ID)),
        zap.String("user_email", user.Email),
        zap.Float64("business_duration_ms", float64(businessDuration.Nanoseconds())/1e6),
        zap.Float64("total_duration_ms", float64(duration.Nanoseconds())/1e6),
        zap.String("operation", "create_user"),
        zap.String("result", "success"),
    )

    // Create response
    response := &dto.CreateUserResponse{}
    response.Body.ID = user.ID
    response.Body.Name = user.Name
    response.Body.Email = user.Email
    response.Body.CreatedAt = user.CreatedAt

    return response, nil
}

// GetUser with detailed logging
func (h *UserHandler) GetUser(ctx context.Context, input *dto.GetUserRequest) (*dto.GetUserResponse, error) {
    start := time.Now()
    requestID := ctx.Value("request_id").(string)

    h.logger.Info("Get user request received",
        zap.String("request_id", requestID),
        zap.String("operation", "get_user"),
        zap.Int("user_id", input.ID),
        zap.String("client_ip", getClientIP(ctx)),
    )

    // Service call
    user, err := h.service.GetUser(ctx, input.ID)
    duration := time.Since(start)

    if err != nil {
        h.logger.Error("Get user request failed",
            zap.Error(err),
            zap.String("request_id", requestID),
            zap.Int("requested_id", input.ID),
            zap.Float64("duration_ms", float64(duration.Nanoseconds())/1e6),
            zap.String("failure_type", "user_not_found"),
        )
        return nil, h.mapServiceErrorToHTTPError(err)
    }

    // Success case
    h.logger.Info("Get user request completed successfully",
        zap.String("request_id", requestID),
        zap.Int("user_id", input.ID),
        zap.String("user_email", user.Email),
        zap.String("user_name", user.Name),
        zap.Float64("duration_ms", float64(duration.Nanoseconds())/1e6),
        zap.String("operation", "get_user"),
        zap.String("result", "success"),
    )

    response := &dto.GetUserResponse{}
    response.Body.ID = user.ID
    response.Body.Name = user.Name
    response.Body.Email = user.Email
    response.Body.CreatedAt = user.CreatedAt
    response.Body.UpdatedAt = user.UpdatedAt

    return response, nil
}

// ListUsers with pagination logging
func (h *UserHandler) ListUsers(ctx context.Context, input *dto.ListUsersRequest) (*dto.ListUsersResponse, error) {
    start := time.Now()
    requestID := ctx.Value("request_id").(string)

    h.logger.Info("List users request received",
        zap.String("request_id", requestID),
        zap.String("operation", "list_users"),
        zap.Int("page", input.Page),
        zap.Int("limit", input.Limit),
        zap.String("search", input.Search),
        zap.String("client_ip", getClientIP(ctx)),
    )

    // Input validation
    if input.Page < 1 {
        h.logger.Warn("Invalid pagination parameters",
            zap.String("request_id", requestID),
            zap.Int("page", input.Page),
            zap.String("validation_error", "page must be >= 1"),
        )
        return nil, huma.Error400BadRequest("Invalid page number")
    }

    if input.Limit < 1 || input.Limit > 100 {
        h.logger.Warn("Invalid limit parameter",
            zap.String("request_id", requestID),
            zap.Int("limit", input.Limit),
            zap.String("validation_error", "limit must be between 1 and 100"),
        )
        return nil, huma.Error400BadRequest("Invalid limit parameter")
    }

    // Service call
    users, total, err := h.service.ListUsers(ctx, input.Page, input.Limit, input.Search)
    duration := time.Since(start)

    if err != nil {
        h.logger.Error("List users request failed",
            zap.Error(err),
            zap.String("request_id", requestID),
            zap.Float64("duration_ms", float64(duration.Nanoseconds())/1e6),
            zap.String("failure_type", getErrorType(err)),
        )
        return nil, h.mapServiceErrorToHTTPError(err)
    }

    h.logger.Info("List users request completed successfully",
        zap.String("request_id", requestID),
        zap.Int("page", input.Page),
        zap.Int("limit", input.Limit),
        zap.String("search", input.Search),
        zap.Int("returned_count", len(users)),
        zap.Int64("total_count", total),
        zap.Bool("has_more", int64(input.Page*input.Limit+len(users)) < total),
        zap.Float64("duration_ms", float64(duration.Nanoseconds())/1e6),
        zap.String("operation", "list_users"),
        zap.String("result", "success"),
    )

    response := &dto.ListUsersResponse{}
    response.Body.Users = make([]*dto.UserListItem, len(users))
    response.Body.Total = total
    response.Body.Page = input.Page
    response.Body.Limit = input.Limit
    response.Body.HasMore = int64(input.Page*input.Limit+len(users)) < total

    for i, user := range users {
        response.Body.Users[i] = &dto.UserListItem{
            ID:        user.ID,
            Name:      user.Name,
            Email:     user.Email,
            CreatedAt: user.CreatedAt,
        }
    }

    return response, nil
}

// Helper functions

func (h *UserHandler) validateCreateUserInput(input *dto.CreateUserRequestBody) error {
    if input.Name == "" {
        return errors.New("name is required")
    }
    if len(input.Name) < 2 {
        return errors.New("name must be at least 2 characters")
    }
    if len(input.Name) > 100 {
        return errors.New("name must be less than 100 characters")
    }

    if input.Email == "" {
        return errors.New("email is required")
    }
    if !isValidEmail(input.Email) {
        return errors.New("invalid email format")
    }

    if input.Password == "" {
        return errors.New("password is required")
    }
    if len(input.Password) < 8 {
        return errors.New("password must be at least 8 characters")
    }

    return nil
}

func (h *UserHandler) mapServiceErrorToHTTPError(err error) *huma.StatusError {
    // Map service errors to HTTP errors
    switch {
    case errors.Is(err, ErrUserNotFound):
        return huma.Error404NotFound("User not found")
    case errors.Is(err, ErrUserAlreadyExists):
        return huma.Error409Conflict("User already exists")
    case errors.Is(err, ErrValidationError):
        return huma.Error400BadRequest(err.Error())
    default:
        return huma.Error500InternalServerError("Internal server error")
    }
}

func getErrorType(err error) string {
    switch {
    case errors.Is(err, ErrUserNotFound):
        return "not_found"
    case errors.Is(err, ErrUserAlreadyExists):
        return "already_exists"
    case errors.Is(err, ErrValidationError):
        return "validation_error"
    case errors.Is(err, ErrDatabaseError):
        return "database_error"
    case errors.Is(err, ErrServiceUnavailable):
        return "service_unavailable"
    default:
        return "unknown_error"
    }
}

func isValidEmail(email string) bool {
    // Simple email validation
    return strings.Contains(email, "@") && strings.Contains(email, ".")
}

func getClientIP(ctx context.Context) string {
    if ip, ok := ctx.Value("client_ip").(string); ok {
        return ip
    }
    return "unknown"
}
```

## Request-Response Flow Example

### 1. **Request Logging**
```json
{
  "level": "info",
  "timestamp": "2025-11-25T08:44:13.898+0700",
  "event": "http_request",
  "request_id": "req_17009190538901",
  "operation": "create_user",
  "user_email": "john@example.com",
  "user_name": "John Doe",
  "client_ip": "192.168.1.100",
  "method": "POST",
  "path": "/api/users"
}
```

### 2. **Business Logic Logging**
```json
{
  "level": "info",
  "timestamp": "2025-11-25T08:44:13.898+0700",
  "request_id": "req_17009190538901",
  "user_id": "12345",
  "user_email": "john@example.com",
  "business_duration_ms": 145.67,
  "operation": "create_user",
  "result": "success"
}
```

### 3. **Response Logging**
```json
{
  "level": "info",
  "timestamp": "2025-11-25T08:44:13.898+0700",
  "request_id": "req_17009190538901",
  "total_duration_ms": 189.23,
  "operation": "create_user",
  "result": "success",
  "status_code": 201
}
```

## Error Handling Examples

### Validation Error
```json
{
  "level": "warn",
  "timestamp": "2025-11-25T08:44:13.898+0700",
  "request_id": "req_17009190538901",
  "validation_type": "input_validation",
  "error": "name must be at least 2 characters",
  "duration_ms": 12.45,
  "operation": "create_user",
  "input_data": {
    "name": "",
    "email": "john@example.com",
    "password": "password123"
  }
}
```

### Business Logic Error
```json
{
  "level": "error",
  "timestamp": "2025-11-25T08:44:13.898+0700",
  "request_id": "req_17009190538901",
  "user_id": "999",
  "failure_type": "user_not_found",
  "business_duration_ms": 45.12,
  "total_duration_ms": 67.89,
  "error": "user not found in database",
  "operation": "get_user"
}
```

## Performance Monitoring

### Response Time Buckets
```go
// Add performance metrics
var durationMs = float64(duration.Nanoseconds()) / 1e6

switch {
case durationMs < 100:
    h.logger.Info("Fast response", zap.Float64("duration_ms", durationMs))
case durationMs < 500:
    h.logger.Info("Normal response", zap.Float64("duration_ms", durationMs))
case durationMs < 1000:
    h.logger.Warn("Slow response", zap.Float64("duration_ms", durationMs))
default:
    h.logger.Error("Very slow response", zap.Float64("duration_ms", durationMs))
}
```

### Success Rate Tracking
```go
// Track operation success/failure rates
successRate := float64(successCount) / float64(totalCount) * 100

h.logger.Info("Operation statistics",
    zap.Int("total_requests", totalCount),
    zap.Int("successful_requests", successCount),
    zap.Int("failed_requests", failedCount),
    zap.Float64("success_rate_percent", successRate),
    zap.String("operation", "create_user"),
)
```

## Best Practices

### 1. **Request Context**
- Always include request_id for correlation
- Log client IP for security monitoring
- Include operation name for easier filtering
- Add user context when available

### 2. **Structured Data**
- Use consistent field names
- Log request/response sizes
- Include timing metrics in milliseconds
- Use appropriate log levels

### 3. **Error Handling**
- Log errors at the point of occurrence
- Include error context and stack traces
- Map to appropriate HTTP status codes
- Use error categorization

### 4. **Performance**
- Track both business and total duration
- Add performance buckets
- Monitor slow queries and operations
- Include metrics for alerting

This comprehensive logging implementation provides excellent observability for your API handlers with full trace correlation and business insights.