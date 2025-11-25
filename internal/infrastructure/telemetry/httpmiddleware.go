package telemetry

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const (
	TracerName = "github.com/go-yippi/http-tracer"
)

// HTTPMiddleware provides enhanced HTTP request tracing with custom attributes
func HTTPMiddleware(serviceName string) fiber.Handler {
	tracer := otel.Tracer(TracerName)

	return func(c *fiber.Ctx) error {
		// Start span
		ctx, span := tracer.Start(
			c.Context(),
			"HTTP "+c.Method()+" "+c.Path(),
			trace.WithAttributes(
				attribute.String("http.method", c.Method()),
				attribute.String("http.url", c.OriginalURL()),
				attribute.String("http.scheme", c.Protocol()),
				attribute.String("http.host", c.Hostname()),
				attribute.String("http.user_agent", c.Get("User-Agent")),
				attribute.String("http.target", c.Path()),
				attribute.String("service.name", serviceName),
			),
		)

		// Add custom attributes
		if clientIP := c.IP(); clientIP != "" {
			span.SetAttributes(attribute.String("client.ip", clientIP))
		}

		// Extract trace context from headers if present
		headers := make(map[string][]string)
		c.Request().Header.VisitAll(func(key, value []byte) {
			headers[string(key)] = append(headers[string(key)], string(value))
		})

		propagator := otel.GetTextMapPropagator()
		extractor := propagation.HeaderCarrier(headers)
		ctx = propagator.Extract(ctx, extractor)

		// Set context for fiber
		c.SetUserContext(ctx)

		// Record request start time
		startTime := time.Now()

		// Process request
		err := c.Next()

		// Calculate duration
		duration := time.Since(startTime)

		// Add response attributes
		span.SetAttributes(
			attribute.Int("http.status_code", c.Response().StatusCode()),
			attribute.Int("http.response_content_length", len(c.Response().Body())),
			attribute.Float64("http.request.duration_ms", float64(duration.Nanoseconds())/1e6),
			attribute.String("http.route", c.Route().Path),
		)

		// Set span status based on HTTP status code
		statusCode := c.Response().StatusCode()
		if statusCode >= 400 {
			spanStatus := codes.Error
			if statusCode >= 500 {
				spanStatus = codes.Error
			}
			span.SetStatus(spanStatus, http.StatusText(statusCode))
		} else {
			span.SetStatus(codes.Ok, "")
		}

		// Add error details if present
		if err != nil {
			span.SetAttributes(
				attribute.String("error.message", err.Error()),
				attribute.String("error.type", "fiber.handler_error"),
			)
			span.SetStatus(codes.Error, err.Error())
		}

		// End span
		span.End()

		return err
	}
}

// AddCustomSpanAttribute adds custom attributes to the current span
func AddCustomSpanAttribute(ctx context.Context, key string, value interface{}) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		switch v := value.(type) {
		case string:
			span.SetAttributes(attribute.String(key, v))
		case int:
			span.SetAttributes(attribute.Int(key, v))
		case int64:
			span.SetAttributes(attribute.Int64(key, v))
		case float64:
			span.SetAttributes(attribute.Float64(key, v))
		case bool:
			span.SetAttributes(attribute.Bool(key, v))
		default:
			span.SetAttributes(attribute.String(key, fmt.Sprintf("%v", v)))
		}
	}
}

// AddErrorToSpan adds error information to the current span
func AddErrorToSpan(ctx context.Context, err error) {
	if err == nil {
		return
	}

	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.SetAttributes(
			attribute.String("error.message", err.Error()),
			attribute.String("error.type", "application_error"),
		)
		span.SetStatus(codes.Error, err.Error())
	}
}

// RecordMetrics records custom metrics for the current span
func RecordMetrics(ctx context.Context, metricName string, value float64, unit string) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.SetAttributes(
			attribute.Float64(metricName, value),
			attribute.String(metricName+".unit", unit),
		)
	}
}