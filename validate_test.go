package result

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

// ============================================================================
// MultiError Tests
// ============================================================================

func TestMultiError_Error_NoErrors(t *testing.T) {
	me := MultiError{Errors: []error{}}

	if me.Error() != "no errors" {
		t.Errorf("MultiError.Error() = %q, want %q", me.Error(), "no errors")
	}
}

func TestMultiError_Error_SingleError(t *testing.T) {
	me := MultiError{Errors: []error{errors.New("single error")}}

	if me.Error() != "single error" {
		t.Errorf("MultiError.Error() = %q, want %q", me.Error(), "single error")
	}
}

func TestMultiError_Error_MultipleErrors(t *testing.T) {
	me := MultiError{Errors: []error{
		errors.New("error 1"),
		errors.New("error 2"),
		errors.New("error 3"),
	}}

	msg := me.Error()
	if !strings.Contains(msg, "3 validation errors") {
		t.Errorf("MultiError.Error() should contain count")
	}
	if !strings.Contains(msg, "error 1") || !strings.Contains(msg, "error 2") || !strings.Contains(msg, "error 3") {
		t.Errorf("MultiError.Error() should contain all error messages")
	}
}

func TestMultiError_Unwrap(t *testing.T) {
	err1 := errors.New("first error")
	me := MultiError{Errors: []error{err1, errors.New("second error")}}

	unwrapped := me.Unwrap()
	if !errors.Is(unwrapped, err1) {
		t.Error("MultiError.Unwrap() should return first error")
	}
}

// ============================================================================
// ValidateAll Tests
// ============================================================================

func TestValidateAll_AllSucceed(t *testing.T) {
	result := ValidateAll(
		Ok(1),
		Ok(2),
		Ok(3),
	)

	values := AssertOk(t, result)
	if len(values) != 3 {
		t.Fatalf("Expected 3 values, got %d", len(values))
	}

	for i, v := range values {
		if v != i+1 {
			t.Errorf("Value %d = %v, want %v", i, v, i+1)
		}
	}
}

func TestValidateAll_OneFails(t *testing.T) {
	result := ValidateAll(
		Ok(1),
		Err[int](errors.New("error 2")),
		Ok(3),
	)

	err := AssertErr(t, result)

	var me MultiError
	if !errors.As(err, &me) {
		t.Fatal("Error should be MultiError")
	}

	if len(me.Errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(me.Errors))
	}
}

func TestValidateAll_MultipleFail(t *testing.T) {
	result := ValidateAll(
		Err[int](errors.New("error 1")),
		Ok(2),
		Err[int](errors.New("error 3")),
	)

	err := AssertErr(t, result)

	var me MultiError
	if !errors.As(err, &me) {
		t.Fatal("Error should be MultiError")
	}

	if len(me.Errors) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(me.Errors))
	}
}

func TestValidateAll_AllFail(t *testing.T) {
	result := ValidateAll(
		Err[int](errors.New("error 1")),
		Err[int](errors.New("error 2")),
		Err[int](errors.New("error 3")),
	)

	err := AssertErr(t, result)

	var me MultiError
	if !errors.As(err, &me) {
		t.Fatal("Error should be MultiError")
	}

	if len(me.Errors) != 3 {
		t.Errorf("Expected 3 errors, got %d", len(me.Errors))
	}
}

// ============================================================================
// ValidateAllMap Tests
// ============================================================================

func TestValidateAllMap_AllSucceed(t *testing.T) {
	values, errs := ValidateAllMap(
		Ok(1),
		Ok(2),
		Ok(3),
	)

	if len(values) != 3 {
		t.Errorf("Expected 3 values, got %d", len(values))
	}
	if len(errs) != 0 {
		t.Errorf("Expected 0 errors, got %d", len(errs))
	}
}

func TestValidateAllMap_SomeFail(t *testing.T) {
	values, errs := ValidateAllMap(
		Ok(1),
		Err[int](errors.New("error 2")),
		Ok(3),
	)

	if len(values) != 2 {
		t.Errorf("Expected 2 values, got %d", len(values))
	}
	if len(errs) != 1 {
		t.Errorf("Expected 1 error, got %d", len(errs))
	}
}

// ============================================================================
// ValidateRules Tests
// ============================================================================

func TestValidateRules_AllPass(t *testing.T) {
	result := ValidateRules(
		10,
		func(n int) error {
			if n < 0 {
				return errors.New("must be positive")
			}
			return nil
		},
		func(n int) error {
			if n > 100 {
				return errors.New("must be <= 100")
			}
			return nil
		},
	)

	value := AssertOk(t, result)
	if value != 10 {
		t.Errorf("ValidateRules() = %v, want 10", value)
	}
}

