package result

import (
	"errors"
	"fmt"
	"testing"
)

// ============================================================================
// Constructor Tests
// ============================================================================

func TestOk(t *testing.T) {
	r := Ok(42)

	if !r.IsOk() {
		t.Error("Ok() should create an Ok result")
	}

	if r.IsErr() {
		t.Error("Ok() result should not be Err")
	}

	if r.value != 42 {
		t.Errorf("Ok() value = %v, want 42", r.value)
	}

	if r.err != nil {
		t.Errorf("Ok() err = %v, want nil", r.err)
	}
}

func TestErr(t *testing.T) {
	testErr := errors.New("test error")
	r := Err[int](testErr)

	if r.IsOk() {
		t.Error("Err() result should not be Ok")
	}

	if !r.IsErr() {
		t.Error("Err() should create an Err result")
	}

	if r.err != testErr {
		t.Errorf("Err() err = %v, want %v", r.err, testErr)
	}
}

func TestFrom(t *testing.T) {
	t.Run("with nil error creates Ok", func(t *testing.T) {
		r := From(42, nil)

		if !r.IsOk() {
			t.Error("From(value, nil) should create Ok result")
		}

		if r.value != 42 {
			t.Errorf("From() value = %v, want 42", r.value)
		}
	})

	t.Run("with error creates Err", func(t *testing.T) {
		testErr := errors.New("test error")
		r := From(0, testErr)

		if !r.IsErr() {
			t.Error("From(value, err) should create Err result")
		}

		if r.err != testErr {
			t.Errorf("From() err = %v, want %v", r.err, testErr)
		}
	})
}

// ============================================================================
// Predicate Tests
// ============================================================================

