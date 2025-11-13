package model

import (
	"time"

	"github.com/0xsj/result"
)

// User represents a user in the system
type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Email represents a validated email address
type Email string

// NewEmail creates and validates an email address
func NewEmail(s string) result.Result[Email] {
	if s == "" {
		return result.Err[Email](
			result.Validation("NewEmail", "email is required", map[string]interface{}{
				"field": "email",
			}),
		)
	}

	// Simplified validation - in production use a proper email validator
	if len(s) < 5 || !containsAt(s) {
		return result.Err[Email](
			result.Validation("NewEmail", "invalid email format", map[string]interface{}{
				"field": "email",
				"value": s,
			}),
		)
	}

	return result.Ok(Email(s))
}

func (e Email) String() string {
	return string(e)
}

func containsAt(s string) bool {
	for _, c := range s {
		if c == '@' {
			return true
		}
	}
	return false
}

// CreateUserRequest represents the request to create a user
type CreateUserRequest struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

// UpdateUserRequest represents the request to update a user
type UpdateUserRequest struct {
	Email string `json:"email,omitempty"`
	Name  string `json:"name,omitempty"`
}

// Validate validates the create user request
func (r CreateUserRequest) Validate() result.Result[CreateUserRequest] {
	if r.Name == "" {
		return result.Err[CreateUserRequest](
			result.Validation("CreateUserRequest", "name is required", map[string]interface{}{
				"field": "name",
			}),
		)
	}

	// Validate email
	emailResult := NewEmail(r.Email)
	if emailResult.IsErr() {
		return result.Err[CreateUserRequest](emailResult.Error())
	}

	return result.Ok(r)
}