package result

import (
	"errors"
	"testing"
)

func TestTry(t *testing.T) {
	t.Run("captures successful execution", func(t *testing.T) {
		result := Try(func() int {
			return 42
		})

		if !result.IsOk() {
			t.Error("Try should succeed for non-panicking function")
		}
		if result.Unwrap() != 42 {
			t.Errorf("value = %d, want 42", result.Unwrap())
		}
	})

	t.Run("captures panic as error", func(t *testing.T) {
		result := Try(func() int {
			panic("something went wrong")
		})

		if !result.IsErr() {
			t.Error("Try should capture panic as error")
		}

		err := result.UnwrapErr()
		if err == nil {
			t.Fatal("error should not be nil")
		}
		if !contains(err.Error(), "panic") {
			t.Errorf("error should mention panic, got: %v", err)
		}
	})

	t.Run("captures panic with error value", func(t *testing.T) {
		result := Try(func() string {
			panic(errors.New("test error"))
		})

		if !result.IsErr() {
			t.Error("Try should capture panic")
		}
	})

	t.Run("handles nil panic", func(t *testing.T) {
		result := Try(func() int {
			panic(nil)
		})

		if !result.IsErr() {
			t.Error("Try should capture nil panic")
		}
	})
}

func TestTryWith(t *testing.T) {
	t.Run("converts successful (T, error) to Ok", func(t *testing.T) {
		result := TryWith(func() (int, error) {
			return 42, nil
		})

		if !result.IsOk() {
			t.Error("TryWith should succeed when error is nil")
		}
		if result.Unwrap() != 42 {
			t.Errorf("value = %d, want 42", result.Unwrap())
		}
	})

	t.Run("converts (T, error) with error to Err", func(t *testing.T) {
		expectedErr := errors.New("test error")
		result := TryWith(func() (int, error) {
			return 0, expectedErr
		})

		if !result.IsErr() {
			t.Error("TryWith should fail when error is non-nil")
		}
		if result.UnwrapErr() != expectedErr {
			t.Errorf("error = %v, want %v", result.UnwrapErr(), expectedErr)
		}
	})

	t.Run("is equivalent to From", func(t *testing.T) {
		fn := func() (string, error) {
			return "hello", nil
		}

		r1 := TryWith(fn)
		r2 := From(fn())

		if r1.IsErr() != r2.IsErr() {
			t.Error("TryWith and From should have same result")
		}
		if r1.IsOk() && r1.Unwrap() != r2.Unwrap() {
			t.Error("TryWith and From should have same value")
		}
	})
}

// Helper function for tests
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr))
}
