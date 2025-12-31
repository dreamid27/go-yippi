package ports

import (
	"context"

	"example.com/go-yippi/internal/domain/entities"
)

// AuthService defines the interface for authentication operations
type AuthService interface {
	Register(ctx context.Context, input RegisterInput) (*entities.User, error)
	Login(ctx context.Context, email, password string) (*AuthResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*AuthResponse, error)
	Logout(ctx context.Context, refreshToken string) error
	ValidateToken(ctx context.Context, accessToken string) (*entities.User, error)
}

// RegisterInput represents the data needed to register a new user
type RegisterInput struct {
	Email    string
	Password string
	Name     string
	Phone    string
}

// AuthResponse represents the response after successful authentication
type AuthResponse struct {
	AccessToken  string
	RefreshToken string
	User         *entities.User
	ExpiresIn    int // seconds until access token expires
}
