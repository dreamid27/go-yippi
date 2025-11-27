package telemetry

import (
	"context"
	"log"
	"runtime"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"google.golang.org/grpc/credentials"

	"example.com/go-yippi/internal/infrastructure/config"
)

var (
	// Global meter for application metrics
	meter metric.Meter

	// HTTP metrics
	httpRequestCounter    metric.Int64Counter
	httpRequestDuration   metric.Float64Histogram
	httpActiveRequests    metric.Int64UpDownCounter
	httpRequestSize       metric.Int64Histogram
	httpResponseSize      metric.Int64Histogram

	// Database metrics
	dbQueryCounter      metric.Int64Counter
	dbQueryDuration     metric.Float64Histogram
	dbActiveConnections metric.Int64UpDownCounter
	dbErrorCounter      metric.Int64Counter

	// Service metrics
	serviceOperationCounter  metric.Int64Counter
	serviceOperationDuration metric.Float64Histogram
	serviceErrorCounter      metric.Int64Counter

	// Business metrics
	businessEventCounter metric.Int64Counter

	// System metrics
	systemCPUUsage    metric.Float64ObservableGauge
	systemMemoryUsage metric.Int64ObservableGauge
	systemGoroutines  metric.Int64ObservableGauge
)

// InitMetrics initializes OpenTelemetry metrics with the given configuration
func InitMetrics(cfg *config.OpenTelemetryConfig) func(context.Context) error {
	if !cfg.Enabled {
		log.Println("OpenTelemetry Metrics is disabled")
		return func(ctx context.Context) error {
			return nil
		}
	}

	// Create the metrics exporter
	var secureOption otlpmetricgrpc.Option

	if !cfg.InsecureMode {
		secureOption = otlpmetricgrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, ""))
	} else {
		secureOption = otlpmetricgrpc.WithInsecure()
	}

	exporter, err := otlpmetricgrpc.New(
		context.Background(),
		secureOption,
		otlpmetricgrpc.WithEndpoint(cfg.CollectorURL),
	)

	if err != nil {
		log.Fatalf("Failed to create OpenTelemetry metrics exporter: %v", err)
	}

	// Create the resource
	resources, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			attribute.String("service.name", cfg.ServiceName),
			attribute.String("library.language", "go"),
		),
	)
	if err != nil {
		log.Fatalf("Could not set OpenTelemetry resources: %v", err)
	}

	// Create the meter provider
	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter,
			sdkmetric.WithInterval(10*time.Second), // Export metrics every 10 seconds
		)),
		sdkmetric.WithResource(resources),
	)

	otel.SetMeterProvider(meterProvider)

	// Create meter
	meter = otel.Meter("github.com/go-yippi/metrics")

	// Initialize HTTP metrics
	if err := initHTTPMetrics(); err != nil {
		log.Fatalf("Failed to initialize HTTP metrics: %v", err)
	}

	// Initialize database metrics
	if err := initDatabaseMetrics(); err != nil {
		log.Fatalf("Failed to initialize database metrics: %v", err)
	}

	// Initialize service metrics
	if err := initServiceMetrics(); err != nil {
		log.Fatalf("Failed to initialize service metrics: %v", err)
	}

	// Initialize business metrics
	if err := initBusinessMetrics(); err != nil {
		log.Fatalf("Failed to initialize business metrics: %v", err)
	}

	// Initialize system metrics
	if err := initSystemMetrics(); err != nil {
		log.Fatalf("Failed to initialize system metrics: %v", err)
	}

	log.Printf("OpenTelemetry Metrics initialized with service name: %s", cfg.ServiceName)

	return meterProvider.Shutdown
}

// initHTTPMetrics initializes HTTP-related metrics
func initHTTPMetrics() error {
	var err error

	httpRequestCounter, err = meter.Int64Counter(
		"http.server.request.count",
		metric.WithDescription("Total number of HTTP requests"),
		metric.WithUnit("{request}"),
	)
	if err != nil {
		return err
	}

	httpRequestDuration, err = meter.Float64Histogram(
		"http.server.request.duration",
		metric.WithDescription("HTTP request duration in milliseconds"),
		metric.WithUnit("ms"),
	)
	if err != nil {
		return err
	}

	httpActiveRequests, err = meter.Int64UpDownCounter(
		"http.server.active_requests",
		metric.WithDescription("Number of active HTTP requests"),
		metric.WithUnit("{request}"),
	)
	if err != nil {
		return err
	}

	httpRequestSize, err = meter.Int64Histogram(
		"http.server.request.size",
		metric.WithDescription("HTTP request body size in bytes"),
		metric.WithUnit("By"),
	)
	if err != nil {
		return err
	}

	httpResponseSize, err = meter.Int64Histogram(
		"http.server.response.size",
		metric.WithDescription("HTTP response body size in bytes"),
		metric.WithUnit("By"),
	)
	if err != nil {
		return err
	}

	return nil
}

// initDatabaseMetrics initializes database-related metrics
func initDatabaseMetrics() error {
	var err error

	dbQueryCounter, err = meter.Int64Counter(
		"db.client.query.count",
		metric.WithDescription("Total number of database queries"),
		metric.WithUnit("{query}"),
	)
	if err != nil {
		return err
	}

	dbQueryDuration, err = meter.Float64Histogram(
		"db.client.query.duration",
		metric.WithDescription("Database query duration in milliseconds"),
		metric.WithUnit("ms"),
	)
	if err != nil {
		return err
	}

	dbActiveConnections, err = meter.Int64UpDownCounter(
		"db.client.connections.active",
		metric.WithDescription("Number of active database connections"),
		metric.WithUnit("{connection}"),
	)
	if err != nil {
		return err
	}

	dbErrorCounter, err = meter.Int64Counter(
		"db.client.errors.count",
		metric.WithDescription("Total number of database errors"),
		metric.WithUnit("{error}"),
	)
	if err != nil {
		return err
	}

	return nil
}

