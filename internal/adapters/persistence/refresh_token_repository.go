package persistence

import (
	"context"
	"time"

	"example.com/go-yippi/internal/adapters/persistence/db/ent"
	"example.com/go-yippi/internal/adapters/persistence/db/ent/refreshtoken"
	"example.com/go-yippi/internal/adapters/persistence/db/ent/user"
	"example.com/go-yippi/internal/domain/entities"
	domainErrors "example.com/go-yippi/internal/domain/errors"
	"example.com/go-yippi/internal/domain/ports"
)

// RefreshTokenRepository implements the ports.RefreshTokenRepository interface using Ent
type RefreshTokenRepository struct {
	client *ent.Client
}

// NewRefreshTokenRepository creates a new RefreshTokenRepository instance
func NewRefreshTokenRepository(client *ent.Client) ports.RefreshTokenRepository {
	return &RefreshTokenRepository{client: client}
}

// Create saves a new refresh token to the database
func (r *RefreshTokenRepository) Create(ctx context.Context, token *entities.RefreshToken) error {
	builder := r.client.RefreshToken.
		Create().
		SetToken(token.Token).
		SetExpiresAt(token.ExpiresAt).
		SetUserID(token.UserID)

	// Set revoked_at if present
	if token.RevokedAt != nil {
		builder.SetRevokedAt(*token.RevokedAt)
	}

	created, err := builder.Save(ctx)
	if err != nil {
		return err
	}

	// Populate the entity with generated values
	token.ID = created.ID
	token.CreatedAt = created.CreatedAt

	return nil
}

// FindByToken retrieves a refresh token by its token string
func (r *RefreshTokenRepository) FindByToken(ctx context.Context, tokenStr string) (*entities.RefreshToken, error) {
	found, err := r.client.RefreshToken.
		Query().
		Where(refreshtoken.TokenEQ(tokenStr)).
		WithUser().
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domainErrors.NewNotFoundError("RefreshToken", tokenStr)
		}
		return nil, err
	}

	// Get UserID from the loaded User edge
	userID := 0
	if found.Edges.User != nil {
		userID = found.Edges.User.ID
	}

	// Convert Ent entity to domain entity
	return &entities.RefreshToken{
		ID:        found.ID,
		UserID:    userID,
		Token:     found.Token,
		ExpiresAt: found.ExpiresAt,
		CreatedAt: found.CreatedAt,
		RevokedAt: found.RevokedAt,
	}, nil
}

// Revoke marks a refresh token as revoked by setting revoked_at to current time
func (r *RefreshTokenRepository) Revoke(ctx context.Context, tokenStr string) error {
	now := time.Now()

	result, err := r.client.RefreshToken.
		Update().
		Where(refreshtoken.TokenEQ(tokenStr)).
		SetRevokedAt(now).
		Save(ctx)

	if err != nil {
		return err
	}

	// If no rows were affected, the token doesn't exist
	if result == 0 {
		return domainErrors.NewNotFoundError("RefreshToken", tokenStr)
	}

	return nil
}

// RevokeAllForUser marks all refresh tokens for a user as revoked
func (r *RefreshTokenRepository) RevokeAllForUser(ctx context.Context, userID int) error {
	now := time.Now()

	_, err := r.client.RefreshToken.
		Update().
		Where(refreshtoken.HasUserWith(user.IDEQ(userID))).
		SetRevokedAt(now).
		Save(ctx)

	return err
}

// DeleteExpired removes all expired refresh tokens from the database
func (r *RefreshTokenRepository) DeleteExpired(ctx context.Context) error {
	now := time.Now()

	_, err := r.client.RefreshToken.
		Delete().
		Where(refreshtoken.ExpiresAtLT(now)).
		Exec(ctx)

	return err
}
