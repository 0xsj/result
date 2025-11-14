package result

import (
	"errors"
	"strings"
	"testing"
)

// Test helpers
type Email struct {
	value string
}

func (e Email) Value() string {
	return e.value
}

var (
	errInvalidEmail       = errors.New("invalid email format")
	errEmailAlreadyExists = errors.New("email already exists")
	errDatabaseError      = errors.New("database error")
)

// ============================================================================
// Extract Tests
// ============================================================================

func TestExtract_Ok(t *testing.T) {
	r := Ok(42)
	value, err := Extract(r, "test.operation")

	if err != nil {
		t.Errorf("Extract() returned error for Ok result: %v", err)
	}
	if value != 42 {
		t.Errorf("Extract() = %v, want %v", value, 42)
	}
}

func TestExtract_Err(t *testing.T) {
	originalErr := Validation("user.validate", "invalid input", nil)
	r := Err[int](originalErr)

	value, err := Extract(r, "test.operation")

	if err == nil {
		t.Fatal("Extract() did not return error for Err result")
	}
	if value != 0 {
		t.Errorf("Extract() value = %v, want zero value", value)
	}

	// Check that error was wrapped
	if !strings.Contains(err.Error(), "test.operation") {
		t.Errorf("Extract() error should contain operation context: %v", err)
	}
}

func TestExtract_PreservesErrorKind(t *testing.T) {
	originalErr := Validation("user.validate", "invalid input", nil)
	r := Err[int](originalErr)

	_, err := Extract(r, "test.operation")

	if KindOf(err) != KindValidation {
		t.Errorf("Extract() changed error kind: got %v, want %v", KindOf(err), KindValidation)
	}
}

// ============================================================================
// ExtractOr Tests
// ============================================================================

func TestExtractOr_Ok(t *testing.T) {
	r := Ok(42)
	value := ExtractOr(r, 99)

	if value != 42 {
		t.Errorf("ExtractOr() = %v, want %v", value, 42)
	}
}

func TestExtractOr_Err(t *testing.T) {
	r := Err[int](errors.New("some error"))
	value := ExtractOr(r, 99)

	if value != 99 {
		t.Errorf("ExtractOr() = %v, want default %v", value, 99)
	}
}

func TestExtractOr_WithStruct(t *testing.T) {
	type Config struct {
		Timeout int
		Host    string
	}

	defaultConfig := Config{Timeout: 30, Host: "localhost"}
	r := Err[Config](errors.New("config not found"))

	value := ExtractOr(r, defaultConfig)

	if value.Timeout != 30 || value.Host != "localhost" {
		t.Errorf("ExtractOr() = %+v, want %+v", value, defaultConfig)
	}
}

// ============================================================================
// Must Tests
// ============================================================================

func TestMust_Ok(t *testing.T) {
	r := Ok(42)

	// Should not panic
	value := Must(r, "test operation failed")

	if value != 42 {
		t.Errorf("Must() = %v, want %v", value, 42)
	}
}

func TestMust_Err_Panics(t *testing.T) {
	r := Err[int](errors.New("something went wrong"))

	defer func() {
		if r := recover(); r == nil {
			t.Error("Must() did not panic for Err result")
		}
	}()

	Must(r, "test operation failed")
}

// ============================================================================
// ValidateAndCheck Tests
// ============================================================================

func TestValidateAndCheck_ValidationFails(t *testing.T) {
	validate := Err[Email](errInvalidEmail)

	result := ValidateAndCheck(
		validate,
		func(e Email) Result[bool] {
			return Ok(false)
		},
		func(e Email) error {
			return errEmailAlreadyExists
		},
		"create_user.email",
	)

	if result.IsOk() {
		t.Fatal("ValidateAndCheck() should return Err when validation fails")
	}

	err := result.UnwrapErr()
	if !strings.Contains(err.Error(), "create_user.email.validate") {
		t.Errorf("Error should contain validation context: %v", err)
	}
}

func TestValidateAndCheck_CheckFails(t *testing.T) {
	validate := Ok(Email{value: "test@example.com"})

	result := ValidateAndCheck(
		validate,
		func(e Email) Result[bool] {
			return Err[bool](errDatabaseError)
		},
		func(e Email) error {
			return errEmailAlreadyExists
		},
		"create_user.email",
	)

	if result.IsOk() {
		t.Fatal("ValidateAndCheck() should return Err when check fails")
	}

	err := result.UnwrapErr()
	if !strings.Contains(err.Error(), "create_user.email.check") {
		t.Errorf("Error should contain check context: %v", err)
	}
}

func TestValidateAndCheck_ConflictDetected(t *testing.T) {
	validate := Ok(Email{value: "test@example.com"})

	result := ValidateAndCheck(
		validate,
		func(e Email) Result[bool] {
			return Ok(true) // Exists = conflict
		},
		func(e Email) error {
			return errEmailAlreadyExists
		},
		"create_user.email",
	)

	if result.IsOk() {
		t.Fatal("ValidateAndCheck() should return Err when conflict detected")
	}

	err := result.UnwrapErr()
	if !strings.Contains(err.Error(), "create_user.email.conflict") {
		t.Errorf("Error should contain conflict context: %v", err)
	}
}

func TestValidateAndCheck_Success(t *testing.T) {
	validate := Ok(Email{value: "test@example.com"})

	result := ValidateAndCheck(
		validate,
		func(e Email) Result[bool] {
			return Ok(false) // Does not exist
		},
		func(e Email) error {
			return errEmailAlreadyExists
		},
		"create_user.email",
	)

	if result.IsErr() {
		t.Fatalf("ValidateAndCheck() should return Ok when successful: %v", result.UnwrapErr())
	}

	email := result.Unwrap()
	if email.Value() != "test@example.com" {
		t.Errorf("ValidateAndCheck() = %v, want test@example.com", email.Value())
	}
}

