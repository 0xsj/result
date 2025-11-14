package result

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

// ============================================================================
// ParMap Tests
// ============================================================================

func TestParMap_AllSucceed(t *testing.T) {
	items := []int{1, 2, 3, 4, 5}

	results := ParMap(items, func(n int) Result[int] {
		return Ok(n * 2)
	})

	if len(results) != 5 {
		t.Fatalf("Expected 5 results, got %d", len(results))
	}

	for i, r := range results {
		if r.IsErr() {
			t.Errorf("Result %d should be Ok: %v", i, r.UnwrapErr())
		}
		expected := items[i] * 2
		if r.Unwrap() != expected {
			t.Errorf("Result %d = %v, want %v", i, r.Unwrap(), expected)
		}
	}
}

func TestParMap_SomeFail(t *testing.T) {
	items := []int{1, 2, 3, 4, 5}

	results := ParMap(items, func(n int) Result[int] {
		if n%2 == 0 {
			return Err[int](errors.New("even number"))
		}
		return Ok(n * 2)
	})

	if len(results) != 5 {
		t.Fatalf("Expected 5 results, got %d", len(results))
	}

	// Check that even numbers failed
	for i, r := range results {
		n := items[i]
		if n%2 == 0 {
			if r.IsOk() {
				t.Errorf("Result %d should be Err for even number", i)
			}
		} else {
			if r.IsErr() {
				t.Errorf("Result %d should be Ok for odd number: %v", i, r.UnwrapErr())
			}
		}
	}
}

func TestParMap_EmptySlice(t *testing.T) {
	items := []int{}

	results := ParMap(items, func(n int) Result[int] {
		return Ok(n * 2)
	})

	if len(results) != 0 {
		t.Errorf("Expected 0 results for empty slice, got %d", len(results))
	}
}

func TestParMap_MaintainsOrder(t *testing.T) {
	items := []int{1, 2, 3, 4, 5}

	results := ParMap(items, func(n int) Result[int] {
		// Add some delay to test order preservation
		time.Sleep(time.Millisecond * time.Duration(10-n))
		return Ok(n * 10)
	})

	// Check order is maintained despite different execution times
	for i, r := range results {
		if r.IsErr() {
			t.Errorf("Result %d should be Ok: %v", i, r.UnwrapErr())
			continue
		}
		expected := items[i] * 10
		if r.Unwrap() != expected {
			t.Errorf("Result %d = %v, want %v (order not maintained)", i, r.Unwrap(), expected)
		}
	}
}

// ============================================================================
// ParMapContext Tests
// ============================================================================

func TestParMapContext_AllSucceed(t *testing.T) {
	ctx := context.Background()
	items := []int{1, 2, 3}

	results := ParMapContext(ctx, items, func(ctx context.Context, n int) Result[int] {
		return Ok(n * 2)
	})

	if len(results) != 3 {
		t.Fatalf("Expected 3 results, got %d", len(results))
	}

	for i, r := range results {
		AssertOk(t, r)
		if r.Unwrap() != items[i]*2 {
			t.Errorf("Result %d = %v, want %v", i, r.Unwrap(), items[i]*2)
		}
	}
}

func TestParMapContext_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	items := []int{1, 2, 3, 4, 5}

	results := ParMapContext(ctx, items, func(ctx context.Context, n int) Result[int] {
		return Ok(n * 2)
	})

	// All results should be errors due to cancelled context
	for i, r := range results {
		if r.IsOk() {
			t.Errorf("Result %d should be Err due to cancelled context", i)
		}
	}
}

func TestParMapContext_TimeoutDuringExecution(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	items := []int{1, 2, 3}
	var completed atomic.Int32

	results := ParMapContext(ctx, items, func(ctx context.Context, n int) Result[int] {
		// Simulate slow operation
		select {
		case <-time.After(100 * time.Millisecond):
			completed.Add(1)
			return Ok(n * 2)
		case <-ctx.Done():
			return Err[int](Internal("timeout", ctx.Err()))
		}
	})

	// Some operations should complete, some should timeout
	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}
}

