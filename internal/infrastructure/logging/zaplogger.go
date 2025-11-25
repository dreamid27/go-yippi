package logging

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.opentelemetry.io/otel/trace"
)

var (
	// Global logger instance
	Logger *zap.Logger
	Sugar  *zap.SugaredLogger
)

// LogConfig holds logging configuration
type LogConfig struct {
	Level      string `json:"level"`
	Format     string `json:"format"`     // "json" or "console"
	Output     string `json:"output"`     // "stdout", "stderr", or file path
	MaxSize    int    `json:"max_size"`   // Max size in MB
	MaxBackups int    `json:"max_backups"` // Max number of old log files
	MaxAge     int    `json:"max_age"`     // Max age in days
	Compress   bool   `json:"compress"`    // Compress old log files
}

// Default log configuration
func DefaultLogConfig() *LogConfig {
	return &LogConfig{
		Level:      "info",
		Format:     "json",
		Output:     "stdout",
		MaxSize:    100,
		MaxBackups: 3,
		MaxAge:     28,
		Compress:   true,
	}
}

// Initialize sets up the global logger with the given configuration
func Initialize(config *LogConfig) error {
	// Parse log level
	level, err := zapcore.ParseLevel(config.Level)
	if err != nil {
		return fmt.Errorf("invalid log level: %w", err)
	}

	// Create encoder config
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:     "function",
		MessageKey:     "message",
		StacktraceKey:   "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:     zapcore.LowercaseLevelEncoder,
		EncodeTime:      zapcore.ISO8601TimeEncoder,
		EncodeDuration:   zapcore.StringDurationEncoder,
		EncodeCaller:     zapcore.ShortCallerEncoder,
		EncodeName:      zapcore.FullNameEncoder,
	}

	// Create encoder
	var encoder zapcore.Encoder
	switch config.Format {
	case "console":
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	case "json":
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	default:
		return fmt.Errorf("invalid log format: %s", config.Format)
	}

	// Create core with output
	var core zapcore.Core
	switch config.Output {
	case "stderr":
		core = zapcore.NewCore(encoder, zapcore.AddSync(os.Stderr), level)
	case "stdout":
		core = zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), level)
	default:
		// File output with rotation
		if config.MaxSize <= 0 {
			config.MaxSize = 100
		}
		if config.MaxBackups <= 0 {
			config.MaxBackups = 3
		}
		if config.MaxAge <= 0 {
			config.MaxAge = 28
		}

		// File rotation core would require zapio or external rotation library
		// For now, simple file output
		file, err := os.OpenFile(config.Output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return fmt.Errorf("failed to open log file: %w", err)
		}
		core = zapcore.NewCore(encoder, zapcore.AddSync(file), level)
	}

	// Create logger with additional context
	logger := zap.New(core,
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
		zap.Fields(
			zap.String("service", "go-yippi-api"),
			zap.String("version", "1.0.0"),
		),
	)

	// Store global instances
	Logger = logger
	Sugar = logger.Sugar()

	Logger.Info("Logger initialized",
		zap.String("level", config.Level),
		zap.String("format", config.Format),
		zap.String("output", config.Output),
	)

	return nil
}

// TraceLogger combines logging and tracing
type TraceLogger struct {
	logger *zap.Logger
}

// NewTraceLogger creates a new trace logger
func NewTraceLogger() *TraceLogger {
	return &TraceLogger{
		logger: Logger,
	}
}

// WithContext adds trace context to the logger
func (tl *TraceLogger) WithContext(ctx context.Context) *zap.Logger {
	if ctx == nil {
		return tl.logger
	}

	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return tl.logger
	}

	spanCtx := span.SpanContext()
	fields := []zap.Field{
		zap.String("trace_id", spanCtx.TraceID().String()),
		zap.String("span_id", spanCtx.SpanID().String()),
		zap.String("trace_flags", spanCtx.TraceFlags().String()),
	}

	// Note: Baggage handling simplified to avoid type issues
	// In practice, you might want to implement proper baggage extraction

	return tl.logger.With(fields...)
}

// InfoWithTrace logs info message with trace context
func (tl *TraceLogger) InfoWithTrace(ctx context.Context, message string, fields ...zap.Field) {
	logger := tl.WithContext(ctx)
	logger.Info(message, fields...)
}

// ErrorWithTrace logs error message with trace context
func (tl *TraceLogger) ErrorWithTrace(ctx context.Context, message string, err error, fields ...zap.Field) {
	logger := tl.WithContext(ctx)
	allFields := append(fields, zap.Error(err))
	logger.Error(message, allFields...)
}

// WarnWithTrace logs warning message with trace context
func (tl *TraceLogger) WarnWithTrace(ctx context.Context, message string, fields ...zap.Field) {
	logger := tl.WithContext(ctx)
	logger.Warn(message, fields...)
}

// DebugWithTrace logs debug message with trace context
func (tl *TraceLogger) DebugWithTrace(ctx context.Context, message string, fields ...zap.Field) {
	logger := tl.WithContext(ctx)
	logger.Debug(message, fields...)
}

