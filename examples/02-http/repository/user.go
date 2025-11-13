package repository

import (
	"context"
	"sync"
	"time"

	"github.com/0xsj/result"
	"github.com/0xsj/result/examples/02-http/model"
)

// UserRepository handles user data access
type UserRepository struct {
	mu    sync.RWMutex
	users map[string]*model.User
}

// NewUserRepository creates a new user repository
func NewUserRepository() *UserRepository {
	return &UserRepository{
		users: make(map[string]*model.User),
	}
}

// FindByID finds a user by ID
func (r *UserRepository) FindByID(ctx context.Context, id string) result.Result[*model.User] {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[id]
	if !exists {
		return result.Err[*model.User](
			result.NotFound("UserRepository.FindByID", "user"),
		).WithContext(ctx)
	}

	return result.Ok(user)
}

// FindByEmail finds a user by email
func (r *UserRepository) FindByEmail(ctx context.Context, email string) result.Result[*model.User] {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, user := range r.users {
		if user.Email == email {
			return result.Ok(user)
		}
	}

	return result.Err[*model.User](
		result.NotFound("UserRepository.FindByEmail", "user"),
	).WithContext(ctx)
}

// Save creates or updates a user
func (r *UserRepository) Save(ctx context.Context, user *model.User) result.Result[*model.User] {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check for email conflict (excluding same user)
	for _, existing := range r.users {
		if existing.Email == user.Email && existing.ID != user.ID {
			return result.Err[*model.User](
				result.Conflict("UserRepository.Save", "email"),
			).WithContext(ctx)
		}
	}

	user.UpdatedAt = time.Now()
	r.users[user.ID] = user

	return result.Ok(user)
}

// Delete removes a user
func (r *UserRepository) Delete(ctx context.Context, id string) result.Result[struct{}] {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.users[id]; !exists {
		return result.Err[struct{}](
			result.NotFound("UserRepository.Delete", "user"),
		).WithContext(ctx)
	}

	delete(r.users, id)
	return result.Ok(struct{}{})
}

// List returns all users
func (r *UserRepository) List(ctx context.Context) result.Result[[]*model.User] {
	r.mu.RLock()
	defer r.mu.RUnlock()

	users := make([]*model.User, 0, len(r.users))
	for _, user := range r.users {
		users = append(users, user)
	}

	return result.Ok(users)
}