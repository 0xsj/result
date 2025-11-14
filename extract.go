package result

// Extract safely extracts a value from Result[T] or returns a wrapped error.
// This simulates Rust's ? operator for extracting values or propagating errors.
//
// Returns (value, nil) if Result is Ok.
// Returns (zero, wrappedError) if Result is Err.
//
// Example:
//
//	email, err := result.Extract(domain.NewEmail(cmd.Email), "create_user.validate_email")
//	if err != nil {
//	    return result.Err[UserDTO](err)
//	}
//	// email is safely extracted, no panic possible
func Extract[T any](r Result[T], op string) (T, error) {
	if r.IsErr() {
		var zero T
		return zero, Wrap(r.UnwrapErr(), op)
	}
	return r.Unwrap(), nil
}

// ExtractOr safely extracts a value from Result[T] or returns a default.
// No error is returned - useful when you have a sensible fallback.
//
// This is equivalent to Result.UnwrapOr but with a more explicit name.
//
// Example:
//
//	count := result.ExtractOr(repo.Count(ctx), 0)
//	config := result.ExtractOr(loadOptionalConfig(), defaultConfig)
func ExtractOr[T any](r Result[T], defaultValue T) T {
	return r.UnwrapOr(defaultValue)
}

// Must extracts a value from Result[T] or panics with a descriptive message.
// Only use this in initialization code or when failure is truly unrecoverable.
//
// This is similar to Result.Expect but with a more explicit name.
//
// Example:
//
//	config := result.Must(loadConfig(), "failed to load required config")
//	db := result.Must(connectDB(), "database connection is required")
func Must[T any](r Result[T], msg string) T {
	return r.Expect(msg)
}

// ValidateAndCheck runs validation and a uniqueness check, returning the validated value.
// This encapsulates the common pattern: validate format, check if exists, return error if conflict.
//
// The validate Result is checked first.
// If Ok, the check function is called with the validated value.
// If check returns Ok(true), conflictErr is called and wrapped.
// If check returns Ok(false), the validated value is returned.
// If check returns Err, that error is wrapped and returned.
//
// Example:
//
//	emailResult := result.ValidateAndCheck(
//	    domain.NewEmail(cmd.Email),
//	    func(e domain.Email) result.Result[bool] {
//	        return repo.ExistsByEmail(ctx, e)
//	    },
//	    func(e domain.Email) error {
//	        return domain.ErrEmailAlreadyExists{Email: e.Value()}
//	    },
//	    "create_user.email",
//	)
func ValidateAndCheck[T any](
	validate Result[T],
	check func(T) Result[bool],
	conflictErr func(T) error,
	op string,
) Result[T] {
	// First, check if validation passed
	if validate.IsErr() {
		return Err[T](Wrap(validate.UnwrapErr(), op+".validate"))
	}
	value := validate.Unwrap()

	// Run the check
	checkResult := check(value)
	if checkResult.IsErr() {
		return Err[T](Wrap(checkResult.UnwrapErr(), op+".check"))
	}

	// If check returns true, it means conflict (e.g., already exists)
	if checkResult.Unwrap() {
		return Err[T](Wrap(conflictErr(value), op+".conflict"))
	}

	return Ok(value)
}

// ValidateWith runs validation and returns the validated value or wrapped error.
// This is a convenience wrapper that just adds operation context to validation.
//
// Example:
//
//	email, err := result.ValidateWith(
//	    domain.NewEmail(cmd.Email),
//	    "create_user.email",
//	)
//	if err != nil {
//	    return result.Err[UserDTO](err)
//	}
func ValidateWith[T any](validate Result[T], op string) (T, error) {
	return Extract(validate, op)
}

// CheckCondition wraps a boolean check into a Result.
// If condition is true, returns Err with the provided error (wrapped with op).
// If condition is false, returns Ok with the provided value.
//
// This is useful for existence checks or business rule validations.
//
// Example:
//
//	result.CheckCondition(
//	    emailExists,
//	    email,
//	    domain.ErrEmailAlreadyExists{Email: email.Value()},
//	    "create_user.check_email",
//	)
func CheckCondition[T any](condition bool, value T, err error, op string) Result[T] {
	if condition {
		return Err[T](Wrap(err, op))
	}
	return Ok(value)
}

// Validate3 validates three values in sequence and returns all three or the first error.
// This is a specialized helper for the common create user pattern.
//
// Example:
//
//	result.Validate3(
//	    domain.NewEmail(cmd.Email),
//	    domain.NewUsername(cmd.Username),
//	    domain.NewPassword(cmd.Password),
//	    "create_user.validate",
//	).AndThenMap(func(v struct{V1 Email; V2 Username; V3 Password}) result.Result[*User] {
//	    return result.Ok(domain.NewUser(v.V1, v.V2, v.V3))
//	})
func Validate3[T1, T2, T3 any](
	r1 Result[T1],
	r2 Result[T2],
	r3 Result[T3],
	op string,
) Result[struct {
	V1 T1
	V2 T2
	V3 T3
}] {
	all := All3(r1, r2, r3)
	if all.IsErr() {
		return Err[struct {
			V1 T1
			V2 T2
			V3 T3
		}](Wrap(all.UnwrapErr(), op))
	}
	return all
}
