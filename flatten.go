package result

// Flatten unwraps a nested Result[Result[T]] into Result[T].
// If the outer Result is Err, the error is propagated.
// If the outer Result is Ok, the inner Result is returned.
//
// This is useful when chaining operations that themselves return Results.
//
// Example:
//
//	nestedResult := getUser(id)  // Returns Result[Result[*User]]
//	flatResult := result.Flatten(nestedResult)  // Returns Result[*User]
func Flatten[T any](r Result[Result[T]]) Result[T] {
	if r.IsErr() {
		return Result[T]{
			err:  r.err,
			op:   r.op,
			meta: r.meta,
		}
	}

	inner := r.value

	// Preserve outer context if inner doesn't have it
	if inner.op == "" && r.op != "" {
		inner.op = r.op
	}
	if inner.meta == nil && r.meta != nil {
		inner.meta = r.meta
	}

	return inner
}

// FlatMap is a convenience that combines Map and Flatten.
// It's equivalent to: Flatten(MapValue(r, f))
//
// This is useful when you have a function that returns a nested Result.
//
// Example:
//
//	result.FlatMap(userResult, func(user *User) result.Result[*Profile] {
//	    return getProfile(user.ID)
//	})
func FlatMap[T, U any](r Result[T], f func(T) Result[U]) Result[U] {
	return AndThenMap(r, f)
}
