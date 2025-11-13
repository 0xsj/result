// Package result provides a Result[T] type for explicit error handling
// with functional programming patterns inspired by Rust.
package result

import (
	"errors"
	"fmt"
)

// Kind represents the category of error
type Kind int

const (
	KindDomain         Kind = iota // Business rule violation
	KindValidation                 // Input validation failure
	KindNotFound                   // Resource not found
	KindConflict                   // Resource conflict (already exists)
	KindUnauthorized               // Authentication failure
	KindForbidden                  // Authorization failure
	KindInfrastructure             // External system failure
	KindInternal                   // Unexpected internal error
)

// String returns the string representation of Kind
func (k Kind) String() string {
	switch k {
	case KindDomain:
		return "domain"
	case KindValidation:
		return "validation"
	case KindNotFound:
		return "not_found"
	case KindConflict:
		return "conflict"
	case KindUnauthorized:
		return "unauthorized"
	case KindForbidden:
		return "forbidden"
	case KindInfrastructure:
		return "infrastructure"
	case KindInternal:
		return "internal"
	default:
		return "unknown"
	}
}

// Error represents a structured error with categorization and context
type Error struct {
	Kind    Kind
	Op      string                 // Operation being performed
	Err     error                  // Underlying error
	Message string                 // User-facing message
	Meta    map[string]interface{} // Additional metadata
}

// Error implements the error interface
func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Op, e.Message, e.Err)
	}
	if e.Op != "" {
		return fmt.Sprintf("%s: %s", e.Op, e.Message)
	}
	return e.Message
}

// Unwrap returns the underlying error for errors.Is and errors.As
func (e *Error) Unwrap() error {
	return e.Err
}

// ============================================================================
// Constructors
// ============================================================================

// Domain creates a domain error (business rule violation)
func Domain(op, message string) error {
	return &Error{Kind: KindDomain, Op: op, Message: message}
}

// Validation creates a validation error with optional metadata
func Validation(op, message string, meta map[string]interface{}) error {
	return &Error{Kind: KindValidation, Op: op, Message: message, Meta: meta}
}

// NotFound creates a not found error
func NotFound(op, resource string) error {
	return &Error{Kind: KindNotFound, Op: op, Message: fmt.Sprintf("%s not found", resource)}
}

// Conflict creates a conflict error (e.g., duplicate resource)
func Conflict(op, resource string) error {
	return &Error{Kind: KindConflict, Op: op, Message: fmt.Sprintf("%s already exists", resource)}
}

// Unauthorized creates an authentication error
func Unauthorized(op, message string) error {
	return &Error{Kind: KindUnauthorized, Op: op, Message: message}
}

// Forbidden creates an authorization error
func Forbidden(op, message string) error {
	return &Error{Kind: KindForbidden, Op: op, Message: message}
}

// Infrastructure creates an infrastructure error (external system failure)
func Infrastructure(op string, err error) error {
	return &Error{Kind: KindInfrastructure, Op: op, Err: err, Message: "infrastructure error"}
}

// Internal creates an internal error (unexpected failure)
func Internal(op string, err error) error {
	return &Error{Kind: KindInternal, Op: op, Err: err, Message: "internal error"}
}

// ============================================================================
// Helpers
// ============================================================================

// KindOf returns the Kind of an error
func KindOf(err error) Kind {
	if err == nil {
		return KindInternal
	}

	var e *Error
	if errors.As(err, &e) {
		return e.Kind
	}
	return KindInternal
}

// Is checks if an error is of a specific Kind
func Is(err error, kind Kind) bool {
	return KindOf(err) == kind
}

// OpOf returns the operation name from an error
func OpOf(err error) string {
	var e *Error
	if errors.As(err, &e) {
		return e.Op
	}
	return ""
}

// MetaOf returns the metadata from an error
func MetaOf(err error) map[string]interface{} {
	var e *Error
	if errors.As(err, &e) {
		return e.Meta
	}
	return nil
}

// Wrap wraps an error with additional context while preserving Kind
func Wrap(err error, op string) error {
	if err == nil {
		return nil
	}

	var e *Error
	if errors.As(err, &e) {
		return &Error{
			Kind:    e.Kind,
			Op:      op,
			Err:     err,
			Message: e.Message,
			Meta:    e.Meta,
		}
	}

	return &Error{
		Kind:    KindInternal,
		Op:      op,
		Err:     err,
		Message: err.Error(),
	}
}