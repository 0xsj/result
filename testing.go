package result

import (
	"fmt"
	"strings"
	"testing"
)

// AssertOk fails the test if Result is Err and returns the unwrapped value.
// This is the most common testing helper - use it when you expect success.
//
// Example:
//
//	user := result.AssertOk(t, repo.FindByID(ctx, id))
//	assert.Equal(t, "test@example.com", user.Email)
func AssertOk[T any](t testing.TB, r Result[T]) T {
	t.Helper()
	if r.IsErr() {
		t.Fatalf("Expected Ok, got Err: %v", r.UnwrapErr())
	}
	return r.Unwrap()
}

// AssertErr fails the test if Result is Ok and returns the unwrapped error.
// Use this when you expect an operation to fail.
//
// Example:
//
//	err := result.AssertErr(t, repo.FindByID(ctx, "invalid-id"))
//	assert.Contains(t, err.Error(), "not found")
func AssertErr[T any](t testing.TB, r Result[T]) error {
	t.Helper()
	if r.IsOk() {
		t.Fatalf("Expected Err, got Ok: %v", r.Unwrap())
	}
	return r.UnwrapErr()
}

// AssertErrKind fails the test if Result is not Err with the expected Kind.
// Use this to verify specific error categories.
//
// Example:
//
//	result.AssertErrKind(t, repo.FindByID(ctx, "invalid-id"), result.KindNotFound)
func AssertErrKind[T any](t testing.TB, r Result[T], expected Kind) {
	t.Helper()
	if r.IsOk() {
		t.Fatalf("Expected Err with kind %v, got Ok: %v", expected, r.Unwrap())
	}
	actual := KindOf(r.UnwrapErr())
	if actual != expected {
		t.Fatalf("Expected error kind %v, got %v: %v", expected, actual, r.UnwrapErr())
	}
}

// AssertErrContains fails the test if the error message doesn't contain the substring.
// Use this for checking specific error messages.
//
// Example:
//
//	result.AssertErrContains(t, domain.NewEmail(""), "cannot be empty")
func AssertErrContains[T any](t testing.TB, r Result[T], substr string) {
	t.Helper()
	err := AssertErr(t, r)
	if !strings.Contains(err.Error(), substr) {
		t.Fatalf("Expected error to contain %q, got: %v", substr, err)
	}
}

// AssertErrOp fails the test if the error doesn't have the expected operation.
// Use this to verify error context propagation.
//
// Example:
//
//	result.AssertErrOp(t, handler.Handle(ctx, cmd), "commands.create_user")
func AssertErrOp[T any](t testing.TB, r Result[T], expectedOp string) {
	t.Helper()
	err := AssertErr(t, r)
	actualOp := OpOf(err)
	if !strings.Contains(actualOp, expectedOp) {
		t.Fatalf("Expected error op to contain %q, got: %q", expectedOp, actualOp)
	}
}

// OkOrFatal is like AssertOk but uses t.Fatal instead of t.Fatalf.
// Use this when you want a simpler error message.
//
// Example:
//
//	user := result.OkOrFatal(t, repo.FindByID(ctx, id))
func OkOrFatal[T any](t testing.TB, r Result[T]) T {
	t.Helper()
	if r.IsErr() {
		t.Fatal(r.UnwrapErr())
	}
	return r.Unwrap()
}

// RequireOk is an alias for AssertOk (for compatibility with testify naming).
func RequireOk[T any](t testing.TB, r Result[T]) T {
	return AssertOk(t, r)
}

// RequireErr is an alias for AssertErr (for compatibility with testify naming).
func RequireErr[T any](t testing.TB, r Result[T]) error {
	return AssertErr(t, r)
}

