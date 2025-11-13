package result

import (
	"errors"
	"fmt"
	"testing"
)

// ============================================================================
// Match Tests
// ============================================================================

func TestMatch_Ok(t *testing.T) {
	okCalled := false
	errCalled := false
	var capturedValue int

	Ok(42).Match(
		func(n int) {
			okCalled = true
			capturedValue = n
		},
		func(err error) {
			errCalled = true
		},
	)

	if !okCalled {
		t.Error("Match should call onOk for Ok result")
	}

	if errCalled {
		t.Error("Match should not call onErr for Ok result")
	}

	if capturedValue != 42 {
		t.Errorf("Match captured value = %v, want 42", capturedValue)
	}
}

func TestMatch_Err(t *testing.T) {
	okCalled := false
	errCalled := false
	testErr := errors.New("test error")
	var capturedErr error

	Err[int](testErr).Match(
		func(n int) {
			okCalled = true
		},
		func(err error) {
			errCalled = true
			capturedErr = err
		},
	)

	if okCalled {
		t.Error("Match should not call onOk for Err result")
	}

	if !errCalled {
		t.Error("Match should call onErr for Err result")
	}

	if capturedErr != testErr {
		t.Errorf("Match captured error = %v, want %v", capturedErr, testErr)
	}
}

func TestMatch_SideEffects(t *testing.T) {
	// Test that Match is useful for side effects
	called := 0

	Ok(10).Match(
		func(n int) {
			called++
		},
		func(err error) {
			called++
		},
	)

	if called != 1 {
		t.Errorf("Match should call handler once, called %d times", called)
	}
}

// ============================================================================
// MatchValue Tests
// ============================================================================

func TestMatchValue_Ok(t *testing.T) {
	result := MatchValue(
		Ok(42),
		func(n int) string {
			return fmt.Sprintf("number: %d", n)
		},
		func(err error) string {
			return "error"
		},
	)

	expected := "number: 42"
	if result != expected {
		t.Errorf("MatchValue result = %v, want %v", result, expected)
	}
}

func TestMatchValue_Err(t *testing.T) {
	testErr := errors.New("test error")
	result := MatchValue(
		Err[int](testErr),
		func(n int) string {
			return "success"
		},
		func(err error) string {
			return fmt.Sprintf("error: %v", err)
		},
	)

	expected := "error: test error"
	if result != expected {
		t.Errorf("MatchValue result = %v, want %v", result, expected)
	}
}

func TestMatchValue_StatusCode(t *testing.T) {
	// Practical example: converting to HTTP status code
	okCode := MatchValue(
		Ok("user"),
		func(s string) int { return 200 },
		func(err error) int { return 500 },
	)

	if okCode != 200 {
		t.Errorf("MatchValue status = %v, want 200", okCode)
	}

	errCode := MatchValue(
		Err[string](errors.New("error")),
		func(s string) int { return 200 },
		func(err error) int { return 500 },
	)

	if errCode != 500 {
		t.Errorf("MatchValue status = %v, want 500", errCode)
	}
}

func TestMatchValue_TypeTransformation(t *testing.T) {
	// int -> bool
	result := MatchValue(
		Ok(42),
		func(n int) bool { return n > 0 },
		func(err error) bool { return false },
	)

	if !result {
		t.Error("MatchValue should return true")
	}
}

func TestMatchValue_StructTransformation(t *testing.T) {
	type Response struct {
		Success bool
		Message string
	}

	okResponse := MatchValue(
		Ok("data"),
		func(s string) Response {
			return Response{Success: true, Message: s}
		},
		func(err error) Response {
			return Response{Success: false, Message: err.Error()}
		},
	)

	if !okResponse.Success {
		t.Error("MatchValue should create success response")
	}

	if okResponse.Message != "data" {
		t.Errorf("MatchValue message = %v, want data", okResponse.Message)
	}
}

// ============================================================================
// MatchKind Tests
// ============================================================================

