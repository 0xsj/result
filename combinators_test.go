package result

import (
	"errors"
	"fmt"
	"testing"
)

// ============================================================================
// AndThen Tests
// ============================================================================

func TestAndThen_Ok(t *testing.T) {
	// Simple chain of Ok results
	r := Ok(10).
		AndThen(func(n int) Result[int] {
			return Ok(n * 2)
		}).
		AndThen(func(n int) Result[int] {
			return Ok(n + 5)
		})

	if !r.IsOk() {
		t.Error("AndThen should preserve Ok through chain")
	}

	// 10 * 2 = 20, 20 + 5 = 25
	if r.value != 25 {
		t.Errorf("AndThen value = %v, want 25", r.value)
	}
}

func TestAndThen_ErrShortCircuits(t *testing.T) {
	called := false

	r := Ok(10).
		AndThen(func(n int) Result[int] {
			return Err[int](errors.New("step 1 failed"))
		}).
		AndThen(func(n int) Result[int] {
			called = true
			return Ok(n * 2)
		})

	if !r.IsErr() {
		t.Error("AndThen should propagate error")
	}

	if called {
		t.Error("AndThen should short-circuit on error")
	}

	if r.err.Error() != "step 1 failed" {
		t.Errorf("AndThen err = %v, want 'step 1 failed'", r.err)
	}
}

func TestAndThen_InitialErr(t *testing.T) {
	testErr := errors.New("initial error")
	called := false

	r := Err[int](testErr).
		AndThen(func(n int) Result[int] {
			called = true
			return Ok(n * 2)
		})

	if !r.IsErr() {
		t.Error("AndThen should preserve initial error")
	}

	if called {
		t.Error("AndThen should not call function on initial error")
	}

	if r.err != testErr {
		t.Errorf("AndThen err = %v, want %v", r.err, testErr)
	}
}

func TestAndThen_PreservesContext(t *testing.T) {
	r := Ok(10).
		WithOp("Service.Process").
		WithMeta("step", "1").
		AndThen(func(n int) Result[int] {
			return Ok(n * 2) // No context set
		})

	if r.op != "Service.Process" {
		t.Errorf("AndThen should preserve op: got %v, want Service.Process", r.op)
	}

	if r.meta["step"] != "1" {
		t.Error("AndThen should preserve meta")
	}
}

func TestAndThen_NextContextOverrides(t *testing.T) {
	r := Ok(10).
		WithOp("Service.Process").
		WithMeta("step", "1").
		AndThen(func(n int) Result[int] {
			return Ok(n*2).
				WithOp("SubService.Process").
				WithMeta("step", "2")
		})

	// Next operation's context should override
	if r.op != "SubService.Process" {
		t.Errorf("AndThen next op should override: got %v", r.op)
	}

	if r.meta["step"] != "2" {
		t.Error("AndThen next meta should override")
	}
}

func TestAndThen_RailwayOriented(t *testing.T) {
	// Simulate a pipeline where step 2 fails
	step1Called := false
	step2Called := false
	step3Called := false

	r := Ok(10).
		AndThen(func(n int) Result[int] {
			step1Called = true
			return Ok(n + 1) // 11
		}).
		AndThen(func(n int) Result[int] {
			step2Called = true
			return Err[int](errors.New("step 2 failed"))
		}).
		AndThen(func(n int) Result[int] {
			step3Called = true
			return Ok(n + 1)
		})

	if !step1Called {
		t.Error("Step 1 should be called")
	}

	if !step2Called {
		t.Error("Step 2 should be called")
	}

	if step3Called {
		t.Error("Step 3 should NOT be called (short-circuited)")
	}

	if !r.IsErr() {
		t.Error("Result should be Err")
	}
}

// ============================================================================
// AndThenMap Tests
// ============================================================================

func TestAndThenMap_Ok(t *testing.T) {
	// Transform int -> string
	r := AndThenMap(
		Ok(42),
		func(n int) Result[string] {
			return Ok(fmt.Sprintf("number: %d", n))
		},
	)

	if !r.IsOk() {
		t.Error("AndThenMap should preserve Ok")
	}

	expected := "number: 42"
	if r.value != expected {
		t.Errorf("AndThenMap value = %v, want %v", r.value, expected)
	}
}

func TestAndThenMap_Err(t *testing.T) {
	testErr := errors.New("test error")
	called := false

	r := AndThenMap(
		Err[int](testErr),
		func(n int) Result[string] {
			called = true
			return Ok("should not be called")
		},
	)

	if !r.IsErr() {
		t.Error("AndThenMap should preserve Err")
	}

	if called {
		t.Error("AndThenMap should not call function on Err")
	}

	if r.err != testErr {
		t.Errorf("AndThenMap err = %v, want %v", r.err, testErr)
	}
}

func TestAndThenMap_ErrInFunction(t *testing.T) {
	testErr := errors.New("function failed")

	r := AndThenMap(
		Ok(42),
		func(n int) Result[string] {
			return Err[string](testErr)
		},
	)

	if !r.IsErr() {
		t.Error("AndThenMap should propagate error from function")
	}

	if r.err != testErr {
		t.Errorf("AndThenMap err = %v, want %v", r.err, testErr)
	}
}

