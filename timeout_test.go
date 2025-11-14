package result

import (
	"context"
	"errors"
	"testing"
	"time"
)

// ============================================================================
// WithTimeout Tests
// ============================================================================

func TestWithTimeout_CompletesInTime(t *testing.T) {
	ctx := context.Background()

	result := WithTimeout(ctx, 100*time.Millisecond, func(ctx context.Context) Result[int] {
		time.Sleep(10 * time.Millisecond)
		return Ok(42)
	})

	value := AssertOk(t, result)
	if value != 42 {
		t.Errorf("WithTimeout() = %v, want 42", value)
	}
}

func TestWithTimeout_ExceedsTimeout(t *testing.T) {
	ctx := context.Background()

	result := WithTimeout(ctx, 50*time.Millisecond, func(ctx context.Context) Result[int] {
		time.Sleep(200 * time.Millisecond)
		return Ok(42)
	})

	AssertErr(t, result)
	AssertErrKind(t, result, KindInternal)
	AssertErrContains(t, result, "deadline exceeded")
}

func TestWithTimeout_ReturnsError(t *testing.T) {
	ctx := context.Background()

	result := WithTimeout(ctx, 100*time.Millisecond, func(ctx context.Context) Result[int] {
		return Err[int](errors.New("operation failed"))
	})

	AssertErr(t, result)
	AssertErrContains(t, result, "operation failed")
}

func TestWithTimeout_ContextCancellation(t *testing.T) {
	ctx := context.Background()

	result := WithTimeout(ctx, 100*time.Millisecond, func(ctx context.Context) Result[int] {
		select {
		case <-time.After(200 * time.Millisecond):
			return Ok(42)
		case <-ctx.Done():
			return Err[int](Internal("cancelled", ctx.Err()))
		}
	})

	AssertErr(t, result)
}

// ============================================================================
// WithDeadline Tests
// ============================================================================

func TestWithDeadline_CompletesBeforeDeadline(t *testing.T) {
	ctx := context.Background()
	deadline := time.Now().Add(100 * time.Millisecond)

	result := WithDeadline(ctx, deadline, func(ctx context.Context) Result[int] {
		time.Sleep(10 * time.Millisecond)
		return Ok(42)
	})

	value := AssertOk(t, result)
	if value != 42 {
		t.Errorf("WithDeadline() = %v, want 42", value)
	}
}

func TestWithDeadline_ExceedsDeadline(t *testing.T) {
	ctx := context.Background()
	deadline := time.Now().Add(50 * time.Millisecond)

	result := WithDeadline(ctx, deadline, func(ctx context.Context) Result[int] {
		time.Sleep(200 * time.Millisecond)
		return Ok(42)
	})

	AssertErr(t, result)
	AssertErrContains(t, result, "deadline exceeded")
}

// ============================================================================
// Retry Tests
// ============================================================================

func TestRetry_SucceedsFirstAttempt(t *testing.T) {
	ctx := context.Background()
	attempts := 0

	result := Retry(ctx, 3, time.Millisecond, func() Result[int] {
		attempts++
		return Ok(42)
	})

	value := AssertOk(t, result)
	if value != 42 {
		t.Errorf("Retry() = %v, want 42", value)
	}
	if attempts != 1 {
		t.Errorf("Expected 1 attempt, got %d", attempts)
	}
}

func TestRetry_SucceedsSecondAttempt(t *testing.T) {
	ctx := context.Background()
	attempts := 0

	result := Retry(ctx, 3, time.Millisecond, func() Result[int] {
		attempts++
		if attempts < 2 {
			return Err[int](errors.New("temporary failure"))
		}
		return Ok(42)
	})

	value := AssertOk(t, result)
	if value != 42 {
		t.Errorf("Retry() = %v, want 42", value)
	}
	if attempts != 2 {
		t.Errorf("Expected 2 attempts, got %d", attempts)
	}
}

func TestRetry_AllAttemptsFail(t *testing.T) {
	ctx := context.Background()
	attempts := 0

	result := Retry(ctx, 3, time.Millisecond, func() Result[int] {
		attempts++
		return Err[int](errors.New("persistent failure"))
	})

	AssertErr(t, result)
	AssertErrContains(t, result, "persistent failure")
	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}
}

func TestRetry_ExponentialBackoff(t *testing.T) {
	ctx := context.Background()
	start := time.Now()
	attempts := 0

	Retry(ctx, 3, 10*time.Millisecond, func() Result[int] {
		attempts++
		return Err[int](errors.New("fail"))
	})

	duration := time.Since(start)

	// Should backoff: 0ms, 10ms, 20ms = ~30ms total
	// Allow some buffer for execution time
	if duration < 25*time.Millisecond || duration > 50*time.Millisecond {
		t.Errorf("Retry backoff took %v, expected ~30ms", duration)
	}
}

func TestRetry_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	attempts := 0

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	result := Retry(ctx, 10, 30*time.Millisecond, func() Result[int] {
		attempts++
		return Err[int](errors.New("fail"))
	})

	AssertErr(t, result)
	AssertErrContains(t, result, "cancelled")

	// Should stop retrying after context cancellation
	if attempts >= 10 {
		t.Errorf("Expected fewer than 10 attempts due to cancellation, got %d", attempts)
	}
}

