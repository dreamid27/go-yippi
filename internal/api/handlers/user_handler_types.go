package handlers

// CreateUserRequest defines the request body for creating a user
type CreateUserRequest struct {
	Body struct {
		Name  string `json:"name" minLength:"1" doc:"User name"`
		Email string `json:"email" format:"email" doc:"User email"`
	}
}

// UserResponse defines the response for user operations
type UserResponse struct {
	Body struct {
		ID    int    `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}
}

// GetUserRequest defines the request for getting a single user
type GetUserRequest struct {
	ID int `path:"id" doc:"User ID"`
}

// ListUsersResponse defines the response for listing users
type ListUsersResponse struct {
	Body struct {
		Users []struct {
			ID    int    `json:"id"`
			Name  string `json:"name"`
			Email string `json:"email"`
		} `json:"users"`
	}
}

// UpdateUserRequest defines the request for updating a user
type UpdateUserRequest struct {
	ID   int `path:"id" doc:"User ID"`
	Body struct {
		Name  string `json:"name" minLength:"1" doc:"User name"`
		Email string `json:"email" format:"email" doc:"User email"`
	}
}

// DeleteUserRequest defines the request for deleting a user
type DeleteUserRequest struct {
	ID int `path:"id" doc:"User ID"`
}