// ============================================================================
// ParCollect Tests
// ============================================================================

func TestParCollect_AllSucceed(t *testing.T) {
	result := ParCollect(
		func() Result[int] { return Ok(1) },
		func() Result[int] { return Ok(2) },
		func() Result[int] { return Ok(3) },
	)

	values := AssertOk(t, result)
	if len(values) != 3 {
		t.Fatalf("Expected 3 values, got %d", len(values))
	}

	// Values should be in order
	for i, v := range values {
		if v != i+1 {
			t.Errorf("Value %d = %v, want %v", i, v, i+1)
		}
	}
}

func TestParCollect_OneFails(t *testing.T) {
	result := ParCollect(
		func() Result[int] { return Ok(1) },
		func() Result[int] { return Err[int](errors.New("failed")) },
		func() Result[int] { return Ok(3) },
	)

	AssertErr(t, result)
}

func TestParCollect_EmptyFunctions(t *testing.T) {
	result := ParCollect[int]()

	values := AssertOk(t, result)
	if len(values) != 0 {
		t.Errorf("Expected empty slice, got %d values", len(values))
	}
}

func TestParCollect_ConcurrencyExecution(t *testing.T) {
	start := time.Now()

	result := ParCollect(
		func() Result[int] {
			time.Sleep(100 * time.Millisecond)
			return Ok(1)
		},
		func() Result[int] {
			time.Sleep(100 * time.Millisecond)
			return Ok(2)
		},
		func() Result[int] {
			time.Sleep(100 * time.Millisecond)
			return Ok(3)
		},
	)

	duration := time.Since(start)

	AssertOk(t, result)

	// Should take ~100ms (concurrent), not ~300ms (sequential)
	if duration > 200*time.Millisecond {
		t.Errorf("ParCollect took %v, expected ~100ms (concurrent execution)", duration)
	}
}

// ============================================================================
// ParCollectContext Tests
// ============================================================================

func TestParCollectContext_AllSucceed(t *testing.T) {
	ctx := context.Background()

	result := ParCollectContext(ctx,
		func(ctx context.Context) Result[int] { return Ok(1) },
		func(ctx context.Context) Result[int] { return Ok(2) },
		func(ctx context.Context) Result[int] { return Ok(3) },
	)

	values := AssertOk(t, result)
	if len(values) != 3 {
		t.Fatalf("Expected 3 values, got %d", len(values))
	}
}

func TestParCollectContext_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	result := ParCollectContext(ctx,
		func(ctx context.Context) Result[int] { return Ok(1) },
		func(ctx context.Context) Result[int] { return Ok(2) },
	)

	AssertErr(t, result)
	AssertErrContains(t, result, "context canceled")
}

// ============================================================================
// Race Tests
// ============================================================================

func TestRace_FirstSucceeds(t *testing.T) {
	result := Race(
		func() Result[int] {
			return Ok(1) // First and fastest
		},
		func() Result[int] {
			time.Sleep(100 * time.Millisecond)
			return Ok(2)
		},
		func() Result[int] {
			time.Sleep(200 * time.Millisecond)
			return Ok(3)
		},
	)

	value := AssertOk(t, result)
	if value != 1 {
		t.Errorf("Race() = %v, want 1 (first success)", value)
	}
}

func TestRace_SecondSucceeds(t *testing.T) {
	result := Race(
		func() Result[int] {
			return Err[int](errors.New("first failed"))
		},
		func() Result[int] {
			return Ok(2) // First success
		},
		func() Result[int] {
			time.Sleep(100 * time.Millisecond)
			return Ok(3)
		},
	)

	value := AssertOk(t, result)
	if value != 2 {
		t.Errorf("Race() = %v, want 2 (first success)", value)
	}
}

