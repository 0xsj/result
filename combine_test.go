package result

import (
	"errors"
	"fmt"
	"testing"
)

func TestAll2(t *testing.T) {
	t.Run("combines two Ok results", func(t *testing.T) {
		r1 := Ok(42)
		r2 := Ok("hello")

		result := All2(r1, r2)

		if !result.IsOk() {
			t.Error("All2 should succeed when both results are Ok")
		}

		combined := result.Unwrap()
		if combined.V1 != 42 {
			t.Errorf("V1 = %d, want 42", combined.V1)
		}
		if combined.V2 != "hello" {
			t.Errorf("V2 = %s, want hello", combined.V2)
		}
	})

	t.Run("fails when first result is Err", func(t *testing.T) {
		err1 := errors.New("error 1")
		r1 := Err[int](err1)
		r2 := Ok("hello")

		result := All2(r1, r2)

		if !result.IsErr() {
			t.Error("All2 should fail when first result is Err")
		}
		if result.UnwrapErr() != err1 {
			t.Errorf("error = %v, want %v", result.UnwrapErr(), err1)
		}
	})

	t.Run("fails when second result is Err", func(t *testing.T) {
		err2 := errors.New("error 2")
		r1 := Ok(42)
		r2 := Err[string](err2)

		result := All2(r1, r2)

		if !result.IsErr() {
			t.Error("All2 should fail when second result is Err")
		}
		if result.UnwrapErr() != err2 {
			t.Errorf("error = %v, want %v", result.UnwrapErr(), err2)
		}
	})

	t.Run("returns first error when both fail", func(t *testing.T) {
		err1 := errors.New("error 1")
		err2 := errors.New("error 2")
		r1 := Err[int](err1)
		r2 := Err[string](err2)

		result := All2(r1, r2)

		if result.UnwrapErr() != err1 {
			t.Errorf("error = %v, want %v (first error)", result.UnwrapErr(), err1)
		}
	})

	t.Run("preserves metadata from failing result", func(t *testing.T) {
		err1 := errors.New("error 1")
		r1 := Err[int](err1).WithOp("operation1").WithMeta("key", "value")
		r2 := Ok("hello")

		result := All2(r1, r2)

		if result.op != "operation1" {
			t.Errorf("op = %s, want operation1", result.op)
		}
		if result.meta["key"] != "value" {
			t.Error("metadata not preserved")
		}
	})
}

func TestAll3(t *testing.T) {
	t.Run("combines three Ok results", func(t *testing.T) {
		r1 := Ok(42)
		r2 := Ok("hello")
		r3 := Ok(true)

		result := All3(r1, r2, r3)

		if !result.IsOk() {
			t.Error("All3 should succeed when all results are Ok")
		}

		combined := result.Unwrap()
		if combined.V1 != 42 {
			t.Errorf("V1 = %d, want 42", combined.V1)
		}
		if combined.V2 != "hello" {
			t.Errorf("V2 = %s, want hello", combined.V2)
		}
		if combined.V3 != true {
			t.Errorf("V3 = %v, want true", combined.V3)
		}
	})

	t.Run("fails when any result is Err", func(t *testing.T) {
		err := errors.New("error")
		r1 := Ok(42)
		r2 := Err[string](err)
		r3 := Ok(true)

		result := All3(r1, r2, r3)

		if !result.IsErr() {
			t.Error("All3 should fail when any result is Err")
		}
		if result.UnwrapErr() != err {
			t.Errorf("error = %v, want %v", result.UnwrapErr(), err)
		}
	})
}

func TestAll4(t *testing.T) {
	t.Run("combines four Ok results", func(t *testing.T) {
		r1 := Ok(1)
		r2 := Ok(2)
		r3 := Ok(3)
		r4 := Ok(4)

		result := All4(r1, r2, r3, r4)

		if !result.IsOk() {
			t.Error("All4 should succeed when all results are Ok")
		}

		combined := result.Unwrap()
		if combined.V1 != 1 || combined.V2 != 2 || combined.V3 != 3 || combined.V4 != 4 {
			t.Error("All4 did not combine values correctly")
		}
	})
}