// initServiceMetrics initializes service-related metrics
func initServiceMetrics() error {
	var err error

	serviceOperationCounter, err = meter.Int64Counter(
		"service.operation.count",
		metric.WithDescription("Total number of service operations"),
		metric.WithUnit("{operation}"),
	)
	if err != nil {
		return err
	}

	serviceOperationDuration, err = meter.Float64Histogram(
		"service.operation.duration",
		metric.WithDescription("Service operation duration in milliseconds"),
		metric.WithUnit("ms"),
	)
	if err != nil {
		return err
	}

	serviceErrorCounter, err = meter.Int64Counter(
		"service.errors.count",
		metric.WithDescription("Total number of service errors"),
		metric.WithUnit("{error}"),
	)
	if err != nil {
		return err
	}

	return nil
}

// initBusinessMetrics initializes business-related metrics
func initBusinessMetrics() error {
	var err error

	businessEventCounter, err = meter.Int64Counter(
		"business.event.count",
		metric.WithDescription("Total number of business events"),
		metric.WithUnit("{event}"),
	)
	if err != nil {
		return err
	}

	return nil
}

// initSystemMetrics initializes system-related metrics
func initSystemMetrics() error {
	var err error

	systemCPUUsage, err = meter.Float64ObservableGauge(
		"system.cpu.usage",
		metric.WithDescription("CPU usage percentage"),
		metric.WithUnit("%"),
		metric.WithFloat64Callback(func(ctx context.Context, o metric.Float64Observer) error {
			// This is a simplified CPU usage - in production use runtime/metrics
			o.Observe(getCPUUsage())
			return nil
		}),
	)
	if err != nil {
		return err
	}

	systemMemoryUsage, err = meter.Int64ObservableGauge(
		"system.memory.usage",
		metric.WithDescription("Memory usage in bytes"),
		metric.WithUnit("By"),
		metric.WithInt64Callback(func(ctx context.Context, o metric.Int64Observer) error {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			o.Observe(int64(m.Alloc))
			return nil
		}),
	)
	if err != nil {
		return err
	}

	systemGoroutines, err = meter.Int64ObservableGauge(
		"system.goroutines.count",
		metric.WithDescription("Number of goroutines"),
		metric.WithUnit("{goroutine}"),
		metric.WithInt64Callback(func(ctx context.Context, o metric.Int64Observer) error {
			o.Observe(int64(runtime.NumGoroutine()))
			return nil
		}),
	)
	if err != nil {
		return err
	}

	return nil
}

// RecordHTTPRequest records HTTP request metrics
func RecordHTTPRequest(ctx context.Context, method, route string, statusCode int, duration time.Duration, requestSize, responseSize int64) {
	if meter == nil {
		return
	}

	attrs := metric.WithAttributes(
		attribute.String("http.method", method),
		attribute.String("http.route", route),
		attribute.Int("http.status_code", statusCode),
	)

	httpRequestCounter.Add(ctx, 1, attrs)
	httpRequestDuration.Record(ctx, float64(duration.Milliseconds()), attrs)

	if requestSize > 0 {
		httpRequestSize.Record(ctx, requestSize, attrs)
	}
	if responseSize > 0 {
		httpResponseSize.Record(ctx, responseSize, attrs)
	}
}

// IncrementActiveHTTPRequests increments active HTTP request count
func IncrementActiveHTTPRequests(ctx context.Context) {
	if httpActiveRequests != nil {
		httpActiveRequests.Add(ctx, 1)
	}
}

// DecrementActiveHTTPRequests decrements active HTTP request count
func DecrementActiveHTTPRequests(ctx context.Context) {
	if httpActiveRequests != nil {
		httpActiveRequests.Add(ctx, -1)
	}
}

// RecordDBQuery records database query metrics
func RecordDBQuery(ctx context.Context, operation, entity string, duration time.Duration, err error) {
	if meter == nil {
		return
	}

	attrs := metric.WithAttributes(
		attribute.String("db.operation", operation),
		attribute.String("db.entity", entity),
		attribute.String("db.system", "postgresql"),
	)

	dbQueryCounter.Add(ctx, 1, attrs)
	dbQueryDuration.Record(ctx, float64(duration.Milliseconds()), attrs)

	if err != nil {
		dbErrorCounter.Add(ctx, 1, attrs)
	}
}

// RecordServiceOperation records service operation metrics
func RecordServiceOperation(ctx context.Context, service, method string, duration time.Duration, err error) {
	if meter == nil {
		return
	}

	attrs := metric.WithAttributes(
		attribute.String("service.name", service),
		attribute.String("service.method", method),
	)

	serviceOperationCounter.Add(ctx, 1, attrs)
	serviceOperationDuration.Record(ctx, float64(duration.Milliseconds()), attrs)

	if err != nil {
		serviceErrorCounter.Add(ctx, 1, attrs)
	}
}

// RecordBusinessEvent records business event metrics
func RecordBusinessEvent(ctx context.Context, eventType, eventName string) {
	if meter == nil {
		return
	}

	attrs := metric.WithAttributes(
		attribute.String("business.event.type", eventType),
		attribute.String("business.event.name", eventName),
	)

	businessEventCounter.Add(ctx, 1, attrs)
}

// getCPUUsage returns a simplified CPU usage metric
// In production, use runtime/metrics or a proper CPU monitoring library
func getCPUUsage() float64 {
	// This is a placeholder - implement proper CPU monitoring
	// You can use github.com/shirou/gopsutil/cpu for accurate CPU metrics
	return 0.0
}

// GetMeter returns the global meter instance
func GetMeter() metric.Meter {
	return meter
}
