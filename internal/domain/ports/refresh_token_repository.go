package ports

import (
	"context"

	"example.com/go-yippi/internal/domain/entities"
)

// RefreshTokenRepository defines the interface for refresh token data access
type RefreshTokenRepository interface {
	Create(ctx context.Context, token *entities.RefreshToken) error
	FindByToken(ctx context.Context, token string) (*entities.RefreshToken, error)
	Revoke(ctx context.Context, token string) error
	RevokeAllForUser(ctx context.Context, userID int) error
	DeleteExpired(ctx context.Context) error
}
