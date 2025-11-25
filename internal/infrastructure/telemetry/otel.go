package telemetry

import (
	"context"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/credentials"

	"example.com/go-yippi/internal/infrastructure/config"
)

// InitTracer initializes OpenTelemetry with the given configuration
// It returns a cleanup function that should be called when the application shuts down
func InitTracer(cfg *config.OpenTelemetryConfig) func(context.Context) error {
	if !cfg.Enabled {
		log.Println("OpenTelemetry is disabled")
		return func(ctx context.Context) error {
			return nil
		}
	}

	// Create the exporter
	var secureOption otlptracegrpc.Option

	if !cfg.InsecureMode {
		secureOption = otlptracegrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, ""))
	} else {
		secureOption = otlptracegrpc.WithInsecure()
	}

	exporter, err := otlptrace.New(
		context.Background(),
		otlptracegrpc.NewClient(
			secureOption,
			otlptracegrpc.WithEndpoint(cfg.CollectorURL),
		),
	)

	if err != nil {
		log.Fatalf("Failed to create OpenTelemetry exporter: %v", err)
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

	// Create the tracer provider
	otel.SetTracerProvider(
		sdktrace.NewTracerProvider(
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
			sdktrace.WithBatcher(exporter),
			sdktrace.WithResource(resources),
		),
	)

	log.Printf("OpenTelemetry initialized with service name: %s, collector: %s", cfg.ServiceName, cfg.CollectorURL)

	return exporter.Shutdown
}

// GetTracer returns a tracer for the current application
func GetTracer(name string) trace.Tracer {
	return otel.Tracer(name)
}

// IsEnabled checks if OpenTelemetry is enabled
func IsEnabled(cfg *config.OpenTelemetryConfig) bool {
	return cfg.Enabled
}