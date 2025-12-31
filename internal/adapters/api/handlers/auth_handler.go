package handlers

import (
	"context"
	"errors"
	"log"
	"net/http"

	"example.com/go-yippi/internal/adapters/api/dto"
	"example.com/go-yippi/internal/domain/entities"
	domainErrors "example.com/go-yippi/internal/domain/errors"
	"example.com/go-yippi/internal/domain/ports"
	"github.com/danielgtaylor/huma/v2"
)

// AuthHandler handles HTTP requests for authentication
type AuthHandler struct {
	authService ports.AuthService
}

// NewAuthHandler creates a new AuthHandler instance
func NewAuthHandler(authService ports.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// RegisterRoutes registers all authentication routes with Huma
func (h *AuthHandler) RegisterRoutes(api huma.API) {
	// Register user
	huma.Register(api, huma.Operation{
		OperationID: "register",
		Method:      http.MethodPost,
		Path:        "/auth/register",
		Summary:     "Register a new user",
		Description: "Creates a new user account and automatically logs them in",
		Tags:        []string{"Authentication"},
		Errors:      []int{http.StatusBadRequest, http.StatusConflict, http.StatusInternalServerError},
	}, h.Register)

	// Login user
	huma.Register(api, huma.Operation{
		OperationID: "login",
		Method:      http.MethodPost,
		Path:        "/auth/login",
		Summary:     "Login user",
		Description: "Authenticates a user with email and password",
		Tags:        []string{"Authentication"},
		Errors:      []int{http.StatusBadRequest, http.StatusUnauthorized, http.StatusInternalServerError},
	}, h.Login)

	// Refresh access token
	huma.Register(api, huma.Operation{
		OperationID: "refresh-token",
		Method:      http.MethodPost,
		Path:        "/auth/refresh",
		Summary:     "Refresh access token",
		Description: "Obtains a new access token using a valid refresh token",
		Tags:        []string{"Authentication"},
		Errors:      []int{http.StatusBadRequest, http.StatusUnauthorized, http.StatusInternalServerError},
	}, h.RefreshToken)

	// Logout user
	huma.Register(api, huma.Operation{
		OperationID: "logout",
		Method:      http.MethodPost,
		Path:        "/auth/logout",
		Summary:     "Logout user",
		Description: "Invalidates the user's refresh token",
		Tags:        []string{"Authentication"},
		Errors:      []int{http.StatusInternalServerError},
	}, h.Logout)

	// Get current user
	huma.Register(api, huma.Operation{
		OperationID: "get-current-user",
		Method:      http.MethodGet,
		Path:        "/auth/me",
		Summary:     "Get current user",
		Description: "Retrieves the authenticated user's information",
		Tags:        []string{"Authentication"},
		Errors:      []int{http.StatusUnauthorized, http.StatusInternalServerError},
	}, h.GetCurrentUser)
}

// Register handles user registration and auto-login
func (h *AuthHandler) Register(ctx context.Context, input *dto.RegisterRequest) (*dto.AuthResponse, error) {
	// Create register input from request
	registerInput := ports.RegisterInput{
		Email:    input.Body.Email,
		Password: input.Body.Password,
		Name:     input.Body.Name,
		Phone:    input.Body.Phone,
	}

	// Register the user
	_, err := h.authService.Register(ctx, registerInput)
	if err != nil {
		// Check for validation errors
		if errors.Is(err, domainErrors.ErrInvalidInput) {
			return nil, huma.Error400BadRequest("Invalid input", err)
		}
		// Check for duplicate errors
		if errors.Is(err, domainErrors.ErrDuplicateEntry) {
			return nil, huma.Error409Conflict("Email already exists")
		}
		return nil, huma.Error500InternalServerError("Failed to register user", err)
	}

	// Auto-login after successful registration
	authResp, err := h.authService.Login(ctx, input.Body.Email, input.Body.Password)
	if err != nil {
		// Registration succeeded but login failed - this shouldn't happen
		// Return error but user is created
		return nil, huma.Error500InternalServerError("User registered but auto-login failed", err)
	}

	// Build response
	resp := &dto.AuthResponse{}
	resp.Body.AccessToken = authResp.AccessToken
	resp.Body.RefreshToken = authResp.RefreshToken
	resp.Body.ExpiresIn = authResp.ExpiresIn
	resp.Body.User = convertUserToDTO(authResp.User)

	return resp, nil
}

// Login handles user authentication
func (h *AuthHandler) Login(ctx context.Context, input *dto.LoginRequest) (*dto.AuthResponse, error) {
	// Authenticate user
	authResp, err := h.authService.Login(ctx, input.Body.Email, input.Body.Password)
	if err != nil {
		// Check for validation errors
		if errors.Is(err, domainErrors.ErrInvalidInput) {
			return nil, huma.Error400BadRequest("Invalid input", err)
		}
		// Check for not found or unauthorized errors
		if errors.Is(err, domainErrors.ErrNotFound) || errors.Is(err, domainErrors.ErrUnauthorized) {
			return nil, huma.Error401Unauthorized("Invalid email or password")
		}
		return nil, huma.Error500InternalServerError("Failed to login", err)
	}

	// Build response
	resp := &dto.AuthResponse{}
	resp.Body.AccessToken = authResp.AccessToken
	resp.Body.RefreshToken = authResp.RefreshToken
	resp.Body.ExpiresIn = authResp.ExpiresIn
	resp.Body.User = convertUserToDTO(authResp.User)

	return resp, nil
}

// RefreshToken handles access token refresh
func (h *AuthHandler) RefreshToken(ctx context.Context, input *dto.RefreshTokenRequest) (*dto.AuthResponse, error) {
	// Refresh the access token
	authResp, err := h.authService.RefreshToken(ctx, input.Body.RefreshToken)
	if err != nil {
		// Check for validation errors
		if errors.Is(err, domainErrors.ErrInvalidInput) {
			return nil, huma.Error400BadRequest("Invalid refresh token", err)
		}
		// Check for not found or unauthorized errors
		if errors.Is(err, domainErrors.ErrNotFound) || errors.Is(err, domainErrors.ErrUnauthorized) {
			return nil, huma.Error401Unauthorized("Invalid or expired refresh token")
		}
		return nil, huma.Error500InternalServerError("Failed to refresh token", err)
	}

	// Build response
	resp := &dto.AuthResponse{}
	resp.Body.AccessToken = authResp.AccessToken
	resp.Body.RefreshToken = authResp.RefreshToken
	resp.Body.ExpiresIn = authResp.ExpiresIn
	resp.Body.User = convertUserToDTO(authResp.User)

	return resp, nil
}

// Logout handles user logout
func (h *AuthHandler) Logout(ctx context.Context, input *dto.LogoutRequest) (*struct{}, error) {
	// Invalidate the refresh token
	err := h.authService.Logout(ctx, input.Body.RefreshToken)
	if err != nil {
		// Log the error but don't fail - logout should always succeed from user perspective
		log.Printf("Logout error (non-critical): %v", err)
	}

	// Always return success
	return &struct{}{}, nil
}

// GetCurrentUser retrieves the authenticated user's information
func (h *AuthHandler) GetCurrentUser(ctx context.Context, input *struct{}) (*struct {
	Body dto.UserData `json:"body"`
}, error) {
	// Extract user from context (set by auth middleware)
	user, ok := ctx.Value("user").(*entities.User)
	if !ok || user == nil {
		return nil, huma.Error401Unauthorized("User not authenticated")
	}

	// Build response
	resp := &struct {
		Body dto.UserData `json:"body"`
	}{}
	resp.Body = convertUserToDTO(user)

	return resp, nil
}

// convertUserToDTO converts a domain User entity to a UserData DTO
func convertUserToDTO(user *entities.User) dto.UserData {
	return dto.UserData{
		ID:       user.ID.String(),
		Email:    user.Email,
		Name:     user.Name,
		Phone:    user.Phone,
		Role:     string(user.Role),
		IsActive: user.IsActive,
	}
}
