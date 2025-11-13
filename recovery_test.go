package result

import (
	"errors"
	"testing"
)

// ============================================================================
// Recover Tests
// ============================================================================

func TestRecover_OkUnchanged(t *testing.T) {
	r := Ok(42).Recover(KindNotFound, func(err error) int {
		return 0
	})

	if !r.IsOk() {
		t.Error("Recover should not change Ok result")
	}

	if r.value != 42 {
		t.Errorf("Recover value = %v, want 42", r.value)
	}
}

func TestRecover_MatchingKind(t *testing.T) {
	called := false
	var capturedErr error

	r := Err[int](NotFound("test", "resource")).
		Recover(KindNotFound, func(err error) int {
			called = true
			capturedErr = err
			return 99
		})

	if !called {
		t.Error("Recover should call recovery function for matching kind")
	}

	if capturedErr == nil {
		t.Error("Recover should pass error to recovery function")
	}

	if !r.IsOk() {
		t.Error("Recover should convert to Ok result")
	}

	if r.value != 99 {
		t.Errorf("Recover value = %v, want 99", r.value)
	}
}

func TestRecover_NonMatchingKind(t *testing.T) {
	called := false
	testErr := Validation("test", "invalid", nil)

	r := Err[int](testErr).
		Recover(KindNotFound, func(err error) int {
			called = true
			return 99
		})

	if called {
		t.Error("Recover should not call recovery function for non-matching kind")
	}

	if !r.IsErr() {
		t.Error("Recover should preserve Err for non-matching kind")
	}

	if r.err != testErr {
		t.Errorf("Recover should preserve original error")
	}
}

func TestRecover_MultipleKinds(t *testing.T) {
	r := Err[int](NotFound("test", "resource")).
		Recover(KindValidation, func(err error) int {
			return 0
		}).
		Recover(KindNotFound, func(err error) int {
			return 99
		})

	if !r.IsOk() {
		t.Error("Recover chain should eventually recover")
	}

	if r.value != 99 {
		t.Errorf("Recover value = %v, want 99", r.value)
	}
}

// ============================================================================
// RecoverWith Tests
// ============================================================================

func TestRecoverWith_OkUnchanged(t *testing.T) {
	r := Ok(42).RecoverWith(KindNotFound, func(err error) Result[int] {
		return Ok(0)
	})

	if !r.IsOk() {
		t.Error("RecoverWith should not change Ok result")
	}

	if r.value != 42 {
		t.Errorf("RecoverWith value = %v, want 42", r.value)
	}
}

func TestRecoverWith_MatchingKindReturnsOk(t *testing.T) {
	called := false

	r := Err[int](NotFound("test", "resource")).
		RecoverWith(KindNotFound, func(err error) Result[int] {
			called = true
			return Ok(99)
		})

	if !called {
		t.Error("RecoverWith should call recovery function")
	}

	if !r.IsOk() {
		t.Error("RecoverWith should convert to Ok")
	}

	if r.value != 99 {
		t.Errorf("RecoverWith value = %v, want 99", r.value)
	}
}

func TestRecoverWith_MatchingKindReturnsErr(t *testing.T) {
	fallbackErr := errors.New("fallback failed")

	r := Err[int](NotFound("test", "resource")).
		RecoverWith(KindNotFound, func(err error) Result[int] {
			// Recovery itself fails
			return Err[int](fallbackErr)
		})

	if !r.IsErr() {
		t.Error("RecoverWith should preserve Err if recovery fails")
	}

	if r.err != fallbackErr {
		t.Errorf("RecoverWith err = %v, want %v", r.err, fallbackErr)
	}
}

func TestRecoverWith_NonMatchingKind(t *testing.T) {
	called := false
	testErr := Validation("test", "invalid", nil)

	r := Err[int](testErr).
		RecoverWith(KindNotFound, func(err error) Result[int] {
			called = true
			return Ok(99)
		})

	if called {
		t.Error("RecoverWith should not call recovery for non-matching kind")
	}

	if !r.IsErr() {
		t.Error("RecoverWith should preserve Err")
	}

	if r.err != testErr {
		t.Error("RecoverWith should preserve original error")
	}
}