// AssertOkWith checks that Result is Ok and the value satisfies a predicate.
//
// Example:
//
//	result.AssertOkWith(t, repo.FindByID(ctx, id), func(u *User) bool {
//	    return u.IsActive
//	}, "expected user to be active")
func AssertOkWith[T any](t testing.TB, r Result[T], predicate func(T) bool, msg string) T {
	t.Helper()
	value := AssertOk(t, r)
	if !predicate(value) {
		t.Fatalf("Predicate failed: %s (value: %v)", msg, value)
	}
	return value
}

// AssertErrWith checks that Result is Err and the error satisfies a predicate.
//
// Example:
//
//	result.AssertErrWith(t, domain.NewEmail(""), func(err error) bool {
//	    return result.KindOf(err) == result.KindValidation
//	}, "expected validation error")
func AssertErrWith[T any](t testing.TB, r Result[T], predicate func(error) bool, msg string) error {
	t.Helper()
	err := AssertErr(t, r)
	if !predicate(err) {
		t.Fatalf("Predicate failed: %s (error: %v)", msg, err)
	}
	return err
}

// ExpectOk is like AssertOk but doesn't fail the test - just logs an error.
// Use this for non-critical assertions.
//
// Example:
//
//	if user, ok := result.ExpectOk(t, repo.FindByID(ctx, id)); ok {
//	    // Continue with test
//	}
func ExpectOk[T any](t testing.TB, r Result[T]) (T, bool) {
	t.Helper()
	if r.IsErr() {
		t.Errorf("Expected Ok, got Err: %v", r.UnwrapErr())
		var zero T
		return zero, false
	}
	return r.Unwrap(), true
}

// ExpectErr is like AssertErr but doesn't fail the test - just logs an error.
//
// Example:
//
//	if err, ok := result.ExpectErr(t, repo.FindByID(ctx, "invalid")); ok {
//	    // Continue with test
//	}
func ExpectErr[T any](t testing.TB, r Result[T]) (error, bool) {
	t.Helper()
	if r.IsOk() {
		t.Errorf("Expected Err, got Ok: %v", r.Unwrap())
		return nil, false
	}
	return r.UnwrapErr(), true
}

// AssertPanics verifies that a function panics when called.
// Use this to test Must() and similar panic-inducing functions.
//
// Example:
//
//	result.AssertPanics(t, func() {
//	    result.Must(result.Err[int](errors.New("fail")), "should panic")
//	})
func AssertPanics(t testing.TB, fn func()) {
	t.Helper()
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("Expected function to panic, but it didn't")
		}
	}()
	fn()
}

// AssertNotPanics verifies that a function does not panic.
//
// Example:
//
//	result.AssertNotPanics(t, func() {
//	    result.Must(result.Ok(42), "should not panic")
//	})
func AssertNotPanics(t testing.TB, fn func()) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Expected function not to panic, but it panicked: %v", r)
		}
	}()
	fn()
}

// MockResult creates a mock Result for testing.
// Use this to simulate different Result scenarios in tests.
//
// Example:
//
//	mockRepo := &MockRepo{
//	    FindByIDFunc: func(id string) result.Result[*User] {
//	        return result.MockResult[*User](result.KindNotFound, "user not found")
//	    },
//	}
func MockResult[T any](kind Kind, message string) Result[T] {
	var err error
	switch kind {
	case KindValidation:
		err = Validation("mock.operation", message, nil)
	case KindNotFound:
		err = NotFound("mock.operation", message)
	case KindConflict:
		err = Conflict("mock.operation", message)
	case KindUnauthorized:
		err = Unauthorized("mock.operation", message)
	case KindForbidden:
		err = Forbidden("mock.operation", message)
	case KindDomain:
		err = Domain("mock.operation", message)
	case KindInfrastructure:
		err = Infrastructure("mock.operation", fmt.Errorf("%s", message))
	case KindInternal:
		err = Internal("mock.operation", fmt.Errorf("%s", message))
	default:
		err = Internal("mock.operation", fmt.Errorf("%s", message))
	}
	return Err[T](err)
}
