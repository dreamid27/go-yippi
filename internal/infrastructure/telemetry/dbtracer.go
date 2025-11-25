package telemetry

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"example.com/go-yippi/internal/adapters/persistence/db/ent"
)

const (
	DBTracerName = "github.com/go-yippi/db-tracer"
)

// TracedEntClient wraps Ent client with OpenTelemetry tracing
type TracedEntClient struct {
	*ent.Client
	tracer trace.Tracer
}

// NewTracedEntClient creates a new traced Ent client
func NewTracedEntClient(client *ent.Client) *TracedEntClient {
	return &TracedEntClient{
		Client: client,
		tracer: otel.Tracer(DBTracerName),
	}
}

// Intercept queries for tracing
func (t *TracedEntClient) Intercept() *TracedEntClient {
	// Hook into the Ent client to add tracing
	hooks := []ent.Hook{
		t.createHook(),
		t.updateHook(),
		t.deleteHook(),
	}

	t.Client.Use(hooks...)
	return t
}


// createHook creates a hook for create operations
func (t *TracedEntClient) createHook() ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			return t.traceQuery(ctx, m.Type(), "create", func() (ent.Value, error) {
				return next.Mutate(ctx, m)
			})
		})
	}
}

// updateHook creates a hook for update operations
func (t *TracedEntClient) updateHook() ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			return t.traceQuery(ctx, m.Type(), "update", func() (ent.Value, error) {
				return next.Mutate(ctx, m)
			})
		})
	}
}

// deleteHook creates a hook for delete operations
func (t *TracedEntClient) deleteHook() ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			return t.traceQuery(ctx, m.Type(), "delete", func() (ent.Value, error) {
				return next.Mutate(ctx, m)
			})
		})
	}
}

// traceQuery traces a database operation
func (t *TracedEntClient) traceQuery(ctx context.Context, entityType, operation string, fn func() (ent.Value, error)) (ent.Value, error) {
	operationName := fmt.Sprintf("db.%s.%s", operation, entityType)

	ctx, span := t.tracer.Start(ctx, operationName,
		trace.WithAttributes(
			attribute.String("db.system", "postgresql"),
			attribute.String("db.operation", operation),
			attribute.String("db.entity", entityType),
			attribute.String("db.statement_type", operation),
		))

	startTime := time.Now()
	result, err := fn()
	duration := time.Since(startTime)

	span.SetAttributes(
		attribute.Float64("db.duration_ms", float64(duration.Nanoseconds())/1e6),
		attribute.String("db.operation.name", operationName),
	)

	if err != nil {
		span.SetAttributes(
			attribute.String("db.error", err.Error()),
			attribute.String("error.type", "database_error"),
		)
		span.SetStatus(codes.Error, err.Error())
	} else {
		span.SetStatus(codes.Ok, "")
	}

	// Add result count for query operations
	if operation == "query" && result != nil {
		if list, ok := result.([]interface{}); ok {
			span.SetAttributes(attribute.Int("db.rows_count", len(list)))
		}
	}

	span.End()
	return result, err
}

// TraceRawQuery traces a raw SQL query
func (t *TracedEntClient) TraceRawQuery(ctx context.Context, query string, args ...interface{}) error {
	ctx, span := t.tracer.Start(ctx, "db.raw_query",
		trace.WithAttributes(
			attribute.String("db.system", "postgresql"),
			attribute.String("db.operation", "raw_query"),
			attribute.String("db.statement", query),
			attribute.Int("db.parameters_count", len(args)),
		))

	startTime := time.Now()
	// In a real implementation, you would execute the actual query here
	// For now, we'll just simulate the operation
	duration := time.Since(startTime)

	span.SetAttributes(
		attribute.Float64("db.duration_ms", float64(duration.Nanoseconds())/1e6),
		attribute.String("db.query_type", "raw"),
	)

	span.SetStatus(codes.Ok, "")
	span.End()
	return nil
}