func TestRetry_ZeroAttempts(t *testing.T) {
	ctx := context.Background()
	attempts := 0

	result := Retry(ctx, 0, time.Millisecond, func() Result[int] {
		attempts++
		return Ok(42)
	})

	AssertOk(t, result)
	if attempts != 1 {
		t.Errorf("Expected at least 1 attempt, got %d", attempts)
	}
}

// ============================================================================
// RetryIf Tests
// ============================================================================

func TestRetryIf_RetriesOnInfrastructureError(t *testing.T) {
	ctx := context.Background()
	attempts := 0

	result := RetryIf(
		ctx,
		3,
		time.Millisecond,
		[]Kind{KindInfrastructure},
		func() Result[int] {
			attempts++
			if attempts < 2 {
				return Err[int](Infrastructure("db.query", errors.New("connection lost")))
			}
			return Ok(42)
		},
	)

	value := AssertOk(t, result)
	if value != 42 {
		t.Errorf("RetryIf() = %v, want 42", value)
	}
	if attempts != 2 {
		t.Errorf("Expected 2 attempts, got %d", attempts)
	}
}

func TestRetryIf_DoesNotRetryOnValidationError(t *testing.T) {
	ctx := context.Background()
	attempts := 0

	result := RetryIf(
		ctx,
		3,
		time.Millisecond,
		[]Kind{KindInfrastructure},
		func() Result[int] {
			attempts++
			return Err[int](Validation("input.validate", "invalid input", nil))
		},
	)

	AssertErr(t, result)
	AssertErrKind(t, result, KindValidation)
	if attempts != 1 {
		t.Errorf("Expected 1 attempt (no retry), got %d", attempts)
	}
}

func TestRetryIf_MultipleRetryableKinds(t *testing.T) {
	ctx := context.Background()
	attempts := 0

	result := RetryIf(
		ctx,
		3,
		time.Millisecond,
		[]Kind{KindInfrastructure, KindInternal},
		func() Result[int] {
			attempts++
			if attempts == 1 {
				return Err[int](Infrastructure("db", errors.New("infra error")))
			}
			if attempts == 2 {
				return Err[int](Internal("system", errors.New("internal error")))
			}
			return Ok(42)
		},
	)

	value := AssertOk(t, result)
	if value != 42 {
		t.Errorf("RetryIf() = %v, want 42", value)
	}
	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}
}

// ============================================================================
// RetryWithBackoff Tests
// ============================================================================

func TestRetryWithBackoff_CustomBackoff(t *testing.T) {
	ctx := context.Background()
	attempts := 0
	backoffCalls := 0

	result := RetryWithBackoff(
		ctx,
		3,
		func(attempt int) time.Duration {
			backoffCalls++
			return time.Duration(attempt*attempt) * time.Millisecond // Quadratic
		},
		func() Result[int] {
			attempts++
			if attempts < 3 {
				return Err[int](errors.New("fail"))
			}
			return Ok(42)
		},
	)

	value := AssertOk(t, result)
	if value != 42 {
		t.Errorf("RetryWithBackoff() = %v, want 42", value)
	}
	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}
	// Backoff called for first 2 attempts (not after last)
	if backoffCalls != 2 {
		t.Errorf("Expected 2 backoff calls, got %d", backoffCalls)
	}
}

func TestRetryWithBackoff_LinearBackoff(t *testing.T) {
	ctx := context.Background()
	start := time.Now()

	RetryWithBackoff(
		ctx,
		3,
		func(attempt int) time.Duration {
			return time.Duration(attempt+1) * 10 * time.Millisecond // Linear
		},
		func() Result[int] {
			return Err[int](errors.New("fail"))
		},
	)

	duration := time.Since(start)

	// Should backoff: 0ms, 10ms, 20ms = ~30ms total
	if duration < 25*time.Millisecond || duration > 50*time.Millisecond {
		t.Errorf("Linear backoff took %v, expected ~30ms", duration)
	}
}

// ============================================================================
// RetryContext Tests
// ============================================================================

func TestRetryContext_SucceedsFirstAttempt(t *testing.T) {
	ctx := context.Background()
	attempts := 0

	result := RetryContext(ctx, 3, time.Millisecond, func(ctx context.Context) Result[int] {
		attempts++
		return Ok(42)
	})

	value := AssertOk(t, result)
	if value != 42 {
		t.Errorf("RetryContext() = %v, want 42", value)
	}
	if attempts != 1 {
		t.Errorf("Expected 1 attempt, got %d", attempts)
	}
}

func TestRetryContext_PassesContextToFunction(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	attempts := 0

	result := RetryContext(ctx, 3, time.Millisecond, func(ctx context.Context) Result[int] {
		attempts++
		// Verify context is passed correctly
		if ctx.Err() != nil {
			return Err[int](Internal("cancelled", ctx.Err()))
		}
		if attempts < 2 {
			return Err[int](errors.New("fail"))
		}
		return Ok(42)
	})

	value := AssertOk(t, result)
	if value != 42 {
		t.Errorf("RetryContext() = %v, want 42", value)
	}
}

func TestRetryContext_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	attempts := 0

	go func() {
		time.Sleep(30 * time.Millisecond)
		cancel()
	}()

	result := RetryContext(ctx, 10, 20*time.Millisecond, func(ctx context.Context) Result[int] {
		attempts++
		return Err[int](errors.New("fail"))
	})

	AssertErr(t, result)
	AssertErrContains(t, result, "cancelled")

	// Should stop early due to cancellation
	if attempts >= 10 {
		t.Errorf("Expected fewer than 10 attempts, got %d", attempts)
	}
}