func TestMatchKind_Ok(t *testing.T) {
	called := false

	Ok(42).MatchKind(
		map[Kind]func(error){
			KindNotFound: func(e error) {
				called = true
			},
		},
		func(e error) {
			called = true
		},
	)

	if called {
		t.Error("MatchKind should not call any handler for Ok result")
	}
}

func TestMatchKind_MatchingCase(t *testing.T) {
	notFoundCalled := false
	validationCalled := false
	defaultCalled := false

	Err[int](NotFound("test", "user")).MatchKind(
		map[Kind]func(error){
			KindNotFound: func(e error) {
				notFoundCalled = true
			},
			KindValidation: func(e error) {
				validationCalled = true
			},
		},
		func(e error) {
			defaultCalled = true
		},
	)

	if !notFoundCalled {
		t.Error("MatchKind should call KindNotFound handler")
	}

	if validationCalled {
		t.Error("MatchKind should not call KindValidation handler")
	}

	if defaultCalled {
		t.Error("MatchKind should not call default handler when case matches")
	}
}

func TestMatchKind_DefaultCase(t *testing.T) {
	notFoundCalled := false
	defaultCalled := false

	// Error is KindValidation, but only KindNotFound handler provided
	Err[int](Validation("test", "invalid", nil)).MatchKind(
		map[Kind]func(error){
			KindNotFound: func(e error) {
				notFoundCalled = true
			},
		},
		func(e error) {
			defaultCalled = true
		},
	)

	if notFoundCalled {
		t.Error("MatchKind should not call non-matching handler")
	}

	if !defaultCalled {
		t.Error("MatchKind should call default handler when no case matches")
	}
}

func TestMatchKind_NoDefaultCase(t *testing.T) {
	called := false

	// No matching case and no default - should not panic
	Err[int](Validation("test", "invalid", nil)).MatchKind(
		map[Kind]func(error){
			KindNotFound: func(e error) {
				called = true
			},
		},
		nil, // No default case
	)

	if called {
		t.Error("MatchKind should not call handler")
	}

	// Test passes if it doesn't panic
}

func TestMatchKind_AllErrorKinds(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected Kind
	}{
		{"Domain", Domain("test", "msg"), KindDomain},
		{"Validation", Validation("test", "msg", nil), KindValidation},
		{"NotFound", NotFound("test", "resource"), KindNotFound},
		{"Conflict", Conflict("test", "resource"), KindConflict},
		{"Unauthorized", Unauthorized("test", "msg"), KindUnauthorized},
		{"Forbidden", Forbidden("test", "msg"), KindForbidden},
		{"Infrastructure", Infrastructure("test", errors.New("db")), KindInfrastructure},
		{"Internal", Internal("test", errors.New("panic")), KindInternal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched := false

			Err[int](tt.err).MatchKind(
				map[Kind]func(error){
					tt.expected: func(e error) {
						matched = true
					},
				},
				func(e error) {
					t.Errorf("MatchKind should not call default for %s", tt.name)
				},
			)

			if !matched {
				t.Errorf("MatchKind should match %s", tt.name)
			}
		})
	}
}

func TestMatchKind_ErrorDetails(t *testing.T) {
	testErr := NotFound("UserService.GetUser", "user")
	var capturedErr error

	Err[int](testErr).MatchKind(
		map[Kind]func(error){
			KindNotFound: func(e error) {
				capturedErr = e
			},
		},
		nil,
	)

	if capturedErr != testErr {
		t.Error("MatchKind should pass original error to handler")
	}

	if OpOf(capturedErr) != "UserService.GetUser" {
		t.Error("MatchKind should preserve error context")
	}
}

// ============================================================================
// MatchKindValue Tests
// ============================================================================

func TestMatchKindValue_Ok(t *testing.T) {
	result := MatchKindValue(
		Ok(42),
		func(n int) string {
			return fmt.Sprintf("success: %d", n)
		},
		map[Kind]func(error) string{
			KindNotFound: func(e error) string {
				return "not found"
			},
		},
		func(e error) string {
			return "error"
		},
	)

	expected := "success: 42"
	if result != expected {
		t.Errorf("MatchKindValue result = %v, want %v", result, expected)
	}
}

