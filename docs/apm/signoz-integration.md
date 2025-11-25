# SigNoz APM Integration

This document describes how to configure and use SigNoz for Application Performance Monitoring (APM) with the Go Yippi API.

## Overview

The application is integrated with OpenTelemetry to send traces, metrics, and logs to a SigNoz backend. This enables you to monitor:

- HTTP request latency and error rates
- Database query performance
- Service dependency mapping
- Distributed tracing across services

## Configuration

### Environment Variables

Add the following environment variables to your `.env` file:

```bash
# Enable OpenTelemetry tracing
OTEL_ENABLED=true

# Service name for identification in SigNoz
OTEL_SERVICE_NAME=go-yippi-api

# SigNoz collector endpoint
OTEL_EXPORTER_OTLP_ENDPOINT=your-signoz-host:4317

# Connection security (set to false for local development)
OTEL_INSECURE_MODE=true
```

### Configuration Options

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `OTEL_ENABLED` | Enable/disable OpenTelemetry | `false` | `true` |
| `OTEL_SERVICE_NAME` | Service name displayed in SigNoz | `go-yippi-api` | `production-api` |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | SigNoz collector address | `localhost:4317` | `signoz.example.com:4317` |
| `OTEL_INSECURE_MODE` | Use insecure connection (no TLS) | `true` | `false` |
| `LOG_LEVEL` | Logging level (debug/info/warn/error) | `info` | `debug` |
| `LOG_FORMAT` | Log format (json/console) | `json` | `console` |
| `LOG_OUTPUT` | Log output (stdout/stderr/file) | `stdout` | `/var/log/go-yippi/app.log` |
| `LOG_MAX_SIZE` | Max log file size in MB | `100` | `500` |
| `LOG_MAX_BACKUPS` | Number of old log files to keep | `3` | `10` |
| `LOG_MAX_AGE` | Max age of log files in days | `28` | `30` |
| `LOG_COMPRESS` | Compress old log files | `true` | `false` |

### Enhanced Tracing Features

The enhanced implementation provides:

#### 1. **HTTP Request Tracing**
- Automatic request/response cycle tracing
- Request duration measurement with millisecond precision
- Rich metadata collection (IP, User-Agent, headers)
- Error correlation with HTTP status codes

#### 2. **Database Query Tracing**
- Automatic Ent ORM operation tracing
- Transaction lifecycle monitoring
- Query performance metrics
- Database error classification

#### 3. **Service Layer Tracing**
- Business operation timing
- Repository call tracing
- Custom attribute support
- Error propagation with context

#### 4. **Custom Tracing Utilities**
- ServiceTracer for business logic tracing
- ServiceOperation for method-level tracing
- Attribute and metric collection
- Context propagation utilities

## SigNoz Setup

### Option 1: Docker Compose (Recommended for Development)

1. Clone the SigNoz repository:
```bash
git clone https://github.com/SigNoz/signoz.git
cd signoz/deploy/
```

2. Start SigNoz with Docker Compose:
```bash
docker-compose -f docker/clickhouse-setup/docker-compose.yaml up -d
```

3. Access the SigNoz UI at `http://localhost:3301`

### Option 2: Self-Hosted Production

