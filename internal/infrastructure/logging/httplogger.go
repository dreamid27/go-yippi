package logging

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// HTTPLoggerMiddleware provides structured HTTP request logging
func HTTPLoggerMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Skip logging for health check endpoints
		if c.Path() == "/health" || c.Path() == "/metrics" {
			return c.Next()
		}

		// Record start time
		start := time.Now()

		// Process request
		err := c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Extract request information
		method := c.Method()
		path := c.Path()
		userAgent := c.Get("User-Agent")
		clientIP := c.IP()
		statusCode := c.Response().StatusCode()
		requestID := c.GetRespHeader("X-Request-ID")
		queryString := string(c.Request().URI().QueryString())

		// Log the request
		fields := []zap.Field{
			zap.String("event", "http_request"),
			zap.String("method", method),
			zap.String("path", path),
			zap.String("query", queryString),
			zap.String("user_agent", userAgent),
			zap.String("client_ip", clientIP),
			zap.Int("status_code", statusCode),
			zap.Float64("duration_ms", float64(duration.Nanoseconds())/1e6),
			zap.Int("response_size", len(c.Response().Body())),
		}

		// Add request ID if present
		if requestID != "" {
			fields = append(fields, zap.String("request_id", requestID))
		}

		// Add content length from request
		if len(c.Request().Body()) > 0 {
			fields = append(fields, zap.Int("request_size", len(c.Request().Body())))
		}

		// Add referer if present
		if referer := c.Get("Referer"); referer != "" {
			fields = append(fields, zap.String("referer", referer))
		}

		// Add custom headers if present
		if apiVersion := c.Get("X-API-Version"); apiVersion != "" {
			fields = append(fields, zap.String("api_version", apiVersion))
		}

		// Categorize by HTTP status
		var logLevel string
		switch {
		case statusCode >= 500:
			logLevel = "error"
		case statusCode >= 400:
			logLevel = "warn"
		case statusCode >= 300:
			logLevel = "info"
		default:
			logLevel = "info"
		}

		// Add error details if present
		if err != nil {
			fields = append(fields, zap.Error(err))
		}

		// Log based on level
		switch logLevel {
		case "error":
			if GetGlobalLogger() != nil {
				GetGlobalLogger().Error("HTTP request completed", fields...)
			}
		case "warn":
			if GetGlobalLogger() != nil {
				GetGlobalLogger().Warn("HTTP request completed", fields...)
			}
		default:
			if GetGlobalLogger() != nil {
				GetGlobalLogger().Info("HTTP request completed", fields...)
			}
		}

		return err
	}
}

// RequestIDMiddleware adds unique request ID to each request
func RequestIDMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Generate request ID if not present
		requestID := c.Get("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}

		// Set request ID in response headers
		c.Set("X-Request-ID", requestID)
		c.Response().Header.Set("X-Request-ID", requestID)

		return c.Next()
	}
}

// SecurityLoggerMiddleware logs security-related events
func SecurityLoggerMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Check for suspicious patterns
		clientIP := c.IP()
		userAgent := c.Get("User-Agent")
		path := c.Path()

		// Log suspicious user agents
		if isSuspiciousUserAgent(userAgent) {
			LogSecurityEvent(c.Context(), "suspicious_user_agent", "medium", map[string]interface{}{
				"client_ip":  clientIP,
				"user_agent": userAgent,
				"path":       path,
			})
		}

		// Log SQL injection attempts
		if isSQLInjectionAttempt(string(c.Request().Body())) {
			LogSecurityEvent(c.Context(), "sql_injection_attempt", "high", map[string]interface{}{
				"client_ip": clientIP,
				"user_agent": userAgent,
				"path":       path,
				"body":        string(c.Request().Body()),
			})
		}

		// Log path traversal attempts
		if isPathTraversalAttempt(path) {
			LogSecurityEvent(c.Context(), "path_traversal_attempt", "high", map[string]interface{}{
				"client_ip": clientIP,
				"user_agent": userAgent,
				"path":       path,
			})
		}

		return c.Next()
	}
}