func TestMatchKindValue_MatchingCase(t *testing.T) {
	result := MatchKindValue(
		Err[int](NotFound("test", "user")),
		func(n int) string {
			return "success"
		},
		map[Kind]func(error) string{
			KindNotFound: func(e error) string {
				return "user not found"
			},
			KindValidation: func(e error) string {
				return "validation failed"
			},
		},
		func(e error) string {
			return "error"
		},
	)

	expected := "user not found"
	if result != expected {
		t.Errorf("MatchKindValue result = %v, want %v", result, expected)
	}
}

func TestMatchKindValue_DefaultCase(t *testing.T) {
	result := MatchKindValue(
		Err[int](Domain("test", "business rule violated")),
		func(n int) string {
			return "success"
		},
		map[Kind]func(error) string{
			KindNotFound: func(e error) string {
				return "not found"
			},
		},
		func(e error) string {
			return "default error"
		},
	)

	expected := "default error"
	if result != expected {
		t.Errorf("MatchKindValue result = %v, want %v", result, expected)
	}
}

func TestMatchKindValue_NoDefaultCase(t *testing.T) {
	result := MatchKindValue(
		Err[int](Domain("test", "msg")),
		func(n int) string {
			return "success"
		},
		map[Kind]func(error) string{
			KindNotFound: func(e error) string {
				return "not found"
			},
		},
		nil, // No default case
	)

	// Should return zero value
	if result != "" {
		t.Errorf("MatchKindValue result = %v, want empty string", result)
	}
}

func TestMatchKindValue_HTTPStatusCodes(t *testing.T) {
	// Practical example: comprehensive HTTP status mapping
	tests := []struct {
		name     string
		result   Result[string]
		expected int
	}{
		{
			name:     "Ok",
			result:   Ok("data"),
			expected: 200,
		},
		{
			name:     "NotFound",
			result:   Err[string](NotFound("test", "resource")),
			expected: 404,
		},
		{
			name:     "Validation",
			result:   Err[string](Validation("test", "invalid", nil)),
			expected: 400,
		},
		{
			name:     "Conflict",
			result:   Err[string](Conflict("test", "resource")),
			expected: 409,
		},
		{
			name:     "Unauthorized",
			result:   Err[string](Unauthorized("test", "msg")),
			expected: 401,
		},
		{
			name:     "Forbidden",
			result:   Err[string](Forbidden("test", "msg")),
			expected: 403,
		},
		{
			name:     "Domain",
			result:   Err[string](Domain("test", "msg")),
			expected: 422,
		},
		{
			name:     "Infrastructure",
			result:   Err[string](Infrastructure("test", errors.New("db"))),
			expected: 503,
		},
		{
			name:     "Internal",
			result:   Err[string](Internal("test", errors.New("panic"))),
			expected: 500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statusCode := MatchKindValue(
				tt.result,
				func(s string) int { return 200 },
				map[Kind]func(error) int{
					KindNotFound:       func(e error) int { return 404 },
					KindValidation:     func(e error) int { return 400 },
					KindConflict:       func(e error) int { return 409 },
					KindUnauthorized:   func(e error) int { return 401 },
					KindForbidden:      func(e error) int { return 403 },
					KindDomain:         func(e error) int { return 422 },
					KindInfrastructure: func(e error) int { return 503 },
					KindInternal:       func(e error) int { return 500 },
				},
				func(e error) int { return 500 },
			)

			if statusCode != tt.expected {
				t.Errorf("status code = %v, want %v", statusCode, tt.expected)
			}
		})
	}
}

