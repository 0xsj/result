package result

// All2 combines two Results into a single Result containing both values.
// If both Results are Ok, returns Ok with a struct containing both values.
// If any Result is Err, returns the first Err encountered (fail-fast).
//
// Example:
//
//	result.All2(
//	    domain.NewEmail(cmd.Email),
//	    domain.NewUsername(cmd.Username),
//	).AndThenMap(func(v struct{V1 domain.Email; V2 domain.Username}) result.Result[*User] {
//	    return createUser(v.V1, v.V2)
//	})
func All2[T1, T2 any](r1 Result[T1], r2 Result[T2]) Result[struct {
	V1 T1
	V2 T2
}] {
	if r1.IsErr() {
		return Result[struct {
			V1 T1
			V2 T2
		}]{
			err:  r1.err,
			op:   r1.op,
			meta: r1.meta,
		}
	}

	if r2.IsErr() {
		return Result[struct {
			V1 T1
			V2 T2
		}]{
			err:  r2.err,
			op:   r2.op,
			meta: r2.meta,
		}
	}

	return Ok(struct {
		V1 T1
		V2 T2
	}{
		V1: r1.value,
		V2: r2.value,
	})
}

// All3 combines three Results into a single Result.
// If all Results are Ok, returns Ok with a struct containing all values.
// If any Result is Err, returns the first Err encountered (fail-fast).
//
// Example:
//
//	result.All3(
//	    domain.NewEmail(cmd.Email),
//	    domain.NewUsername(cmd.Username),
//	    domain.NewPassword(cmd.Password),
//	).AndThenMap(func(v struct{V1 domain.Email; V2 domain.Username; V3 domain.Password}) result.Result[*User] {
//	    user := domain.NewUser(v.V1, v.V2, v.V3)
//	    return result.Ok(user)
//	})
func All3[T1, T2, T3 any](r1 Result[T1], r2 Result[T2], r3 Result[T3]) Result[struct {
	V1 T1
	V2 T2
	V3 T3
}] {
	if r1.IsErr() {
		return Result[struct {
			V1 T1
			V2 T2
			V3 T3
		}]{
			err:  r1.err,
			op:   r1.op,
			meta: r1.meta,
		}
	}
	if r2.IsErr() {
		return Result[struct {
			V1 T1
			V2 T2
			V3 T3
		}]{
			err:  r2.err,
			op:   r2.op,
			meta: r2.meta,
		}
	}
	if r3.IsErr() {
		return Result[struct {
			V1 T1
			V2 T2
			V3 T3
		}]{
			err:  r3.err,
			op:   r3.op,
			meta: r3.meta,
		}
	}

	return Ok(struct {
		V1 T1
		V2 T2
		V3 T3
	}{
		V1: r1.value,
		V2: r2.value,
		V3: r3.value,
	})
}

// All4 combines four Results into a single Result.
// If all Results are Ok, returns Ok with a struct containing all values.
// If any Result is Err, returns the first Err encountered (fail-fast).
func All4[T1, T2, T3, T4 any](
	r1 Result[T1],
	r2 Result[T2],
	r3 Result[T3],
	r4 Result[T4],
) Result[struct {
	V1 T1
	V2 T2
	V3 T3
	V4 T4
}] {
	if r1.IsErr() {
		return Result[struct {
			V1 T1
			V2 T2
			V3 T3
			V4 T4
		}]{
			err:  r1.err,
			op:   r1.op,
			meta: r1.meta,
		}
	}
	if r2.IsErr() {
		return Result[struct {
			V1 T1
			V2 T2
			V3 T3
			V4 T4
		}]{
			err:  r2.err,
			op:   r2.op,
			meta: r2.meta,
		}
	}
	if r3.IsErr() {
		return Result[struct {
			V1 T1
			V2 T2
			V3 T3
			V4 T4
		}]{
			err:  r3.err,
			op:   r3.op,
			meta: r3.meta,
		}
	}
	if r4.IsErr() {
		return Result[struct {
			V1 T1
			V2 T2
			V3 T3
			V4 T4
		}]{
			err:  r4.err,
			op:   r4.op,
			meta: r4.meta,
		}
	}

	return Ok(struct {
		V1 T1
		V2 T2
		V3 T3
		V4 T4
	}{
		V1: r1.value,
		V2: r2.value,
		V3: r3.value,
		V4: r4.value,
	})
}

