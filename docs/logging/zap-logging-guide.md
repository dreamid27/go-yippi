# Zap Structured Logging Implementation

This document describes the comprehensive Zap logging implementation for the Go Yippi API, providing structured, high-performance logging with OpenTelemetry integration.

## 🚀 **Features Implemented**

### Core Logging Infrastructure
- **Structured Logging**: JSON format with proper field structure
- **Performance Optimized**: High-performance logging with Zap
- **Configurable Levels**: Debug, Info, Warn, Error, Fatal
- **Multiple Outputs**: Console, file, stderr support
- **Log Rotation**: Automatic file rotation with compression
- **Context Propagation**: Automatic trace ID and span ID injection

### OpenTelemetry Integration
- **Trace Context**: Automatic trace ID and span ID injection
- **Correlation Links**: Log and trace correlation
- **Baggage Support**: Distributed context propagation
- **Service Attributes**: Service name, version, and metadata

### Middleware Stack
- **HTTP Request Logging**: Request/response cycle logging
- **Security Logging**: Suspicious activity detection
- **Audit Logging**: Sensitive operation tracking
- **Request ID Middleware**: Unique request tracking

## ⚙️ **Configuration**

### Environment Variables

```bash
# Logging Configuration
LOG_LEVEL=info                    # debug, info, warn, error
LOG_FORMAT=json                   # json or console
LOG_OUTPUT=stdout                  # stdout, stderr, or file path
LOG_MAX_SIZE=100                   # Max log file size in MB
LOG_MAX_BACKUPS=3                  # Number of old log files to keep
LOG_MAX_AGE=28                     # Max age in days
LOG_COMPRESS=true                   # Compress old log files
```

### Log Levels
- **debug**: Detailed development information
- **info**: General application flow (default)
- **warn**: Warning conditions that don't stop execution
- **error**: Error conditions that are recoverable
- **fatal**: Critical errors that stop the application

### Output Formats

#### JSON Format (Production)
```json
{
  "level": "info",
  "timestamp": "2025-11-25T08:44:13.898+0700",
  "caller": "logging/zaplogger.go:122",
  "function": "example.com/go-yippi/internal/infrastructure/logging.Initialize",
  "message": "Logger initialized",
  "service": "go-yippi-api",
  "version": "1.0.0",
  "trace_id": "abc123def456",
  "span_id": "xyz789uvw012"
}
```

#### Console Format (Development)
```
2025-11-25T08:44:13.898+0700	info	logging/zaplogger.go	Example message	{"service": "go-yippi-api"}
```

## 🔄 **Middleware Implementation**

### HTTP Request Logging
Captures complete request/response lifecycle:

```go
app.Use(logging.HTTPLoggerMiddleware())
```

**Fields Logged:**
- `event`: "http_request"
- `method`: HTTP method (GET, POST, etc.)
- `path`: Request path
- `query`: Query parameters
- `user_agent`: Client User-Agent
- `client_ip`: Client IP address
- `status_code`: HTTP status code
- `duration_ms`: Request duration
- `request_size`: Request body size
- `response_size`: Response body size
- `request_id`: Unique request identifier

### Security Logging
Detects suspicious activities:

```go
app.Use(logging.SecurityLoggerMiddleware())
```

**Security Events:**
- Suspicious User-Agent detection
- SQL injection attempt detection
- Path traversal attempt detection

**Fields:**
- `event`: "security_event"
- `security_event`: Event type
- `severity`: high/medium/low
- `client_ip`: Attacker IP
- `user_agent`: User-Agent
- `path`: Requested path

### Audit Logging
Tracks sensitive operations:

```go
app.Use(logging.AuditLoggerMiddleware())
```

**Audited Operations:**
- User creation/deletion/modification
- Product operations
- Authentication events
- Configuration changes

**Fields:**
- `event`: "audit_event"
- `method`: HTTP method
- `path`: Requested path
- `status_code`: Response code
- `user_id`: User ID (if available)
- `client_ip`: Client IP
- `timestamp`: ISO8601 timestamp

## 📊 **Structured Logging Functions**

### Core Logging

```go
// Global logger usage
logging.GetGlobalLogger().Info("Message", zap.String("key", "value"))

// Trace-aware logging
tl := logging.NewTraceLogger()
tl.InfoWithTrace(ctx, "Operation completed", zap.String("operation", "user_created"))
```

### Domain-Specific Logging

#### HTTP Request Logging
```go
logging.LogHTTPRequest(ctx, method, path, userAgent, clientIP, statusCode, duration, err)
```

#### Database Operation Logging
```go
logging.LogDatabaseOperation(ctx, "create", "User", duration, err, 1)
```

#### Service Operation Logging
```go
logging.LogServiceOperation(ctx, "UserService", "CreateUser", duration, err, user)
```