// TraceTransaction traces a database transaction
func (t *TracedEntClient) TraceTransaction(ctx context.Context, fn func(context.Context, *ent.Tx) error) error {
	ctx, span := t.tracer.Start(ctx, "db.transaction",
		trace.WithAttributes(
			attribute.String("db.system", "postgresql"),
			attribute.String("db.operation", "transaction"),
		))

	startTime := time.Now()
	tx, err := t.Client.Tx(ctx)
	if err != nil {
		span.SetAttributes(
			attribute.String("db.error", err.Error()),
			attribute.String("transaction.status", "failed_to_begin"),
		)
		span.SetStatus(codes.Error, err.Error())
		span.End()
		return err
	}

	err = fn(ctx, tx)
	duration := time.Since(startTime)

	if err != nil {
		_ = tx.Rollback()
		span.SetAttributes(
			attribute.String("db.error", err.Error()),
			attribute.String("transaction.status", "rollback"),
			attribute.Float64("db.duration_ms", float64(duration.Nanoseconds())/1e6),
		)
		span.SetStatus(codes.Error, err.Error())
	} else {
		err = tx.Commit()
		if err != nil {
			span.SetAttributes(
				attribute.String("db.error", err.Error()),
				attribute.String("transaction.status", "commit_failed"),
			)
			span.SetStatus(codes.Error, err.Error())
		} else {
			span.SetAttributes(
				attribute.String("transaction.status", "commit"),
				attribute.Float64("db.duration_ms", float64(duration.Nanoseconds())/1e6),
			)
			span.SetStatus(codes.Ok, "")
		}
	}

	span.End()
	return err
}

// ExtractQueryInfo extracts information from SQL query for better tracing
func ExtractQueryInfo(query string) (operation string, table string) {
	query = strings.ToLower(strings.TrimSpace(query))

	// Extract operation type
	switch {
	case strings.HasPrefix(query, "select"):
		operation = "SELECT"
	case strings.HasPrefix(query, "insert"):
		operation = "INSERT"
	case strings.HasPrefix(query, "update"):
		operation = "UPDATE"
	case strings.HasPrefix(query, "delete"):
		operation = "DELETE"
	case strings.HasPrefix(query, "create"):
		operation = "CREATE"
	case strings.HasPrefix(query, "alter"):
		operation = "ALTER"
	case strings.HasPrefix(query, "drop"):
		operation = "DROP"
	default:
		operation = "UNKNOWN"
	}

	// Extract table name (simplified)
	if operation == "SELECT" {
		if idx := strings.Index(query, "from"); idx != -1 {
			tablePart := strings.TrimSpace(query[idx+4:])
			if spaceIdx := strings.Index(tablePart, " "); spaceIdx != -1 {
				table = tablePart[:spaceIdx]
			} else {
				table = tablePart
			}
		}
	} else if operation == "INSERT" {
		if idx := strings.Index(query, "into"); idx != -1 {
			tablePart := strings.TrimSpace(query[idx+4:])
			if spaceIdx := strings.Index(tablePart, " "); spaceIdx != -1 {
				table = tablePart[:spaceIdx]
			} else {
				table = tablePart
			}
		}
	} else if operation == "UPDATE" || operation == "DELETE" {
		if idx := strings.Index(query, operation+" "); idx != -1 {
			tablePart := strings.TrimSpace(query[idx+len(operation)+1:])
			if spaceIdx := strings.Index(tablePart, " "); spaceIdx != -1 {
				table = tablePart[:spaceIdx]
			} else {
				table = tablePart
			}
		}
	}

	return operation, table
}

// DBStats holds database performance statistics
type DBStats struct {
	QueryCount    int64
	TotalDuration time.Duration
	ErrorCount    int64
	AvgDuration   time.Duration
}

// CollectDBStats collects database statistics for monitoring
func (t *TracedEntClient) CollectDBStats(ctx context.Context) DBStats {
	// This is a placeholder for actual stats collection
	// In a real implementation, you would aggregate metrics from spans
	return DBStats{
		QueryCount:    0,
		TotalDuration: 0,
		ErrorCount:    0,
		AvgDuration:   0,
	}
}