package telemetry

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const (
	ServiceTracerName = "github.com/go-yippi/service-tracer"
)

// ServiceTracer provides tracing utilities for service layer operations
type ServiceTracer struct {
	tracer trace.Tracer
}

// NewServiceTracer creates a new service tracer
func NewServiceTracer() *ServiceTracer {
	return &ServiceTracer{
		tracer: otel.Tracer(ServiceTracerName),
	}
}

// TraceServiceMethod traces a service method execution
func (st *ServiceTracer) TraceServiceMethod(ctx context.Context, serviceName, methodName string, input interface{}, fn func(context.Context) error) error {
	spanName := serviceName + "." + methodName

	ctx, span := st.tracer.Start(ctx, spanName,
		trace.WithAttributes(
			attribute.String("service.name", serviceName),
			attribute.String("service.method", methodName),
			attribute.String("service.layer", "application"),
		))

	// Add input metadata if provided
	if input != nil {
		switch v := input.(type) {
		case map[string]interface{}:
			for key, value := range v {
				span.SetAttributes(attribute.String("input."+key, toString(value)))
			}
		case []interface{}:
			span.SetAttributes(attribute.Int("input.count", len(v)))
		default:
			span.SetAttributes(attribute.String("input.type", typeof(input)))
		}
	}

	startTime := time.Now()
	err := fn(ctx)
	duration := time.Since(startTime)

	span.SetAttributes(
		attribute.Float64("service.duration_ms", float64(duration.Nanoseconds())/1e6),
		attribute.String("service.operation", spanName),
	)

	if err != nil {
		span.SetAttributes(
			attribute.String("service.error", err.Error()),
			attribute.String("error.type", "service_error"),
		)
		span.SetStatus(codes.Error, err.Error())
	} else {
		span.SetStatus(codes.Ok, "")
	}

	span.End()
	return err
}

// TraceRepositoryOperation traces repository operations
func (st *ServiceTracer) TraceRepositoryOperation(ctx context.Context, operation, entity string, fn func(context.Context) error) error {
	spanName := "repository." + operation + "." + entity

	ctx, span := st.tracer.Start(ctx, spanName,
		trace.WithAttributes(
			attribute.String("repository.operation", operation),
			attribute.String("repository.entity", entity),
			attribute.String("repository.layer", "persistence"),
		))

	startTime := time.Now()
	err := fn(ctx)
	duration := time.Since(startTime)

	span.SetAttributes(
		attribute.Float64("repository.duration_ms", float64(duration.Nanoseconds())/1e6),
	)

	if err != nil {
		span.SetAttributes(
			attribute.String("repository.error", err.Error()),
			attribute.String("error.type", "repository_error"),
		)
		span.SetStatus(codes.Error, err.Error())
	} else {
		span.SetStatus(codes.Ok, "")
	}

	span.End()
	return err
}

// TraceBusinessOperation traces business logic operations
func (st *ServiceTracer) TraceBusinessOperation(ctx context.Context, operationName string, fn func(context.Context) error) error {
	ctx, span := st.tracer.Start(ctx, "business."+operationName,
		trace.WithAttributes(
			attribute.String("business.operation", operationName),
			attribute.String("business.layer", "domain"),
		))

	startTime := time.Now()
	err := fn(ctx)
	duration := time.Since(startTime)

	span.SetAttributes(
		attribute.Float64("business.duration_ms", float64(duration.Nanoseconds())/1e6),
	)

	if err != nil {
		span.SetAttributes(
			attribute.String("business.error", err.Error()),
			attribute.String("error.type", "business_error"),
		)
		span.SetStatus(codes.Error, err.Error())
	} else {
		span.SetStatus(codes.Ok, "")
	}

	span.End()
	return err
}

// AddBusinessAttribute adds business-related attributes to the current span
func (st *ServiceTracer) AddBusinessAttribute(ctx context.Context, key string, value interface{}) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.SetAttributes(
			attribute.String("business."+key, toString(value)),
		)
	}
}

// AddServiceMetric adds service metrics to the current span
func (st *ServiceTracer) AddServiceMetric(ctx context.Context, metricName string, value float64) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.SetAttributes(
			attribute.Float64("service."+metricName, value),
		)
	}
}

// WithUserContext adds user context to the span
func (st *ServiceTracer) WithUserContext(ctx context.Context, userID, userEmail, userRole string) context.Context {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.SetAttributes(
			attribute.String("user.id", userID),
			attribute.String("user.email", userEmail),
			attribute.String("user.role", userRole),
		)
	}
	return ctx
}

// WithRequestContext adds request context to the span
func (st *ServiceTracer) WithRequestContext(ctx context.Context, requestID, clientIP, userAgent string) context.Context {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.SetAttributes(
			attribute.String("request.id", requestID),
			attribute.String("client.ip", clientIP),
			attribute.String("user.agent", userAgent),
		)
	}
	return ctx
}

// GetServiceTracer returns a global service tracer instance
func GetServiceTracer() *ServiceTracer {
	return &ServiceTracer{
		tracer: otel.Tracer(ServiceTracerName),
	}
}

// Helper functions

func toString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case int, int32, int64, float32, float64:
		return fmt.Sprintf("%v", v)
	case bool:
		return fmt.Sprintf("%t", v)
	default:
		return fmt.Sprintf("%T", value)
	}
}

func typeof(value interface{}) string {
	return fmt.Sprintf("%T", value)
}

// ServiceOperation represents a traceable service operation
type ServiceOperation struct {
	tracer   *ServiceTracer
	service  string
	method   string
	ctx      context.Context
	span     trace.Span
	startTime time.Time
}

// NewServiceOperation creates a new service operation trace
func NewServiceOperation(ctx context.Context, serviceName, methodName string) *ServiceOperation {
	st := GetServiceTracer()
	spanName := serviceName + "." + methodName

	ctx, span := st.tracer.Start(ctx, spanName,
		trace.WithAttributes(
			attribute.String("service.name", serviceName),
			attribute.String("service.method", methodName),
			attribute.String("service.layer", "application"),
		))

	return &ServiceOperation{
		tracer:   st,
		service:  serviceName,
		method:   methodName,
		ctx:      ctx,
		span:     span,
		startTime: time.Now(),
	}
}

// AddAttribute adds an attribute to the operation
func (so *ServiceOperation) AddAttribute(key string, value interface{}) {
	if so.span.IsRecording() {
		so.span.SetAttributes(attribute.String("service."+key, toString(value)))
	}
}

// AddError adds error information to the operation
func (so *ServiceOperation) AddError(err error) {
	if err != nil && so.span.IsRecording() {
		so.span.SetAttributes(
			attribute.String("service.error", err.Error()),
			attribute.String("error.type", "service_error"),
		)
		so.span.SetStatus(codes.Error, err.Error())
	}
}

// Context returns the operation context
func (so *ServiceOperation) Context() context.Context {
	return so.ctx
}

// End completes the operation trace
func (so *ServiceOperation) End(err error) {
	duration := time.Since(so.startTime)

	if so.span.IsRecording() {
		so.span.SetAttributes(
			attribute.Float64("service.duration_ms", float64(duration.Nanoseconds())/1e6),
		)

		if err != nil {
			so.AddError(err)
		} else {
			so.span.SetStatus(codes.Ok, "")
		}

		so.span.End()
	}
}