func TestRecoverWith_FallbackChain(t *testing.T) {
	// Simulate cache -> redis -> db fallback
	cacheCalled := false
	redisCalled := false
	dbCalled := false

	r := Err[int](NotFound("cache", "key")).
		RecoverWith(KindNotFound, func(err error) Result[int] {
			cacheCalled = true
			// Redis also fails
			return Err[int](NotFound("redis", "key"))
		}).
		RecoverWith(KindNotFound, func(err error) Result[int] {
			redisCalled = true
			// DB succeeds
			dbCalled = true
			return Ok(42)
		})

	if !cacheCalled || !redisCalled || !dbCalled {
		t.Error("RecoverWith chain should try all fallbacks")
	}

	if !r.IsOk() {
		t.Error("RecoverWith chain should eventually succeed")
	}

	if r.value != 42 {
		t.Errorf("RecoverWith value = %v, want 42", r.value)
	}
}

// ============================================================================
// RecoverAll Tests
// ============================================================================

func TestRecoverAll_OkUnchanged(t *testing.T) {
	r := Ok(42).RecoverAll(func(err error) int {
		return 0
	})

	if !r.IsOk() {
		t.Error("RecoverAll should not change Ok result")
	}

	if r.value != 42 {
		t.Errorf("RecoverAll value = %v, want 42", r.value)
	}
}

func TestRecoverAll_AnyError(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{"NotFound", NotFound("test", "resource")},
		{"Validation", Validation("test", "invalid", nil)},
		{"Domain", Domain("test", "msg")},
		{"Infrastructure", Infrastructure("test", errors.New("db"))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			var capturedErr error

			r := Err[int](tt.err).
				RecoverAll(func(err error) int {
					called = true
					capturedErr = err
					return 99
				})

			if !called {
				t.Errorf("RecoverAll should call recovery for %s", tt.name)
			}

			if capturedErr != tt.err {
				t.Error("RecoverAll should pass error to recovery")
			}

			if !r.IsOk() {
				t.Error("RecoverAll should convert to Ok")
			}

			if r.value != 99 {
				t.Errorf("RecoverAll value = %v, want 99", r.value)
			}
		})
	}
}

func TestRecoverAll_GuaranteesOk(t *testing.T) {
	// RecoverAll ensures result is always Ok
	r := Err[int](errors.New("any error")).
		RecoverAll(func(err error) int {
			return 0 // Default value
		})

	if !r.IsOk() {
		t.Error("RecoverAll should guarantee Ok result")
	}
}

// ============================================================================
// RecoverAllWith Tests
// ============================================================================

func TestRecoverAllWith_OkUnchanged(t *testing.T) {
	r := Ok(42).RecoverAllWith(func(err error) Result[int] {
		return Ok(0)
	})

	if !r.IsOk() {
		t.Error("RecoverAllWith should not change Ok result")
	}

	if r.value != 42 {
		t.Errorf("RecoverAllWith value = %v, want 42", r.value)
	}
}

func TestRecoverAllWith_AnyErrorReturnsOk(t *testing.T) {
	r := Err[int](Validation("test", "invalid", nil)).
		RecoverAllWith(func(err error) Result[int] {
			return Ok(99)
		})

	if !r.IsOk() {
		t.Error("RecoverAllWith should convert to Ok")
	}

	if r.value != 99 {
		t.Errorf("RecoverAllWith value = %v, want 99", r.value)
	}
}

func TestRecoverAllWith_AnyErrorReturnsErr(t *testing.T) {
	fallbackErr := errors.New("fallback failed")

	r := Err[int](NotFound("test", "resource")).
		RecoverAllWith(func(err error) Result[int] {
			return Err[int](fallbackErr)
		})

	if !r.IsErr() {
		t.Error("RecoverAllWith should preserve Err if recovery fails")
	}

	if r.err != fallbackErr {
		t.Errorf("RecoverAllWith err = %v, want %v", r.err, fallbackErr)
	}
}

// ============================================================================
// Inspect Tests
// ============================================================================

func TestInspect_Ok(t *testing.T) {
	called := false
	var capturedValue int

	r := Ok(42).Inspect(func(n int) {
		called = true
		capturedValue = n
	})

	if !called {
		t.Error("Inspect should call function on Ok")
	}

	if capturedValue != 42 {
		t.Errorf("Inspect captured = %v, want 42", capturedValue)
	}

	if !r.IsOk() {
		t.Error("Inspect should preserve Ok")
	}

	if r.value != 42 {
		t.Errorf("Inspect value = %v, want 42", r.value)
	}
}

