package result

// Recover recovers from a specific error kind with a fallback value.
// If the result is Ok, it is returned unchanged.
// If the result is Err with a matching Kind, the recovery function is called
// with the error and its return value is wrapped in Ok.
// If the result is Err with a non-matching Kind, it is returned unchanged.
//
// This is useful for graceful degradation where certain errors can be recovered
// from with a default or fallback value.
//
// Example:
//
//	userResult.Recover(result.KindNotFound, func(err error) *User {
//	    return NewGuestUser()
//	})
func (r Result[T]) Recover(kind Kind, recovery func(error) T) Result[T] {
	if r.IsErr() && KindOf(r.err) == kind {
		return Ok(recovery(r.err))
	}
	return r
}

// RecoverWith recovers from a specific error kind with a Result-returning function.
// If the result is Ok, it is returned unchanged.
// If the result is Err with a matching Kind, the recovery function is called
// with the error and its returned Result is used.
// If the result is Err with a non-matching Kind, it is returned unchanged.
//
// This is useful for fallback chains where the fallback itself might fail.
//
// Example:
//
//	cache.Get(id).RecoverWith(result.KindNotFound, func(err error) result.Result[*User] {
//	    return db.FindByID(id)
//	})
func (r Result[T]) RecoverWith(kind Kind, recovery func(error) Result[T]) Result[T] {
	if r.IsErr() && KindOf(r.err) == kind {
		return recovery(r.err)
	}
	return r
}

// RecoverAll recovers from any error with a fallback value.
// If the result is Ok, it is returned unchanged.
// If the result is Err (any kind), the recovery function is called
// with the error and its return value is wrapped in Ok.
//
// This is useful when you want to ensure a Result is always Ok,
// regardless of the error type.
//
// Example:
//
//	userResult.RecoverAll(func(err error) *User {
//	    log.Warn("failed to get user, using guest", "error", err)
//	    return NewGuestUser()
//	})
func (r Result[T]) RecoverAll(recovery func(error) T) Result[T] {
	if r.IsErr() {
		return Ok(recovery(r.err))
	}
	return r
}

// RecoverAllWith recovers from any error with a Result-returning function.
// If the result is Ok, it is returned unchanged.
// If the result is Err (any kind), the recovery function is called
// with the error and its returned Result is used.
//
// This is useful for universal fallback mechanisms that might themselves fail.
//
// Example:
//
//	primaryService.Get(id).RecoverAllWith(func(err error) result.Result[*Data] {
//	    return backupService.Get(id)
//	})
func (r Result[T]) RecoverAllWith(recovery func(error) Result[T]) Result[T] {
	if r.IsErr() {
		return recovery(r.err)
	}
	return r
}

// ============================================================================
// Inspection (Side Effects Without Consuming)
// ============================================================================

// Inspect performs a side effect on the Ok value without consuming it.
// If the result is Ok, the function f is called with the value, and the
// original result is returned unchanged.
// If the result is Err, nothing happens and the error is returned unchanged.
//
// This is useful for logging, metrics, debugging, or other observability
// without breaking the chain.
//
// Example:
//
//	userResult.
//	    Inspect(func(user *User) {
//	        log.Info("user retrieved", "id", user.ID)
//	        metrics.RecordUserFetch("success")
//	    }).
//	    AndThen(validate)
func (r Result[T]) Inspect(f func(T)) Result[T] {
	if r.IsOk() {
		f(r.value)
	}
	return r
}

// InspectErr performs a side effect on the Err without consuming it.
// If the result is Err, the function f is called with the error, and the
// original result is returned unchanged.
// If the result is Ok, nothing happens and the value is returned unchanged.
//
// This is useful for error logging, metrics, or alerting without breaking
// the error propagation chain.
//
// Example:
//
//	userResult.
//	    InspectErr(func(err error) {
//	        log.Error("failed to get user", "error", err)
//	        metrics.RecordUserFetch("error")
//	    }).
//	    OrElse(tryFallback)
func (r Result[T]) InspectErr(f func(error)) Result[T] {
	if r.IsErr() {
		f(r.err)
	}
	return r
}

// Tap performs side effects on both Ok and Err without consuming the result.
// If the result is Ok, onOk is called with the value.
// If the result is Err, onErr is called with the error.
// The original result is returned unchanged in both cases.
//
// This is useful when you want to perform observability actions regardless
// of success or failure.
//
// Example:
//
//	userResult.
//	    Tap(
//	        func(user *User) {
//	            log.Info("success", "user_id", user.ID)
//	        },
//	        func(err error) {
//	            log.Error("failure", "error", err)
//	        },
//	    ).
//	    AndThen(process)
func (r Result[T]) Tap(onOk func(T), onErr func(error)) Result[T] {
	if r.IsOk() {
		onOk(r.value)
	} else {
		onErr(r.err)
	}
	return r
}