func TestValidateRules_FirstFails(t *testing.T) {
	result := ValidateRules(
		-5,
		func(n int) error {
			if n < 0 {
				return errors.New("must be positive")
			}
			return nil
		},
		func(n int) error {
			if n > 100 {
				return errors.New("must be <= 100")
			}
			return nil
		},
	)

	AssertErr(t, result)
	AssertErrContains(t, result, "must be positive")
}

func TestValidateRules_SecondFails(t *testing.T) {
	result := ValidateRules(
		150,
		func(n int) error {
			if n < 0 {
				return errors.New("must be positive")
			}
			return nil
		},
		func(n int) error {
			if n > 100 {
				return errors.New("must be <= 100")
			}
			return nil
		},
	)

	AssertErr(t, result)
	AssertErrContains(t, result, "must be <= 100")
}

// ============================================================================
// ValidateAllRules Tests
// ============================================================================

func TestValidateAllRules_AllPass(t *testing.T) {
	result := ValidateAllRules(
		10,
		func(n int) error {
			if n < 0 {
				return errors.New("must be positive")
			}
			return nil
		},
		func(n int) error {
			if n > 100 {
				return errors.New("must be <= 100")
			}
			return nil
		},
	)

	value := AssertOk(t, result)
	if value != 10 {
		t.Errorf("ValidateAllRules() = %v, want 10", value)
	}
}

func TestValidateAllRules_MultipleFail(t *testing.T) {
	result := ValidateAllRules(
		-5,
		func(n int) error {
			if n < 0 {
				return errors.New("must be positive")
			}
			return nil
		},
		func(n int) error {
			if n%2 != 0 {
				return errors.New("must be even")
			}
			return nil
		},
	)

	err := AssertErr(t, result)

	var me MultiError
	if !errors.As(err, &me) {
		t.Fatal("Error should be MultiError")
	}

	if len(me.Errors) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(me.Errors))
	}
}

// ============================================================================
// ValidateField Tests
// ============================================================================

func TestValidateField_Ok(t *testing.T) {
	result := ValidateField("email", Ok("test@example.com"))

	value := AssertOk(t, result)
	if value != "test@example.com" {
		t.Errorf("ValidateField() = %v, want test@example.com", value)
	}
}

func TestValidateField_Err(t *testing.T) {
	result := ValidateField("email", Err[string](errors.New("invalid format")))

	err := AssertErr(t, result)
	op := OpOf(err)
	if !strings.Contains(op, "field.email") {
		t.Errorf("Error op should contain field name: %v", op)
	}
}

// ============================================================================
// ValidateStruct Tests
// ============================================================================

func TestValidateStruct_AllFieldsValid(t *testing.T) {
	result := ValidateStruct(
		ValidateField("email", Ok("test@example.com")),
		ValidateField("username", Ok("testuser")),
		ValidateField("password", Ok("secretpass")),
	)

	AssertOk(t, result)
}

func TestValidateStruct_OneFieldInvalid(t *testing.T) {
	result := ValidateStruct(
		ValidateField("email", Ok("test@example.com")),
		ValidateField("username", Err[string](errors.New("too short"))),
		ValidateField("password", Ok("secretpass")),
	)

	err := AssertErr(t, result)

	var me MultiError
	if !errors.As(err, &me) {
		t.Fatal("Error should be MultiError")
	}

	if len(me.Errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(me.Errors))
	}
}

func TestValidateStruct_MultipleFieldsInvalid(t *testing.T) {
	result := ValidateStruct(
		ValidateField("email", Err[string](errors.New("invalid format"))),
		ValidateField("username", Err[string](errors.New("too short"))),
		ValidateField("password", Ok("secretpass")),
	)

	err := AssertErr(t, result)

	var me MultiError
	if !errors.As(err, &me) {
		t.Fatal("Error should be MultiError")
	}

	if len(me.Errors) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(me.Errors))
	}
}

// ============================================================================
// ValidateMap Tests
// ============================================================================

func TestValidateMap_AllValid(t *testing.T) {
	items := map[string]int{
		"a": 1,
		"b": 2,
		"c": 3,
	}

	result := ValidateMap(items, func(key string, value int) error {
		if value < 0 {
			return fmt.Errorf("%s must be positive", key)
		}
		return nil
	})

	validated := AssertOk(t, result)
	if len(validated) != 3 {
		t.Errorf("Expected 3 items, got %d", len(validated))
	}
}

