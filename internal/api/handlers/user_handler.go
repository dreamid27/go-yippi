package handlers

import (
	"context"
	"net/http"

	"example.com/go-yippi/internal/application/services"
	"example.com/go-yippi/internal/domain/entities"
	"github.com/danielgtaylor/huma/v2"
)

// UserHandler handles HTTP requests for users
type UserHandler struct {
	service *services.UserService
}

func NewUserHandler(service *services.UserService) *UserHandler {
	return &UserHandler{service: service}
}

// RegisterRoutes registers all user routes with Huma
func (h *UserHandler) RegisterRoutes(api huma.API) {
	// Register operations with Huma
	huma.Register(api, huma.Operation{
		OperationID: "create-user",
		Method:      http.MethodPost,
		Path:        "/users",
		Summary:     "Create a new user",
		Description: "Creates a new user with the provided name and age",
		Tags:        []string{"Users"},
	}, h.CreateUser)

	huma.Register(api, huma.Operation{
		OperationID: "list-users",
		Method:      http.MethodGet,
		Path:        "/users",
		Summary:     "List all users",
		Description: "Retrieves a list of all users in the system",
		Tags:        []string{"Users"},
	}, h.GetUsers)

	huma.Register(api, huma.Operation{
		OperationID: "get-user",
		Method:      http.MethodGet,
		Path:        "/users/{id}",
		Summary:     "Get a user by ID",
		Description: "Retrieves a double user by their ID",
		Tags:        []string{"Users"},
	}, h.GetUser)

	huma.Register(api, huma.Operation{
		OperationID: "update-user",
		Method:      http.MethodPut,
		Path:        "/users/{id}",
		Summary:     "Update a user",
		Description: "Updates an existing user's information",
		Tags:        []string{"Users"},
	}, h.UpdateUser)

	huma.Register(api, huma.Operation{
		OperationID:   "delete-user",
		Method:        http.MethodDelete,
		Path:          "/users/{id}",
		Summary:       "Delete a user",
		Description:   "Deletes a user from the system",
		Tags:          []string{"Users"},
		DefaultStatus: http.StatusNoContent,
	}, h.DeleteUser)
}

func (h *UserHandler) CreateUser(ctx context.Context, input *CreateUserRequest) (*UserResponse, error) {
	user := &entities.User{
		Name:  input.Body.Name,
		Age: input.Body.Age,
	}

	err := h.service.CreateUser(ctx, user)
	if err != nil {
		return nil, err
	}

	resp := &UserResponse{}
	resp.Body.ID = user.ID
	resp.Body.Name = user.Name
	resp.Body.Age = user.Age

	return resp, nil
}

func (h *UserHandler) GetUsers(ctx context.Context, input *struct{}) (*ListUsersResponse, error) {
	users, err := h.service.ListUsers(ctx)
	if err != nil {
		return nil, err
	}

	resp := &ListUsersResponse{}
	resp.Body.Users = make([]struct {
		ID    int    `json:"id"`
		Name  string `json:"name"`
		Age int `json:"age"`
	}, len(users))

	for i, user := range users {
		resp.Body.Users[i].ID = user.ID
		resp.Body.Users[i].Name = user.Name
		resp.Body.Users[i].Age = user.Age
	}

	return resp, nil
}

func (h *UserHandler) GetUser(ctx context.Context, input *GetUserRequest) (*UserResponse, error) {
	user, err := h.service.GetUser(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	resp := &UserResponse{}
	resp.Body.ID = user.ID
	resp.Body.Name = user.Name
	resp.Body.Age = user.Age

	return resp, nil
}

func (h *UserHandler) UpdateUser(ctx context.Context, input *UpdateUserRequest) (*UserResponse, error) {
	user := &entities.User{
		ID:    input.ID,
		Name:  input.Body.Name,
		Age: input.Body.Age,
	}

	err := h.service.UpdateUser(ctx, user)
	if err != nil {
		return nil, err
	}

	resp := &UserResponse{}
	resp.Body.ID = user.ID
	resp.Body.Name = user.Name
	resp.Body.Age = user.Age

	return resp, nil
}

func (h *UserHandler) DeleteUser(ctx context.Context, input *DeleteUserRequest) (*struct{}, error) {
	err := h.service.DeleteUser(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	return &struct{}{}, nil
}
