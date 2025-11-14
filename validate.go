package result

import (
	"fmt"
	"strings"
)

// MultiError holds multiple validation errors.
// Useful for collecting all validation failures instead of failing fast.
type MultiError struct {
	Errors []error
}

// Error implements the error interface.
func (m MultiError) Error() string {
	if len(m.Errors) == 0 {
		return "no errors"
	}
	if len(m.Errors) == 1 {
		return m.Errors[0].Error()
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%d validation errors: ", len(m.Errors)))
	for i, err := range m.Errors {
		if i > 0 {
			sb.WriteString("; ")
		}
		sb.WriteString(err.Error())
	}
	return sb.String()
}

// Unwrap returns the first error for errors.Is and errors.As compatibility.
func (m MultiError) Unwrap() error {
	if len(m.Errors) == 0 {
		return nil
	}
	return m.Errors[0]
}

// Validator is an interface that any Result can implement for validation purposes.
// All Result[T] types automatically implement this interface.
type Validator interface {
	IsErr() bool
	UnwrapErr() error
}

// ValidateAll validates multiple Results and collects ALL errors (not fail-fast).
// Returns Ok with all values if all succeed, or Err with MultiError containing all failures.
//
// This is useful for form validation where you want to show all errors at once.
//
// Example:
//
//	result := result.ValidateAll(
//	    domain.NewEmail(cmd.Email),
//	    domain.NewUsername(cmd.Username),
//	    domain.NewPassword(cmd.Password),
//	)
//	// If multiple validations fail, all errors are collected
func ValidateAll[T any](results ...Result[T]) Result[[]T] {
	var values []T
	var errors []error

	for _, r := range results {
		if r.IsOk() {
			values = append(values, r.Unwrap())
		} else {
			errors = append(errors, r.UnwrapErr())
		}
	}

	if len(errors) > 0 {
		return Err[[]T](MultiError{Errors: errors})
	}

	return Ok(values)
}

// ValidateAllMap validates Results of different types and returns all values or all errors.
// Returns a struct with slices of Ok values and errors.
//
// Example:
//
//	results := result.ValidateAllMap(
//	    domain.NewEmail(cmd.Email),
//	    domain.NewUsername(cmd.Username),
//	    domain.NewPassword(cmd.Password),
//	)
//	if len(results.Errors) > 0 {
//	    // Handle multiple validation errors
//	}
func ValidateAllMap[T any](results ...Result[T]) (values []T, errors []error) {
	for _, r := range results {
		if r.IsOk() {
			values = append(values, r.Unwrap())
		} else {
			errors = append(errors, r.UnwrapErr())
		}
	}
	return values, errors
}

// ValidateRules applies multiple validation rules to a value.
// Returns Ok if all rules pass, or Err with the first failing rule.
//
// Example:
//
//	result := result.ValidateRules(
//	    email,
//	    func(e Email) error {
//	        if isBlacklisted(e) { return ErrBlacklisted }
//	        return nil
//	    },
//	    func(e Email) error {
//	        if isDisposable(e) { return ErrDisposable }
//	        return nil
//	    },
//	)
func ValidateRules[T any](value T, rules ...func(T) error) Result[T] {
	for _, rule := range rules {
		if err := rule(value); err != nil {
			return Err[T](err)
		}
	}
	return Ok(value)
}

// ValidateAllRules applies multiple validation rules and collects ALL errors.
// Returns Ok if all rules pass, or Err with MultiError containing all failures.
//
// Example:
//
//	result := result.ValidateAllRules(
//	    email,
//	    func(e Email) error {
//	        if isBlacklisted(e) { return ErrBlacklisted }
//	        return nil
//	    },
//	    func(e Email) error {
//	        if isDisposable(e) { return ErrDisposable }
//	        return nil
//	    },
//	)
//	// Both errors will be collected if both rules fail
func ValidateAllRules[T any](value T, rules ...func(T) error) Result[T] {
	var errors []error

	for _, rule := range rules {
		if err := rule(value); err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return Err[T](MultiError{Errors: errors})
	}

	return Ok(value)
}

// ValidateField wraps a field validation with field name context.
// Useful for building structured validation error messages.
//
// Example:
//
//	emailResult := result.ValidateField("email", domain.NewEmail(cmd.Email))
//	usernameResult := result.ValidateField("username", domain.NewUsername(cmd.Username))
func ValidateField[T any](fieldName string, r Result[T]) Result[T] {
	if r.IsErr() {
		return Err[T](Wrap(r.UnwrapErr(), "field."+fieldName))
	}
	return r
}

// ValidateStruct validates multiple fields of a struct and collects all errors.
// Accepts any type that implements the Validator interface (which all Result[T] do).
// Returns Ok if all validations pass, or Err with MultiError.
//
// Example:
//
//	result.ValidateStruct(
//	    result.ValidateField("email", domain.NewEmail(input.Email)),
//	    result.ValidateField("username", domain.NewUsername(input.Username)),
//	    result.ValidateField("password", domain.NewPassword(input.Password)),
//	)
func ValidateStruct(results ...Validator) Result[struct{}] {
	var errors []error

	for _, r := range results {
		if r.IsErr() {
			errors = append(errors, r.UnwrapErr())
		}
	}

	if len(errors) > 0 {
		return Err[struct{}](MultiError{Errors: errors})
	}

	return Ok(struct{}{})
}

// ValidateMap validates all values in a map and collects errors by key.
// Returns Ok if all validations pass, or Err with map of key -> error.
//
// Example:
//
//	fields := map[string]string{
//	    "email": "test@example.com",
//	    "username": "testuser",
//	}
//	result := result.ValidateMap(fields, func(key, value string) error {
//	    if value == "" {
//	        return fmt.Errorf("%s cannot be empty", key)
//	    }
//	    return nil
//	})
func ValidateMap[K comparable, V any](
	items map[K]V,
	validator func(K, V) error,
) Result[map[K]V] {
	errorMap := make(map[K]error)

	for key, value := range items {
		if err := validator(key, value); err != nil {
			errorMap[key] = err
		}
	}

	if len(errorMap) > 0 {
		// Convert map errors to MultiError
		var errors []error
		for key, err := range errorMap {
			errors = append(errors, fmt.Errorf("%v: %w", key, err))
		}
		return Err[map[K]V](MultiError{Errors: errors})
	}

	return Ok(items)
}

// ValidateSlice validates all elements in a slice and collects all errors.
// Returns Ok if all validations pass, or Err with MultiError.
//
// Example:
//
//	emails := []string{"test1@example.com", "invalid", "test2@example.com"}
//	result := result.ValidateSlice(emails, func(email string) error {
//	    _, err := result.Extract(domain.NewEmail(email), "validate")
//	    return err
//	})
func ValidateSlice[T any](items []T, validator func(T) error) Result[[]T] {
	var errors []error

	for i, item := range items {
		if err := validator(item); err != nil {
			errors = append(errors, fmt.Errorf("index %d: %w", i, err))
		}
	}

	if len(errors) > 0 {
		return Err[[]T](MultiError{Errors: errors})
	}

	return Ok(items)
}

// ValidateSliceResults validates a slice of Results and collects all errors.
// This is like ValidateAll but specifically for slices.
//
// Example:
//
//	emailResults := []result.Result[Email]{
//	    domain.NewEmail("test1@example.com"),
//	    domain.NewEmail("invalid"),
//	    domain.NewEmail("test2@example.com"),
//	}
//	result := result.ValidateSliceResults(emailResults)
func ValidateSliceResults[T any](results []Result[T]) Result[[]T] {
	return ValidateAll(results...)
}
