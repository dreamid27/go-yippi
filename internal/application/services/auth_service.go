package services

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"example.com/go-yippi/internal/domain/entities"
	domainErrors "example.com/go-yippi/internal/domain/errors"
	"example.com/go-yippi/internal/domain/ports"
)

// AuthServiceImpl implements the AuthService interface
type AuthServiceImpl struct {
	userRepo         ports.UserRepository
	refreshTokenRepo ports.RefreshTokenRepository
	jwtSecret        string
	accessTokenTTL   time.Duration
	refreshTokenTTL  time.Duration
}

// NewAuthService creates a new instance of AuthService
func NewAuthService(
	userRepo ports.UserRepository,
	refreshTokenRepo ports.RefreshTokenRepository,
	jwtSecret string,
) ports.AuthService {
	return &AuthServiceImpl{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		jwtSecret:        jwtSecret,
		accessTokenTTL:   15 * time.Minute,
		refreshTokenTTL:  30 * 24 * time.Hour, // 30 days
	}
}

// Register registers a new user
func (s *AuthServiceImpl) Register(ctx context.Context, input ports.RegisterInput) (*entities.User, error) {
	// Check if email already exists
	existingUser, err := s.userRepo.FindByEmail(ctx, input.Email)
	if err == nil && existingUser != nil {
		return nil, domainErrors.NewDuplicateError("user", "email", input.Email)
	}

	// Hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user entity
	user := &entities.User{
		Email:        input.Email,
		PasswordHash: string(passwordHash),
		Name:         input.Name,
		Phone:        input.Phone,
		Role:         entities.UserRoleCustomer,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Save user to repository
	err = s.userRepo.Create(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// Login authenticates a user and returns access and refresh tokens
func (s *AuthServiceImpl) Login(ctx context.Context, email, password string) (*ports.AuthResponse, error) {
	// Find user by email
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, domainErrors.NewValidationError("credentials", "invalid email or password")
	}

	// Check if user is active
	if !user.IsActive {
		return nil, domainErrors.NewValidationError("account", "account is not active")
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, domainErrors.NewValidationError("credentials", "invalid email or password")
	}

	// Generate access token
	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token
	refreshToken, err := s.generateRefreshToken(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &ports.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user,
		ExpiresIn:    int(s.accessTokenTTL.Seconds()),
	}, nil
}

// RefreshToken generates a new access token using a valid refresh token
func (s *AuthServiceImpl) RefreshToken(ctx context.Context, tokenString string) (*ports.AuthResponse, error) {
	// Find refresh token
	refreshToken, err := s.refreshTokenRepo.FindByToken(ctx, tokenString)
	if err != nil {
		return nil, domainErrors.NewValidationError("refresh_token", "invalid refresh token")
	}

	// Check if token is valid (not revoked and not expired)
	if !refreshToken.IsValid() {
		return nil, domainErrors.NewValidationError("refresh_token", "refresh token is invalid or expired")
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, refreshToken.UserID)
	if err != nil {
		return nil, domainErrors.NewValidationError("user", "user not found")
	}

	// Check if user is active
	if !user.IsActive {
		return nil, domainErrors.NewValidationError("account", "account is not active")
	}

	// Generate new access token
	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	return &ports.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: tokenString, // Return the same refresh token
		User:         user,
		ExpiresIn:    int(s.accessTokenTTL.Seconds()),
	}, nil
}

// Logout revokes a refresh token
func (s *AuthServiceImpl) Logout(ctx context.Context, refreshToken string) error {
	// Find refresh token to validate it exists
	_, err := s.refreshTokenRepo.FindByToken(ctx, refreshToken)
	if err != nil {
		return domainErrors.NewValidationError("refresh_token", "invalid refresh token")
	}

	// Revoke the token
	err = s.refreshTokenRepo.Revoke(ctx, refreshToken)
	if err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}

	return nil
}

// ValidateToken validates an access token and returns the associated user
func (s *AuthServiceImpl) ValidateToken(ctx context.Context, tokenString string) (*entities.User, error) {
	// Parse and validate JWT token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, domainErrors.NewValidationError("access_token", "invalid or expired token")
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, domainErrors.NewValidationError("access_token", "invalid token claims")
	}

	// Extract user ID from claims
	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return nil, domainErrors.NewValidationError("access_token", "invalid user_id in token")
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, domainErrors.NewValidationError("access_token", "invalid user_id format in token")
	}

	// Get user from repository
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, domainErrors.NewValidationError("user", "user not found")
	}

	// Check if user is active
	if !user.IsActive {
		return nil, domainErrors.NewValidationError("account", "account is not active")
	}

	return user, nil
}

// generateAccessToken generates a JWT access token for a user
func (s *AuthServiceImpl) generateAccessToken(user *entities.User) (string, error) {
	now := time.Now()
	expiresAt := now.Add(s.accessTokenTTL)

	// Create JWT claims
	claims := jwt.MapClaims{
		"user_id": user.ID.String(),
		"email":   user.Email,
		"role":    user.Role,
		"exp":     expiresAt.Unix(),
		"iat":     now.Unix(),
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token with secret
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// generateRefreshToken generates a refresh token and stores it in the database
func (s *AuthServiceImpl) generateRefreshToken(ctx context.Context, userID uuid.UUID) (string, error) {
	// Generate UUID token
	tokenString := uuid.New().String()

	// Create refresh token entity
	refreshToken := &entities.RefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		Token:     tokenString, // TODO: hash for production (storing plaintext for MVP)
		ExpiresAt: time.Now().Add(s.refreshTokenTTL),
		CreatedAt: time.Now(),
	}

	// Save to repository
	err := s.refreshTokenRepo.Create(ctx, refreshToken)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