func TestValidateMap_SomeInvalid(t *testing.T) {
	items := map[string]int{
		"a": 1,
		"b": -2,
		"c": 3,
	}

	result := ValidateMap(items, func(key string, value int) error {
		if value < 0 {
			return fmt.Errorf("must be positive")
		}
		return nil
	})

	err := AssertErr(t, result)

	var me MultiError
	if !errors.As(err, &me) {
		t.Fatal("Error should be MultiError")
	}

	if len(me.Errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(me.Errors))
	}
}

// ============================================================================
// ValidateSlice Tests
// ============================================================================

func TestValidateSlice_AllValid(t *testing.T) {
	items := []int{1, 2, 3, 4, 5}

	result := ValidateSlice(items, func(n int) error {
		if n < 0 {
			return errors.New("must be positive")
		}
		return nil
	})

	validated := AssertOk(t, result)
	if len(validated) != 5 {
		t.Errorf("Expected 5 items, got %d", len(validated))
	}
}

func TestValidateSlice_SomeInvalid(t *testing.T) {
	items := []int{1, -2, 3, -4, 5}

	result := ValidateSlice(items, func(n int) error {
		if n < 0 {
			return errors.New("must be positive")
		}
		return nil
	})

	err := AssertErr(t, result)

	var me MultiError
	if !errors.As(err, &me) {
		t.Fatal("Error should be MultiError")
	}

	if len(me.Errors) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(me.Errors))
	}

	// Check that error messages include index
	for _, e := range me.Errors {
		if !strings.Contains(e.Error(), "index") {
			t.Errorf("Error should include index: %v", e)
		}
	}
}

// ============================================================================
// ValidateSliceResults Tests
// ============================================================================

func TestValidateSliceResults_AllValid(t *testing.T) {
	results := []Result[int]{
		Ok(1),
		Ok(2),
		Ok(3),
	}

	result := ValidateSliceResults(results)

	values := AssertOk(t, result)
	if len(values) != 3 {
		t.Errorf("Expected 3 values, got %d", len(values))
	}
}

func TestValidateSliceResults_SomeInvalid(t *testing.T) {
	results := []Result[int]{
		Ok(1),
		Err[int](errors.New("error at index 1")),
		Ok(3),
		Err[int](errors.New("error at index 3")),
	}

	result := ValidateSliceResults(results)

	err := AssertErr(t, result)

	var me MultiError
	if !errors.As(err, &me) {
		t.Fatal("Error should be MultiError")
	}

	if len(me.Errors) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(me.Errors))
	}
}

// ============================================================================
// Integration Tests
// ============================================================================

func TestValidation_IntegrationFlow(t *testing.T) {
	// Simulate validating a user registration form

	type UserInput struct {
		Email    string
		Username string
		Password string
		Age      int
	}

	input := UserInput{
		Email:    "test@example.com",
		Username: "testuser",
		Password: "secret123",
		Age:      25,
	}

	// Validate all fields
	result := ValidateStruct(
		ValidateField("email", ValidateRules(
			input.Email,
			func(email string) error {
				if email == "" {
					return errors.New("cannot be empty")
				}
				if !strings.Contains(email, "@") {
					return errors.New("invalid format")
				}
				return nil
			},
		)),
		ValidateField("username", ValidateRules(
			input.Username,
			func(username string) error {
				if len(username) < 3 {
					return errors.New("too short")
				}
				if len(username) > 20 {
					return errors.New("too long")
				}
				return nil
			},
		)),
		ValidateField("password", ValidateRules(
			input.Password,
			func(password string) error {
				if len(password) < 8 {
					return errors.New("too short")
				}
				return nil
			},
		)),
		ValidateField("age", ValidateRules(
			input.Age,
			func(age int) error {
				if age < 18 {
					return errors.New("must be 18 or older")
				}
				return nil
			},
		)),
	)

	AssertOk(t, result)
}

func TestValidation_IntegrationFlow_MultipleErrors(t *testing.T) {
	type UserInput struct {
		Email    string
		Username string
		Password string
	}

	input := UserInput{
		Email:    "invalid-email",
		Username: "ab",
		Password: "short",
	}

	// Validate all fields and collect all errors
	result := ValidateStruct(
		ValidateField("email", ValidateRules(
			input.Email,
			func(email string) error {
				if !strings.Contains(email, "@") {
					return errors.New("invalid format")
				}
				return nil
			},
		)),
		ValidateField("username", ValidateRules(
			input.Username,
			func(username string) error {
				if len(username) < 3 {
					return errors.New("too short")
				}
				return nil
			},
		)),
		ValidateField("password", ValidateRules(
			input.Password,
			func(password string) error {
				if len(password) < 8 {
					return errors.New("too short")
				}
				return nil
			},
		)),
	)

	err := AssertErr(t, result)

	var me MultiError
	if !errors.As(err, &me) {
		t.Fatal("Error should be MultiError")
	}

	// All three fields should have failed
	if len(me.Errors) != 3 {
		t.Errorf("Expected 3 errors, got %d", len(me.Errors))
	}

	// Check that errors contain field names
	errStr := err.Error()
	if !strings.Contains(errStr, "email") {
		t.Error("Error should contain 'email'")
	}
	if !strings.Contains(errStr, "username") {
		t.Error("Error should contain 'username'")
	}
	if !strings.Contains(errStr, "password") {
		t.Error("Error should contain 'password'")
	}
}

