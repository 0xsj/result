package result

import (
	"context"
	"time"
)

// WithTimeout runs a function with a timeout, returning Err if it exceeds the duration.
// The context passed to fn will be cancelled after timeout.
//
// Example:
//
//	result := result.WithTimeout(ctx, 5*time.Second, func(ctx context.Context) result.Result[*User] {
//	    return repo.FindByID(ctx, id)
//	})
func WithTimeout[T any](
	ctx context.Context,
	timeout time.Duration,
	fn func(context.Context) Result[T],
) Result[T] {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	resultChan := make(chan Result[T], 1)

	go func() {
		resultChan <- fn(ctx)
	}()

	select {
	case result := <-resultChan:
		return result
	case <-ctx.Done():
		return Err[T](Internal("timeout", ctx.Err()))
	}
}

// WithDeadline runs a function with an absolute deadline.
// The context passed to fn will be cancelled at the deadline.
//
// Example:
//
//	deadline := time.Now().Add(10 * time.Second)
//	result := result.WithDeadline(ctx, deadline, func(ctx context.Context) result.Result[*User] {
//	    return repo.FindByID(ctx, id)
//	})
func WithDeadline[T any](
	ctx context.Context,
	deadline time.Time,
	fn func(context.Context) Result[T],
) Result[T] {
	ctx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()

	resultChan := make(chan Result[T], 1)

	go func() {
		resultChan <- fn(ctx)
	}()

	select {
	case result := <-resultChan:
		return result
	case <-ctx.Done():
		return Err[T](Internal("deadline", ctx.Err()))
	}
}

// Retry retries a function up to maxAttempts with exponential backoff.
// Returns the first successful Result or the last error if all attempts fail.
//
// Backoff formula: initialBackoff * 2^attempt
//
// Example:
//
//	result := result.Retry(ctx, 3, time.Second, func() result.Result[*User] {
//	    return externalAPI.GetUser(id)
//	})
//	// Tries at: 0s, 1s, 2s (exponential backoff)
func Retry[T any](
	ctx context.Context,
	maxAttempts int,
	initialBackoff time.Duration,
	fn func() Result[T],
) Result[T] {
	if maxAttempts < 1 {
		maxAttempts = 1
	}

	var lastResult Result[T]

	for attempt := 0; attempt < maxAttempts; attempt++ {
		// Check if context is cancelled before retry
		if ctx.Err() != nil {
			return Err[T](Internal("retry.cancelled", ctx.Err()))
		}

		// Execute function
		lastResult = fn()
		if lastResult.IsOk() {
			return lastResult // Success!
		}

		// Don't sleep after last attempt
		if attempt < maxAttempts-1 {
			backoff := initialBackoff * time.Duration(1<<uint(attempt))

			select {
			case <-time.After(backoff):
				// Continue to next attempt
			case <-ctx.Done():
				return Err[T](Internal("retry.cancelled", ctx.Err()))
			}
		}
	}

	// All attempts failed, return last error
	return lastResult
}

// RetryIf retries only if the error matches specific kinds.
// This is useful when you only want to retry transient errors (e.g., infrastructure failures).
//
// Example:
//
//	result := result.RetryIf(
//	    ctx,
//	    3,
//	    time.Second,
//	    []result.Kind{result.KindInfrastructure, result.KindInternal},
//	    func() result.Result[*User] {
//	        return externalAPI.GetUser(id)
//	    },
//	)
func RetryIf[T any](
	ctx context.Context,
	maxAttempts int,
	initialBackoff time.Duration,
	retryableKinds []Kind,
	fn func() Result[T],
) Result[T] {
	if maxAttempts < 1 {
		maxAttempts = 1
	}

	// Create a map for O(1) lookup
	retryableMap := make(map[Kind]bool)
	for _, kind := range retryableKinds {
		retryableMap[kind] = true
	}

	var lastResult Result[T]

	for attempt := 0; attempt < maxAttempts; attempt++ {
		// Check if context is cancelled
		if ctx.Err() != nil {
			return Err[T](Internal("retry.cancelled", ctx.Err()))
		}

		// Execute function
		lastResult = fn()
		if lastResult.IsOk() {
			return lastResult // Success!
		}

		// Check if error is retryable
		errorKind := KindOf(lastResult.UnwrapErr())
		if !retryableMap[errorKind] {
			// Non-retryable error, return immediately
			return lastResult
		}

		// Don't sleep after last attempt
		if attempt < maxAttempts-1 {
			backoff := initialBackoff * time.Duration(1<<uint(attempt))

			select {
			case <-time.After(backoff):
				// Continue to next attempt
			case <-ctx.Done():
				return Err[T](Internal("retry.cancelled", ctx.Err()))
			}
		}
	}

	// All attempts failed, return last error
	return lastResult
}

// RetryWithBackoff retries with a custom backoff strategy.
//
// Example:
//
//	result := result.RetryWithBackoff(
//	    ctx,
//	    3,
//	    func(attempt int) time.Duration {
//	        return time.Duration(attempt*attempt) * time.Second // Quadratic backoff
//	    },
//	    func() result.Result[*User] {
//	        return externalAPI.GetUser(id)
//	    },
//	)
func RetryWithBackoff[T any](
	ctx context.Context,
	maxAttempts int,
	backoffFn func(attempt int) time.Duration,
	fn func() Result[T],
) Result[T] {
	if maxAttempts < 1 {
		maxAttempts = 1
	}

	var lastResult Result[T]

	for attempt := 0; attempt < maxAttempts; attempt++ {
		// Check if context is cancelled
		if ctx.Err() != nil {
			return Err[T](Internal("retry.cancelled", ctx.Err()))
		}

		// Execute function
		lastResult = fn()
		if lastResult.IsOk() {
			return lastResult // Success!
		}

		// Don't sleep after last attempt
		if attempt < maxAttempts-1 {
			backoff := backoffFn(attempt)

			select {
			case <-time.After(backoff):
				// Continue to next attempt
			case <-ctx.Done():
				return Err[T](Internal("retry.cancelled", ctx.Err()))
			}
		}
	}

	// All attempts failed, return last error
	return lastResult
}

// RetryContext is like Retry but passes context to the function.
//
// Example:
//
//	result := result.RetryContext(ctx, 3, time.Second, func(ctx context.Context) result.Result[*User] {
//	    return repo.FindByID(ctx, id)
//	})
func RetryContext[T any](
	ctx context.Context,
	maxAttempts int,
	initialBackoff time.Duration,
	fn func(context.Context) Result[T],
) Result[T] {
	if maxAttempts < 1 {
		maxAttempts = 1
	}

	var lastResult Result[T]

	for attempt := 0; attempt < maxAttempts; attempt++ {
		// Check if context is cancelled
		if ctx.Err() != nil {
			return Err[T](Internal("retry.cancelled", ctx.Err()))
		}

		// Execute function
		lastResult = fn(ctx)
		if lastResult.IsOk() {
			return lastResult // Success!
		}

		// Don't sleep after last attempt
		if attempt < maxAttempts-1 {
			backoff := initialBackoff * time.Duration(1<<uint(attempt))

			select {
			case <-time.After(backoff):
				// Continue to next attempt
			case <-ctx.Done():
				return Err[T](Internal("retry.cancelled", ctx.Err()))
			}
		}
	}

	// All attempts failed, return last error
	return lastResult
}
