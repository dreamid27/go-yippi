# OpenTelemetry Tracing Examples

This document provides practical examples of how to add OpenTelemetry tracing to your Go Yippi API services and handlers.

## Service Layer Tracing Examples

### Enhanced UserService with Tracing

```go
package services

import (
    "context"
    "example.com/go-yippi/internal/domain/entities"
    "example.com/go-yippi/internal/domain/ports"
    "example.com/go-yippi/internal/infrastructure/telemetry"
)

// UserService handles business logic for users
type UserService struct {
    repo  ports.UserRepository
    tracer *telemetry.ServiceTracer
}

func NewUserService(repo ports.UserRepository) *UserService {
    return &UserService{
        repo:  repo,
        tracer: telemetry.GetServiceTracer(),
    }
}

// CreateUser with enhanced tracing
func (s *UserService) CreateUser(ctx context.Context, user *entities.User) error {
    // Method 1: Using ServiceOperation
    op := telemetry.NewServiceOperation(ctx, "UserService", "CreateUser")
    defer op.End(nil)

    // Add custom attributes
    op.AddAttribute("user.email", user.Email)
    op.AddAttribute("user.role", user.Role)

    // Trace business logic
    return s.tracer.TraceBusinessOperation(ctx, "user.creation", func(ctx context.Context) error {
        // Validation logic
        if err := s.validateUser(user); err != nil {
            op.AddError(err)
            return err
        }

        // Repository operation with tracing
        return s.tracer.TraceRepositoryOperation(ctx, "create", "User", func(ctx context.Context) error {
            return s.repo.Create(ctx, user)
        })
    })
}

// GetUser with detailed tracing
func (s *UserService) GetUser(ctx context.Context, id int) (*entities.User, error) {
    return s.tracer.TraceServiceMethod(ctx, "UserService", "GetUser", map[string]interface{}{
        "user_id": id,
    }, func(ctx context.Context) error {
        var err error
        var user *entities.User

        // Trace repository call
        err = s.tracer.TraceRepositoryOperation(ctx, "get_by_id", "User", func(ctx context.Context) error {
            user, err = s.repo.GetByID(ctx, id)
            return err
        })

        // Add result metrics if successful
        if err == nil && user != nil {
            s.tracer.AddServiceMetric(ctx, "user.found", 1.0)
        }

        return err
    })
}

// ListUsers with pagination tracing
func (s *UserService) ListUsers(ctx context.Context) ([]*entities.User, error) {
    op := telemetry.NewServiceOperation(ctx, "UserService", "ListUsers")
    defer op.End(nil)

    var users []*entities.User
    var err error

    // Trace the actual database operation
    err = s.tracer.TraceRepositoryOperation(ctx, "list", "User", func(ctx context.Context) error {
        users, err = s.repo.List(ctx)
        return err
    })

    if err == nil {
        // Add result metrics
        s.tracer.AddServiceMetric(ctx, "users.count", float64(len(users)))
        op.AddAttribute("result.count", len(users))
    }

    return users, err
}

// UpdateUser with validation tracing
func (s *UserService) UpdateUser(ctx context.Context, user *entities.User) error {
    return s.tracer.TraceServiceMethod(ctx, "UserService", "UpdateUser", map[string]interface{}{
        "user_id": user.ID,
        "updates": len(user.GetChanges()), // Assuming you have a method to track changes
    }, func(ctx context.Context) error {
        return s.tracer.TraceBusinessOperation(ctx, "user.update", func(ctx context.Context) error {
            // Business validation
            if err := s.validateUpdate(user); err != nil {
                s.tracer.AddBusinessAttribute(ctx, "validation.error", err.Error())
                return err
            }

            // Repository operation
            return s.tracer.TraceRepositoryOperation(ctx, "update", "User", func(ctx context.Context) error {
                return s.repo.Update(ctx, user)
            })
        })
    })
}

// Private validation method with tracing
func (s *UserService) validateUser(user *entities.User) error {
    return s.tracer.TraceBusinessOperation(ctx, "user.validation", func(ctx context.Context) error {
        // Add validation metrics
        s.tracer.AddBusinessAttribute(ctx, "validation.type", "user_creation")

        if user.Email == "" {
            return errors.New("email is required")
        }

        if user.Password == "" {
            return errors.New("password is required")
        }

        return nil
    })
}

func (s *UserService) validateUpdate(user *entities.User) error {
    return s.tracer.TraceBusinessOperation(ctx, "user.validation.update", func(ctx context.Context) error {
        // Business validation logic
        s.tracer.AddBusinessAttribute(ctx, "validation.type", "user_update")

        if user.ID == 0 {
            return errors.New("user ID is required for update")
        }

        return nil
    })
}
```