// AuditLoggerMiddleware logs sensitive operations
func AuditLoggerMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		err := c.Next()

		// Log sensitive operations
		method := c.Method()
		path := c.Path()

		// Define sensitive operations
		sensitiveOperations := map[string][]string{
			"POST": {
				"/users",
				"/auth/login",
				"/auth/register",
				"/products",
				"/categories",
			},
			"PUT": {
				"/users/",
				"/products/",
				"/categories/",
			},
			"DELETE": {
				"/users/",
				"/products/",
				"/categories/",
			},
		}

		// Check if current operation is sensitive
		if paths, ok := sensitiveOperations[method]; ok {
			for _, sensitivePath := range paths {
				if path == sensitivePath || (len(sensitivePath) > 1 && path == sensitivePath[:len(sensitivePath)-1]) {
					LogAuditEvent(c.Context(), method, path, c.Response().StatusCode(), err)
					break
				}
			}
		}

		return err
	}
}

// LogAuditEvent logs audit events with detailed information
func LogAuditEvent(ctx context.Context, method, path string, statusCode int, err error) {
	fields := []zap.Field{
		zap.String("event", "audit_event"),
		zap.String("method", method),
		zap.String("path", path),
		zap.Int("status_code", statusCode),
		zap.String("timestamp", time.Now().UTC().Format(time.RFC3339)),
	}

	// Add client information
	if clientIP := getClientIP(ctx); clientIP != "" {
		fields = append(fields, zap.String("client_ip", clientIP))
	}

	// Add user information if available
	if userID := getUserID(ctx); userID != "" {
		fields = append(fields, zap.String("user_id", userID))
	}

	// Add error if present
	if err != nil {
		fields = append(fields, zap.Error(err))
	}

	// Log audit event
	if GetGlobalLogger() != nil {
		GetGlobalLogger().Info("Audit event", fields...)
	}
}

// Helper functions

func generateRequestID() string {
	// Generate a simple request ID (in practice, use UUID library)
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}

func isSuspiciousUserAgent(userAgent string) bool {
	suspiciousPatterns := []string{
		"sqlmap", "nikto", "dirb", "nmap", "masscan",
		"nessus", "burp", "owasp", "zap", "acunetix",
		"python-requests", "curl", "wget", "scrapy", "bot",
	}

	for _, pattern := range suspiciousPatterns {
		if userAgent != "" && containsIgnoreCase(userAgent, pattern) {
			return true
		}
	}
	return false
}

func isSQLInjectionAttempt(body string) bool {
	sqlPatterns := []string{
		"' OR", "' OR '", "--", "/*", "*/", "xp_", "sp_",
		"UNION", "SELECT", "INSERT", "UPDATE", "DELETE", "DROP",
		"CREATE", "ALTER", "EXEC", "EXECUTE", "WAITFOR", "DELAY",
	}

	for _, pattern := range sqlPatterns {
		if containsIgnoreCase(body, pattern) {
			return true
		}
	}
	return false
}

func isPathTraversalAttempt(path string) bool {
	traversalPatterns := []string{
		"..", "%2e%2e%2f", "%2e%2e", "..%2f", "..%5c",
		"%2f..%2f", "%2f%2e%2e%2f", "../", "..\\",
	}

	for _, pattern := range traversalPatterns {
		if containsIgnoreCase(path, pattern) {
			return true
		}
	}
	return false
}

func containsIgnoreCase(str, substr string) bool {
	return len(str) >= len(substr) &&
		(str[:len(substr)] == substr || str[len(str)-len(substr):] == substr)
}

func getClientIP(ctx context.Context) string {
	// In a real implementation, extract client IP from context
	// This is a placeholder
	return ""
}

func getUserID(ctx context.Context) string {
	// In a real implementation, extract user ID from context
	// This is a placeholder
	return ""
}