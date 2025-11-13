package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/0xsj/result"
	"github.com/0xsj/result/examples/02-http/model"
	"github.com/0xsj/result/examples/02-http/service"
)

// UserHandler handles HTTP requests for users
type UserHandler struct {
	service *service.UserService
}

// NewUserHandler creates a new user handler
func NewUserHandler(service *service.UserService) *UserHandler {
	return &UserHandler{service: service}
}

// ServeHTTP implements http.Handler
func (h *UserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Simple routing based on method and path
	path := strings.TrimPrefix(r.URL.Path, "/users")

	switch {
	case r.Method == http.MethodGet && path == "":
		h.ListUsers(w, r)
	case r.Method == http.MethodGet && path != "":
		h.GetUser(w, r, strings.TrimPrefix(path, "/"))
	case r.Method == http.MethodPost && path == "":
		h.CreateUser(w, r)
	case r.Method == http.MethodPut && path != "":
		h.UpdateUser(w, r, strings.TrimPrefix(path, "/"))
	case r.Method == http.MethodDelete && path != "":
		h.DeleteUser(w, r, strings.TrimPrefix(path, "/"))
	default:
		http.Error(w, "Not Found", http.StatusNotFound)
	}
}

// GetUser handles GET /users/{id}
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request, id string) {
	fmt.Printf("[Handler] GET /users/%s\n", id)

	h.service.GetUser(r.Context(), id).Match(
		func(user *model.User) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(user)
		},
		func(err error) {
			result.WriteHTTPError(w, err)
		},
	)
}

// CreateUser handles POST /users
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	fmt.Println("[Handler] POST /users")

	var req model.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	h.service.CreateUser(r.Context(), req).Match(
		func(user *model.User) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(user)
		},
		func(err error) {
			result.WriteHTTPError(w, err)
		},
	)
}

// UpdateUser handles PUT /users/{id}
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request, id string) {
	fmt.Printf("[Handler] PUT /users/%s\n", id)

	var req model.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	h.service.UpdateUser(r.Context(), id, req).Match(
		func(user *model.User) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(user)
		},
		func(err error) {
			result.WriteHTTPError(w, err)
		},
	)
}

// DeleteUser handles DELETE /users/{id}
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request, id string) {
	fmt.Printf("[Handler] DELETE /users/%s\n", id)

	h.service.DeleteUser(r.Context(), id).Match(
		func(_ struct{}) {
			w.WriteHeader(http.StatusNoContent)
		},
		func(err error) {
			result.WriteHTTPError(w, err)
		},
	)
}

// ListUsers handles GET /users
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	fmt.Println("[Handler] GET /users")

	h.service.ListUsers(r.Context()).Match(
		func(users []*model.User) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"users": users,
				"count": len(users),
			})
		},
		func(err error) {
			result.WriteHTTPError(w, err)
		},
	)
}