## Handler Layer Tracing Examples

### Enhanced User Handler

```go
package handlers

import (
    "context"
    "example.com/go-yippi/internal/application/services"
    "example.com/go-yippi/internal/infrastructure/telemetry"
    "github.com/danielgtaylor/huma/v2"
)

type UserHandler struct {
    service ports.UserService
    tracer  *telemetry.ServiceTracer
}

func NewUserHandler(service ports.UserService) *UserHandler {
    return &UserHandler{
        service: service,
        tracer:  telemetry.GetServiceTracer(),
    }
}

// CreateUser with enhanced tracing
func (h *UserHandler) CreateUser(ctx context.Context, input *CreateUserRequest) (*CreateUserResponse, error) {
    // Add request context to span
    ctx = h.tracer.WithRequestContext(ctx,
        c.GetRespHeader("X-Request-ID"),
        c.IP(),
        c.Get("User-Agent"),
    )

    return h.tracer.TraceServiceMethod(ctx, "UserHandler", "CreateUser", input.Body, func(ctx context.Context) error {
        // Business logic execution
        err := h.service.CreateUser(ctx, input.Body)
        if err != nil {
            // Add error to current span
            telemetry.AddErrorToSpan(ctx, err)
            return err
        }

        // Add success metrics
        h.tracer.AddServiceMetric(ctx, "user.created", 1.0)
        return nil
    })
}

// GetUser with detailed error tracing
func (h *UserHandler) GetUser(ctx context.Context, input *GetUserRequest) (*GetUserResponse, error) {
    return h.tracer.TraceServiceMethod(ctx, "UserHandler", "GetUser", map[string]interface{}{
        "user_id": input.ID,
    }, func(ctx context.Context) error {
        user, err := h.service.GetUser(ctx, input.ID)
        if err != nil {
            // Classify error type
            telemetry.AddCustomSpanAttribute(ctx, "error.classification", "user_not_found")
            telemetry.AddErrorToSpan(ctx, err)
            return huma.Error404NotFound("User not found")
        }

        // Add success attributes
        telemetry.AddCustomSpanAttribute(ctx, "user.found", true)
        h.tracer.AddServiceMetric(ctx, "user.retrieved", 1.0)

        return nil
    })
}
```

## Repository Layer Tracing Examples

### Enhanced User Repository

```go
package persistence

import (
    "context"
    "example.com/go-yippi/internal/domain/entities"
    "example.com/go-yippi/internal/domain/ports"
    "example.com/go-yippi/internal/infrastructure/telemetry"
)

type UserRepository struct {
    client *ent.Client
    tracer *telemetry.ServiceTracer
}

func NewUserRepository(client *ent.Client) ports.UserRepository {
    return &UserRepository{
        client: client,
        tracer: telemetry.GetServiceTracer(),
    }
}

// Create with database tracing
func (r *UserRepository) Create(ctx context.Context, user *entities.User) error {
    return r.tracer.TraceRepositoryOperation(ctx, "create", "User", func(ctx context.Context) error {
        // The Ent hooks will automatically trace the database operations
        // We add additional business context here
        r.tracer.AddBusinessAttribute(ctx, "operation.type", "user_creation")

        // Add custom timing
        startTime := time.Now()
        err := r.client.User.Create().
            SetName(user.Name).
            SetEmail(user.Email).
            Exec(ctx)

        duration := time.Since(startTime)
        r.tracer.AddServiceMetric(ctx, "db.create.duration_ms", float64(duration.Nanoseconds())/1e6)

        return err
    })
}

// GetByID with cache simulation
func (r *UserRepository) GetByID(ctx context.Context, id int) (*entities.User, error) {
    return r.tracer.TraceRepositoryOperation(ctx, "get_by_id", "User", func(ctx context.Context) error {
        // Simulate cache check
        cacheHit := r.checkCache(id)
        r.tracer.AddBusinessAttribute(ctx, "cache.hit", cacheHit)

        var user *entities.User
        var err error

        if cacheHit {
            // Simulate cache retrieval
            user = r.getFromCache(id)
            r.tracer.AddServiceMetric(ctx, "cache.hit_count", 1.0)
        } else {
            // Database query (automatically traced by hooks)
            user, err = r.client.User.Get(ctx, id)
            if err == nil {
                r.cacheUser(user)
                r.tracer.AddServiceMetric(ctx, "cache.miss_count", 1.0)
            }
        }

        return err
    })
}
```

