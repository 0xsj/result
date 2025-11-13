package result

// Match provides basic pattern matching with side effects.
// If the result is Ok, onOk is called with the value.
// If the result is Err, onErr is called with the error.
//
// This is useful for handling results at boundaries (like HTTP handlers)
// where you need to perform side effects based on success or failure.
//
// Example:
//
//	userResult.Match(
//	    func(user *User) {
//	        json.NewEncoder(w).Encode(user)
//	    },
//	    func(err error) {
//	        http.Error(w, err.Error(), 500)
//	    },
//	)
func (r Result[T]) Match(onOk func(T), onErr func(error)) {
	if r.IsOk() {
		onOk(r.value)
	} else {
		onErr(r.err)
	}
}

// MatchValue transforms Result into a single value (fold/catamorphism).
// If the result is Ok, onOk is called with the value and its return is returned.
// If the result is Err, onErr is called with the error and its return is returned.
//
// This is useful when you need to convert a Result into a single type,
// such as an HTTP status code or a response object.
//
// Example:
//
//	statusCode := result.MatchValue(
//	    userResult,
//	    func(user *User) int { return 200 },
//	    func(err error) int { return 500 },
//	)
func MatchValue[T, U any](r Result[T], onOk func(T) U, onErr func(error) U) U {
	if r.IsOk() {
		return onOk(r.value)
	}
	return onErr(r.err)
}

// MatchKind pattern matches on error kinds with handlers.
// If the result is Ok, nothing happens.
// If the result is Err, the error's Kind is extracted and matched against the cases map.
// If a matching case is found, that handler is called.
// Otherwise, the defaultCase handler is called (if provided).
//
// This is particularly useful for HTTP error handling where you want different
// responses based on error type.
//
// Example:
//
//	userResult.MatchKind(
//	    map[result.Kind]func(error){
//	        result.KindNotFound: func(e error) {
//	            http.Error(w, "Not found", 404)
//	        },
//	        result.KindValidation: func(e error) {
//	            http.Error(w, e.Error(), 400)
//	        },
//	    },
//	    func(e error) {
//	        http.Error(w, "Internal error", 500)
//	    },
//	)
func (r Result[T]) MatchKind(cases map[Kind]func(error), defaultCase func(error)) {
	if r.IsErr() {
		kind := KindOf(r.err)
		if handler, ok := cases[kind]; ok {
			handler(r.err)
			return
		}
		if defaultCase != nil {
			defaultCase(r.err)
		}
	}
}

// MatchKindValue pattern matches on error kinds and returns a value.
// This combines MatchValue and MatchKind - it can handle both Ok and Err results,
// and for Err results, it matches on the error Kind.
//
// If the result is Ok, onOk is called and its return value is returned.
// If the result is Err, the error's Kind is matched against cases.
// If a match is found, that handler is called and its return value is returned.
// Otherwise, defaultCase is called and its return value is returned.
//
// This is useful for creating structured API responses that differ based on
// both success/failure and error type.
//
// Example:
//
//	response := result.MatchKindValue(
//	    userResult,
//	    func(user *User) Response {
//	        return Response{Success: true, Data: user}
//	    },
//	    map[result.Kind]func(error) Response{
//	        result.KindNotFound: func(e error) Response {
//	            return Response{Success: false, Error: "not_found"}
//	        },
//	        result.KindValidation: func(e error) Response {
//	            return Response{Success: false, Error: "validation_failed"}
//	        },
//	    },
//	    func(e error) Response {
//	        return Response{Success: false, Error: "internal_error"}
//	    },
//	)
func MatchKindValue[T, U any](
	r Result[T],
	onOk func(T) U,
	cases map[Kind]func(error) U,
	defaultCase func(error) U,
) U {
	if r.IsOk() {
		return onOk(r.value)
	}

	kind := KindOf(r.err)
	if handler, ok := cases[kind]; ok {
		return handler(r.err)
	}

	if defaultCase != nil {
		return defaultCase(r.err)
	}

	// Should never reach here if defaultCase is provided
	var zero U
	return zero
}