func TestInspect_Err(t *testing.T) {
	called := false
	testErr := errors.New("test error")

	r := Err[int](testErr).Inspect(func(n int) {
		called = true
	})

	if called {
		t.Error("Inspect should not call function on Err")
	}

	if !r.IsErr() {
		t.Error("Inspect should preserve Err")
	}

	if r.err != testErr {
		t.Error("Inspect should preserve error")
	}
}

func TestInspect_Chain(t *testing.T) {
	inspections := []int{}

	r := Ok(10).
		Inspect(func(n int) {
			inspections = append(inspections, n)
		}).
		Map(func(n int) int {
			return n * 2
		}).
		Inspect(func(n int) {
			inspections = append(inspections, n)
		})

	if len(inspections) != 2 {
		t.Errorf("Inspect chain called %d times, want 2", len(inspections))
	}

	if inspections[0] != 10 {
		t.Errorf("First inspection = %v, want 10", inspections[0])
	}

	if inspections[1] != 20 {
		t.Errorf("Second inspection = %v, want 20", inspections[1])
	}

	if r.value != 20 {
		t.Errorf("Final value = %v, want 20", r.value)
	}
}

// ============================================================================
// InspectErr Tests
// ============================================================================

func TestInspectErr_Err(t *testing.T) {
	called := false
	testErr := errors.New("test error")
	var capturedErr error

	r := Err[int](testErr).InspectErr(func(err error) {
		called = true
		capturedErr = err
	})

	if !called {
		t.Error("InspectErr should call function on Err")
	}

	if capturedErr != testErr {
		t.Errorf("InspectErr captured = %v, want %v", capturedErr, testErr)
	}

	if !r.IsErr() {
		t.Error("InspectErr should preserve Err")
	}

	if r.err != testErr {
		t.Error("InspectErr should preserve error")
	}
}

func TestInspectErr_Ok(t *testing.T) {
	called := false

	r := Ok(42).InspectErr(func(err error) {
		called = true
	})

	if called {
		t.Error("InspectErr should not call function on Ok")
	}

	if !r.IsOk() {
		t.Error("InspectErr should preserve Ok")
	}

	if r.value != 42 {
		t.Errorf("InspectErr value = %v, want 42", r.value)
	}
}

func TestInspectErr_Chain(t *testing.T) {
	inspections := []error{}
	err1 := errors.New("error 1")
	err2 := errors.New("error 2")

	r := Err[int](err1).
		InspectErr(func(err error) {
			inspections = append(inspections, err)
		}).
		RecoverWith(KindInternal, func(err error) Result[int] {
			return Err[int](err2)
		}).
		InspectErr(func(err error) {
			inspections = append(inspections, err)
		})

	if len(inspections) != 2 {
		t.Errorf("InspectErr chain called %d times, want 2", len(inspections))
	}

	if inspections[0] != err1 {
		t.Error("First inspection should be err1")
	}

	if inspections[1] != err2 {
		t.Error("Second inspection should be err2")
	}

	if !r.IsErr() {
		t.Error("Result should still be Err")
	}

	if r.err != err2 {
		t.Error("Final error should be err2")
	}
}

// ============================================================================
// Tap Tests
// ============================================================================

func TestTap_Ok(t *testing.T) {
	okCalled := false
	errCalled := false
	var capturedValue int

	r := Ok(42).Tap(
		func(n int) {
			okCalled = true
			capturedValue = n
		},
		func(err error) {
			errCalled = true
		},
	)

	if !okCalled {
		t.Error("Tap should call onOk for Ok result")
	}

	if errCalled {
		t.Error("Tap should not call onErr for Ok result")
	}

	if capturedValue != 42 {
		t.Errorf("Tap captured = %v, want 42", capturedValue)
	}

	if !r.IsOk() {
		t.Error("Tap should preserve Ok")
	}

	if r.value != 42 {
		t.Errorf("Tap value = %v, want 42", r.value)
	}
}

func TestTap_Err(t *testing.T) {
	okCalled := false
	errCalled := false
	testErr := errors.New("test error")
	var capturedErr error

	r := Err[int](testErr).Tap(
		func(n int) {
			okCalled = true
		},
		func(err error) {
			errCalled = true
			capturedErr = err
		},
	)

	if okCalled {
		t.Error("Tap should not call onOk for Err result")
	}

	if !errCalled {
		t.Error("Tap should call onErr for Err result")
	}

	if capturedErr != testErr {
		t.Errorf("Tap captured = %v, want %v", capturedErr, testErr)
	}

	if !r.IsErr() {
		t.Error("Tap should preserve Err")
	}

	if r.err != testErr {
		t.Error("Tap should preserve error")
	}
}

