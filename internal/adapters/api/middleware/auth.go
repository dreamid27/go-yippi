package middleware

import (
	"context"
	"strings"

	"example.com/go-yippi/internal/domain/ports"
	"github.com/gofiber/fiber/v2"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	// UserContextKey is the key used to store user in context
	UserContextKey contextKey = "user"
)

// RequireAuth validates the JWT token and requires authentication
func RequireAuth(authService ports.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Extract token from Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Missing authorization header",
			})
		}

		// Check if it's a Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid authorization header format",
			})
		}

		token := parts[1]
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Missing token",
			})
		}

		// Validate token and get user
		user, err := authService.ValidateToken(c.Context(), token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid or expired token",
			})
		}

		// Store user in Fiber's locals (Fiber-specific context)
		c.Locals(string(UserContextKey), user)

		// Also store in standard context for service layer usage
		ctx := context.WithValue(c.Context(), UserContextKey, user)
		c.SetUserContext(ctx)

		return c.Next()
	}
}

// RequireAdmin checks if the authenticated user has admin role
// Must be used after RequireAuth middleware
func RequireAdmin() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user from Fiber locals
		userVal := c.Locals(string(UserContextKey))
		if userVal == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authentication required",
			})
		}

		user, ok := userVal.(*struct {
			ID           int
			Email        string
			PasswordHash string
			Name         string
			Phone        string
			Role         string
			IsActive     bool
		})
		if !ok {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Invalid user context",
			})
		}

		// Check if user is admin
		if user.Role != "admin" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Admin access required",
			})
		}

		return c.Next()
	}
}

// OptionalAuth validates the JWT token if present, but doesn't require it
func OptionalAuth(authService ports.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Extract token from Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			// No token provided, continue without authentication
			return c.Next()
		}

		// Check if it's a Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			// Invalid format, but optional so continue
			return c.Next()
		}

		token := parts[1]
		if token == "" {
			// Empty token, continue
			return c.Next()
		}

		// Try to validate token and get user
		user, err := authService.ValidateToken(c.Context(), token)
		if err != nil {
			// Invalid token, but optional so continue
			return c.Next()
		}

		// Store user in Fiber's locals
		c.Locals(string(UserContextKey), user)

		// Also store in standard context
		ctx := context.WithValue(c.Context(), UserContextKey, user)
		c.SetUserContext(ctx)

		return c.Next()
	}
}
