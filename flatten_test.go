package result

import (
	"errors"
	"fmt"
	"testing"
)

func TestFlatten(t *testing.T) {
	t.Run("flattens Ok(Ok(value))", func(t *testing.T) {
		inner := Ok(42)
		outer := Ok(inner)

		result := Flatten(outer)

		if !result.IsOk() {
			t.Error("Flatten should succeed for Ok(Ok(value))")
		}
		if result.Unwrap() != 42 {
			t.Errorf("value = %d, want 42", result.Unwrap())
		}
	})

	t.Run("propagates outer error", func(t *testing.T) {
		outerErr := errors.New("outer error")
		outer := Err[Result[int]](outerErr)

		result := Flatten(outer)

		if !result.IsErr() {
			t.Error("Flatten should propagate outer error")
		}
		if result.UnwrapErr() != outerErr {
			t.Errorf("error = %v, want %v", result.UnwrapErr(), outerErr)
		}
	})

	t.Run("flattens Ok(Err(error))", func(t *testing.T) {
		innerErr := errors.New("inner error")
		inner := Err[int](innerErr)
		outer := Ok(inner)

		result := Flatten(outer)

		if !result.IsErr() {
			t.Error("Flatten should propagate inner error")
		}
		if result.UnwrapErr() != innerErr {
			t.Errorf("error = %v, want %v", result.UnwrapErr(), innerErr)
		}
	})

	t.Run("preserves outer metadata when outer fails", func(t *testing.T) {
		err := errors.New("error")
		outer := Err[Result[int]](err).
			WithOp("outer").
			WithMeta("key", "value")

		result := Flatten(outer)

		if result.op != "outer" {
			t.Errorf("op = %s, want outer", result.op)
		}
		if result.meta["key"] != "value" {
			t.Error("metadata not preserved")
		}
	})

	t.Run("preserves inner metadata when inner fails", func(t *testing.T) {
		err := errors.New("error")
		inner := Err[int](err).WithOp("inner").WithMeta("key", "inner_value")
		outer := Ok(inner)

		result := Flatten(outer)

		if result.op != "inner" {
			t.Errorf("op = %s, want inner", result.op)
		}
		if result.meta["key"] != "inner_value" {
			t.Error("inner metadata not preserved")
		}
	})

	t.Run("outer metadata takes precedence when inner has no metadata", func(t *testing.T) {
		inner := Ok(42)
		outer := Ok(inner).WithOp("outer").WithMeta("key", "outer_value")

		result := Flatten(outer)

		if result.op != "outer" {
			t.Errorf("op = %s, want outer", result.op)
		}
		if result.meta["key"] != "outer_value" {
			t.Error("outer metadata should be preserved")
		}
	})
}

func TestFlatMap(t *testing.T) {
	t.Run("is equivalent to AndThenMap", func(t *testing.T) {
		r := Ok(42)

		result := FlatMap(r, func(n int) Result[string] {
			return Ok(fmt.Sprintf("number: %d", n))
		})

		if !result.IsOk() {
			t.Error("FlatMap should succeed")
		}
		if result.Unwrap() != "number: 42" {
			t.Errorf("result = %s, want 'number: 42'", result.Unwrap())
		}
	})
}
