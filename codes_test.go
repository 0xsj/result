package result

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// ============================================================================
// HTTPStatus Tests
// ============================================================================

func TestHTTPStatus_AllKinds(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected int
	}{
		{
			name:     "Domain error",
			err:      Domain("test", "business rule violated"),
			expected: http.StatusUnprocessableEntity, // 422
		},
		{
			name:     "Validation error",
			err:      Validation("test", "invalid input", nil),
			expected: http.StatusBadRequest, // 400
		},
		{
			name:     "NotFound error",
			err:      NotFound("test", "resource"),
			expected: http.StatusNotFound, // 404
		},
		{
			name:     "Conflict error",
			err:      Conflict("test", "resource"),
			expected: http.StatusConflict, // 409
		},
		{
			name:     "Unauthorized error",
			err:      Unauthorized("test", "invalid token"),
			expected: http.StatusUnauthorized, // 401
		},
		{
			name:     "Forbidden error",
			err:      Forbidden("test", "insufficient permissions"),
			expected: http.StatusForbidden, // 403
		},
		{
			name:     "Infrastructure error",
			err:      Infrastructure("test", errors.New("db down")),
			expected: http.StatusServiceUnavailable, // 503
		},
		{
			name:     "Internal error",
			err:      Internal("test", errors.New("panic")),
			expected: http.StatusInternalServerError, // 500
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := HTTPStatus(tt.err)
			if status != tt.expected {
				t.Errorf("HTTPStatus() = %d, want %d", status, tt.expected)
			}
		})
	}
}

func TestHTTPStatus_NilError(t *testing.T) {
	status := HTTPStatus(nil)
	if status != http.StatusInternalServerError {
		t.Errorf("HTTPStatus(nil) = %d, want %d", status, http.StatusInternalServerError)
	}
}

func TestHTTPStatus_StandardError(t *testing.T) {
	// Standard Go error (not result.Error)
	err := errors.New("standard error")
	status := HTTPStatus(err)

	// Should default to Internal (500)
	if status != http.StatusInternalServerError {
		t.Errorf("HTTPStatus(standard error) = %d, want %d", status, http.StatusInternalServerError)
	}
}

// ============================================================================
// HTTPStatusResult Tests
// ============================================================================

func TestHTTPStatusResult_Ok(t *testing.T) {
	r := Ok(42)
	status := HTTPStatusResult(r)

	if status != http.StatusOK {
		t.Errorf("HTTPStatusResult(Ok) = %d, want %d", status, http.StatusOK)
	}
}

func TestHTTPStatusResult_Err(t *testing.T) {
	tests := []struct {
		name     string
		result   Result[int]
		expected int
	}{
		{
			name:     "NotFound",
			result:   Err[int](NotFound("test", "resource")),
			expected: http.StatusNotFound,
		},
		{
			name:     "Validation",
			result:   Err[int](Validation("test", "invalid", nil)),
			expected: http.StatusBadRequest,
		},
		{
			name:     "Internal",
			result:   Err[int](Internal("test", errors.New("error"))),
			expected: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := HTTPStatusResult(tt.result)
			if status != tt.expected {
				t.Errorf("HTTPStatusResult() = %d, want %d", status, tt.expected)
			}
		})
	}
}

// ============================================================================
// HTTPStatusOr Tests
// ============================================================================

func TestHTTPStatusOr_Ok(t *testing.T) {
	tests := []struct {
		name          string
		result        Result[int]
		successStatus int
		expected      int
	}{
		{
			name:          "Ok with 200",
			result:        Ok(42),
			successStatus: http.StatusOK,
			expected:      http.StatusOK,
		},
		{
			name:          "Ok with 201 Created",
			result:        Ok(42),
			successStatus: http.StatusCreated,
			expected:      http.StatusCreated,
		},
		{
			name:          "Ok with 204 No Content",
			result:        Ok(42),
			successStatus: http.StatusNoContent,
			expected:      http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := HTTPStatusOr(tt.result, tt.successStatus)
			if status != tt.expected {
				t.Errorf("HTTPStatusOr() = %d, want %d", status, tt.expected)
			}
		})
	}
}

func TestHTTPStatusOr_Err(t *testing.T) {
	r := Err[int](NotFound("test", "resource"))
	status := HTTPStatusOr(r, http.StatusCreated)

	// Should return error status, not success status
	if status != http.StatusNotFound {
		t.Errorf("HTTPStatusOr(Err) = %d, want %d", status, http.StatusNotFound)
	}
}

// ============================================================================
// WriteHTTPError Tests
// ============================================================================

func TestWriteHTTPError_WithoutMetadata(t *testing.T) {
	err := NotFound("test", "resource")
	w := httptest.NewRecorder()

	WriteHTTPError(w, err)

	if w.Code != http.StatusNotFound {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusNotFound)
	}

	// Should be plain text (no JSON)
	contentType := w.Header().Get("Content-Type")
	if contentType == "application/json" {
		t.Error("Should not set JSON content type for errors without metadata")
	}

	body := w.Body.String()
	if body == "" {
		t.Error("Body should not be empty")
	}
}