// ============================================================================
// ValidateWith Tests
// ============================================================================

func TestValidateWith_Ok(t *testing.T) {
	validate := Ok(Email{value: "test@example.com"})

	email, err := ValidateWith(validate, "create_user.email")

	if err != nil {
		t.Errorf("ValidateWith() returned error for Ok result: %v", err)
	}
	if email.Value() != "test@example.com" {
		t.Errorf("ValidateWith() = %v, want test@example.com", email.Value())
	}
}

func TestValidateWith_Err(t *testing.T) {
	validate := Err[Email](errInvalidEmail)

	_, err := ValidateWith(validate, "create_user.email")

	if err == nil {
		t.Fatal("ValidateWith() did not return error for Err result")
	}
	if !strings.Contains(err.Error(), "create_user.email") {
		t.Errorf("Error should contain operation context: %v", err)
	}
}

// ============================================================================
// CheckCondition Tests
// ============================================================================

func TestCheckCondition_ConditionTrue_ReturnsErr(t *testing.T) {
	email := Email{value: "test@example.com"}

	result := CheckCondition(
		true, // Condition true = error
		email,
		errEmailAlreadyExists,
		"create_user.check_email",
	)

	if result.IsOk() {
		t.Fatal("CheckCondition() should return Err when condition is true")
	}

	err := result.UnwrapErr()
	if !strings.Contains(err.Error(), "create_user.check_email") {
		t.Errorf("Error should contain operation context: %v", err)
	}
}

func TestCheckCondition_ConditionFalse_ReturnsOk(t *testing.T) {
	email := Email{value: "test@example.com"}

	result := CheckCondition(
		false, // Condition false = success
		email,
		errEmailAlreadyExists,
		"create_user.check_email",
	)

	if result.IsErr() {
		t.Fatalf("CheckCondition() should return Ok when condition is false: %v", result.UnwrapErr())
	}

	value := result.Unwrap()
	if value.Value() != "test@example.com" {
		t.Errorf("CheckCondition() = %v, want test@example.com", value.Value())
	}
}

// ============================================================================
// Validate3 Tests
// ============================================================================

func TestValidate3_AllOk(t *testing.T) {
	r1 := Ok(Email{value: "test@example.com"})
	r2 := Ok("testuser")
	r3 := Ok("password123")

	result := Validate3(r1, r2, r3, "create_user.validate")

	if result.IsErr() {
		t.Fatalf("Validate3() should return Ok when all results are Ok: %v", result.UnwrapErr())
	}

	values := result.Unwrap()
	if values.V1.Value() != "test@example.com" {
		t.Errorf("Validate3() V1 = %v, want test@example.com", values.V1.Value())
	}
	if values.V2 != "testuser" {
		t.Errorf("Validate3() V2 = %v, want testuser", values.V2)
	}
	if values.V3 != "password123" {
		t.Errorf("Validate3() V3 = %v, want password123", values.V3)
	}
}

func TestValidate3_FirstErr(t *testing.T) {
	r1 := Err[Email](errInvalidEmail)
	r2 := Ok("testuser")
	r3 := Ok("password123")

	result := Validate3(r1, r2, r3, "create_user.validate")

	if result.IsOk() {
		t.Fatal("Validate3() should return Err when first result is Err")
	}

	err := result.UnwrapErr()
	if !strings.Contains(err.Error(), "create_user.validate") {
		t.Errorf("Error should contain operation context: %v", err)
	}
}

func TestValidate3_SecondErr(t *testing.T) {
	r1 := Ok(Email{value: "test@example.com"})
	r2 := Err[string](errors.New("invalid username"))
	r3 := Ok("password123")

	result := Validate3(r1, r2, r3, "create_user.validate")

	if result.IsOk() {
		t.Fatal("Validate3() should return Err when second result is Err")
	}
}

func TestValidate3_ThirdErr(t *testing.T) {
	r1 := Ok(Email{value: "test@example.com"})
	r2 := Ok("testuser")
	r3 := Err[string](errors.New("weak password"))

	result := Validate3(r1, r2, r3, "create_user.validate")

	if result.IsOk() {
		t.Fatal("Validate3() should return Err when third result is Err")
	}
}

// ============================================================================
// Integration Test
// ============================================================================

func TestExtractHelpers_IntegrationFlow(t *testing.T) {
	// Simulates a typical command handler flow

	// Step 1: Validate email
	emailResult := Ok(Email{value: "test@example.com"})
	email, err := Extract(emailResult, "create_user.email")
	if err != nil {
		t.Fatalf("Failed to extract email: %v", err)
	}

	// Step 2: Check email doesn't exist (using CheckCondition)
	emailExists := false
	checkResult := CheckCondition(
		emailExists,
		email,
		errEmailAlreadyExists,
		"create_user.check_email",
	)
	if checkResult.IsErr() {
		t.Fatalf("Email check failed: %v", checkResult.UnwrapErr())
	}

	// Step 3: Use ValidateAndCheck for username
	usernameResult := ValidateAndCheck(
		Ok("testuser"),
		func(username string) Result[bool] {
			return Ok(false) // Username doesn't exist
		},
		func(username string) error {
			return errors.New("username already exists")
		},
		"create_user.username",
	)
	if usernameResult.IsErr() {
		t.Fatalf("Username validation failed: %v", usernameResult.UnwrapErr())
	}

	// All validations passed
	t.Log("Integration flow completed successfully")
}