func TestAndThenMap_PreservesContext(t *testing.T) {
	r := AndThenMap(
		Ok(42).WithOp("Service.Get").WithMeta("key", "value"),
		func(n int) Result[string] {
			return Ok(fmt.Sprintf("%d", n)) // No context
		},
	)

	if r.op != "Service.Get" {
		t.Error("AndThenMap should preserve op")
	}

	if r.meta["key"] != "value" {
		t.Error("AndThenMap should preserve meta")
	}
}

func TestAndThenMap_Chain(t *testing.T) {
	// int -> string -> bool
	r := AndThenMap(
		Ok(42),
		func(n int) Result[string] {
			return Ok(fmt.Sprintf("%d", n))
		},
	)

	r2 := AndThenMap(
		r,
		func(s string) Result[bool] {
			return Ok(len(s) > 0)
		},
	)

	if !r2.IsOk() {
		t.Error("Chained AndThenMap should preserve Ok")
	}

	if !r2.value {
		t.Error("Chained AndThenMap value should be true")
	}
}

func TestAndThenMap_MixedWithAndThen(t *testing.T) {
	// Mixing AndThen and AndThenMap
	type User struct {
		ID   int
		Name string
	}

	r := AndThenMap(
		Ok(42),
		func(id int) Result[*User] {
			return Ok(&User{ID: id, Name: "Alice"})
		},
	)

	// Now use regular AndThen since we're staying with User type
	r2 := r.AndThen(func(user *User) Result[*User] {
		user.Name = user.Name + " Smith"
		return Ok(user)
	})

	if !r2.IsOk() {
		t.Error("Mixed chain should work")
	}

	if r2.value.Name != "Alice Smith" {
		t.Errorf("Mixed chain value = %v, want Alice Smith", r2.value.Name)
	}
}

// ============================================================================
// OrElse Tests
// ============================================================================

func TestOrElse_OkUnchanged(t *testing.T) {
	called := false

	r := Ok(42).OrElse(func(err error) Result[int] {
		called = true
		return Ok(0)
	})

	if !r.IsOk() {
		t.Error("OrElse should preserve Ok")
	}

	if called {
		t.Error("OrElse should not call function on Ok")
	}

	if r.value != 42 {
		t.Errorf("OrElse value = %v, want 42", r.value)
	}
}

func TestOrElse_ErrCallsFunction(t *testing.T) {
	testErr := errors.New("test error")
	called := false
	var capturedErr error

	r := Err[int](testErr).OrElse(func(err error) Result[int] {
		called = true
		capturedErr = err
		return Ok(99)
	})

	if !called {
		t.Error("OrElse should call function on Err")
	}

	if capturedErr != testErr {
		t.Errorf("OrElse should pass error to function: got %v", capturedErr)
	}

	if !r.IsOk() {
		t.Error("OrElse should return alternative Ok result")
	}

	if r.value != 99 {
		t.Errorf("OrElse value = %v, want 99", r.value)
	}
}

func TestOrElse_FallbackChain(t *testing.T) {
	// Simulate cache miss -> DB miss -> default
	cacheErr := errors.New("cache miss")
	dbErr := errors.New("db miss")

	cacheCalled := false
	dbCalled := false
	defaultCalled := false

	r := Err[int](cacheErr).
		OrElse(func(err error) Result[int] {
			cacheCalled = true
			// Simulate DB lookup failure
			return Err[int](dbErr)
		}).
		OrElse(func(err error) Result[int] {
			dbCalled = true
			// Return default
			defaultCalled = true
			return Ok(0)
		})

	if !cacheCalled || !dbCalled || !defaultCalled {
		t.Error("OrElse chain should call all fallbacks")
	}

	if !r.IsOk() {
		t.Error("OrElse chain should eventually succeed")
	}

	if r.value != 0 {
		t.Errorf("OrElse chain value = %v, want 0", r.value)
	}
}

func TestOrElse_StopsAtFirstSuccess(t *testing.T) {
	fallback1Called := false
	fallback2Called := false

	r := Err[int](errors.New("error")).
		OrElse(func(err error) Result[int] {
			fallback1Called = true
			return Ok(42) // Success here
		}).
		OrElse(func(err error) Result[int] {
			fallback2Called = true
			return Ok(99)
		})

	if !fallback1Called {
		t.Error("First OrElse should be called")
	}

	if fallback2Called {
		t.Error("Second OrElse should NOT be called (chain succeeded)")
	}

	if !r.IsOk() {
		t.Error("OrElse should succeed at first successful fallback")
	}

	if r.value != 42 {
		t.Errorf("OrElse value = %v, want 42", r.value)
	}
}

func TestOrElse_AllFallbacksFail(t *testing.T) {
	finalErr := errors.New("final error")

	r := Err[int](errors.New("initial error")).
		OrElse(func(err error) Result[int] {
			return Err[int](errors.New("fallback 1 failed"))
		}).
		OrElse(func(err error) Result[int] {
			return Err[int](finalErr)
		})

	if !r.IsErr() {
		t.Error("OrElse should remain Err if all fallbacks fail")
	}

	if r.err != finalErr {
		t.Errorf("OrElse err = %v, want %v", r.err, finalErr)
	}
}

