package result

import "fmt"

// Try wraps a function that might panic into a Result.
// If the function panics, the panic is caught and returned as an error.
// If the function completes successfully, its return value is wrapped in Ok.
//
// This is useful for calling functions that might panic or working with
// third-party code.
//
// Example:
//
//	result := result.Try(func() *User {
//	    return parseUser(jsonData)  // Might panic
//	})
func Try[T any](fn func() T) Result[T] {
	var value T
	var err error

	func() {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("panic: %v", r)
			}
		}()
		value = fn()
	}()

	if err != nil {
		return Err[T](err)
	}
	return Ok(value)
}

// TryWith wraps a function that returns (T, error) into a Result.
// This is an alias for From but with a more explicit name.
//
// Example:
//
//	result := result.TryWith(func() (*User, error) {
//	    return repo.FindByID(ctx, id)
//	})
func TryWith[T any](fn func() (T, error)) Result[T] {
	value, err := fn()
	return From(value, err)
}
