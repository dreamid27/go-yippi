package observability

import (
	"context"
	"time"

	"example.com/go-yippi/internal/infrastructure/logging"
	"example.com/go-yippi/internal/infrastructure/telemetry"
	"go.uber.org/zap"
)

// ServiceObserver combines tracing, logging, and metrics for service operations
type ServiceObserver struct {
	serviceName string
	tracer      *telemetry.ServiceTracer
	logger      *logging.TraceLogger
}

// NewServiceObserver creates a new service observer
func NewServiceObserver(serviceName string) *ServiceObserver {
	return &ServiceObserver{
		serviceName: serviceName,
		tracer:      telemetry.NewServiceTracer(),
		logger:      logging.NewTraceLogger(),
	}
}

// Operation represents a service operation with full observability
type Operation struct {
	ctx         context.Context
	serviceName string
	methodName  string
	startTime   time.Time
	tracer      *telemetry.ServiceTracer
	logger      *logging.TraceLogger
	operation   *telemetry.ServiceOperation
}

// StartOperation starts a new observable service operation
func (so *ServiceObserver) StartOperation(ctx context.Context, methodName string) *Operation {
	// Create service operation (starts tracing span)
	serviceOp := telemetry.NewServiceOperation(ctx, so.serviceName, methodName)

	// Log operation start
	so.logger.InfoWithTrace(serviceOp.Context(), "Service operation started",
		zap.String("service", so.serviceName),
		zap.String("method", methodName),
	)

	return &Operation{
		ctx:         serviceOp.Context(),
		serviceName: so.serviceName,
		methodName:  methodName,
		startTime:   time.Now(),
		tracer:      so.tracer,
		logger:      so.logger,
		operation:   serviceOp,
	}
}

// Context returns the operation context (with trace propagation)
func (o *Operation) Context() context.Context {
	return o.ctx
}

// AddAttribute adds an attribute to both trace and logs
func (o *Operation) AddAttribute(key string, value interface{}) {
	o.operation.AddAttribute(key, value)
}

// LogInfo logs informational message during operation
func (o *Operation) LogInfo(message string, fields ...zap.Field) {
	o.logger.InfoWithTrace(o.ctx, message, fields...)
}

// LogWarn logs warning message during operation
func (o *Operation) LogWarn(message string, fields ...zap.Field) {
	o.logger.WarnWithTrace(o.ctx, message, fields...)
}

// LogError logs error message during operation
func (o *Operation) LogError(message string, err error, fields ...zap.Field) {
	o.logger.ErrorWithTrace(o.ctx, message, err, fields...)
}

// RecordBusinessEvent records a business event with logging and metrics
func (o *Operation) RecordBusinessEvent(eventType, eventName string, fields ...zap.Field) {
	// Record metric
	telemetry.RecordBusinessEvent(o.ctx, eventType, eventName)

	// Log business event
	allFields := append(fields,
		zap.String("service", o.serviceName),
		zap.String("business_event_type", eventType),
		zap.String("business_event_name", eventName),
	)
	o.logger.InfoWithTrace(o.ctx, "Business event occurred", allFields...)
}

// End completes the operation with full observability
func (o *Operation) End(err error) {
	duration := time.Since(o.startTime)

	// Record metrics
	telemetry.RecordServiceOperation(o.ctx, o.serviceName, o.methodName, duration, err)

	// End tracing span
	o.operation.End(err)

	// Log operation completion
	fields := []zap.Field{
		zap.String("service", o.serviceName),
		zap.String("method", o.methodName),
		zap.Float64("duration_ms", float64(duration.Nanoseconds())/1e6),
	}

	if err != nil {
		fields = append(fields, zap.Error(err))
		o.logger.ErrorWithTrace(o.ctx, "Service operation failed", err, fields...)
	} else {
		o.logger.InfoWithTrace(o.ctx, "Service operation completed", fields...)
	}
}

// TraceFunc wraps a function with complete observability (convenience method)
func (so *ServiceObserver) TraceFunc(ctx context.Context, methodName string, fn func(context.Context) error) error {
	op := so.StartOperation(ctx, methodName)
	err := fn(op.Context())
	op.End(err)
	return err
}

// TraceFuncWithResult wraps a function that returns a result with complete observability
func TraceFuncWithResult[T any](so *ServiceObserver, ctx context.Context, methodName string, fn func(context.Context) (T, error)) (T, error) {
	op := so.StartOperation(ctx, methodName)
	result, err := fn(op.Context())
	op.End(err)
	return result, err
}

// Helper functions for common logging patterns

// LogValidationError logs validation errors with proper context
func LogValidationError(ctx context.Context, field, message string) {
	logger := logging.NewTraceLogger()
	logger.WarnWithTrace(ctx, "Validation error",
		zap.String("error_type", "validation"),
		zap.String("field", field),
		zap.String("message", message),
	)
}

// LogNotFoundError logs not found errors
func LogNotFoundError(ctx context.Context, entity string, identifier interface{}) {
	logger := logging.NewTraceLogger()
	logger.WarnWithTrace(ctx, "Entity not found",
		zap.String("error_type", "not_found"),
		zap.String("entity", entity),
		zap.Any("identifier", identifier),
	)
}

// LogDuplicateError logs duplicate entry errors
func LogDuplicateError(ctx context.Context, entity, field string, value interface{}) {
	logger := logging.NewTraceLogger()
	logger.WarnWithTrace(ctx, "Duplicate entry",
		zap.String("error_type", "duplicate"),
		zap.String("entity", entity),
		zap.String("field", field),
		zap.Any("value", value),
	)
}

// LogBusinessRule logs business rule violations
func LogBusinessRule(ctx context.Context, rule, message string) {
	logger := logging.NewTraceLogger()
	logger.InfoWithTrace(ctx, "Business rule applied",
		zap.String("event_type", "business_rule"),
		zap.String("rule", rule),
		zap.String("message", message),
	)
}

// LogSlowOperation logs operations that exceed a threshold
func LogSlowOperation(ctx context.Context, operation string, duration time.Duration, threshold time.Duration) {
	if duration > threshold {
		logger := logging.NewTraceLogger()
		logger.WarnWithTrace(ctx, "Slow operation detected",
			zap.String("operation", operation),
			zap.Float64("duration_ms", float64(duration.Milliseconds())),
			zap.Float64("threshold_ms", float64(threshold.Milliseconds())),
		)
	}
}