#### Business Event Logging
```go
logging.LogBusinessEvent(ctx, "user_registered", map[string]interface{}{
    "user_id": user.ID,
    "email": user.Email,
    "role": user.Role,
})
```

#### Security Event Logging
```go
logging.LogSecurityEvent(ctx, "suspicious_activity", "high", map[string]interface{}{
    "activity": "sql_injection_attempt",
    "pattern": detectedPattern,
    "client_ip": clientIP,
})
```

## 🔧 **Usage Examples**

### In Services

```go
package services

import (
    "example.com/go-yippi/internal/infrastructure/logging"
    "go.uber.org/zap"
)

type UserService struct {
    repo   ports.UserRepository
    logger *zap.Logger
}

func NewUserService(repo ports.UserRepository) *UserService {
    return &UserService{
        repo:   repo,
        logger: logging.GetGlobalLogger(),
    }
}

func (s *UserService) CreateUser(ctx context.Context, user *entities.User) error {
    // Log service operation start
    start := time.Now()

    // Validate input
    if err := s.validateUser(user); err != nil {
        s.logger.Error("User validation failed",
            zap.Error(err),
            zap.String("email", user.Email),
            zap.Float64("validation_duration_ms", float64(time.Since(start).Nanoseconds())/1e6),
        )
        return err
    }

    // Create user
    err := s.repo.Create(ctx, user)

    // Log operation result
    duration := time.Since(start)
    if err != nil {
        s.logger.Error("User creation failed",
            zap.Error(err),
            zap.String("email", user.Email),
            zap.Float64("duration_ms", float64(duration.Nanoseconds())/1e6),
            zap.String("operation", "create_user"),
        )
    } else {
        s.logger.Info("User created successfully",
            zap.String("user_id", fmt.Sprintf("%d", user.ID)),
            zap.String("email", user.Email),
            zap.Float64("duration_ms", float64(duration.Nanoseconds())/1e6),
            zap.String("operation", "create_user"),
        )
    }

    return err
}

func (s *UserService) validateUser(user *entities.User) error {
    start := time.Now()

    if user.Email == "" {
        s.logger.Warn("Email validation failed",
            zap.String("field", "email"),
            zap.String("reason", "required"),
            zap.Float64("validation_duration_ms", float64(time.Since(start).Nanoseconds())/1e6),
        )
        return errors.New("email is required")
    }

    if len(user.Password) < 8 {
        s.logger.Warn("Password validation failed",
            zap.String("field", "password"),
            zap.String("reason", "too_short"),
            zap.Int("min_length", 8),
            zap.Int("actual_length", len(user.Password)),
            zap.Float64("validation_duration_ms", float64(time.Since(start).Nanoseconds())/1e6),
        )
        return errors.New("password too short")
    }

    return nil
}
```

### In Handlers

```go
package handlers

import (
    "example.com/go-yippi/internal/infrastructure/logging"
    "go.uber.org/zap"
    "github.com/danielgtaylor/huma/v2"
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

func (h *UserHandler) CreateUser(ctx context.Context, input *CreateUserRequest) (*CreateUserResponse, error) {
    start := time.Now()

    // Log incoming request
    h.logger.Info("Create user request received",
        zap.String("request_id", ctx.Value("request_id").(string)),
        zap.String("user_email", input.Body.Email),
        zap.String("operation", "create_user_request"),
    )

    // Call service
    err := h.service.CreateUser(ctx, &entities.User{
        Name:  input.Body.Name,
        Email: input.Body.Email,
        Password: input.Body.Password,
    })

    // Log result
    duration := time.Since(start)
    if err != nil {
        h.logger.Error("Create user request failed",
            zap.Error(err),
            zap.String("request_id", ctx.Value("request_id").(string)),
            zap.Float64("duration_ms", float64(duration.Nanoseconds())/1e6),
            zap.String("operation", "create_user_request"),
        )
        return nil, huma.Error400BadRequest("User creation failed")
    }

    h.logger.Info("Create user request completed",
        zap.String("request_id", ctx.Value("request_id").(string)),
        zap.String("user_email", input.Body.Email),
        zap.Float64("duration_ms", float64(duration.Nanoseconds())/1e6),
        zap.String("operation", "create_user_request"),
        zap.String("result", "success"),
    )

    response := &CreateUserResponse{}
    response.Body.ID = 123 // Placeholder ID
    return response, nil
}
```

## 🔍 **Log Analysis Examples**

### Request Flow Trace
```json
{
  "level": "info",
  "timestamp": "2025-11-25T08:44:13.898+0700",
  "event": "http_request",
  "method": "POST",
  "path": "/api/users",
  "user_agent": "curl/7.68.0",
  "client_ip": "192.168.1.100",
  "status_code": 201,
  "duration_ms": 245.67,
  "request_size": 156,
  "response_size": 89,
  "request_id": "req_17009190538901"
}
```