func TestAll5(t *testing.T) {
	t.Run("combines five Ok results", func(t *testing.T) {
		r1 := Ok(1)
		r2 := Ok(2)
		r3 := Ok(3)
		r4 := Ok(4)
		r5 := Ok(5)

		result := All5(r1, r2, r3, r4, r5)

		if !result.IsOk() {
			t.Error("All5 should succeed when all results are Ok")
		}

		combined := result.Unwrap()
		if combined.V1 != 1 || combined.V2 != 2 || combined.V3 != 3 || combined.V4 != 4 || combined.V5 != 5 {
			t.Error("All5 did not combine values correctly")
		}
	})
}

func TestZip2(t *testing.T) {
	t.Run("combines and transforms two Ok results", func(t *testing.T) {
		r1 := Ok(10)
		r2 := Ok(20)

		result := Zip2(r1, r2, func(a, b int) int {
			return a + b
		})

		if !result.IsOk() {
			t.Error("Zip2 should succeed when both results are Ok")
		}
		if result.Unwrap() != 30 {
			t.Errorf("result = %d, want 30", result.Unwrap())
		}
	})

	t.Run("fails when any result is Err", func(t *testing.T) {
		err := errors.New("error")
		r1 := Ok(10)
		r2 := Err[int](err)

		result := Zip2(r1, r2, func(a, b int) int {
			return a + b
		})

		if !result.IsErr() {
			t.Error("Zip2 should fail when any result is Err")
		}
		if result.UnwrapErr() != err {
			t.Errorf("error = %v, want %v", result.UnwrapErr(), err)
		}
	})

	t.Run("transforms to different type", func(t *testing.T) {
		r1 := Ok(42)
		r2 := Ok("answer")

		result := Zip2(r1, r2, func(n int, s string) string {
			return fmt.Sprintf("%s: %d", s, n)
		})

		if !result.IsOk() {
			t.Error("Zip2 should succeed")
		}
		if result.Unwrap() != "answer: 42" {
			t.Errorf("result = %s, want 'answer: 42'", result.Unwrap())
		}
	})
}

func TestZip3(t *testing.T) {
	t.Run("combines and transforms three Ok results", func(t *testing.T) {
		r1 := Ok(1)
		r2 := Ok(2)
		r3 := Ok(3)

		result := Zip3(r1, r2, r3, func(a, b, c int) int {
			return a + b + c
		})

		if !result.IsOk() {
			t.Error("Zip3 should succeed when all results are Ok")
		}
		if result.Unwrap() != 6 {
			t.Errorf("result = %d, want 6", result.Unwrap())
		}
	})
}

func TestZip4(t *testing.T) {
	t.Run("combines and transforms four Ok results", func(t *testing.T) {
		r1 := Ok(1)
		r2 := Ok(2)
		r3 := Ok(3)
		r4 := Ok(4)

		result := Zip4(r1, r2, r3, r4, func(a, b, c, d int) int {
			return a + b + c + d
		})

		if !result.IsOk() {
			t.Error("Zip4 should succeed when all results are Ok")
		}
		if result.Unwrap() != 10 {
			t.Errorf("result = %d, want 10", result.Unwrap())
		}
	})
}

func TestZip5(t *testing.T) {
	t.Run("combines and transforms five Ok results", func(t *testing.T) {
		r1 := Ok(1)
		r2 := Ok(2)
		r3 := Ok(3)
		r4 := Ok(4)
		r5 := Ok(5)

		result := Zip5(r1, r2, r3, r4, r5, func(a, b, c, d, e int) int {
			return a + b + c + d + e
		})

		if !result.IsOk() {
			t.Error("Zip5 should succeed when all results are Ok")
		}
		if result.Unwrap() != 15 {
			t.Errorf("result = %d, want 15", result.Unwrap())
		}
	})
}