// ============================================================================
// Integration Tests (Combining Combinators)
// ============================================================================

func TestCombinators_ComplexPipeline(t *testing.T) {
	// Simulate: validate -> process -> save with fallback
	type Data struct {
		value int
		valid bool
	}

	validate := func(n int) Result[*Data] {
		if n <= 0 {
			return Err[*Data](errors.New("invalid input"))
		}
		return Ok(&Data{value: n, valid: true})
	}

	process := func(d *Data) Result[*Data] {
		d.value = d.value * 2
		return Ok(d)
	}

	save := func(d *Data) Result[*Data] {
		// Simulate save failure
		return Err[*Data](errors.New("save failed"))
	}

	fallbackSave := func(err error) Result[*Data] {
		// Fallback: save to cache
		return Ok(&Data{value: 42, valid: true})
	}

	r := AndThenMap(
		Ok(21),
		validate,
	).
		AndThen(process).
		AndThen(save).
		OrElse(fallbackSave)

	if !r.IsOk() {
		t.Error("Pipeline should succeed with fallback")
	}

	if r.value.value != 42 {
		t.Errorf("Pipeline value = %v, want 42", r.value.value)
	}
}

func TestCombinators_RailwayWithRecovery(t *testing.T) {
	step1 := func(n int) Result[int] {
		return Ok(n + 10)
	}

	step2 := func(n int) Result[int] {
		return Err[int](errors.New("step 2 failed"))
	}

	step3 := func(n int) Result[int] {
		return Ok(n + 20)
	}

	recover := func(err error) Result[int] {
		return Ok(100) // Recovered value
	}

	r := Ok(5).
		AndThen(step1). // Ok(15)
		AndThen(step2). // Err
		AndThen(step3). // Skipped
		OrElse(recover) // Recovered to 100

	if !r.IsOk() {
		t.Error("Pipeline should recover from error")
	}

	if r.value != 100 {
		t.Errorf("Recovered value = %v, want 100", r.value)
	}
}

func TestCombinators_ContextPropagation(t *testing.T) {
	r := Ok(10).
		WithOp("Pipeline.Start").
		WithMeta("trace_id", "abc123").
		AndThen(func(n int) Result[int] {
			return Ok(n * 2) // Context should be preserved
		}).
		AndThen(func(n int) Result[int] {
			return Ok(n + 5) // Context should be preserved
		})

	if r.op != "Pipeline.Start" {
		t.Error("Context should propagate through AndThen chain")
	}

	if r.meta["trace_id"] != "abc123" {
		t.Error("Meta should propagate through AndThen chain")
	}

	if r.value != 25 {
		t.Errorf("Value = %v, want 25", r.value)
	}
}

func TestCombinators_MixedChain(t *testing.T) {
	// Test mixing AndThen, AndThenMap, and OrElse
	type User struct{ ID int }
	type Profile struct{ UserID int }

	r := AndThenMap(
		Ok(42),
		func(id int) Result[*User] {
			return Ok(&User{ID: id})
		},
	).
		AndThen(func(user *User) Result[*User] {
			if user.ID == 0 {
				return Err[*User](errors.New("invalid user"))
			}
			return Ok(user)
		}).
		OrElse(func(err error) Result[*User] {
			return Ok(&User{ID: -1}) // Guest user
		})

	r2 := AndThenMap(
		r,
		func(user *User) Result[*Profile] {
			return Ok(&Profile{UserID: user.ID})
		},
	)

	if !r2.IsOk() {
		t.Error("Mixed chain should succeed")
	}

	if r2.value.UserID != 42 {
		t.Errorf("Profile UserID = %v, want 42", r2.value.UserID)
	}
}

// ============================================================================
// Edge Cases
// ============================================================================

func TestAndThen_NilFunction(t *testing.T) {
	// This would panic in real code, but documenting behavior
	defer func() {
		if r := recover(); r == nil {
			t.Error("AndThen with nil function should panic")
		}
	}()

	_ = Ok(42).AndThen(nil)
}

func TestOrElse_MultipleErrors(t *testing.T) {
	// Track error progression through fallback chain
	errorMessages := []string{}

	r := Err[int](errors.New("error1")).
		OrElse(func(err error) Result[int] {
			errorMessages = append(errorMessages, err.Error())
			return Err[int](errors.New("error2"))
		}).
		OrElse(func(err error) Result[int] {
			errorMessages = append(errorMessages, err.Error())
			return Err[int](errors.New("error3"))
		})

	if len(errorMessages) != 2 {
		t.Errorf("Should have tracked 2 errors, got %d", len(errorMessages))
	}

	if errorMessages[0] != "error1" {
		t.Errorf("First error = %v, want error1", errorMessages[0])
	}

	if errorMessages[1] != "error2" {
		t.Errorf("Second error = %v, want error2", errorMessages[1])
	}

	if r.err.Error() != "error3" {
		t.Errorf("Final error = %v, want error3", r.err.Error())
	}
}
