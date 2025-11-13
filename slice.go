package result

import "context"

// Collect transforms a slice of Results into a Result containing a slice.
// If all Results are Ok, it returns Ok with a slice of all values.
// If any Result is Err, it returns the first Err encountered (fail-fast).
//
// This is useful for batch operations where you want to ensure all operations
// succeeded before proceeding.
//
// Example:
//
//	userResults := []result.Result[*User]{
//	    repo.FindByID(ctx, "1"),
//	    repo.FindByID(ctx, "2"),
//	    repo.FindByID(ctx, "3"),
//	}
//
//	allUsers := result.Collect(userResults)
//	allUsers.Match(
//	    func(users []*User) {
//	        // All succeeded - process all users
//	    },
//	    func(err error) {
//	        // At least one failed
//	    },
//	)
func Collect[T any](results []Result[T]) Result[[]T] {
	values := make([]T, 0, len(results))

	for i, r := range results {
		if r.IsErr() {
			// Preserve context from the failing result
			return Result[[]T]{
				err:  r.err,
				op:   r.op,
				meta: enrichMetaWithIndex(r.meta, i),
			}
		}
		values = append(values, r.value)
	}

	return Ok(values)
}

// Partition splits a slice of Results into separate slices of values and errors.
// This is useful when you want to process all results regardless of individual
// failures, collecting both successes and failures.
//
// Example:
//
//	userResults := []result.Result[*User]{
//	    repo.FindByID(ctx, "1"),  // Ok
//	    repo.FindByID(ctx, "2"),  // Err
//	    repo.FindByID(ctx, "3"),  // Ok
//	}
//
//	users, errs := result.Partition(userResults)
//	// users = [user1, user3]
//	// errs = [error2]
//
//	for _, user := range users {
//	    process(user)
//	}
//
//	for _, err := range errs {
//	    log.Error("failed to fetch user", "error", err)
//	}
func Partition[T any](results []Result[T]) ([]T, []error) {
	values := make([]T, 0, len(results))
	errors := make([]error, 0)

	for _, r := range results {
		if r.IsOk() {
			values = append(values, r.value)
		} else {
			errors = append(errors, r.err)
		}
	}

	return values, errors
}

// CollectWithContext is like Collect but enriches errors with context.
// This is useful for batch operations where you want correlation IDs
// attached to any failures.
//
// Example:
//
//	userResults := fetchMultipleUsers(ctx, ids)
//	allUsers := result.CollectWithContext(ctx, userResults)
func CollectWithContext[T any](ctx context.Context, results []Result[T]) Result[[]T] {
	values := make([]T, 0, len(results))

	for i, r := range results {
		if r.IsErr() {
			return Result[[]T]{
				err:  r.err,
				op:   r.op,
				meta: enrichMetaWithContext(enrichMetaWithIndex(r.meta, i), ctx),
			}
		}
		values = append(values, r.value)
	}

	return Ok(values)
}

// PartitionWithContext is like Partition but enriches errors with context.
//
// Example:
//
//	userResults := fetchMultipleUsers(ctx, ids)
//	users, errs := result.PartitionWithContext(ctx, userResults)
func PartitionWithContext[T any](ctx context.Context, results []Result[T]) ([]T, []error) {
	values := make([]T, 0, len(results))
	errors := make([]error, 0)

	for i, r := range results {
		if r.IsOk() {
			values = append(values, r.value)
		} else {
			// Enrich error with context and index
			enrichedErr := &Error{
				Kind:    KindOf(r.err),
				Op:      OpOf(r.err),
				Err:     r.err,
				Message: r.err.Error(),
				Meta:    enrichMetaWithContext(enrichMetaWithIndex(r.meta, i), ctx),
			}
			errors = append(errors, enrichedErr)
		}
	}

	return values, errors
}

// ============================================================================
// Helper Functions
// ============================================================================

// enrichMetaWithIndex adds an "index" field to metadata
func enrichMetaWithIndex(meta map[string]any, index int) map[string]any {
	if meta == nil {
		meta = make(map[string]any)
	}

	// Create a copy to avoid mutating original
	enriched := make(map[string]any, len(meta)+1)
	for k, v := range meta {
		enriched[k] = v
	}

	enriched["index"] = index
	return enriched
}

// enrichMetaWithContext adds context values to metadata
func enrichMetaWithContext(meta map[string]any, ctx context.Context) map[string]any {
	if meta == nil {
		meta = make(map[string]any)
	}

	// Create a copy to avoid mutating original
	enriched := make(map[string]any, len(meta))
	for k, v := range meta {
		enriched[k] = v
	}

	// Add context values if not already present
	if _, exists := enriched["request_id"]; !exists {
		if requestID := GetRequestID(ctx); requestID != "" {
			enriched["request_id"] = requestID
		}
	}

	if _, exists := enriched["trace_id"]; !exists {
		if traceID := GetTraceID(ctx); traceID != "" {
			enriched["trace_id"] = traceID
		}
	}

	if _, exists := enriched["user_id"]; !exists {
		if userID := GetUserID(ctx); userID != "" {
			enriched["user_id"] = userID
		}
	}

	if _, exists := enriched["tenant_id"]; !exists {
		if tenantID := GetTenantID(ctx); tenantID != "" {
			enriched["tenant_id"] = tenantID
		}
	}

	return enriched
}
