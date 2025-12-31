package dto

// CreateUserRequest defines the request body for creating a user
type CreateUserRequest struct {
	Body struct {
		Email    string `json:"email" format:"email" doc:"User email"`
		Password string `json:"password" minLength:"6" doc:"User password"`
		Name     string `json:"name" minLength:"1" doc:"User name"`
		Phone    string `json:"phone,omitempty" doc:"User phone (optional)"`
	}
}

// UserResponse defines the response for user operations
type UserResponse struct {
	Body struct {
		ID       string `json:"id" doc:"User ID (UUID)"`
		Email    string `json:"email"`
		Name     string `json:"name"`
		Phone    string `json:"phone,omitempty"`
		Role     string `json:"role"`
		IsActive bool   `json:"is_active"`
	}
}

// GetUserRequest defines the request for getting a single user
type GetUserRequest struct {
	ID string `path:"id" doc:"User ID (UUID)"`
}

// UserListItem represents a user in a list response
type UserListItem struct {
	ID       string `json:"id" doc:"User ID (UUID)"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Phone    string `json:"phone,omitempty"`
	Role     string `json:"role"`
	IsActive bool   `json:"is_active"`
}

// ListUsersResponse defines the response for listing users
type ListUsersResponse struct {
	Body struct {
		Users []UserListItem `json:"users"`
	}
}

// UpdateUserRequest defines the request for updating a user
type UpdateUserRequest struct {
	ID   string `path:"id" doc:"User ID (UUID)"`
	Body struct {
		Email    string `json:"email" format:"email" doc:"User email"`
		Name     string `json:"name" minLength:"1" doc:"User name"`
		Phone    string `json:"phone,omitempty" doc:"User phone (optional)"`
		Role     string `json:"role,omitempty" enum:"customer,admin" doc:"User role (optional)"`
		IsActive *bool  `json:"is_active,omitempty" doc:"User active status (optional)"`
	}
}

// DeleteUserRequest defines the request for deleting a user
type DeleteUserRequest struct {
	ID string `path:"id" doc:"User ID (UUID)"`
}