func TestMatchKindValue_APIResponse(t *testing.T) {
	type APIResponse struct {
		Success bool   `json:"success"`
		Data    string `json:"data,omitempty"`
		Error   string `json:"error,omitempty"`
		Code    string `json:"code,omitempty"`
	}

	okResponse := MatchKindValue(
		Ok("user data"),
		func(s string) APIResponse {
			return APIResponse{
				Success: true,
				Data:    s,
			}
		},
		map[Kind]func(error) APIResponse{
			KindNotFound: func(e error) APIResponse {
				return APIResponse{
					Success: false,
					Code:    "NOT_FOUND",
					Error:   e.Error(),
				}
			},
		},
		func(e error) APIResponse {
			return APIResponse{
				Success: false,
				Code:    "ERROR",
				Error:   e.Error(),
			}
		},
	)

	if !okResponse.Success {
		t.Error("APIResponse should be success")
	}

	if okResponse.Data != "user data" {
		t.Errorf("APIResponse data = %v, want user data", okResponse.Data)
	}

	errResponse := MatchKindValue(
		Err[string](NotFound("test", "user")),
		func(s string) APIResponse {
			return APIResponse{Success: true, Data: s}
		},
		map[Kind]func(error) APIResponse{
			KindNotFound: func(e error) APIResponse {
				return APIResponse{
					Success: false,
					Code:    "NOT_FOUND",
					Error:   "Resource not found",
				}
			},
		},
		func(e error) APIResponse {
			return APIResponse{Success: false, Code: "ERROR"}
		},
	)

	if errResponse.Success {
		t.Error("APIResponse should not be success")
	}

	if errResponse.Code != "NOT_FOUND" {
		t.Errorf("APIResponse code = %v, want NOT_FOUND", errResponse.Code)
	}
}

// ============================================================================
// Integration Tests
// ============================================================================

func TestMatching_ChainedUsage(t *testing.T) {
	// Test using Match after a chain
	callCount := 0

	Ok(10).
		Map(func(n int) int { return n * 2 }).
		Map(func(n int) int { return n + 5 }).
		Match(
			func(n int) {
				callCount++
				if n != 25 {
					t.Errorf("Expected 25, got %d", n)
				}
			},
			func(err error) {
				t.Error("Should not call error handler")
			},
		)

	if callCount != 1 {
		t.Errorf("Match should be called once, called %d times", callCount)
	}
}

func TestMatching_CombinedWithAndThen(t *testing.T) {
	result := Ok(10).
		AndThen(func(n int) Result[int] {
			if n < 0 {
				return Err[int](Validation("test", "negative", nil))
			}
			return Ok(n * 2)
		})

	statusCode := MatchValue(
		result,
		func(n int) int { return 200 },
		func(err error) int { return 400 },
	)

	if statusCode != 200 {
		t.Errorf("Expected 200, got %d", statusCode)
	}
}

func TestMatching_MultiplePatternMatches(t *testing.T) {
	// Using multiple pattern matching styles on the same result
	r := Err[int](NotFound("test", "user"))

	// Style 1: Match
	matchCalled := false
	r.Match(
		func(n int) { t.Error("Should not call onOk") },
		func(err error) { matchCalled = true },
	)

	if !matchCalled {
		t.Error("Match should be called")
	}

	// Style 2: MatchKind
	matchKindCalled := false
	r.MatchKind(
		map[Kind]func(error){
			KindNotFound: func(e error) { matchKindCalled = true },
		},
		nil,
	)

	if !matchKindCalled {
		t.Error("MatchKind should be called")
	}

	// Style 3: MatchValue
	value := MatchValue(
		r,
		func(n int) string { return "ok" },
		func(err error) string { return "error" },
	)

	if value != "error" {
		t.Errorf("MatchValue = %v, want error", value)
	}

	// Style 4: MatchKindValue
	code := MatchKindValue(
		r,
		func(n int) int { return 200 },
		map[Kind]func(error) int{
			KindNotFound: func(e error) int { return 404 },
		},
		func(e error) int { return 500 },
	)

	if code != 404 {
		t.Errorf("MatchKindValue = %v, want 404", code)
	}
}