Follow the official SigNoz deployment guide for production setup:
[https://signoz.io/docs/install/](https://signoz.io/docs/install/)

## Running the Application with SigNoz

### Development

1. Start your SigNoz instance
2. Update your `.env` file with the configuration
3. Run the application:

```bash
# Using make
make run

# Or directly
go run cmd/api/main.go
```

4. Make API calls to generate traces
5. View traces in SigNoz UI at `http://localhost:3301`

### Production

1. Deploy SigNoz in your infrastructure
2. Update environment variables with your SigNoz endpoint
3. Set `OTEL_INSECURE_MODE=false` for production
4. Deploy your application

## Features

### Automatic Instrumentation

The application automatically captures:

#### Logging (Zap)
- **HTTP Request/Response**: Complete request lifecycle with structured logging
- **Security Events**: Suspicious activity detection and alerting
- **Audit Events**: Sensitive operation tracking
- **Database Operations**: Query execution logging
- **Service Operations**: Business logic execution logging
- **Business Events**: Domain-specific event logging

#### Tracing (OpenTelemetry)
- HTTP request/response cycles
- Request/response headers
- Database queries (if using OpenTelemetry database instrumentation)
- Custom spans from business logic
- Cross-service context propagation (when available)

### Custom Spans

You can add custom spans to your business logic using the service tracer:

```go
import (
    "example.com/go-yippi/internal/infrastructure/telemetry"
)

// Method 1: Using ServiceTracer
func businessFunction(ctx context.Context) error {
    st := telemetry.GetServiceTracer()
    return st.TraceBusinessOperation(ctx, "business-operation", func(ctx context.Context) error {
        // Your business logic here
        st.AddBusinessAttribute(ctx, "operation.type", "validation")
        st.AddServiceMetric(ctx, "validation.count", 1.0)
        return nil
    })
}

// Method 2: Using ServiceOperation
func anotherBusinessFunction(ctx context.Context) error {
    op := telemetry.NewServiceOperation(ctx, "UserService", "CreateUser")
    defer op.End(nil)

    // Add custom attributes
    op.AddAttribute("user.role", "admin")

    // Your business logic here
    return nil
}

// Method 3: Simple span with utility
func simpleBusinessFunction(ctx context.Context) error {
    tracer := telemetry.GetTracer("business-logic")
    ctx, span := tracer.Start(ctx, "business-operation")
    defer span.End()

    // Your business logic here
    return nil
}
```

### What Gets Tracked

- **HTTP Requests**: All incoming HTTP requests with detailed metrics:
  - Request duration in milliseconds
  - HTTP method, URL, and status code
  - Client IP and User-Agent
  - Response content length
  - Error details with stack traces

- **Database Operations**: Comprehensive database tracing:
  - All CRUD operations (Create, Read, Update, Delete)
  - Transaction lifecycle (begin, commit, rollback)
  - Query execution time
  - Database error details
  - Entity types being accessed

- **Service Layer**: Business logic execution:
  - Service method execution time
  - Repository operation tracing
  - Business operation metrics
  - Custom attributes and error tracking

- **File Operations**: MinIO/S3 operations (if configured)
- **Custom Business Logic**: Any manually added spans with rich metadata

## Viewing Traces in SigNoz

1. Open SigNoz UI (`http://localhost:3301`)
2. Go to the **Services** tab to see your service
3. Click on your service name (`go-yippi-api`)
4. Use the **Traces** tab to view individual request traces
5. Use **Metrics** tab for performance graphs
6. Use **Logs** tab for correlated logs

## Troubleshooting

### Common Issues

1. **No traces appearing**:
   - Verify `OTEL_ENABLED=true`
   - Check SigNoz collector is running
   - Verify network connectivity to `OTEL_EXPORTER_OTLP_ENDPOINT`

2. **Connection refused errors**:
   - Ensure SigNoz collector is running on the expected port
   - Check firewall rules between application and SigNoz

3. **TLS errors**:
   - Set `OTEL_INSECURE_MODE=true` for development
   - Configure proper certificates for production

### Debug Mode

Enable debug logging for OpenTelemetry by setting:
```bash
export OTEL_LOG_LEVEL=debug
```

### Health Check

Verify the SigNoz collector is accessible:
```bash
curl -I http://localhost:4317/health
```

## Performance Impact

OpenTelemetry has minimal performance impact:
- ~1-2% overhead for HTTP tracing
- Asynchronous batch sending reduces blocking
- Configurable sampling rates for high-traffic scenarios

For high-traffic production environments, consider using probabilistic sampling by modifying the tracer configuration.

## Next Steps

1. Set up alerts in SigNoz for error rates and latency
2. Create custom dashboards for business metrics
3. Add custom spans to critical business operations
4. Configure logging integration for correlated logs

## Additional Resources

- [SigNoz Documentation](https://signoz.io/docs/)
- [OpenTelemetry Go Documentation](https://opentelemetry.io/docs/instrumentation/go/)
- [Fiber OpenTelemetry Integration](https://github.com/gofiber/contrib/tree/main/otelfiber)