package service

import (
	"context"
	"fmt"
	"time"

	"github.com/0xsj/result"
	"github.com/0xsj/result/examples/02-http/model"
	"github.com/0xsj/result/examples/02-http/repository"
)

// UserService handles user business logic
type UserService struct {
	repo *repository.UserRepository
}

// NewUserService creates a new user service
func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

// GetUser retrieves a user by ID
func (s *UserService) GetUser(ctx context.Context, id string) result.Result[*model.User] {
	return s.repo.FindByID(ctx, id).
		WithContextOp(ctx, "UserService.GetUser").
		Inspect(func(user *model.User) {
			fmt.Printf("[Service] Found user: %s (%s)\n", user.Name, user.Email)
		}).
		InspectErr(func(err error) {
			fmt.Printf("[Service] Failed to find user: %v\n", err)
		})
}

// CreateUser creates a new user
func (s *UserService) CreateUser(ctx context.Context, req model.CreateUserRequest) result.Result[*model.User] {
	// Use AndThenMap to transform CreateUserRequest -> *User
	return result.AndThenMap(
		req.Validate().WithContextOp(ctx, "UserService.CreateUser"),
		func(validReq model.CreateUserRequest) result.Result[*model.User] {
			// Check if email already exists
			emailCheck := s.repo.FindByEmail(ctx, validReq.Email)
			
			if emailCheck.IsOk() {
				// Email exists - conflict
				return result.Err[*model.User](
					result.Conflict("UserService.CreateUser", "email"),
				)
			}

			// Check if it's a NotFound error (expected) or real error
			if !result.Is(emailCheck.Error(), result.KindNotFound) {
				return result.Err[*model.User](emailCheck.Error())
			}

			// Email doesn't exist - create user
			user := &model.User{
				ID:        generateID(),
				Email:     validReq.Email,
				Name:      validReq.Name,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			
			return s.repo.Save(ctx, user)
		},
	).Inspect(func(user *model.User) {
		fmt.Printf("[Service] Created user: %s (%s)\n", user.Name, user.Email)
	})
}

// UpdateUser updates an existing user
func (s *UserService) UpdateUser(ctx context.Context, id string, req model.UpdateUserRequest) result.Result[*model.User] {
	return s.repo.FindByID(ctx, id).
		WithContextOp(ctx, "UserService.UpdateUser").
		AndThen(func(user *model.User) result.Result[*model.User] {
			// Update name
			if req.Name != "" {
				user.Name = req.Name
			}

			// Update email if provided
			if req.Email != "" {
				emailResult := model.NewEmail(req.Email)
				if emailResult.IsErr() {
					return result.Err[*model.User](emailResult.Error())
				}
				user.Email = emailResult.Unwrap().String()
			}

			return result.Ok(user)
		}).
		AndThen(func(user *model.User) result.Result[*model.User] {
			return s.repo.Save(ctx, user)
		}).
		Inspect(func(user *model.User) {
			fmt.Printf("[Service] Updated user: %s (%s)\n", user.Name, user.Email)
		})
}

// DeleteUser deletes a user
func (s *UserService) DeleteUser(ctx context.Context, id string) result.Result[struct{}] {
	return s.repo.Delete(ctx, id).
		WithContextOp(ctx, "UserService.DeleteUser").
		Inspect(func(_ struct{}) {
			fmt.Printf("[Service] Deleted user: %s\n", id)
		})
}

// ListUsers returns all users
func (s *UserService) ListUsers(ctx context.Context) result.Result[[]*model.User] {
	return s.repo.List(ctx).
		WithContextOp(ctx, "UserService.ListUsers").
		Inspect(func(users []*model.User) {
			fmt.Printf("[Service] Listed %d users\n", len(users))
		})
}

// generateID generates a simple ID (in production, use UUID or similar)
func generateID() string {
	return fmt.Sprintf("user_%d", time.Now().UnixNano())
}