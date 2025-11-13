package result

// AndThen chains operations that return Results (monadic bind/flatMap)
// This is the core operation for railway-oriented programming.
//
// If the result is Ok, the function f is applied to the value and its result is returned.
// If the result is Err, the error is propagated without calling f.
//
// Context (op, meta) is preserved and propagated through the chain.
func (r Result[T]) AndThen(f func(T) Result[T]) Result[T] {
	if r.IsErr() {
		return r
	}

	next := f(r.value)

	// Preserve context if next doesn't have it
	if next.op == "" && r.op != "" {
		next.op = r.op
	}
	if next.meta == nil && r.meta != nil {
		next.meta = r.meta
	}

	return next
}

// AndThenMap chains operations that return Results of different types.
// This is AndThen but with a type transformation.
//
// Example:
//
//	result.Ok(42).AndThenMap(func(n int) result.Result[string] {
//	    return result.Ok(fmt.Sprintf("number: %d", n))
//	})
func AndThenMap[T, U any](r Result[T], f func(T) Result[U]) Result[U] {
	if r.IsErr() {
		return Result[U]{
			err:  r.err,
			op:   r.op,
			meta: r.meta,
		}
	}

	next := f(r.value)

	// Preserve context
	if next.op == "" && r.op != "" {
		next.op = r.op
	}
	if next.meta == nil && r.meta != nil {
		next.meta = r.meta
	}

	return next
}

// OrElse provides an alternative Result if this one is Err.
// If the result is Ok, it is returned unchanged.
// If the result is Err, the function f is called with the error to produce an alternative.
//
// This is useful for fallback chains:
//
//	cache.Get(id).
//	    OrElse(func(err error) result.Result[User] {
//	        return db.FindByID(id)
//	    }).
//	    OrElse(func(err error) result.Result[User] {
//	        return result.Ok(defaultUser)
//	    })
func (r Result[T]) OrElse(f func(error) Result[T]) Result[T] {
	if r.IsOk() {
		return r
	}
	return f(r.err)
}
