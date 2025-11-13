package result

// Filter checks a predicate on the Ok value.
// If the result is Ok and the predicate returns true, the result is returned unchanged.
// If the result is Ok and the predicate returns false, returns Err with the provided error.
// If the result is already Err, it is returned unchanged.
//
// This is useful for adding validation checks to a successful result.
//
// Example:
//
//	result.Ok(email).
//	    Filter(func(e Email) bool {
//	        return !isBlacklisted(e)
//	    }, ErrEmailBlacklisted{})
func (r Result[T]) Filter(predicate func(T) bool, err error) Result[T] {
	if r.IsErr() {
		return r
	}
	if !predicate(r.value) {
		return Result[T]{
			err:  err,
			op:   r.op,
			meta: r.meta,
		}
	}
	return r
}

// FilterNot is the inverse of Filter - fails if predicate returns true.
// This is useful when you want to check that something is NOT the case.
//
// Example:
//
//	result.Ok(email).
//	    FilterNot(func(e Email) bool {
//	        return emailExists(e)
//	    }, ErrEmailAlreadyExists{})
func (r Result[T]) FilterNot(predicate func(T) bool, err error) Result[T] {
	return r.Filter(func(v T) bool { return !predicate(v) }, err)
}

// When converts a boolean condition into a Result check.
// If condition is true, the result is returned unchanged.
// If condition is false, returns Err with the provided error.
// If the result is already Err, it is returned unchanged.
//
// This is useful when you have an external condition to check.
//
// Example:
//
//	exists := repo.ExistsByEmail(ctx, email)
//	result.Ok(email).
//	    When(!exists, ErrEmailAlreadyExists{})
func (r Result[T]) When(condition bool, err error) Result[T] {
	if r.IsErr() {
		return r
	}
	if !condition {
		return Result[T]{
			err:  err,
			op:   r.op,
			meta: r.meta,
		}
	}
	return r
}

// Unless is the inverse of When - fails if condition is true.
//
// Example:
//
//	exists := repo.ExistsByEmail(ctx, email)
//	result.Ok(email).
//	    Unless(exists, ErrEmailAlreadyExists{})
func (r Result[T]) Unless(condition bool, err error) Result[T] {
	return r.When(!condition, err)
}