func TestRace_AllFail(t *testing.T) {
	result := Race(
		func() Result[int] {
			return Err[int](errors.New("first failed"))
		},
		func() Result[int] {
			return Err[int](errors.New("second failed"))
		},
		func() Result[int] {
			return Err[int](errors.New("third failed"))
		},
	)

	AssertErr(t, result)
	// Should return last error
	AssertErrContains(t, result, "failed")
}

func TestRace_EmptyFunctions(t *testing.T) {
	result := Race[int]()

	AssertErr(t, result)
}

func TestRace_SingleFunction(t *testing.T) {
	result := Race(
		func() Result[int] {
			return Ok(42)
		},
	)

	value := AssertOk(t, result)
	if value != 42 {
		t.Errorf("Race() = %v, want 42", value)
	}
}

// ============================================================================
// RaceContext Tests
// ============================================================================

func TestRaceContext_FirstSucceeds(t *testing.T) {
	ctx := context.Background()

	result := RaceContext(ctx,
		func(ctx context.Context) Result[int] {
			return Ok(1)
		},
		func(ctx context.Context) Result[int] {
			time.Sleep(100 * time.Millisecond)
			return Ok(2)
		},
	)

	value := AssertOk(t, result)
	if value != 1 {
		t.Errorf("RaceContext() = %v, want 1", value)
	}
}

func TestRaceContext_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	result := RaceContext(ctx,
		func(ctx context.Context) Result[int] {
			// Check context immediately
			if ctx.Err() != nil {
				return Err[int](Internal("cancelled", ctx.Err()))
			}
			time.Sleep(100 * time.Millisecond)
			return Ok(1)
		},
		func(ctx context.Context) Result[int] {
			// Check context immediately
			if ctx.Err() != nil {
				return Err[int](Internal("cancelled", ctx.Err()))
			}
			time.Sleep(100 * time.Millisecond)
			return Ok(2)
		},
	)

	AssertErr(t, result)
	AssertErrContains(t, result, "context canceled")
}

func TestRaceContext_TimeoutBeforeSuccess(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	result := RaceContext(ctx,
		func(ctx context.Context) Result[int] {
			time.Sleep(100 * time.Millisecond) // Slower than timeout
			return Ok(1)
		},
		func(ctx context.Context) Result[int] {
			time.Sleep(100 * time.Millisecond) // Slower than timeout
			return Ok(2)
		},
	)

	AssertErr(t, result)
	AssertErrContains(t, result, "deadline exceeded")
}

// ============================================================================
// ParMapCollect Tests
// ============================================================================

func TestParMapCollect_AllSucceed(t *testing.T) {
	items := []int{1, 2, 3}

	result := ParMapCollect(items, func(n int) Result[int] {
		return Ok(n * 2)
	})

	values := AssertOk(t, result)
	if len(values) != 3 {
		t.Fatalf("Expected 3 values, got %d", len(values))
	}

	for i, v := range values {
		expected := items[i] * 2
		if v != expected {
			t.Errorf("Value %d = %v, want %v", i, v, expected)
		}
	}
}

func TestParMapCollect_OneFails(t *testing.T) {
	items := []int{1, 2, 3}

	result := ParMapCollect(items, func(n int) Result[int] {
		if n == 2 {
			return Err[int](errors.New("failed on 2"))
		}
		return Ok(n * 2)
	})

	AssertErr(t, result)
	AssertErrContains(t, result, "failed on 2")
}

// ============================================================================
// ParMapCollectContext Tests
// ============================================================================

func TestParMapCollectContext_AllSucceed(t *testing.T) {
	ctx := context.Background()
	items := []int{1, 2, 3}

	result := ParMapCollectContext(ctx, items, func(ctx context.Context, n int) Result[int] {
		return Ok(n * 2)
	})

	values := AssertOk(t, result)
	if len(values) != 3 {
		t.Fatalf("Expected 3 values, got %d", len(values))
	}
}

func TestParMapCollectContext_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	items := []int{1, 2, 3}

	result := ParMapCollectContext(ctx, items, func(ctx context.Context, n int) Result[int] {
		return Ok(n * 2)
	})

	AssertErr(t, result)
}