// All5 combines five Results into a single Result.
func All5[T1, T2, T3, T4, T5 any](
	r1 Result[T1],
	r2 Result[T2],
	r3 Result[T3],
	r4 Result[T4],
	r5 Result[T5],
) Result[struct {
	V1 T1
	V2 T2
	V3 T3
	V4 T4
	V5 T5
}] {
	if r1.IsErr() {
		return Result[struct {
			V1 T1
			V2 T2
			V3 T3
			V4 T4
			V5 T5
		}]{
			err:  r1.err,
			op:   r1.op,
			meta: r1.meta,
		}
	}
	if r2.IsErr() {
		return Result[struct {
			V1 T1
			V2 T2
			V3 T3
			V4 T4
			V5 T5
		}]{
			err:  r2.err,
			op:   r2.op,
			meta: r2.meta,
		}
	}
	if r3.IsErr() {
		return Result[struct {
			V1 T1
			V2 T2
			V3 T3
			V4 T4
			V5 T5
		}]{
			err:  r3.err,
			op:   r3.op,
			meta: r3.meta,
		}
	}
	if r4.IsErr() {
		return Result[struct {
			V1 T1
			V2 T2
			V3 T3
			V4 T4
			V5 T5
		}]{
			err:  r4.err,
			op:   r4.op,
			meta: r4.meta,
		}
	}
	if r5.IsErr() {
		return Result[struct {
			V1 T1
			V2 T2
			V3 T3
			V4 T4
			V5 T5
		}]{
			err:  r5.err,
			op:   r5.op,
			meta: r5.meta,
		}
	}

	return Ok(struct {
		V1 T1
		V2 T2
		V3 T3
		V4 T4
		V5 T5
	}{
		V1: r1.value,
		V2: r2.value,
		V3: r3.value,
		V4: r4.value,
		V5: r5.value,
	})
}

// Zip2 combines two Results using a function.
// If both Results are Ok, the function is called with both values.
// If any Result is Err, returns the first Err encountered.
//
// This is like All2 but applies a transformation immediately.
//
// Example:
//
//	result.Zip2(
//	    parseEmail(input.Email),
//	    parseUsername(input.Username),
//	    func(email Email, username Username) *User {
//	        return NewUser(email, username)
//	    },
//	)
func Zip2[T1, T2, R any](
	r1 Result[T1],
	r2 Result[T2],
	fn func(T1, T2) R,
) Result[R] {
	return AndThenMap(All2(r1, r2), func(v struct {
		V1 T1
		V2 T2
	}) Result[R] {
		return Ok(fn(v.V1, v.V2))
	})
}

// Zip3 combines three Results using a function.
func Zip3[T1, T2, T3, R any](
	r1 Result[T1],
	r2 Result[T2],
	r3 Result[T3],
	fn func(T1, T2, T3) R,
) Result[R] {
	return AndThenMap(All3(r1, r2, r3), func(v struct {
		V1 T1
		V2 T2
		V3 T3
	}) Result[R] {
		return Ok(fn(v.V1, v.V2, v.V3))
	})
}

// Zip4 combines four Results using a function.
func Zip4[T1, T2, T3, T4, R any](
	r1 Result[T1],
	r2 Result[T2],
	r3 Result[T3],
	r4 Result[T4],
	fn func(T1, T2, T3, T4) R,
) Result[R] {
	return AndThenMap(All4(r1, r2, r3, r4), func(v struct {
		V1 T1
		V2 T2
		V3 T3
		V4 T4
	}) Result[R] {
		return Ok(fn(v.V1, v.V2, v.V3, v.V4))
	})
}

// Zip5 combines five Results using a function.
func Zip5[T1, T2, T3, T4, T5, R any](
	r1 Result[T1],
	r2 Result[T2],
	r3 Result[T3],
	r4 Result[T4],
	r5 Result[T5],
	fn func(T1, T2, T3, T4, T5) R,
) Result[R] {
	return AndThenMap(All5(r1, r2, r3, r4, r5), func(v struct {
		V1 T1
		V2 T2
		V3 T3
		V4 T4
		V5 T5
	}) Result[R] {
		return Ok(fn(v.V1, v.V2, v.V3, v.V4, v.V5))
	})
}
