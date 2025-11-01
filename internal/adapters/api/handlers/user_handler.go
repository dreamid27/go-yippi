package handlers

import (
	"context"
	"errors"
	"net/http"

	"example.com/go-yippi/internal/adapters/api/dto"
	"example.com/go-yippi/internal/application/services"
	"example.com/go-yippi/internal/domain/entities"
	domainErrors "example.com/go-yippi/internal/domain/errors"
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
		Errors:      []int{http.StatusBadRequest, http.StatusInternalServerError},
	}, h.CreateUser)

	huma.Register(api, huma.Operation{
		OperationID: "list-users",
		Method:      http.MethodGet,
		Path:        "/users",
		Summary:     "List all users",
		Description: "Retrieves a list of all users in the system",
		Tags:        []string{"Users"},
		Errors:      []int{http.StatusInternalServerError},
	}, h.GetUsers)

	huma.Register(api, huma.Operation{
		OperationID: "get-user",
		Method:      http.MethodGet,
		Path:        "/users/{id}",
		Summary:     "Get a user by ID",
		Description: "Retrieves a user by their ID",
		Tags:        []string{"Users"},
		Errors:      []int{http.StatusNotFound, http.StatusInternalServerError},
	}, h.GetUser)

	huma.Register(api, huma.Operation{
		OperationID: "update-user",
		Method:      http.MethodPut,
		Path:        "/users/{id}",
		Summary:     "Update a user",
		Description: "Updates an existing user's information",
		Tags:        []string{"Users"},
		Errors:      []int{http.StatusNotFound, http.StatusInternalServerError},
	}, h.UpdateUser)

	huma.Register(api, huma.Operation{
		OperationID:   "delete-user",
		Method:        http.MethodDelete,
		Path:          "/users/{id}",
		Summary:       "Delete a user",
		Description:   "Deletes a user from the system",
		Tags:          []string{"Users"},
		DefaultStatus: http.StatusNoContent,
		Errors:        []int{http.StatusNotFound, http.StatusInternalServerError},
	}, h.DeleteUser)
}

func (h *UserHandler) CreateUser(ctx context.Context, input *dto.CreateUserRequest) (*dto.UserResponse, error) {
	user := &entities.User{
		Name:  input.Body.Name,
		Age: input.Body.Age,
	}

	err := h.service.CreateUser(ctx, user)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to create user", err)
	}

	resp := &dto.UserResponse{}
	resp.Body.ID = user.ID
	resp.Body.Name = user.Name
	resp.Body.Age = user.Age

	return resp, nil
}

func (h *UserHandler) GetUsers(ctx context.Context, input *struct{}) (*dto.ListUsersResponse, error) {
	users, err := h.service.ListUsers(ctx)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to list users", err)
	}

	resp := &dto.ListUsersResponse{}
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

func (h *UserHandler) GetUser(ctx context.Context, input *dto.GetUserRequest) (*dto.UserResponse, error) {
	user, err := h.service.GetUser(ctx, input.ID)
	if err != nil {
		if errors.Is(err, domainErrors.ErrNotFound) {
			return nil, huma.Error404NotFound("User not found")
		}
		return nil, huma.Error500InternalServerError("Failed to get user", err)
	}

	resp := &dto.UserResponse{}
	resp.Body.ID = user.ID
	resp.Body.Name = user.Name
	resp.Body.Age = user.Age

	return resp, nil
}

func (h *UserHandler) UpdateUser(ctx context.Context, input *dto.UpdateUserRequest) (*dto.UserResponse, error) {
	user := &entities.User{
		ID:    input.ID,
		Name:  input.Body.Name,
		Age: input.Body.Age,
	}

	err := h.service.UpdateUser(ctx, user)
	if err != nil {
		if errors.Is(err, domainErrors.ErrNotFound) {
			return nil, huma.Error404NotFound("User not found")
		}
		return nil, huma.Error500InternalServerError("Failed to update user", err)
	}

	resp := &dto.UserResponse{}
	resp.Body.ID = user.ID
	resp.Body.Name = user.Name
	resp.Body.Age = user.Age

	return resp, nil
}

func (h *UserHandler) DeleteUser(ctx context.Context, input *dto.DeleteUserRequest) (*struct{}, error) {
	err := h.service.DeleteUser(ctx, input.ID)
	if err != nil {
		if errors.Is(err, domainErrors.ErrNotFound) {
			return nil, huma.Error404NotFound("User not found")
		}
		return nil, huma.Error500InternalServerError("Failed to delete user", err)
	}

	return &struct{}{}, nil
}