## Complex Business Operation Example

### User Registration Workflow

```go
// RegisterUser with multi-step tracing
func (s *UserService) RegisterUser(ctx context.Context, registration *UserRegistration) error {
    op := telemetry.NewServiceOperation(ctx, "UserService", "RegisterUser")
    defer op.End(nil)

    op.AddAttribute("registration.source", registration.Source)
    op.AddAttribute("registration.marketing_consent", registration.MarketingConsent)

    // Step 1: Validate input
    validationStart := time.Now()
    if err := s.validateRegistration(registration); err != nil {
        op.AddAttribute("validation.failed", true)
        op.AddError(err)
        return err
    }
    validationDuration := time.Since(validationStart)
    s.tracer.AddServiceMetric(ctx, "validation.duration_ms", float64(validationDuration.Nanoseconds())/1e6)

    // Step 2: Check for existing user
    var existingUser *entities.User
    err := s.tracer.TraceBusinessOperation(ctx, "user.duplicate_check", func(ctx context.Context) error {
        existingUser, err = s.repo.GetByEmail(ctx, registration.Email)
        return err
    })

    if existingUser != nil {
        op.AddAttribute("user.duplicate", true)
        return errors.New("user already exists")
    }

    // Step 3: Create user with transaction
    return s.tracer.TraceTransaction(ctx, func(ctx context.Context, tx *ent.Tx) error {
        // This will trace the entire transaction
        user := &entities.User{
            Name:  registration.Name,
            Email: registration.Email,
            Role:  "user",
        }

        // Create main user record
        if err := s.repo.CreateWithTx(ctx, tx, user); err != nil {
            return err
        }

        // Step 4: Create user profile
        profile := &entities.UserProfile{
            UserID:    user.ID,
            FirstName:  registration.FirstName,
            LastName:   registration.LastName,
        }

        err := s.tracer.TraceBusinessOperation(ctx, "profile.creation", func(ctx context.Context) error {
            return s.createProfileWithTx(ctx, tx, profile)
        })

        if err != nil {
            return err
        }

        // Step 5: Send welcome email (simulated)
        err = s.tracer.TraceBusinessOperation(ctx, "email.send_welcome", func(ctx context.Context) error {
            return s.sendWelcomeEmail(ctx, user.Email)
        })

        if err != nil {
            // Email failure doesn't fail registration
            s.tracer.AddBusinessAttribute(ctx, "email.failed", true)
            return nil
        }

        op.AddAttribute("registration.completed", true)
        s.tracer.AddServiceMetric(ctx, "user.registered", 1.0)
        return nil
    })
}
```

## Key Tracing Patterns

### 1. **Service Method Pattern**
```go
func (s *Service) Method(ctx context.Context, input Input) error {
    op := telemetry.NewServiceOperation(ctx, "ServiceName", "MethodName")
    defer op.End(nil)

    op.AddAttribute("input.type", "user_data")

    // Business logic here
    err := s.someOperation(ctx, input)
    if err != nil {
        op.AddError(err)
    }

    return err
}
```

### 2. **Repository Pattern**
```go
func (r *Repository) Method(ctx context.Context, param Param) error {
    return r.tracer.TraceRepositoryOperation(ctx, "operation", "Entity", func(ctx context.Context) error {
        // Database operation (automatically traced by Ent hooks)
        return r.client.Entity.DoSomething(ctx, param)
    })
}
```

### 3. **Business Logic Pattern**
```go
func (s *Service) businessLogic(ctx context.Context) error {
    return s.tracer.TraceBusinessOperation(ctx, "business.operation", func(ctx context.Context) error {
        s.tracer.AddBusinessAttribute(ctx, "business.rule", "validation")
        s.tracer.AddServiceMetric(ctx, "validation.attempts", 1.0)

        // Business logic here
        return nil
    })
}
```

## Viewing Traces in SigNoz

With these examples, you'll see rich traces in SigNoz including:

1. **HTTP Request Layer**: Request/response timing, headers, errors
2. **Handler Layer**: Request processing, validation, response creation
3. **Service Layer**: Business logic execution, validation, business rules
4. **Repository Layer**: Database operations, transactions, query performance
5. **Business Operations**: Multi-step workflows with custom attributes

Each trace will show:
- Execution time for each layer
- Custom attributes you've added
- Error details with context
- Performance metrics
- Parent-child relationship between operations

This comprehensive tracing helps you understand exactly where performance bottlenecks occur and how requests flow through your application.