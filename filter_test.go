package result

import (
	"errors"
	"testing"
)

func TestFilter(t *testing.T) {
	t.Run("passes when predicate is true", func(t *testing.T) {
		result := Ok(42).Filter(func(n int) bool {
			return n > 0
		}, errors.New("number must be positive"))

		if !result.IsOk() {
			t.Error("Filter should pass when predicate returns true")
		}
		if result.Unwrap() != 42 {
			t.Errorf("value = %d, want 42", result.Unwrap())
		}
	})

	t.Run("fails when predicate is false", func(t *testing.T) {
		expectedErr := errors.New("number must be positive")
		result := Ok(-5).Filter(func(n int) bool {
			return n > 0
		}, expectedErr)

		if !result.IsErr() {
			t.Error("Filter should fail when predicate returns false")
		}
		if result.UnwrapErr() != expectedErr {
			t.Errorf("error = %v, want %v", result.UnwrapErr(), expectedErr)
		}
	})

	t.Run("propagates existing error", func(t *testing.T) {
		originalErr := errors.New("original error")
		result := Err[int](originalErr).Filter(func(n int) bool {
			return true
		}, errors.New("new error"))

		if !result.IsErr() {
			t.Error("Filter should propagate error")
		}
		if result.UnwrapErr() != originalErr {
			t.Errorf("error = %v, want %v", result.UnwrapErr(), originalErr)
		}
	})

	t.Run("preserves metadata", func(t *testing.T) {
		result := Ok(42).
			WithOp("test").
			WithMeta("key", "value").
			Filter(func(n int) bool {
				return n < 0
			}, errors.New("failed"))

		if result.op != "test" {
			t.Errorf("op = %s, want test", result.op)
		}
		if result.meta["key"] != "value" {
			t.Error("metadata not preserved")
		}
	})
}

func TestFilterNot(t *testing.T) {
	t.Run("passes when predicate is false", func(t *testing.T) {
		result := Ok(42).FilterNot(func(n int) bool {
			return n < 0
		}, errors.New("number must not be negative"))

		if !result.IsOk() {
			t.Error("FilterNot should pass when predicate returns false")
		}
	})

	t.Run("fails when predicate is true", func(t *testing.T) {
		result := Ok(-5).FilterNot(func(n int) bool {
			return n < 0
		}, errors.New("number must not be negative"))

		if !result.IsErr() {
			t.Error("FilterNot should fail when predicate returns true")
		}
	})
}

func TestWhen(t *testing.T) {
	t.Run("passes when condition is true", func(t *testing.T) {
		result := Ok(42).When(true, errors.New("condition failed"))

		if !result.IsOk() {
			t.Error("When should pass when condition is true")
		}
		if result.Unwrap() != 42 {
			t.Errorf("value = %d, want 42", result.Unwrap())
		}
	})

	t.Run("fails when condition is false", func(t *testing.T) {
		expectedErr := errors.New("condition failed")
		result := Ok(42).When(false, expectedErr)

		if !result.IsErr() {
			t.Error("When should fail when condition is false")
		}
		if result.UnwrapErr() != expectedErr {
			t.Errorf("error = %v, want %v", result.UnwrapErr(), expectedErr)
		}
	})

	t.Run("propagates existing error", func(t *testing.T) {
		originalErr := errors.New("original error")
		result := Err[int](originalErr).When(true, errors.New("new error"))

		if result.UnwrapErr() != originalErr {
			t.Errorf("error = %v, want %v", result.UnwrapErr(), originalErr)
		}
	})
}

func TestUnless(t *testing.T) {
	t.Run("passes when condition is false", func(t *testing.T) {
		result := Ok(42).Unless(false, errors.New("condition must be false"))

		if !result.IsOk() {
			t.Error("Unless should pass when condition is false")
		}
	})

	t.Run("fails when condition is true", func(t *testing.T) {
		result := Ok(42).Unless(true, errors.New("condition must be false"))

		if !result.IsErr() {
			t.Error("Unless should fail when condition is true")
		}
	})
}

func TestFilterChaining(t *testing.T) {
	t.Run("chains multiple filters successfully", func(t *testing.T) {
		result := Ok(42).
			Filter(func(n int) bool { return n > 0 }, errors.New("must be positive")).
			Filter(func(n int) bool { return n < 100 }, errors.New("must be less than 100")).
			Filter(func(n int) bool { return n%2 == 0 }, errors.New("must be even"))

		if !result.IsOk() {
			t.Errorf("chained filters should pass, got error: %v", result.UnwrapErr())
		}
	})

	t.Run("fails at first filter violation", func(t *testing.T) {
		err1 := errors.New("must be positive")
		err2 := errors.New("must be less than 100")

		result := Ok(-5).
			Filter(func(n int) bool { return n > 0 }, err1).
			Filter(func(n int) bool { return n < 100 }, err2)

		if !result.IsErr() {
			t.Error("should fail at first filter")
		}
		if result.UnwrapErr() != err1 {
			t.Errorf("error = %v, want %v", result.UnwrapErr(), err1)
		}
	})
}
