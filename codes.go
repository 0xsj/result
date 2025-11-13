package result

import (
	"encoding/json"
	"net/http"
)

// HTTPStatus returns the appropriate HTTP status code for an error.
// If the error is a result.Error, the status is determined by its Kind.
// If the error is not a result.Error or is nil, it returns 500 Internal Server Error.
//
// This is a convenience function for HTTP handlers to map errors to status codes.
//
// Example:
//
//	userResult.Match(
//	    func(user *User) {
//	        w.WriteHeader(http.StatusOK)
//	        json.NewEncoder(w).Encode(user)
//	    },
//	    func(err error) {
//	        status := result.HTTPStatus(err)
//	        http.Error(w, err.Error(), status)
//	    },
//	)
func HTTPStatus(err error) int {
	if err == nil {
		return http.StatusInternalServerError
	}

	kind := KindOf(err)

	switch kind {
	case KindDomain:
		return http.StatusUnprocessableEntity // 422
	case KindValidation:
		return http.StatusBadRequest // 400
	case KindNotFound:
		return http.StatusNotFound // 404
	case KindConflict:
		return http.StatusConflict // 409
	case KindUnauthorized:
		return http.StatusUnauthorized // 401
	case KindForbidden:
		return http.StatusForbidden // 403
	case KindInfrastructure:
		return http.StatusServiceUnavailable // 503
	case KindInternal:
		return http.StatusInternalServerError // 500
	default:
		return http.StatusInternalServerError // 500
	}
}

// HTTPStatusResult returns the HTTP status code for a Result.
// If the Result is Ok, it returns 200 OK.
// If the Result is Err, it returns the appropriate status code for the error.
//
// This is useful when you want to determine the status code without
// using pattern matching.
//
// Example:
//
//	userResult := service.GetUser(ctx, id)
//	status := result.HTTPStatusResult(userResult)
//	w.WriteHeader(status)
func HTTPStatusResult[T any](r Result[T]) int {
	if r.IsOk() {
		return http.StatusOK
	}
	return HTTPStatus(r.err)
}

// HTTPStatusOr returns the HTTP status code for a Result, with a custom
// success status code.
//
// If the Result is Ok, it returns the provided successStatus.
// If the Result is Err, it returns the appropriate status code for the error.
//
// This is useful for operations that don't return 200 OK on success
// (e.g., 201 Created, 204 No Content).
//
// Example:
//
//	createResult := service.CreateUser(ctx, data)
//	status := result.HTTPStatusOr(createResult, http.StatusCreated)
//	w.WriteHeader(status)
func HTTPStatusOr[T any](r Result[T], successStatus int) int {
	if r.IsOk() {
		return successStatus
	}
	return HTTPStatus(r.err)
}

// WriteHTTPError is a convenience function that writes an error response
// with the appropriate HTTP status code and error message.
//
// If the error has metadata, it includes it in a structured JSON response.
// Otherwise, it uses http.Error for a plain text response.
//
// Example:
//
//	userResult.Match(
//	    func(user *User) {
//	        json.NewEncoder(w).Encode(user)
//	    },
//	    func(err error) {
//	        result.WriteHTTPError(w, err)
//	    },
//	)
func WriteHTTPError(w http.ResponseWriter, err error) {
	status := HTTPStatus(err)
	meta := MetaOf(err)

	// If error has metadata, write structured JSON response
	if meta != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)

		response := map[string]any{
			"error":   err.Error(),
			"kind":    KindOf(err).String(),
			"op":      OpOf(err),
			"details": meta,
		}

		_ = json.NewEncoder(w).Encode(response)
		return
	}

	// Simple text error response
	http.Error(w, err.Error(), status)
}
