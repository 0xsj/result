package result

import (
	"errors"
	"fmt"
	"testing"
)

// ============================================================================
// Kind Tests
// ============================================================================

func TestKind_String(t *testing.T) {
	tests := []struct {
		kind     Kind
		expected string
	}{
		{KindDomain, "domain"},
		{KindValidation, "validation"},
		{KindNotFound, "not_found"},
		{KindConflict, "conflict"},
		{KindUnauthorized, "unauthorized"},
		{KindForbidden, "forbidden"},
		{KindInfrastructure, "infrastructure"},
		{KindInternal, "internal"},
		{Kind(999), "unknown"}, // Invalid kind
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := tt.kind.String()
			if got != tt.expected {
				t.Errorf("Kind.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// ============================================================================
// Error Tests
// ============================================================================

func TestError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *Error
		expected string
	}{
		{
			name: "with op and message only",
			err: &Error{
				Kind:    KindDomain,
				Op:      "UserService.CreateUser",
				Message: "email already exists",
			},
			expected: "UserService.CreateUser: email already exists",
		},
		{
			name: "with op, message, and underlying error",
			err: &Error{
				Kind:    KindInfrastructure,
				Op:      "DB.Query",
				Message: "database error",
				Err:     errors.New("connection refused"),
			},
			expected: "DB.Query: database error: connection refused",
		},
		{
			name: "with message only (no op)",
			err: &Error{
				Kind:    KindValidation,
				Message: "invalid input",
			},
			expected: "invalid input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.expected {
				t.Errorf("Error.Error() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestError_Unwrap(t *testing.T) {
	underlying := errors.New("underlying error")
	err := &Error{
		Kind:    KindInternal,
		Op:      "test",
		Message: "wrapped",
		Err:     underlying,
	}

	unwrapped := err.Unwrap()
	if unwrapped != underlying {
		t.Errorf("Error.Unwrap() = %v, want %v", unwrapped, underlying)
	}
}

func TestError_Unwrap_Nil(t *testing.T) {
	err := &Error{
		Kind:    KindDomain,
		Op:      "test",
		Message: "no underlying error",
	}

	unwrapped := err.Unwrap()
	if unwrapped != nil {
		t.Errorf("Error.Unwrap() = %v, want nil", unwrapped)
	}
}

// ============================================================================
// Constructor Tests
// ============================================================================

func TestDomain(t *testing.T) {
	err := Domain("UserService.ChangeEmail", "email unchanged")

	var resultErr *Error
	if !errors.As(err, &resultErr) {
		t.Fatal("Domain() should return *Error")
	}

	if resultErr.Kind != KindDomain {
		t.Errorf("Kind = %v, want %v", resultErr.Kind, KindDomain)
	}
	if resultErr.Op != "UserService.ChangeEmail" {
		t.Errorf("Op = %v, want UserService.ChangeEmail", resultErr.Op)
	}
	if resultErr.Message != "email unchanged" {
		t.Errorf("Message = %v, want email unchanged", resultErr.Message)
	}
}

func TestValidation(t *testing.T) {
	meta := map[string]interface{}{
		"field": "email",
		"value": "invalid",
	}
	err := Validation("CreateUser", "invalid email format", meta)

	var resultErr *Error
	if !errors.As(err, &resultErr) {
		t.Fatal("Validation() should return *Error")
	}

	if resultErr.Kind != KindValidation {
		t.Errorf("Kind = %v, want %v", resultErr.Kind, KindValidation)
	}
	if resultErr.Meta == nil {
		t.Fatal("Meta should not be nil")
	}
	if resultErr.Meta["field"] != "email" {
		t.Errorf("Meta[field] = %v, want email", resultErr.Meta["field"])
	}
}

func TestNotFound(t *testing.T) {
	err := NotFound("Repository.FindByID", "user")

	var resultErr *Error
	if !errors.As(err, &resultErr) {
		t.Fatal("NotFound() should return *Error")
	}

	if resultErr.Kind != KindNotFound {
		t.Errorf("Kind = %v, want %v", resultErr.Kind, KindNotFound)
	}
	if resultErr.Message != "user not found" {
		t.Errorf("Message = %v, want 'user not found'", resultErr.Message)
	}
}

func TestConflict(t *testing.T) {
	err := Conflict("CreateUser", "email")

	var resultErr *Error
	if !errors.As(err, &resultErr) {
		t.Fatal("Conflict() should return *Error")
	}

	if resultErr.Kind != KindConflict {
		t.Errorf("Kind = %v, want %v", resultErr.Kind, KindConflict)
	}
	if resultErr.Message != "email already exists" {
		t.Errorf("Message = %v, want 'email already exists'", resultErr.Message)
	}
}

func TestUnauthorized(t *testing.T) {
	err := Unauthorized("Auth.Login", "invalid token")

	var resultErr *Error
	if !errors.As(err, &resultErr) {
		t.Fatal("Unauthorized() should return *Error")
	}

	if resultErr.Kind != KindUnauthorized {
		t.Errorf("Kind = %v, want %v", resultErr.Kind, KindUnauthorized)
	}
}

func TestForbidden(t *testing.T) {
	err := Forbidden("Auth.Access", "insufficient permissions")

	var resultErr *Error
	if !errors.As(err, &resultErr) {
		t.Fatal("Forbidden() should return *Error")
	}

	if resultErr.Kind != KindForbidden {
		t.Errorf("Kind = %v, want %v", resultErr.Kind, KindForbidden)
	}
}

func TestInfrastructure(t *testing.T) {
	underlying := errors.New("connection refused")
	err := Infrastructure("DB.Connect", underlying)

	var resultErr *Error
	if !errors.As(err, &resultErr) {
		t.Fatal("Infrastructure() should return *Error")
	}

	if resultErr.Kind != KindInfrastructure {
		t.Errorf("Kind = %v, want %v", resultErr.Kind, KindInfrastructure)
	}
	if resultErr.Err != underlying {
		t.Errorf("Err = %v, want %v", resultErr.Err, underlying)
	}
}

func TestInternal(t *testing.T) {
	underlying := errors.New("unexpected panic")
	err := Internal("Service.Process", underlying)

	var resultErr *Error
	if !errors.As(err, &resultErr) {
		t.Fatal("Internal() should return *Error")
	}

	if resultErr.Kind != KindInternal {
		t.Errorf("Kind = %v, want %v", resultErr.Kind, KindInternal)
	}
	if resultErr.Err != underlying {
		t.Errorf("Err = %v, want %v", resultErr.Err, underlying)
	}
}

// ============================================================================
// Helper Tests
// ============================================================================

func TestKindOf(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected Kind
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: KindInternal,
		},
		{
			name:     "result.Error with KindNotFound",
			err:      NotFound("test", "resource"),
			expected: KindNotFound,
		},
		{
			name:     "standard error (not result.Error)",
			err:      errors.New("standard error"),
			expected: KindInternal,
		},
		{
			name:     "wrapped result.Error",
			err:      fmt.Errorf("wrapped: %w", Domain("test", "message")),
			expected: KindDomain,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := KindOf(tt.err)
			if got != tt.expected {
				t.Errorf("KindOf() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIs(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		kind     Kind
		expected bool
	}{
		{
			name:     "matching kind",
			err:      NotFound("test", "resource"),
			kind:     KindNotFound,
			expected: true,
		},
		{
			name:     "non-matching kind",
			err:      NotFound("test", "resource"),
			kind:     KindValidation,
			expected: false,
		},
		{
			name:     "standard error",
			err:      errors.New("standard"),
			kind:     KindInternal,
			expected: true,
		},
		{
			name:     "nil error",
			err:      nil,
			kind:     KindInternal,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Is(tt.err, tt.kind)
			if got != tt.expected {
				t.Errorf("Is() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestOpOf(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "result.Error with Op",
			err:      Domain("UserService.CreateUser", "message"),
			expected: "UserService.CreateUser",
		},
		{
			name:     "standard error",
			err:      errors.New("standard"),
			expected: "",
		},
		{
			name:     "nil error",
			err:      nil,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := OpOf(tt.err)
			if got != tt.expected {
				t.Errorf("OpOf() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestMetaOf(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected map[string]interface{}
	}{
		{
			name: "result.Error with Meta",
			err: Validation("test", "message", map[string]interface{}{
				"field": "email",
			}),
			expected: map[string]interface{}{
				"field": "email",
			},
		},
		{
			name:     "result.Error without Meta",
			err:      Domain("test", "message"),
			expected: nil,
		},
		{
			name:     "standard error",
			err:      errors.New("standard"),
			expected: nil,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MetaOf(tt.err)
			if tt.expected == nil && got != nil {
				t.Errorf("MetaOf() = %v, want nil", got)
			}
			if tt.expected != nil {
				if got == nil {
					t.Fatal("MetaOf() = nil, want non-nil")
				}
				for k, v := range tt.expected {
					if got[k] != v {
						t.Errorf("MetaOf()[%s] = %v, want %v", k, got[k], v)
					}
				}
			}
		})
	}
}

func TestWrap(t *testing.T) {
	t.Run("wrap result.Error preserves Kind", func(t *testing.T) {
		original := NotFound("Repository.FindByID", "user")
		wrapped := Wrap(original, "Service.GetUser")

		if KindOf(wrapped) != KindNotFound {
			t.Errorf("Wrap() Kind = %v, want %v", KindOf(wrapped), KindNotFound)
		}

		if OpOf(wrapped) != "Service.GetUser" {
			t.Errorf("Wrap() Op = %v, want Service.GetUser", OpOf(wrapped))
		}
	})

	t.Run("wrap standard error creates KindInternal", func(t *testing.T) {
		original := errors.New("standard error")
		wrapped := Wrap(original, "Service.Process")

		if KindOf(wrapped) != KindInternal {
			t.Errorf("Wrap() Kind = %v, want %v", KindOf(wrapped), KindInternal)
		}

		if OpOf(wrapped) != "Service.Process" {
			t.Errorf("Wrap() Op = %v, want Service.Process", OpOf(wrapped))
		}
	})

	t.Run("wrap nil returns nil", func(t *testing.T) {
		wrapped := Wrap(nil, "test")
		if wrapped != nil {
			t.Errorf("Wrap(nil) = %v, want nil", wrapped)
		}
	})

	t.Run("wrap preserves metadata", func(t *testing.T) {
		original := Validation("test", "message", map[string]interface{}{
			"field": "email",
		})
		wrapped := Wrap(original, "Handler.Process")

		meta := MetaOf(wrapped)
		if meta == nil {
			t.Fatal("Wrap() should preserve metadata")
		}
		if meta["field"] != "email" {
			t.Errorf("Wrap() meta[field] = %v, want email", meta["field"])
		}
	})
}

// ============================================================================
// Integration Tests (errors.Is / errors.As compatibility)
// ============================================================================

func TestStdlibErrorsIs(t *testing.T) {
	domainErr := errors.New("domain specific error")
	wrapped := Infrastructure("DB.Query", domainErr)

	if !errors.Is(wrapped, domainErr) {
		t.Error("errors.Is() should work with wrapped errors")
	}
}

func TestStdlibErrorsAs(t *testing.T) {
	err := NotFound("test", "resource")

	var resultErr *Error
	if !errors.As(err, &resultErr) {
		t.Error("errors.As() should extract *Error")
	}

	if resultErr.Kind != KindNotFound {
		t.Errorf("extracted error Kind = %v, want %v", resultErr.Kind, KindNotFound)
	}
}