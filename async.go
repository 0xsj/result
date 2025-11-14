package result

import (
	"context"
	"sync"
)

// ParMap applies a function to each element concurrently and returns a slice of Results.
// Each element is processed in its own goroutine.
// The order of results matches the order of input items.
//
// Example:
//
//	userIDs := []string{"id1", "id2", "id3"}
//	results := result.ParMap(userIDs, func(id string) result.Result[*User] {
//	    return repo.FindByID(ctx, id)
//	})
func ParMap[T, U any](items []T, fn func(T) Result[U]) []Result[U] {
	results := make([]Result[U], len(items))
	var wg sync.WaitGroup
	wg.Add(len(items))

	for i, item := range items {
		i, item := i, item // Capture loop variables
		go func() {
			defer wg.Done()
			results[i] = fn(item)
		}()
	}

	wg.Wait()
	return results
}

// ParMapContext is like ParMap but with context support for cancellation.
// If context is cancelled, remaining operations are not started.
//
// Example:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//	results := result.ParMapContext(ctx, userIDs, func(ctx context.Context, id string) result.Result[*User] {
//	    return repo.FindByID(ctx, id)
//	})
func ParMapContext[T, U any](ctx context.Context, items []T, fn func(context.Context, T) Result[U]) []Result[U] {
	results := make([]Result[U], len(items))
	var wg sync.WaitGroup

	for i, item := range items {
		// Check if context is already cancelled before starting
		if ctx.Err() != nil {
			results[i] = Err[U](Internal("async.parmap", ctx.Err()))
			continue
		}

		i, item := i, item // Capture loop variables
		wg.Add(1)
		go func() {
			defer wg.Done()
			results[i] = fn(ctx, item)
		}()
	}

	wg.Wait()
	return results
}

// ParCollect runs multiple functions concurrently and collects their Results.
// Returns Ok with all values if all succeed, or Err with the first error encountered.
// This is fail-fast - if any operation fails, the error is returned immediately.
//
// Example:
//
//	results := result.ParCollect(
//	    func() result.Result[*User] { return fetchUser(ctx) },
//	    func() result.Result[*Profile] { return fetchProfile(ctx) },
//	    func() result.Result[*Settings] { return fetchSettings(ctx) },
//	)
func ParCollect[T any](fns ...func() Result[T]) Result[[]T] {
	if len(fns) == 0 {
		return Ok([]T{})
	}

	results := make([]Result[T], len(fns))
	var wg sync.WaitGroup
	wg.Add(len(fns))

	for i, fn := range fns {
		i, fn := i, fn // Capture loop variables
		go func() {
			defer wg.Done()
			results[i] = fn()
		}()
	}

	wg.Wait()

	// Collect all results - fail fast on first error
	return Collect(results)
}

// ParCollectContext is like ParCollect but with context support.
//
// Example:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
//	defer cancel()
//	results := result.ParCollectContext(ctx,
//	    func(ctx context.Context) result.Result[*User] { return fetchUser(ctx) },
//	    func(ctx context.Context) result.Result[*Profile] { return fetchProfile(ctx) },
//	)
func ParCollectContext[T any](ctx context.Context, fns ...func(context.Context) Result[T]) Result[[]T] {
	if len(fns) == 0 {
		return Ok([]T{})
	}

	results := make([]Result[T], len(fns))
	var wg sync.WaitGroup

	for i, fn := range fns {
		// Check if context is already cancelled
		if ctx.Err() != nil {
			return Err[[]T](Internal("async.collect", ctx.Err()))
		}

		i, fn := i, fn // Capture loop variables
		wg.Add(1)
		go func() {
			defer wg.Done()
			results[i] = fn(ctx)
		}()
	}

	wg.Wait()

	return Collect(results)
}

// Race runs multiple functions concurrently and returns the first successful Result.
// If all fail, returns the last error encountered.
// This is useful for fallback scenarios (cache → primary DB → replica DB).
//
// Example:
//
//	result := result.Race(
//	    func() result.Result[*User] { return cache.Get(id) },
//	    func() result.Result[*User] { return primaryDB.Find(id) },
//	    func() result.Result[*User] { return replicaDB.Find(id) },
//	)
func Race[T any](fns ...func() Result[T]) Result[T] {
	if len(fns) == 0 {
		return Err[T](Internal("async.race", nil))
	}

	if len(fns) == 1 {
		return fns[0]()
	}

	// Channel to collect results
	resultChan := make(chan Result[T], len(fns))

	// Start all operations
	for _, fn := range fns {
		fn := fn // Capture loop variable
		go func() {
			resultChan <- fn()
		}()
	}

	// Wait for first success or collect all failures
	var lastErr error
	for i := 0; i < len(fns); i++ {
		result := <-resultChan
		if result.IsOk() {
			return result // First success wins
		}
		lastErr = result.UnwrapErr() // Keep track of errors
	}

	// All failed, return last error
	return Err[T](lastErr)
}

// RaceContext is like Race but with context support for cancellation.
//
// Example:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
//	defer cancel()
//	result := result.RaceContext(ctx,
//	    func(ctx context.Context) result.Result[*User] { return cache.Get(ctx, id) },
//	    func(ctx context.Context) result.Result[*User] { return db.Find(ctx, id) },
//	)
func RaceContext[T any](ctx context.Context, fns ...func(context.Context) Result[T]) Result[T] {
	if len(fns) == 0 {
		return Err[T](Internal("async.race", nil))
	}

	if len(fns) == 1 {
		return fns[0](ctx)
	}

	// Check if context is already cancelled
	if ctx.Err() != nil {
		return Err[T](Internal("async.race", ctx.Err()))
	}

	// Channel to collect results
	resultChan := make(chan Result[T], len(fns))
	doneChan := make(chan struct{})

	// Start all operations
	for _, fn := range fns {
		fn := fn // Capture loop variable
		go func() {
			select {
			case <-doneChan:
				return // Stop if we already found a winner
			default:
				resultChan <- fn(ctx)
			}
		}()
	}

	// Wait for first success or collect all failures
	var lastErr error
	for i := 0; i < len(fns); i++ {
		select {
		case <-ctx.Done():
			close(doneChan)
			return Err[T](Internal("async.race", ctx.Err()))
		case result := <-resultChan:
			if result.IsOk() {
				close(doneChan)
				return result // First success wins
			}
			lastErr = result.UnwrapErr()
		}
	}

	// All failed, return last error
	close(doneChan)
	return Err[T](lastErr)
}

// ParMapCollect is a convenience that combines ParMap and Collect.
// It maps concurrently and returns Ok with all values or the first Err.
//
// Example:
//
//	result := result.ParMapCollect(userIDs, func(id string) result.Result[*User] {
//	    return repo.FindByID(ctx, id)
//	})
//	// result is Result[[]*User] - Ok if all succeed, Err if any fail
func ParMapCollect[T, U any](items []T, fn func(T) Result[U]) Result[[]U] {
	results := ParMap(items, fn)
	return Collect(results)
}

// ParMapCollectContext is like ParMapCollect but with context support.
func ParMapCollectContext[T, U any](ctx context.Context, items []T, fn func(context.Context, T) Result[U]) Result[[]U] {
	results := ParMapContext(ctx, items, fn)
	return Collect(results)
}
