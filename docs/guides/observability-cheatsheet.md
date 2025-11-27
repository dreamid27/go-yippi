# Observability Quick Reference

> 🚀 Quick copy-paste examples for adding observability to your code

## Table of Contents
- [Service Setup](#service-setup)
- [Basic Operation](#basic-operation)
- [With Validation](#with-validation)
- [With Business Rules](#with-business-rules)
- [Complex Operations](#complex-operations)
- [Helper Functions](#helper-functions)
- [Common Patterns](#common-patterns)

---

## Service Setup

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
```

---

## Basic Operation

```go
func (s *YourService) GetItem(ctx context.Context, id int) (*entities.Item, error) {
    op := s.observer.StartOperation(ctx, "GetItem")
    defer op.End(nil)

    op.AddAttribute("item_id", id)

    item, err := s.repo.GetByID(op.Context(), id)
    if err != nil {
        observability.LogNotFoundError(op.Context(), "Item", id)
        op.End(err)
        return nil, err
    }

    op.End(nil)
    return item, nil
}
```

---

## With Validation

```go
func (s *YourService) CreateItem(ctx context.Context, item *entities.Item) error {
    op := s.observer.StartOperation(ctx, "CreateItem")
    defer op.End(nil)

    op.AddAttribute("name", item.Name)

    // Validate
    if item.Name == "" {
        err := domainErrors.NewValidationError("name", "Name is required")
        observability.LogValidationError(op.Context(), "name", "Name is required")
        op.End(err)
        return err
    }

    // Create
    err := s.repo.Create(op.Context(), item)
    if err != nil {
        op.End(err)
        return err
    }

    // Record event
    op.RecordBusinessEvent("item", "item_created",
        zap.String("item_id", fmt.Sprintf("%d", item.ID)),
        zap.String("name", item.Name))

    op.End(nil)
    return nil
}
```

---

## With Business Rules

```go
func (s *YourService) ActivateItem(ctx context.Context, id int) error {
    op := s.observer.StartOperation(ctx, "ActivateItem")
    defer op.End(nil)

    op.AddAttribute("item_id", id)

    item, err := s.repo.GetByID(op.Context(), id)
    if err != nil {
        observability.LogNotFoundError(op.Context(), "Item", id)
        op.End(err)
        return err
    }

    // Business rule check
    if item.Status != entities.StatusPending {
        err := domainErrors.NewValidationError("status", "Only pending items can be activated")
        observability.LogBusinessRule(op.Context(), "activate_pending_only",
            fmt.Sprintf("Cannot activate item with status: %s", item.Status))
        op.End(err)
        return err
    }

    oldStatus := item.Status
    item.Status = entities.StatusActive

    err = s.repo.Update(op.Context(), item)
    if err != nil {
        op.End(err)
        return err
    }

    // Log state change
    op.LogInfo("Item status changed",
        zap.String("from", string(oldStatus)),
        zap.String("to", string(item.Status)))

    // Record event
    op.RecordBusinessEvent("item", "item_activated",
        zap.String("item_id", fmt.Sprintf("%d", id)))

    op.End(nil)
    return nil
}
```

---

## Complex Operations

```go
func (s *YourService) ProcessBatch(ctx context.Context, ids []int) error {
    op := s.observer.StartOperation(ctx, "ProcessBatch")
    defer op.End(nil)

    op.AddAttribute("batch_size", len(ids))

    successCount := 0
    errorCount := 0

    for _, id := range ids {
        op.AddAttribute("processing_id", id)

        err := s.processItem(op.Context(), id)
        if err != nil {
            errorCount++
            op.LogWarn("Failed to process item",
                zap.Int("item_id", id),
                zap.Error(err))
            continue
        }
        successCount++
    }

    op.AddAttribute("success_count", successCount)
    op.AddAttribute("error_count", errorCount)

    op.LogInfo("Batch processing completed",
        zap.Int("total", len(ids)),
        zap.Int("success", successCount),
        zap.Int("errors", errorCount))

    op.End(nil)
    return nil
}
```

---

## Helper Functions

### Log Validation Error
```go
observability.LogValidationError(ctx, "field_name", "Error message")
```

### Log Not Found
```go
observability.LogNotFoundError(ctx, "EntityName", entityID)
```

### Log Duplicate Error
```go
observability.LogDuplicateError(ctx, "EntityName", "field_name", value)
```

### Log Business Rule
```go
observability.LogBusinessRule(ctx, "rule_name", "Explanation message")
```

### Log Slow Operation
```go
startTime := time.Now()
// ... operation ...
observability.LogSlowOperation(ctx, "OperationName", time.Since(startTime), 100*time.Millisecond)
```

### Record Business Event
```go
op.RecordBusinessEvent("category", "event_name",
    zap.String("field1", value1),
    zap.Int("field2", value2))
```

---

## Common Patterns

### Pattern: Simple CRUD

```go
// Create
func (s *Service) Create(ctx context.Context, item *Item) error {
    op := s.observer.StartOperation(ctx, "Create")
    defer op.End(nil)

    err := s.repo.Create(op.Context(), item)
    if err != nil {
        op.End(err)
        return err
    }

    op.RecordBusinessEvent("item", "created", zap.String("id", item.ID))
    op.End(nil)
    return nil
}

// Read
func (s *Service) GetByID(ctx context.Context, id int) (*Item, error) {
    op := s.observer.StartOperation(ctx, "GetByID")
    defer op.End(nil)

    op.AddAttribute("id", id)

    item, err := s.repo.GetByID(op.Context(), id)
    if err != nil {
        observability.LogNotFoundError(op.Context(), "Item", id)
        op.End(err)
        return nil, err
    }

    op.End(nil)
    return item, nil
}

// Update
func (s *Service) Update(ctx context.Context, item *Item) error {
    op := s.observer.StartOperation(ctx, "Update")
    defer op.End(nil)

    op.AddAttribute("id", item.ID)

    err := s.repo.Update(op.Context(), item)
    if err != nil {
        op.End(err)
        return err
    }

    op.RecordBusinessEvent("item", "updated", zap.String("id", item.ID))
    op.End(nil)
    return nil
}

// Delete
func (s *Service) Delete(ctx context.Context, id int) error {
    op := s.observer.StartOperation(ctx, "Delete")
    defer op.End(nil)

    op.AddAttribute("id", id)

    err := s.repo.Delete(op.Context(), id)
    if err != nil {
        op.End(err)
        return err
    }

    op.RecordBusinessEvent("item", "deleted", zap.Int("id", id))
    op.End(nil)
    return nil
}
```

### Pattern: Multi-step Operation

```go
func (s *Service) ComplexOperation(ctx context.Context, input Input) (*Result, error) {
    op := s.observer.StartOperation(ctx, "ComplexOperation")
    defer op.End(nil)

    // Step 1
    op.LogInfo("Starting step 1")
    result1, err := s.step1(op.Context(), input)
    if err != nil {
        op.LogError("Step 1 failed", err)
        op.End(err)
        return nil, err
    }

    // Step 2
    op.LogInfo("Starting step 2")
    result2, err := s.step2(op.Context(), result1)
    if err != nil {
        op.LogError("Step 2 failed", err)
        op.End(err)
        return nil, err
    }

    // Step 3
    op.LogInfo("Starting step 3")
    finalResult, err := s.step3(op.Context(), result2)
    if err != nil {
        op.LogError("Step 3 failed", err)
        op.End(err)
        return nil, err
    }

    op.LogInfo("All steps completed successfully")
    op.End(nil)
    return finalResult, nil
}
```

### Pattern: Conditional Logic

```go
func (s *Service) ConditionalProcess(ctx context.Context, item *Item) error {
    op := s.observer.StartOperation(ctx, "ConditionalProcess")
    defer op.End(nil)

    if item.IsSpecial {
        op.LogInfo("Processing special item")
        err := s.specialProcess(op.Context(), item)
        if err != nil {
            op.End(err)
            return err
        }
        op.RecordBusinessEvent("item", "special_processed")
    } else {
        op.LogInfo("Processing normal item")
        err := s.normalProcess(op.Context(), item)
        if err != nil {
            op.End(err)
            return err
        }
        op.RecordBusinessEvent("item", "normal_processed")
    }

    op.End(nil)
    return nil
}
```

---

## Checklist for New Service Method

- [ ] Add `observer` field to service struct
- [ ] Initialize observer in `New*Service()`
- [ ] Start operation: `op := s.observer.StartOperation(ctx, "MethodName")`
- [ ] Add `defer op.End(nil)` at start
- [ ] Add input attributes: `op.AddAttribute("key", value)`
- [ ] Use `op.Context()` for all downstream calls
- [ ] Log validation errors with `observability.LogValidationError()`
- [ ] Log business rules with `observability.LogBusinessRule()`
- [ ] Log not found with `observability.LogNotFoundError()`
- [ ] Call `op.End(err)` before returning errors
- [ ] Record business events with `op.RecordBusinessEvent()`
- [ ] Log important state changes with `op.LogInfo()`
- [ ] End operation: `op.End(nil)` on success

---

## Log Level Guide

| Level | Use For | Example |
|-------|---------|---------|
| Debug | Development details | `op.LogDebug("Variable value", zap.Any("var", v))` |
| Info | Normal events | `op.LogInfo("Operation started")` |
| Warn | Recoverable issues | `op.LogWarn("Item not found", ...)` |
| Error | Errors needing attention | `op.LogError("DB connection failed", err)` |

---

## Context Usage

**✅ ALWAYS use `op.Context()`:**
```go
op := s.observer.StartOperation(ctx, "Method")
result := s.repo.Query(op.Context()) // Correct
```

**❌ NEVER reuse original context:**
```go
op := s.observer.StartOperation(ctx, "Method")
result := s.repo.Query(ctx) // Wrong! Trace not propagated
```

---

## Common Mistakes

### ❌ Forgetting to End Operation
```go
// Wrong
func (s *Service) Method(ctx context.Context) error {
    op := s.observer.StartOperation(ctx, "Method")
    return s.repo.Create(op.Context(), item) // Missing op.End()
}
```

```go
// Correct
func (s *Service) Method(ctx context.Context) error {
    op := s.observer.StartOperation(ctx, "Method")
    defer op.End(nil)

    err := s.repo.Create(op.Context(), item)
    if err != nil {
        op.End(err)
        return err
    }

    op.End(nil)
    return nil
}
```

### ❌ Missing Error Logging
```go
// Wrong
if err != nil {
    return err // No context about the error
}
```

```go
// Correct
if err != nil {
    observability.LogNotFoundError(op.Context(), "Item", id)
    op.End(err)
    return err
}
```

### ❌ Forgetting Business Events
```go
// Wrong
product.Status = StatusPublished
s.repo.Update(ctx, product) // Important event not tracked
```

```go
// Correct
product.Status = StatusPublished
s.repo.Update(op.Context(), product)
op.RecordBusinessEvent("product", "published",
    zap.String("product_id", id))
```

---

## Need More Examples?

See `internal/application/services/product_service.go` for complete, production-ready examples.

See `docs/guides/observability-guide.md` for full documentation.