func TestIsOk(t *testing.T) {
	tests := []struct {
		name     string
		result   Result[int]
		expected bool
	}{
		{
			name:     "Ok result",
			result:   Ok(42),
			expected: true,
		},
		{
			name:     "Err result",
			result:   Err[int](errors.New("error")),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.result.IsOk()
			if got != tt.expected {
				t.Errorf("IsOk() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIsErr(t *testing.T) {
	tests := []struct {
		name     string
		result   Result[int]
		expected bool
	}{
		{
			name:     "Ok result",
			result:   Ok(42),
			expected: false,
		},
		{
			name:     "Err result",
			result:   Err[int](errors.New("error")),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.result.IsErr()
			if got != tt.expected {
				t.Errorf("IsErr() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIsErrKind(t *testing.T) {
	tests := []struct {
		name     string
		result   Result[int]
		kind     Kind
		expected bool
	}{
		{
			name:     "Ok result",
			result:   Ok(42),
			kind:     KindNotFound,
			expected: false,
		},
		{
			name:     "Err with matching kind",
			result:   Err[int](NotFound("test", "resource")),
			kind:     KindNotFound,
			expected: true,
		},
		{
			name:     "Err with non-matching kind",
			result:   Err[int](NotFound("test", "resource")),
			kind:     KindValidation,
			expected: false,
		},
		{
			name:     "Err with standard error",
			result:   Err[int](errors.New("standard error")),
			kind:     KindInternal,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.result.IsErrKind(tt.kind)
			if got != tt.expected {
				t.Errorf("IsErrKind() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// ============================================================================
// Extractor Tests
// ============================================================================

func TestUnwrap(t *testing.T) {
	t.Run("Ok result returns value", func(t *testing.T) {
		r := Ok(42)
		value := r.Unwrap()

		if value != 42 {
			t.Errorf("Unwrap() = %v, want 42", value)
		}
	})

	t.Run("Err result panics", func(t *testing.T) {
		r := Err[int](errors.New("test error"))

		defer func() {
			if r := recover(); r == nil {
				t.Error("Unwrap() should panic on Err result")
			}
		}()

		_ = r.Unwrap()
	})
}

func TestUnwrapOr(t *testing.T) {
	t.Run("Ok result returns value", func(t *testing.T) {
		r := Ok(42)
		value := r.UnwrapOr(0)

		if value != 42 {
			t.Errorf("UnwrapOr() = %v, want 42", value)
		}
	})

	t.Run("Err result returns default", func(t *testing.T) {
		r := Err[int](errors.New("test error"))
		value := r.UnwrapOr(99)

		if value != 99 {
			t.Errorf("UnwrapOr() = %v, want 99", value)
		}
	})
}

func TestUnwrapOrElse(t *testing.T) {
	t.Run("Ok result returns value", func(t *testing.T) {
		r := Ok(42)
		value := r.UnwrapOrElse(func(err error) int {
			return -1
		})

		if value != 42 {
			t.Errorf("UnwrapOrElse() = %v, want 42", value)
		}
	})

	t.Run("Err result computes default", func(t *testing.T) {
		testErr := errors.New("test error")
		r := Err[int](testErr)

		var capturedErr error
		value := r.UnwrapOrElse(func(err error) int {
			capturedErr = err
			return 99
		})

		if value != 99 {
			t.Errorf("UnwrapOrElse() = %v, want 99", value)
		}

		if capturedErr != testErr {
			t.Errorf("UnwrapOrElse() error = %v, want %v", capturedErr, testErr)
		}
	})
}

func TestExpect(t *testing.T) {
	t.Run("Ok result returns value", func(t *testing.T) {
		r := Ok(42)
		value := r.Expect("should have value")

		if value != 42 {
			t.Errorf("Expect() = %v, want 42", value)
		}
	})

	t.Run("Err result panics with custom message", func(t *testing.T) {
		r := Err[int](errors.New("test error"))

		defer func() {
			if r := recover(); r == nil {
				t.Error("Expect() should panic on Err result")
			} else {
				msg := fmt.Sprintf("%v", r)
				if msg != "custom panic message: test error" {
					t.Errorf("Expect() panic message = %v, want custom message", msg)
				}
			}
		}()

		_ = r.Expect("custom panic message")
	})
}

func TestValue(t *testing.T) {
	t.Run("Ok result", func(t *testing.T) {
		r := Ok(42)
		value, err := r.Value()

		if value != 42 {
			t.Errorf("Value() value = %v, want 42", value)
		}

		if err != nil {
			t.Errorf("Value() err = %v, want nil", err)
		}
	})

	t.Run("Err result", func(t *testing.T) {
		testErr := errors.New("test error")
		r := Err[int](testErr)
		value, err := r.Value()

		if value != 0 {
			t.Errorf("Value() value = %v, want zero value", value)
		}

		if err != testErr {
			t.Errorf("Value() err = %v, want %v", err, testErr)
		}
	})
}

func TestError(t *testing.T) {
	t.Run("Ok result returns nil", func(t *testing.T) {
		r := Ok(42)
		err := r.Error()
		if err != nil {
			t.Errorf("Error() = %v, want nil", err)
		}
	})

	t.Run("Err result returns error", func(t *testing.T) {
		testErr := errors.New("test error")
		r := Err[int](testErr)
		err := r.Error()

		if err != testErr {
			t.Errorf("Error() = %v, want %v", err, testErr)
		}
	})
}

// ============================================================================
// Context Builder Tests
// ============================================================================

func TestWithOp(t *testing.T) {
	r := Ok(42).WithOp("Service.GetUser")

	if r.op != "Service.GetUser" {
		t.Errorf("WithOp() op = %v, want Service.GetUser", r.op)
	}

	// Value should be preserved
	if r.value != 42 {
		t.Errorf("WithOp() value = %v, want 42", r.value)
	}
}

func TestWithMeta(t *testing.T) {
	r := Ok(42).
		WithMeta("user_id", "123").
		WithMeta("request_id", "req_abc")

	if r.meta == nil {
		t.Fatal("WithMeta() should initialize meta map")
	}

	if r.meta["user_id"] != "123" {
		t.Errorf("WithMeta() meta[user_id] = %v, want 123", r.meta["user_id"])
	}

	if r.meta["request_id"] != "req_abc" {
		t.Errorf("WithMeta() meta[request_id] = %v, want req_abc", r.meta["request_id"])
	}

	// Value should be preserved
	if r.value != 42 {
		t.Errorf("WithMeta() value = %v, want 42", r.value)
	}
}

func TestWithMetaMap(t *testing.T) {
	meta := map[string]any{
		"user_id":    "123",
		"request_id": "req_abc",
	}

	r := Ok(42).WithMetaMap(meta)

	if r.meta == nil {
		t.Fatal("WithMetaMap() should initialize meta map")
	}

	if r.meta["user_id"] != "123" {
		t.Errorf("WithMetaMap() meta[user_id] = %v, want 123", r.meta["user_id"])
	}

	if r.meta["request_id"] != "req_abc" {
		t.Errorf("WithMetaMap() meta[request_id] = %v, want req_abc", r.meta["request_id"])
	}

	// Value should be preserved
	if r.value != 42 {
		t.Errorf("WithMetaMap() value = %v, want 42", r.value)
	}
}

func TestWithMetaMap_PreservesExisting(t *testing.T) {
	r := Ok(42).
		WithMeta("key1", "value1").
		WithMetaMap(map[string]any{
			"key2": "value2",
			"key3": "value3",
		})

	if r.meta["key1"] != "value1" {
		t.Error("WithMetaMap() should preserve existing metadata")
	}

	if r.meta["key2"] != "value2" {
		t.Error("WithMetaMap() should add new metadata")
	}
}

// ============================================================================
// Transformer Tests
// ============================================================================

func TestMap(t *testing.T) {
	t.Run("Ok result transforms value", func(t *testing.T) {
		r := Ok(21).Map(func(n int) int {
			return n * 2
		})

		if !r.IsOk() {
			t.Error("Map() should preserve Ok status")
		}

		if r.value != 42 {
			t.Errorf("Map() value = %v, want 42", r.value)
		}
	})

	t.Run("Err result unchanged", func(t *testing.T) {
		testErr := errors.New("test error")
		r := Err[int](testErr).Map(func(n int) int {
			return n * 2
		})

		if !r.IsErr() {
			t.Error("Map() should preserve Err status")
		}

		if r.err != testErr {
			t.Errorf("Map() err = %v, want %v", r.err, testErr)
		}
	})

	t.Run("preserves context", func(t *testing.T) {
		r := Ok(21).
			WithOp("test").
			WithMeta("key", "value").
			Map(func(n int) int {
				return n * 2
			})

		if r.op != "test" {
			t.Error("Map() should preserve op")
		}

		if r.meta["key"] != "value" {
			t.Error("Map() should preserve meta")
		}
	})
}

func TestMapErr(t *testing.T) {
	t.Run("Err result transforms error", func(t *testing.T) {
		originalErr := errors.New("original")
		r := Err[int](originalErr).MapErr(func(err error) error {
			return fmt.Errorf("wrapped: %w", err)
		})

		if !r.IsErr() {
			t.Error("MapErr() should preserve Err status")
		}

		expectedMsg := "wrapped: original"
		if r.err.Error() != expectedMsg {
			t.Errorf("MapErr() err = %v, want %v", r.err.Error(), expectedMsg)
		}
	})

	t.Run("Ok result unchanged", func(t *testing.T) {
		r := Ok(42).MapErr(func(err error) error {
			return errors.New("should not be called")
		})

		if !r.IsOk() {
			t.Error("MapErr() should preserve Ok status")
		}

		if r.value != 42 {
			t.Errorf("MapErr() value = %v, want 42", r.value)
		}
	})

	t.Run("preserves context", func(t *testing.T) {
		r := Err[int](errors.New("original")).
			WithOp("test").
			WithMeta("key", "value").
			MapErr(func(err error) error {
				return fmt.Errorf("wrapped: %w", err)
			})

		if r.op != "test" {
			t.Error("MapErr() should preserve op")
		}

		if r.meta["key"] != "value" {
			t.Error("MapErr() should preserve meta")
		}
	})
}

func TestMapValue(t *testing.T) {
	t.Run("Ok result transforms to different type", func(t *testing.T) {
		r := MapValue(
			Ok(42),
			func(n int) string {
				return fmt.Sprintf("number: %d", n)
			},
		)

		if !r.IsOk() {
			t.Error("MapValue() should preserve Ok status")
		}

		expected := "number: 42"
		if r.value != expected {
			t.Errorf("MapValue() value = %v, want %v", r.value, expected)
		}
	})

	t.Run("Err result preserves error", func(t *testing.T) {
		testErr := errors.New("test error")
		r := MapValue(
			Err[int](testErr),
			func(n int) string {
				return fmt.Sprintf("number: %d", n)
			},
		)

		if !r.IsErr() {
			t.Error("MapValue() should preserve Err status")
		}

		if r.err != testErr {
			t.Errorf("MapValue() err = %v, want %v", r.err, testErr)
		}
	})

	t.Run("preserves context", func(t *testing.T) {
		r := MapValue(
			Ok(42).WithOp("test").WithMeta("key", "value"),
			func(n int) string {
				return fmt.Sprintf("%d", n)
			},
		)

		if r.op != "test" {
			t.Error("MapValue() should preserve op")
		}

		if r.meta["key"] != "value" {
			t.Error("MapValue() should preserve meta")
		}
	})
}

// ============================================================================
// Integration Tests
// ============================================================================

func TestResultChaining(t *testing.T) {
	// Test fluent API chaining
	r := Ok(42).
		WithOp("Service.Process").
		WithMeta("step", "1").
		Map(func(n int) int {
			return n * 2
		}).
		WithMeta("step", "2")

	if !r.IsOk() {
		t.Error("Chaining should preserve Ok status")
	}

	if r.value != 84 {
		t.Errorf("Chaining value = %v, want 84", r.value)
	}

	if r.op != "Service.Process" {
		t.Error("Chaining should preserve op")
	}

	if r.meta["step"] != "2" {
		t.Error("Chaining should update meta")
	}
}

func TestResultWithGoIdioms(t *testing.T) {
	// Simulate a function that returns (T, error)
	getUserByID := func(id int) (string, error) {
		if id == 0 {
			return "", errors.New("invalid id")
		}
		return fmt.Sprintf("user_%d", id), nil
	}

	t.Run("convert success to Result", func(t *testing.T) {
		name, err := getUserByID(42)
		r := From(name, err)

		if !r.IsOk() {
			t.Error("Should convert successful call to Ok")
		}

		if r.value != "user_42" {
			t.Errorf("Value = %v, want user_42", r.value)
		}
	})

	t.Run("convert error to Result", func(t *testing.T) {
		name, err := getUserByID(0)
		r := From(name, err)

		if !r.IsErr() {
			t.Error("Should convert error call to Err")
		}
	})

	t.Run("convert Result back to (T, error)", func(t *testing.T) {
		r := Ok("test")
		value, err := r.Value()
		if err != nil {
			t.Error("Value() should return nil error for Ok")
		}

		if value != "test" {
			t.Errorf("Value() = %v, want test", value)
		}
	})
}