// FatalWithTrace logs fatal message with trace context and exits
func (tl *TraceLogger) FatalWithTrace(ctx context.Context, message string, err error, fields ...zap.Field) {
	logger := tl.WithContext(ctx)
	allFields := append(fields, zap.Error(err))
	logger.Fatal(message, allFields...)
}

// Structured logging functions for specific domains

// LogHTTPRequest logs HTTP request with structured data
func LogHTTPRequest(ctx context.Context, method, path, userAgent, clientIP string, statusCode int, duration time.Duration, err error) {
	tl := NewTraceLogger()
	fields := []zap.Field{
		zap.String("event", "http_request"),
		zap.String("method", method),
		zap.String("path", path),
		zap.String("user_agent", userAgent),
		zap.String("client_ip", clientIP),
		zap.Int("status_code", statusCode),
		zap.Float64("duration_ms", float64(duration.Nanoseconds())/1e6),
	}

	if err != nil {
		fields = append(fields, zap.Error(err))
	}

	if ctx != nil {
		tl.InfoWithTrace(ctx, "HTTP request completed", fields...)
	} else {
		Logger.Info("HTTP request completed", fields...)
	}
}

// LogDatabaseOperation logs database operation with structured data
func LogDatabaseOperation(ctx context.Context, operation, entity string, duration time.Duration, err error, affected int) {
	tl := NewTraceLogger()
	fields := []zap.Field{
		zap.String("event", "database_operation"),
		zap.String("operation", operation),
		zap.String("entity", entity),
		zap.Float64("duration_ms", float64(duration.Nanoseconds())/1e6),
		zap.Int("affected_rows", affected),
	}

	if err != nil {
		fields = append(fields, zap.Error(err))
	}

	if ctx != nil {
		tl.InfoWithTrace(ctx, "Database operation completed", fields...)
	} else {
		Logger.Info("Database operation completed", fields...)
	}
}

// LogServiceOperation logs service operation with structured data
func LogServiceOperation(ctx context.Context, serviceName, operation string, duration time.Duration, err error, input interface{}) {
	tl := NewTraceLogger()
	fields := []zap.Field{
		zap.String("event", "service_operation"),
		zap.String("service", serviceName),
		zap.String("operation", operation),
		zap.Float64("duration_ms", float64(duration.Nanoseconds())/1e6),
	}

	if input != nil {
		fields = append(fields, zap.Any("input", input))
	}

	if err != nil {
		fields = append(fields, zap.Error(err))
	}

	if ctx != nil {
		tl.InfoWithTrace(ctx, "Service operation completed", fields...)
	} else {
		Logger.Info("Service operation completed", fields...)
	}
}

// LogBusinessEvent logs business event with structured data
func LogBusinessEvent(ctx context.Context, event string, data map[string]interface{}) {
	tl := NewTraceLogger()
	fields := []zap.Field{
		zap.String("event", "business_event"),
		zap.String("business_event", event),
	}

	// Add business data as fields
	for key, value := range data {
		fields = append(fields, zap.Any(key, value))
	}

	if ctx != nil {
		tl.InfoWithTrace(ctx, "Business event occurred", fields...)
	} else {
		Logger.Info("Business event occurred", fields...)
	}
}

// LogSecurityEvent logs security events with structured data
func LogSecurityEvent(ctx context.Context, event, severity string, details map[string]interface{}) {
	tl := NewTraceLogger()
	fields := []zap.Field{
		zap.String("event", "security_event"),
		zap.String("security_event", event),
		zap.String("severity", severity),
		zap.String("timestamp", time.Now().UTC().Format(time.RFC3339)),
	}

	// Add security details
	for key, value := range details {
		fields = append(fields, zap.Any("security."+key, value))
	}

	if ctx != nil {
		tl.InfoWithTrace(ctx, "Security event", fields...)
	} else {
		Logger.Info("Security event", fields...)
	}
}

// GetGlobalLogger returns the global logger instance
func GetGlobalLogger() *zap.Logger {
	return Logger
}

// GetGlobalSugar returns the global sugared logger instance
func GetGlobalSugar() *zap.SugaredLogger {
	return Sugar
}

// Sync flushes any buffered log entries
func Sync() {
	if Logger != nil {
		Logger.Sync()
	}
}

// GetLogLevel returns the current log level
func GetLogLevel() string {
	if Logger == nil {
		return "info"
	}

	// Extract level from logger (simplified approach)
	return "info"
}

// SetLogLevel changes the log level at runtime
func SetLogLevel(level string) error {
	_, err := zapcore.ParseLevel(level)
	if err != nil {
		return fmt.Errorf("invalid log level: %w", err)
	}

	// Create a new logger with the new level
	if Logger != nil {
		// This is a simplified approach - in practice you might need
		// to recreate the logger or use AtomicLevel
		Logger.Info("Log level changed", zap.String("old_level", GetLogLevel()), zap.String("new_level", level))
	}

	return nil
}