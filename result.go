package result

import "fmt"

// Result represents either a successful value or an error
type Result[T any] struct {
	value T
	err   error
	op    string
	meta  map[string]any
}

// ============================================================================
// Constructors
// ============================================================================

// Ok creates a successful result
func Ok[T any](value T) Result[T] {
	return Result[T]{value: value}
}

// Err creates a failed result
func Err[T any](err error) Result[T] {
	return Result[T]{err: err}
}

// From converts a (T, error) tuple to Result[T]
// This is useful for wrapping standard Go functions
func From[T any](value T, err error) Result[T] {
	if err != nil {
		return Err[T](err)
	}
	return Ok(value)
}

// ============================================================================
// Predicates
// ============================================================================

// IsOk returns true if the result contains a value
func (r Result[T]) IsOk() bool {
	return r.err == nil
}

// IsErr returns true if the result contains an error
func (r Result[T]) IsErr() bool {
	return r.err != nil
}

// IsErrKind checks if the error is of a specific Kind
func (r Result[T]) IsErrKind(kind Kind) bool {
	return r.IsErr() && KindOf(r.err) == kind
}

// ============================================================================
// Extractors
// ============================================================================

// Unwrap returns the value or panics if error
// Use this only when you're certain the result is Ok (e.g., in tests)
func (r Result[T]) Unwrap() T {
	if r.err != nil {
		panic(fmt.Sprintf("called Unwrap on Err result: %v", r.err))
	}
	return r.value
}

// UnwrapOr returns the value or a default if error
func (r Result[T]) UnwrapOr(defaultValue T) T {
	if r.err != nil {
		return defaultValue
	}
	return r.value
}

// UnwrapOrElse returns the value or computes a default from the error
func (r Result[T]) UnwrapOrElse(f func(error) T) T {
	if r.err != nil {
		return f(r.err)
	}
	return r.value
}

// Expect returns the value or panics with a custom message
func (r Result[T]) Expect(msg string) T {
	if r.err != nil {
		panic(fmt.Sprintf("%s: %v", msg, r.err))
	}
	return r.value
}

// Value returns both the value and error (for compatibility with Go idioms)
func (r Result[T]) Value() (T, error) {
	return r.value, r.err
}

// Error returns the underlying error
func (r Result[T]) Error() error {
	return r.err
}

// UnwrapErr returns the underlying error or nil if Ok.
// This is the counterpart to Unwrap() for accessing the error.
func (r Result[T]) UnwrapErr() error {
	return r.err
}

// ============================================================================
// Context Builders (Fluent API)
// ============================================================================

// WithOp adds operation context for observability
func (r Result[T]) WithOp(op string) Result[T] {
	r.op = op
	return r
}

// WithMeta adds a metadata key-value pair
func (r Result[T]) WithMeta(key string, value any) Result[T] {
	if r.meta == nil {
		r.meta = make(map[string]any)
	}
	r.meta[key] = value
	return r
}

// WithMetaMap adds multiple metadata entries
func (r Result[T]) WithMetaMap(meta map[string]any) Result[T] {
	if r.meta == nil {
		r.meta = make(map[string]any)
	}
	for k, v := range meta {
		r.meta[k] = v
	}
	return r
}

// ============================================================================
// Transformers
// ============================================================================

// Map transforms the Ok value, propagates Err unchanged
func (r Result[T]) Map(f func(T) T) Result[T] {
	if r.IsErr() {
		return r
	}
	return Result[T]{
		value: f(r.value),
		op:    r.op,
		meta:  r.meta,
	}
}

// MapErr transforms the Err, propagates Ok unchanged
func (r Result[T]) MapErr(f func(error) error) Result[T] {
	if r.IsOk() {
		return r
	}
	return Result[T]{
		err:  f(r.err),
		op:   r.op,
		meta: r.meta,
	}
}

// MapValue transforms the Ok value to a different type
func MapValue[T, U any](r Result[T], f func(T) U) Result[U] {
	if r.IsErr() {
		return Result[U]{
			err:  r.err,
			op:   r.op,
			meta: r.meta,
		}
	}
	return Result[U]{
		value: f(r.value),
		op:    r.op,
		meta:  r.meta,
	}
}