func TestValidateAllRules_ComplexValidation(t *testing.T) {
	type Email string

	email := Email("test@example.com")

	result := ValidateAllRules(
		email,
		func(e Email) error {
			if !strings.Contains(string(e), "@") {
				return errors.New("must contain @")
			}
			return nil
		},
		func(e Email) error {
			if len(e) > 254 {
				return errors.New("exceeds max length")
			}
			return nil
		},
		func(e Email) error {
			if strings.HasPrefix(string(e), ".") {
				return errors.New("cannot start with dot")
			}
			return nil
		},
		func(e Email) error {
			parts := strings.Split(string(e), "@")
			if len(parts) != 2 {
				return errors.New("invalid format")
			}
			if len(parts[0]) == 0 || len(parts[1]) == 0 {
				return errors.New("empty local or domain part")
			}
			return nil
		},
	)

	AssertOk(t, result)
}

func TestValidateMap_WithComplexTypes(t *testing.T) {
	type Config struct {
		Host string
		Port int
	}

	configs := map[string]Config{
		"primary": {Host: "localhost", Port: 8080},
		"backup":  {Host: "backup.example.com", Port: 8081},
		"invalid": {Host: "", Port: -1},
	}

	result := ValidateMap(configs, func(name string, cfg Config) error {
		if cfg.Host == "" {
			return fmt.Errorf("host cannot be empty")
		}
		if cfg.Port < 0 || cfg.Port > 65535 {
			return fmt.Errorf("invalid port")
		}
		return nil
	})

	err := AssertErr(t, result)

	var me MultiError
	if !errors.As(err, &me) {
		t.Fatal("Error should be MultiError")
	}

	if len(me.Errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(me.Errors))
	}

	// Error should mention the "invalid" config
	if !strings.Contains(err.Error(), "invalid") {
		t.Errorf("Error should mention the invalid config: %v", err)
	}
}

func TestValidateSlice_EarlyExit(t *testing.T) {
	// ValidateSlice collects all errors, not early exit
	items := []int{1, 2, -3, 4, -5}
	errorCount := 0

	result := ValidateSlice(items, func(n int) error {
		if n < 0 {
			errorCount++
			return errors.New("negative")
		}
		return nil
	})

	AssertErr(t, result)

	// Should have validated all items despite errors
	if errorCount != 2 {
		t.Errorf("Expected to validate all items (2 errors), got %d errors", errorCount)
	}
}

func TestValidateAllMap_PreservesOrder(t *testing.T) {
	results := []Result[int]{
		Ok(1),
		Err[int](errors.New("error 2")),
		Ok(3),
		Err[int](errors.New("error 4")),
		Ok(5),
	}

	values, errs := ValidateAllMap(results...)

	// Values should be in order
	if len(values) != 3 {
		t.Fatalf("Expected 3 values, got %d", len(values))
	}
	if values[0] != 1 || values[1] != 3 || values[2] != 5 {
		t.Errorf("Values not in correct order: %v", values)
	}

	// Errors should be collected
	if len(errs) != 2 {
		t.Fatalf("Expected 2 errors, got %d", len(errs))
	}
}

func TestMultiError_NestedErrors(t *testing.T) {
	innerErr1 := errors.New("inner error 1")
	innerErr2 := errors.New("inner error 2")

	me := MultiError{
		Errors: []error{
			fmt.Errorf("outer error 1: %w", innerErr1),
			fmt.Errorf("outer error 2: %w", innerErr2),
		},
	}

	// Check that inner errors are accessible
	if !errors.Is(me.Unwrap(), innerErr1) {
		t.Error("MultiError should unwrap to first inner error")
	}
}

func TestValidateRules_EmptyRules(t *testing.T) {
	result := ValidateRules(42)

	value := AssertOk(t, result)
	if value != 42 {
		t.Errorf("ValidateRules with no rules should return value unchanged")
	}
}

func TestValidateAllRules_EmptyRules(t *testing.T) {
	result := ValidateAllRules(42)

	value := AssertOk(t, result)
	if value != 42 {
		t.Errorf("ValidateAllRules with no rules should return value unchanged")
	}
}