### Security Event Log
```json
{
  "level": "info",
  "timestamp": "2025-11-25T08:44:13.898+0700",
  "event": "security_event",
  "security_event": "sql_injection_attempt",
  "severity": "high",
  "client_ip": "10.0.0.1",
  "user_agent": "sqlmap/1.0",
  "path": "/api/users",
  "security.pattern": "UNION SELECT",
  "security.body": "' OR 1=1 --"
}
```

### Audit Event Log
```json
{
  "level": "info",
  "timestamp": "2025-11-25T08:44:13.898+0700",
  "event": "audit_event",
  "method": "POST",
  "path": "/api/users",
  "status_code": 201,
  "user_id": "user_12345",
  "client_ip": "192.168.1.100",
  "timestamp": "2025-11-25T08:44:13.898+0700"
}
```

## 📈 **Performance Considerations**

### Logging Performance
- **Zap**: Zero-allocation, 10x faster than standard library
- **Sampling**: Configurable sampling for high-traffic scenarios
- **Async**: Buffered writes for improved throughput
- **Compression**: Automatic log rotation with gzip compression

### Memory Usage
- **Field Allocation**: Reuse field builders for memory efficiency
- **Context Propagation**: Minimal overhead with efficient string handling
- **String Formatting**: Pre-computed timestamps and IDs

### CPU Optimization
- **JSON Encoding**: Efficient JSON marshaling
- **Caller Lookup**: Optimized caller information
- **Level Filtering**: Early level filtering to avoid unnecessary processing

## 🛠️ **Advanced Configuration**

### Custom Log Configuration

```go
// Custom log configuration
config := &logging.LogConfig{
    Level:      "debug",
    Format:     "json",
    Output:     "/var/log/go-yippi/app.log",
    MaxSize:    500,  // 500MB
    MaxBackups: 10,
    MaxAge:     30,    // 30 days
    Compress:   true,
}

// Initialize custom logger
err := logging.Initialize(config)
```

### Development vs Production

#### Development Settings
```bash
LOG_LEVEL=debug
LOG_FORMAT=console
LOG_OUTPUT=stdout
```

#### Production Settings
```bash
LOG_LEVEL=info
LOG_FORMAT=json
LOG_OUTPUT=/var/log/go-yippi/app.log
LOG_MAX_SIZE=100
LOG_MAX_BACKUPS=10
LOG_MAX_AGE=30
LOG_COMPRESS=true
```

## 🔧 **Integration with Monitoring Systems**

### SigNoz Integration
Logs automatically correlate with traces:

```go
// Traced logging with correlation
logging.LogServiceOperation(ctx, "UserService", "CreateUser", duration, err, user)
// Automatically includes: trace_id, span_id, service.name
```

### Log Aggregation
Compatible with log aggregation tools:

1. **ELK Stack** (Elasticsearch + Logstash + Kibana)
2. **Grafana Loki** (JSON format)
3. **Datadog** (JSON with proper tagging)
4. **Splunk** (Structured log ingestion)
5. **Sumo Logic** (Cloud-based log management)

### Alerting Examples

Set up alerts based on log patterns:

```yaml
# Example: Rate limit alerts
groups:
  - name: Security Events
    rules:
      - alert: HighSecurityEvents
        expr: rate(security_events[5m]) > 10
        for: 2m
        labels:
          severity: critical
          service: go-yippi-api

  - name: High Error Rate
    rules:
      - alert: HighErrorRate
        expr: rate(error_logs[5m]) > 5
        for: 1m
        labels:
          severity: warning
          service: go-yippi-api
```

## 🧪 **Best Practices**

### 1. **Structured Data**
- Always use structured fields for log context
- Use consistent field names across services
- Include trace context when available
- Add business domain information

### 2. **Error Handling**
- Log errors at the point of occurrence
- Include error context and stack traces
- Use appropriate log levels
- Don't log sensitive data

### 3. **Performance**
- Avoid string formatting in hot paths
- Use field builders for repeated patterns
- Sample logs in high-traffic scenarios
- Use appropriate output formats

### 4. **Security**
- Never log passwords, tokens, or API keys
- Sanitize PII before logging
- Log security events separately
- Use correlation IDs for investigation

## 📚 **Migration from Standard Library**

### Before (Standard Library)
```go
log.Printf("User %s created: %v", user.Email, err)
```

### After (Zap Structured)
```go
logger.Info("User created",
    zap.String("email", user.Email),
    zap.Error(err),
    zap.String("operation", "create_user"),
    zap.Duration("processing_time", duration),
)
```

### Benefits of Migration
- **Searchable**: JSON logs are easily searchable
- **Filterable**: Structured data enables precise filtering
- **Correlatable**: Trace context links related events
- **Analyzable**: Rich data for analysis and alerting
- **Performance**: Much higher logging throughput

This comprehensive Zap implementation provides production-ready logging with full observability stack integration for the Go Yippi API.