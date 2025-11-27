# Complete Observability Guide

This guide explains how to implement comprehensive APM, Distributed Tracing, and Logging in the go-yippi application using OpenTelemetry, Zap, and SigNoz.

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Where to Add Observability](#where-to-add-observability)
- [What to Log](#what-to-log)
- [How to Add Tracing](#how-to-add-tracing)
- [Metrics Collection](#metrics-collection)
- [Best Practices](#best-practices)
- [Examples](#examples)
- [Viewing in SigNoz](#viewing-in-signoz)

---

## Overview

The application uses a comprehensive observability stack:

- **Distributed Tracing**: OpenTelemetry traces propagate across HTTP → Service → Repository → Database
- **Structured Logging**: Zap logger with JSON format, correlated with trace IDs
- **APM Metrics**: OpenTelemetry metrics for request rates, latencies, error rates, system metrics
- **Backend**: SigNoz for unified visualization

**Key Features:**
- ✅ Automatic HTTP request tracing
- ✅ Automatic database query tracing
- ✅ Service layer tracing with helper utilities
- ✅ Logs correlated with traces (trace_id, span_id)
- ✅ Business event tracking
- ✅ Slow operation detection
- ✅ Error tracking with context

---

## Architecture

```
┌─────────────┐
│ HTTP Request│
└──────┬──────┘
       │
       ▼
┌────────────────────────┐
│ HTTP Middleware        │
│ - Start Span           │
│ - Record Metrics       │
│ - Log Request          │
└──────┬─────────────────┘
       │ (context with trace)
       ▼
┌────────────────────────┐
│ Handler (API Layer)    │
│ - Extract context      │
│ - Call service         │
└──────┬─────────────────┘
       │ (context propagation)
       ▼
┌────────────────────────┐
│ Service Layer          │
│ - StartOperation()     │
│ - Business Logic       │
│ - Log Events           │
│ - Record Metrics       │
└──────┬─────────────────┘
       │ (context propagation)
       ▼
┌────────────────────────┐
│ Repository Layer       │
│ - Database Operations  │
│ - (Auto-traced)        │
└──────┬─────────────────┘
       │
       ▼
┌────────────────────────┐
│ Database               │
│ - Query Execution      │
│ - (Auto-traced)        │
└────────────────────────┘
```

**Data Flow:**
1. HTTP request → HTTP middleware creates root span + logs
2. Handler extracts context → calls service
3. Service starts operation → creates child span + logs + metrics
4. Repository uses context → database auto-traced
5. All logs include trace_id for correlation

---

## Where to Add Observability

### 1. HTTP Layer (✅ Already Done)

**Location**: `internal/infrastructure/telemetry/httpmiddleware.go`

**What's tracked:**
- Request/response tracing
- HTTP metrics (request count, duration, sizes)
- Status codes
- Client IP, User-Agent
- Errors

**No action needed** - automatically applied to all routes.

---

### 2. Service Layer (⚠️ You Must Add)

**Location**: `internal/application/services/*.go`

**When to add:**
- ✅ **All public service methods** that implement business logic
- ✅ Methods that make decisions or transformations
- ✅ Methods that coordinate multiple repository calls
- ✅ Methods that enforce business rules

**How to add:**

```go
import (
    "example.com/go-yippi/internal/infrastructure/observability"
    "go.uber.org/zap"
)

type YourService struct {
    repo     ports.YourRepository
    observer *observability.ServiceObserver
}

func NewYourService(repo ports.YourRepository) *YourService {
    return &YourService{
        repo:     repo,
        observer: observability.NewServiceObserver("YourService"),
    }
}

func (s *YourService) YourMethod(ctx context.Context, params ...) error {
    // Start observable operation
    op := s.observer.StartOperation(ctx, "YourMethod")
    defer op.End(nil)

    // Add input attributes
    op.AddAttribute("param1", param1)

    // Your business logic here
    result, err := s.repo.SomeOperation(op.Context())

    if err != nil {
        op.End(err)
        return err
    }

    // Log important events
    op.LogInfo("Operation completed", zap.String("result", result))

    op.End(nil)
    return nil
}
```

**See** `internal/application/services/product_service.go` for complete examples.

---

### 3. Database Layer (✅ Already Done)

**Location**: `internal/infrastructure/telemetry/dbtracer.go`

**What's tracked:**
- All CRUD operations (create, update, delete)
- Query duration
- Entity types
- Errors

**No action needed** - automatically tracked through Ent hooks.

---

## What to Log

### Log Levels

| Level | When to Use | Example |
|-------|------------|---------|
| `Debug` | Development debugging, verbose info | Variable values, detailed flow |
| `Info` | Normal operation events | "Product created", "User logged in" |
| `Warn` | Recoverable issues, business rule violations | "Validation failed", "Slow query detected" |
| `Error` | Errors that need attention | Database errors, external API failures |
| `Fatal` | Unrecoverable errors, app shutdown | Config load failed, DB connection failed |

### What to Log in Services

#### ✅ **Always Log**

1. **Business Events** - Important state changes
   ```go
   op.RecordBusinessEvent("product", "product_created",
       zap.String("product_id", id),
       zap.String("sku", sku))
   ```

2. **Validation Errors** - Why a request was rejected
   ```go
   observability.LogValidationError(ctx, "price", "Price must be positive")
   ```

3. **Business Rule Violations** - Why an operation was denied
   ```go
   observability.LogBusinessRule(ctx, "publish_draft_only",
       "Cannot publish non-draft products")
   ```

4. **Not Found Errors** - When resources don't exist
   ```go
   observability.LogNotFoundError(ctx, "Product", productID)
   ```

5. **Slow Operations** - Performance issues
   ```go
   observability.LogSlowOperation(ctx, "QueryProducts", duration, 200*time.Millisecond)
   ```

6. **State Transitions** - When entity status changes
   ```go
   op.LogInfo("Product status changed",
       zap.String("from", oldStatus),
       zap.String("to", newStatus))
   ```

#### ❌ **Don't Log**

1. **Sensitive Data** - Passwords, tokens, credit cards, PII
2. **High-frequency verbose logs** - Every loop iteration
3. **Redundant information** - Already in traces or metrics
4. **Success of trivial operations** - Simple getters

---

## How to Add Tracing

### Pattern 1: Simple Service Method

```go
func (s *ProductService) GetProduct(ctx context.Context, id int) (*entities.Product, error) {
    // Start operation
    op := s.observer.StartOperation(ctx, "GetProduct")
    defer op.End(nil)

    // Add input for tracing
    op.AddAttribute("product_id", id)

    // Call repository with traced context
    product, err := s.repo.GetByID(op.Context(), id)
    if err != nil {
        observability.LogNotFoundError(op.Context(), "Product", id)
        op.End(err)
        return nil, err
    }

    op.End(nil)
    return product, nil
}
```

### Pattern 2: Method with Business Logic

```go
func (s *ProductService) CreateProduct(ctx context.Context, product *entities.Product) error {
    op := s.observer.StartOperation(ctx, "CreateProduct")
    defer op.End(nil)

    // Add input attributes
    op.AddAttribute("sku", product.SKU)
    op.AddAttribute("name", product.Name)

    // Validation
    if product.Price <= 0 {
        err := domainErrors.NewValidationError("price", "Price must be positive")
        observability.LogValidationError(op.Context(), "price", "Price must be positive")
        op.End(err)
        return err
    }

    // Business logic
    if product.Status == "" {
        product.Status = entities.ProductStatusDraft
        op.LogInfo("Set default status to draft")
    }

    // Repository call
    err := s.repo.Create(op.Context(), product)
    if err != nil {
        op.End(err)
        return err
    }

    // Record business event
    op.RecordBusinessEvent("product", "product_created",
        zap.String("product_id", fmt.Sprintf("%d", product.ID)),
        zap.String("sku", product.SKU))

    op.End(nil)
    return nil
}
```

### Pattern 3: Method with Business Rules

```go
func (s *ProductService) PublishProduct(ctx context.Context, id int) error {
    op := s.observer.StartOperation(ctx, "PublishProduct")
    defer op.End(nil)

    op.AddAttribute("product_id", id)

    product, err := s.repo.GetByID(op.Context(), id)
    if err != nil {
        observability.LogNotFoundError(op.Context(), "Product", id)
        op.End(err)
        return err
    }

    // Business rule enforcement
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

    // Log state transition
    op.LogInfo("Product status changed",
        zap.String("from_status", string(oldStatus)),
        zap.String("to_status", string(product.Status)))

    // Record business event
    op.RecordBusinessEvent("product", "product_published",
        zap.String("product_id", fmt.Sprintf("%d", id)))

    op.End(nil)
    return nil
}
```

### Pattern 4: Complex Operation with Multiple Steps

```go
func (s *ProductService) QueryProducts(ctx context.Context, params *entities.QueryParams) (*entities.QueryResult, error) {
    op := s.observer.StartOperation(ctx, "QueryProducts")
    defer op.End(nil)

    // Add metadata
    op.AddAttribute("filters_count", len(params.Filters))
    op.AddAttribute("sort_count", len(params.Sort))

    // Validation
    if len(params.Filters) > 10 {
        err := domainErrors.NewValidationError("filters", "Max 10 filters allowed")
        observability.LogValidationError(op.Context(), "filters", "Max 10 filters allowed")
        op.End(err)
        return nil, err
    }

    // Complex business logic
    if len(categoryIDs) > 0 {
        expandedIDs, err := s.categoryRepo.GetDescendantIDs(op.Context(), categoryIDs)
        if err != nil {
            op.End(err)
            return nil, err
        }

        // Log important transformations
        if len(expandedIDs) > len(categoryIDs) {
            op.LogInfo("Expanded category filter",
                zap.Int("original", len(categoryIDs)),
                zap.Int("expanded", len(expandedIDs)))
        }
    }

    // Execute with performance monitoring
    startTime := time.Now()
    result, err := s.repo.Query(op.Context(), params)
    observability.LogSlowOperation(op.Context(), "QueryProducts",
        time.Since(startTime), 200*time.Millisecond)

    if err != nil {
        op.End(err)
        return nil, err
    }

    // Log results
    op.AddAttribute("results_count", len(result.Products))
    op.LogInfo("Query completed",
        zap.Int("count", len(result.Products)),
        zap.Bool("has_next", result.PageInfo.HasNextPage))

    op.End(nil)
    return result, nil
}
```

---

## Metrics Collection

### Automatic Metrics (Already Collected)

| Metric | Type | Description |
|--------|------|-------------|
| `http.server.request.count` | Counter | Total HTTP requests |
| `http.server.request.duration` | Histogram | Request latency (ms) |
| `http.server.active_requests` | Gauge | Current active requests |
| `http.server.request.size` | Histogram | Request body size (bytes) |
| `http.server.response.size` | Histogram | Response body size (bytes) |
| `db.client.query.count` | Counter | Total database queries |
| `db.client.query.duration` | Histogram | Query latency (ms) |
| `db.client.errors.count` | Counter | Database errors |
| `service.operation.count` | Counter | Total service operations |
| `service.operation.duration` | Histogram | Service operation latency (ms) |
| `service.errors.count` | Counter | Service errors |
| `business.event.count` | Counter | Business events |
| `system.cpu.usage` | Gauge | CPU usage % |
| `system.memory.usage` | Gauge | Memory usage (bytes) |
| `system.goroutines.count` | Gauge | Number of goroutines |

**All metrics are automatically tagged with:**
- Service name
- Operation type
- HTTP method, route, status code
- Database entity, operation
- Error types

---

## Best Practices

### 1. Context Propagation

**✅ DO:**
```go
op := s.observer.StartOperation(ctx, "Method")
result, err := s.repo.Query(op.Context()) // Use op.Context()
```

**❌ DON'T:**
```go
op := s.observer.StartOperation(ctx, "Method")
result, err := s.repo.Query(ctx) // Wrong! Trace not propagated
```

### 2. Error Handling

**✅ DO:**
```go
product, err := s.repo.GetByID(op.Context(), id)
if err != nil {
    observability.LogNotFoundError(op.Context(), "Product", id)
    op.End(err) // End with error
    return nil, err
}
```

**❌ DON'T:**
```go
product, err := s.repo.GetByID(op.Context(), id)
if err != nil {
    return nil, err // Missing logging and op.End()
}
```

### 3. Defer End()

**✅ DO:**
```go
func (s *Service) Method(ctx context.Context) error {
    op := s.observer.StartOperation(ctx, "Method")
    defer op.End(nil) // Safe fallback

    // ... logic ...

    if err != nil {
        op.End(err) // Explicit error
        return err
    }

    op.End(nil)
    return nil
}
```

### 4. Meaningful Attributes

**✅ DO:**
```go
op.AddAttribute("product_id", id)
op.AddAttribute("sku", product.SKU)
op.AddAttribute("status", string(product.Status))
```

**❌ DON'T:**
```go
op.AddAttribute("data", fmt.Sprintf("%+v", product)) // Too verbose
op.AddAttribute("x", id) // Unclear naming
```

### 5. Business Events

**✅ DO:**
```go
op.RecordBusinessEvent("product", "product_published",
    zap.String("product_id", id),
    zap.String("sku", sku),
    zap.String("status", string(status)))
```

Use business events for:
- State transitions
- User actions
- Financial transactions
- Important milestones

### 6. Slow Operation Thresholds

```go
// API operations
observability.LogSlowOperation(ctx, "CreateProduct", duration, 100*time.Millisecond)

// Database queries
observability.LogSlowOperation(ctx, "QueryProducts", duration, 200*time.Millisecond)

// External API calls
observability.LogSlowOperation(ctx, "FetchExternalData", duration, 1*time.Second)
```

---

## Examples

### Complete Service Example

See `internal/application/services/product_service.go` for:
- ✅ Full observability implementation
- ✅ All patterns demonstrated
- ✅ Error handling
- ✅ Business events
- ✅ Performance monitoring

### Adding to Your Service

```go
package services

import (
    "context"
    "example.com/go-yippi/internal/domain/entities"
    "example.com/go-yippi/internal/domain/ports"
    "example.com/go-yippi/internal/infrastructure/observability"
    "go.uber.org/zap"
)

type UserService struct {
    repo     ports.UserRepository
    observer *observability.ServiceObserver
}

func NewUserService(repo ports.UserRepository) *UserService {
    return &UserService{
        repo:     repo,
        observer: observability.NewServiceObserver("UserService"),
    }
}

func (s *UserService) CreateUser(ctx context.Context, user *entities.User) error {
    op := s.observer.StartOperation(ctx, "CreateUser")
    defer op.End(nil)

    op.AddAttribute("email", user.Email)

    // Validation
    if user.Email == "" {
        err := domainErrors.NewValidationError("email", "Email required")
        observability.LogValidationError(op.Context(), "email", "Email required")
        op.End(err)
        return err
    }

    // Business logic
    err := s.repo.Create(op.Context(), user)
    if err != nil {
        op.End(err)
        return err
    }

    // Record event
    op.RecordBusinessEvent("user", "user_created",
        zap.String("user_id", fmt.Sprintf("%d", user.ID)),
        zap.String("email", user.Email))

    op.End(nil)
    return nil
}
```

---

## Viewing in SigNoz

### Accessing SigNoz

1. **Start SigNoz** (if not running):
   ```bash
   cd signoz
   docker-compose up -d
   ```

2. **Open dashboard**: http://localhost:3301

### Exploring Traces

1. Navigate to **Traces** tab
2. Filter by:
   - Service: `go-yippi-api`
   - Operation: `ProductService.CreateProduct`
   - Status: Error/Success
   - Duration: > 100ms

3. Click on a trace to see:
   - Full request flow (HTTP → Service → DB)
   - Duration breakdown
   - Attributes (product_id, sku, etc.)
   - Logs correlated to the trace

### Viewing Logs

1. Navigate to **Logs** tab
2. Logs are automatically tagged with:
   - `trace_id` - Click to see full trace
   - `span_id` - Current operation
   - `service` - Service name
   - `method` - Method name

3. Filter by:
   ```
   service="go-yippi-api" AND event="business_event"
   ```

### Checking Metrics

1. Navigate to **Metrics** tab
2. Query examples:
   - `http_server_request_duration_ms{route="/products"}` - API latency
   - `service_operation_duration_ms{service="ProductService"}` - Service performance
   - `db_client_query_duration_ms{entity="Product"}` - Database query times
   - `business_event_count{event_name="product_created"}` - Business events

3. Create dashboards with:
   - Request rate (requests/second)
   - Error rate (%)
   - P50, P95, P99 latencies
   - Active requests

### Creating Alerts

1. Go to **Alerts** → **New Alert**
2. Example alerts:
   - High error rate: `service_errors_count > 10 in 5 minutes`
   - Slow requests: `http_server_request_duration_ms p95 > 1000ms`
   - Database issues: `db_client_errors_count > 5 in 1 minute`

---

## Troubleshooting

### Traces not appearing in SigNoz

1. Check OTEL is enabled:
   ```bash
   # In .env
   OTEL_ENABLED=true
   OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4317
   ```

2. Verify SigNoz is running:
   ```bash
   docker ps | grep signoz
   ```

3. Check application logs for OTEL initialization:
   ```
   OpenTelemetry initialized with service name: go-yippi-api
   ```

### Logs missing trace_id

Ensure you're using the observability helpers:
```go
op := s.observer.StartOperation(ctx, "Method")
op.LogInfo("Message", ...) // Has trace_id
```

### Metrics not showing

1. Ensure metrics are initialized in `main.go`:
   ```go
   metricsCleanup := telemetry.InitMetrics(&cfg.OpenTelemetry)
   defer metricsCleanup(context.Background())
   ```

2. Check SigNoz metrics collector is running on port 4317

---

## Next Steps

1. **Add observability to remaining services**:
   - UserService
   - CategoryService
   - BrandService
   - StorageService

2. **Create custom dashboards in SigNoz**:
   - Product creation rate
   - Query performance by filter type
   - Error breakdown by service

3. **Set up alerts**:
   - High error rate
   - Slow operations
   - Database connection issues

4. **Review and optimize**:
   - Identify slow operations from traces
   - Optimize based on metrics
   - Reduce log noise

---

## Summary

✅ **You now have:**
- Distributed tracing across all layers
- Structured logging with trace correlation
- Comprehensive APM metrics
- Business event tracking
- Performance monitoring
- Unified observability in SigNoz

✅ **To add observability to new code:**
1. Add `observer` to service struct
2. Use `op := s.observer.StartOperation(ctx, "Method")`
3. Add attributes, log events, record metrics
4. Always call `op.End(err)`

✅ **For production:**
- Set appropriate log levels (`info` or `warn`)
- Configure log rotation
- Set up alerts for critical issues
- Monitor dashboard regularly

Happy observing! 🔍
