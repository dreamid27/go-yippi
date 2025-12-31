package dto

// RegisterRequest defines the request body for user registration
type RegisterRequest struct {
	Body struct {
		Email    string `json:"email" format:"email" doc:"User email address" example:"user@example.com"`
		Password string `json:"password" minLength:"8" doc:"User password (minimum 8 characters)" example:"securepass123"`
		Name     string `json:"name" minLength:"2" doc:"User full name" example:"John Doe"`
		Phone    string `json:"phone" minLength:"1" doc:"User phone number" example:"+628123456789"`
	}
}

// LoginRequest defines the request body for user login
type LoginRequest struct {
	Body struct {
		Email    string `json:"email" format:"email" doc:"User email address" example:"user@example.com"`
		Password string `json:"password" minLength:"1" doc:"User password" example:"securepass123"`
	}
}

// RefreshTokenRequest defines the request body for refreshing access token
type RefreshTokenRequest struct {
	Body struct {
		RefreshToken string `json:"refresh_token" minLength:"1" doc:"Refresh token to obtain new access token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	}
}

// LogoutRequest defines the request body for user logout
type LogoutRequest struct {
	Body struct {
		RefreshToken string `json:"refresh_token" minLength:"1" doc:"Refresh token to invalidate" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	}
}

// UserData represents user information in auth responses
type UserData struct {
	ID       string `json:"id" doc:"User unique identifier (UUID)" example:"550e8400-e29b-41d4-a716-446655440000"`
	Email    string `json:"email" doc:"User email address" example:"user@example.com"`
	Name     string `json:"name" doc:"User full name" example:"John Doe"`
	Phone    string `json:"phone" doc:"User phone number" example:"+628123456789"`
	Role     string `json:"role" doc:"User role (customer or admin)" example:"customer"`
	IsActive bool   `json:"is_active" doc:"User account active status" example:"true"`
}

// AuthResponse defines the response body for authentication operations
type AuthResponse struct {
	Body struct {
		AccessToken  string   `json:"access_token" doc:"JWT access token for API authentication" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
		RefreshToken string   `json:"refresh_token" doc:"Refresh token for obtaining new access tokens" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
		ExpiresIn    int      `json:"expires_in" doc:"Access token expiration time in seconds" example:"3600"`
		User         UserData `json:"user" doc:"Authenticated user information"`
	}
}