func TestTap_Metrics(t *testing.T) {
	// Simulate metrics recording
	successCount := 0
	errorCount := 0

	Ok(42).Tap(
		func(n int) { successCount++ },
		func(err error) { errorCount++ },
	)

	if successCount != 1 {
		t.Errorf("Success count = %d, want 1", successCount)
	}

	if errorCount != 0 {
		t.Errorf("Error count = %d, want 0", errorCount)
	}

	Err[int](errors.New("error")).Tap(
		func(n int) { successCount++ },
		func(err error) { errorCount++ },
	)

	if successCount != 1 {
		t.Errorf("Success count = %d, want 1", successCount)
	}

	if errorCount != 1 {
		t.Errorf("Error count = %d, want 1", errorCount)
	}
}

// ============================================================================
// Integration Tests
// ============================================================================

func TestRecovery_WithAndThen(t *testing.T) {
	r := Err[int](NotFound("test", "resource")).
		Recover(KindNotFound, func(err error) int {
			return 10
		}).
		AndThen(func(n int) Result[int] {
			return Ok(n * 2)
		})

	if !r.IsOk() {
		t.Error("Recovery + AndThen should work")
	}

	if r.value != 20 {
		t.Errorf("Value = %v, want 20", r.value)
	}
}

func TestInspection_WithRecovery(t *testing.T) {
	inspected := false
	recovered := false

	r := Err[int](NotFound("test", "resource")).
		InspectErr(func(err error) {
			inspected = true
		}).
		Recover(KindNotFound, func(err error) int {
			recovered = true
			return 99
		})

	if !inspected {
		t.Error("Error should be inspected before recovery")
	}

	if !recovered {
		t.Error("Error should be recovered")
	}

	if !r.IsOk() {
		t.Error("Result should be Ok after recovery")
	}
}

func TestComplexPipeline(t *testing.T) {
	inspections := []string{}

	r := Err[int](NotFound("cache", "key")).
		InspectErr(func(err error) {
			inspections = append(inspections, "cache_miss")
		}).
		RecoverWith(KindNotFound, func(err error) Result[int] {
			inspections = append(inspections, "trying_db")
			return Err[int](Infrastructure("db", errors.New("db down")))
		}).
		InspectErr(func(err error) {
			inspections = append(inspections, "db_error")
		}).
		RecoverAll(func(err error) int {
			inspections = append(inspections, "using_default")
			return 0
		}).
		Inspect(func(n int) {
			inspections = append(inspections, "recovered")
		})

	expected := []string{
		"cache_miss",
		"trying_db",
		"db_error",
		"using_default",
		"recovered",
	}

	if len(inspections) != len(expected) {
		t.Errorf("Inspections count = %d, want %d", len(inspections), len(expected))
	}

	for i, exp := range expected {
		if i >= len(inspections) || inspections[i] != exp {
			t.Errorf("Inspection[%d] = %v, want %v", i, inspections[i], exp)
		}
	}

	if !r.IsOk() {
		t.Error("Pipeline should end with Ok")
	}

	if r.value != 0 {
		t.Errorf("Value = %v, want 0", r.value)
	}
}

func TestRecovery_SelectiveByKind(t *testing.T) {
	// Recover NotFound but not Validation
	r1 := Err[int](NotFound("test", "resource")).
		Recover(KindNotFound, func(err error) int { return 1 }).
		Recover(KindValidation, func(err error) int { return 2 })

	if !r1.IsOk() || r1.value != 1 {
		t.Error("Should recover NotFound with first handler")
	}

	r2 := Err[int](Validation("test", "invalid", nil)).
		Recover(KindNotFound, func(err error) int { return 1 }).
		Recover(KindValidation, func(err error) int { return 2 })

	if !r2.IsOk() || r2.value != 2 {
		t.Error("Should recover Validation with second handler")
	}

	r3 := Err[int](Domain("test", "msg")).
		Recover(KindNotFound, func(err error) int { return 1 }).
		Recover(KindValidation, func(err error) int { return 2 })

	if !r3.IsErr() {
		t.Error("Should not recover unhandled kinds")
	}
}