func TestWriteHTTPError_WithMetadata(t *testing.T) {
	err := Validation("test", "invalid email", map[string]interface{}{
		"field": "email",
		"value": "invalid",
	})
	w := httptest.NewRecorder()

	WriteHTTPError(w, err)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusBadRequest)
	}

	// Should be JSON
	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("content type = %s, want application/json", contentType)
	}

	// Parse JSON response
	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode JSON response: %v", err)
	}

	// Check response structure
	if response["error"] == nil {
		t.Error("Response should have 'error' field")
	}

	if response["kind"] != "validation" {
		t.Errorf("Response kind = %v, want validation", response["kind"])
	}

	if response["op"] != "test" {
		t.Errorf("Response op = %v, want test", response["op"])
	}

	details, ok := response["details"].(map[string]interface{})
	if !ok {
		t.Fatal("Response should have 'details' map")
	}

	if details["field"] != "email" {
		t.Errorf("details[field] = %v, want email", details["field"])
	}
}

func TestWriteHTTPError_NilMetadata(t *testing.T) {
	err := Domain("test", "business rule violated")
	w := httptest.NewRecorder()

	WriteHTTPError(w, err)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusUnprocessableEntity)
	}

	// Should be plain text (no metadata)
	contentType := w.Header().Get("Content-Type")
	if contentType == "application/json" {
		t.Error("Should not set JSON content type for errors without metadata")
	}
}

func TestWriteHTTPError_EmptyMetadata(t *testing.T) {
	err := Validation("test", "invalid", map[string]interface{}{})
	w := httptest.NewRecorder()

	WriteHTTPError(w, err)

	// Empty metadata is still metadata, should return JSON
	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("content type = %s, want application/json", contentType)
	}
}

func TestWriteHTTPError_AllErrorKinds(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		expectedCode int
	}{
		{"Domain", Domain("test", "msg"), http.StatusUnprocessableEntity},
		{"Validation", Validation("test", "msg", nil), http.StatusBadRequest},
		{"NotFound", NotFound("test", "resource"), http.StatusNotFound},
		{"Conflict", Conflict("test", "resource"), http.StatusConflict},
		{"Unauthorized", Unauthorized("test", "msg"), http.StatusUnauthorized},
		{"Forbidden", Forbidden("test", "msg"), http.StatusForbidden},
		{"Infrastructure", Infrastructure("test", errors.New("db")), http.StatusServiceUnavailable},
		{"Internal", Internal("test", errors.New("panic")), http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			WriteHTTPError(w, tt.err)

			if w.Code != tt.expectedCode {
				t.Errorf("status code = %d, want %d", w.Code, tt.expectedCode)
			}
		})
	}
}

// ============================================================================
// Integration Tests
// ============================================================================

func TestHTTPStatus_WithWrappedError(t *testing.T) {
	original := NotFound("repo", "user")
	wrapped := Wrap(original, "service")

	status := HTTPStatus(wrapped)

	// Should preserve NotFound kind
	if status != http.StatusNotFound {
		t.Errorf("HTTPStatus(wrapped) = %d, want %d", status, http.StatusNotFound)
	}
}

func TestWriteHTTPError_RealWorldResponse(t *testing.T) {
	// Simulate a validation error with rich metadata
	err := Validation("UserService.CreateUser", "invalid email format", map[string]interface{}{
		"field":      "email",
		"value":      "invalid@",
		"constraint": "email",
		"request_id": "req_123",
	})

	w := httptest.NewRecorder()
	WriteHTTPError(w, err)

	// Should be 400
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}

	// Should be JSON
	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	// Verify structure
	if response["error"] == nil {
		t.Error("Missing error field")
	}

	if response["kind"] != "validation" {
		t.Error("Wrong kind")
	}

	if response["op"] != "UserService.CreateUser" {
		t.Error("Wrong op")
	}

	details := response["details"].(map[string]interface{})
	if details["field"] != "email" || details["request_id"] != "req_123" {
		t.Error("Missing or wrong metadata")
	}
}

func TestHTTPStatusOr_CreateEndpoint(t *testing.T) {
	// Simulate create operation
	createResult := Ok(&struct{ ID string }{ID: "user_123"})

	status := HTTPStatusOr(createResult, http.StatusCreated)

	if status != http.StatusCreated {
		t.Errorf("Create success should return 201, got %d", status)
	}
}

func TestHTTPStatusOr_DeleteEndpoint(t *testing.T) {
	// Simulate delete operation
	deleteResult := Ok(struct{}{})

	status := HTTPStatusOr(deleteResult, http.StatusNoContent)

	if status != http.StatusNoContent {
		t.Errorf("Delete success should return 204, got %d", status)
	}
}
