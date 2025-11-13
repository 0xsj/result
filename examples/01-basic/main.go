// Package main demonstrates basic usage of the result package.
//
// This example covers:
// - Creating Ok and Err results
// - Checking result status with IsOk/IsErr
// - Extracting values with Unwrap, UnwrapOr, UnwrapOrElse
// - Pattern matching with Match
// - Basic error handling
//
// Run:
//
//	go run main.go
package main

import (
	"errors"
	"fmt"

	"github.com/0xsj/result"
)

// ============================================================================
// Example 1: Creating Results
// ============================================================================

// divide demonstrates creating Ok and Err results
func divide(a, b int) result.Result[int] {
	if b == 0 {
		return result.Err[int](errors.New("division by zero"))
	}
	return result.Ok(a / b)
}

func example1_Creating() {
	fmt.Println("=== Example 1: Creating Results ===")

	// Success case
	r1 := divide(10, 2)
	fmt.Printf("10 / 2 = %v (IsOk: %v)\n", r1.Unwrap(), r1.IsOk())

	// Error case
	r2 := divide(10, 0)
	fmt.Printf("10 / 0 = Error (IsErr: %v)\n", r2.IsErr())

	fmt.Println()
}

// ============================================================================
// Example 2: Extracting Values
// ============================================================================

func example2_Extracting() {
	fmt.Println("=== Example 2: Extracting Values ===")

	success := divide(20, 4)
	failure := divide(10, 0)

	// Unwrap - panics on error (use only when certain it's Ok)
	fmt.Printf("Unwrap success: %d\n", success.Unwrap())

	// UnwrapOr - returns default on error
	fmt.Printf("UnwrapOr failure: %d (default: 0)\n", failure.UnwrapOr(0))

	// UnwrapOrElse - computes default from error
	defaultValue := failure.UnwrapOrElse(func(err error) int {
		fmt.Printf("  Computing default because: %v\n", err)
		return -1
	})
	fmt.Printf("UnwrapOrElse failure: %d\n", defaultValue)

	// Value - returns (T, error) tuple (Go idiom compatible)
	value, err := failure.Value()
	if err != nil {
		fmt.Printf("Value failure: error = %v\n", err)
	} else {
		fmt.Printf("Value success: %d\n", value)
	}

	fmt.Println()
}

// ============================================================================
// Example 3: Pattern Matching
// ============================================================================

func example3_PatternMatching() {
	fmt.Println("=== Example 3: Pattern Matching ===")

	results := []result.Result[int]{
		divide(100, 10),
		divide(50, 5),
		divide(30, 0),
	}

	for i, r := range results {
		fmt.Printf("Result %d: ", i+1)

		r.Match(
			func(value int) {
				fmt.Printf("Success = %d\n", value)
			},
			func(err error) {
				fmt.Printf("Error = %v\n", err)
			},
		)
	}

	fmt.Println()
}

// ============================================================================
// Example 4: Error Types with result.Error
// ============================================================================

// findUser demonstrates using typed errors
func findUser(id string) result.Result[string] {
	if id == "" {
		return result.Err[string](
			result.Validation("findUser", "user ID is required", map[string]interface{}{
				"field": "id",
			}),
		)
	}

	if id == "404" {
		return result.Err[string](
			result.NotFound("findUser", "user"),
		)
	}

	if id == "conflict" {
		return result.Err[string](
			result.Conflict("findUser", "user"),
		)
	}

	return result.Ok(fmt.Sprintf("User(%s)", id))
}

func example4_ErrorTypes() {
	fmt.Println("=== Example 4: Error Types ===")

	testCases := []string{"123", "", "404", "conflict"}

	for _, id := range testCases {
		r := findUser(id)

		fmt.Printf("findUser(%q): ", id)

		r.Match(
			func(user string) {
				fmt.Printf("Found = %s\n", user)
			},
			func(err error) {
				kind := result.KindOf(err)
				fmt.Printf("Error (Kind: %s) = %v\n", kind.String(), err)
			},
		)
	}

	fmt.Println()
}

// ============================================================================
// Example 5: Transforming Results
// ============================================================================

func example5_Transforming() {
	fmt.Println("=== Example 5: Transforming Results ===")

	// Map - transform the Ok value
	doubled := divide(10, 2).Map(func(n int) int {
		return n * 2
	})
	fmt.Printf("(10 / 2) * 2 = %d\n", doubled.Unwrap())

	// Map on error - no transformation
	errorResult := divide(10, 0).Map(func(n int) int {
		fmt.Println("  This won't be called")
		return n * 2
	})
	fmt.Printf("(10 / 0) * 2 = Error: %v\n", errorResult.Error())

	// MapErr - transform the error
	betterError := divide(10, 0).MapErr(func(err error) error {
		return fmt.Errorf("math error: %w", err)
	})
	fmt.Printf("Better error: %v\n", betterError.Error())

	fmt.Println()
}

// ============================================================================
// Example 6: Converting to/from Go Idioms
// ============================================================================

// parseNumber is a standard Go function returning (T, error)
func parseNumber(s string) (int, error) {
	if s == "" {
		return 0, errors.New("empty string")
	}
	// Simplified parse (real code would use strconv.Atoi)
	return len(s), nil
}

func example6_GoIdioms() {
	fmt.Println("=== Example 6: Converting to/from Go Idioms ===")

	// From - convert (T, error) to Result[T]
	r1 := result.From(parseNumber("hello"))
	fmt.Printf("From(parseNumber('hello')): IsOk = %v, Value = %d\n",
		r1.IsOk(), r1.UnwrapOr(0))

	r2 := result.From(parseNumber(""))
	fmt.Printf("From(parseNumber('')): IsErr = %v, Error = %v\n",
		r2.IsErr(), r2.Error())

	// Value - convert Result[T] back to (T, error)
	value, err := divide(20, 4).Value()
	if err != nil {
		fmt.Printf("Value(): error = %v\n", err)
	} else {
		fmt.Printf("Value(): success = %d\n", value)
	}

	fmt.Println()
}

// ============================================================================
// Example 7: Real-World Usage Pattern
// ============================================================================

type User struct {
	ID    string
	Name  string
	Email string
}

// validateEmail validates an email address
func validateEmail(email string) result.Result[string] {
	if email == "" {
		return result.Err[string](
			result.Validation("validateEmail", "email is required", map[string]interface{}{
				"field": "email",
			}),
		)
	}

	// Simplified validation
	if len(email) < 5 {
		return result.Err[string](
			result.Validation("validateEmail", "email too short", map[string]interface{}{
				"field":  "email",
				"length": len(email),
			}),
		)
	}

	return result.Ok(email)
}

// createUser demonstrates a real-world function
func createUser(name, email string) result.Result[*User] {
	// Validate email
	emailResult := validateEmail(email)
	if emailResult.IsErr() {
		return result.Err[*User](emailResult.Error())
	}

	// Create user
	user := &User{
		ID:    "user_123",
		Name:  name,
		Email: emailResult.Unwrap(),
	}

	return result.Ok(user)
}

func example7_RealWorld() {
	fmt.Println("=== Example 7: Real-World Usage ===")

	testCases := []struct {
		name  string
		email string
	}{
		{"Alice", "alice@example.com"},
		{"Bob", ""},
		{"Charlie", "bad"},
	}

	for _, tc := range testCases {
		fmt.Printf("Creating user: name=%s, email=%s\n", tc.name, tc.email)

		createUser(tc.name, tc.email).Match(
			func(user *User) {
				fmt.Printf("  ✓ Success: User(id=%s, name=%s, email=%s)\n",
					user.ID, user.Name, user.Email)
			},
			func(err error) {
				fmt.Printf("  ✗ Error: %v\n", err)

				// Check error metadata
				if meta := result.MetaOf(err); meta != nil {
					fmt.Printf("    Metadata: %v\n", meta)
				}
			},
		)
	}

	fmt.Println()
}

// ============================================================================
// Example 8: Inspection (Side Effects Without Consuming)
// ============================================================================

func example8_Inspection() {
	fmt.Println("=== Example 8: Inspection ===")

	// Inspect - side effect on success
	result.Ok(42).
		Inspect(func(n int) {
			fmt.Printf("Inspecting success: value = %d\n", n)
		}).
		Map(func(n int) int {
			return n * 2
		}).
		Inspect(func(n int) {
			fmt.Printf("After Map: value = %d\n", n)
		})

	// InspectErr - side effect on error
	divide(10, 0).
		InspectErr(func(err error) {
			fmt.Printf("Inspecting error: %v\n", err)
		}).
		MapErr(func(err error) error {
			return fmt.Errorf("wrapped: %w", err)
		}).
		InspectErr(func(err error) {
			fmt.Printf("After MapErr: %v\n", err)
		})

	// Tap - side effect on both success and error
	fmt.Println("\nUsing Tap:")
	divide(100, 10).Tap(
		func(n int) {
			fmt.Printf("  Success handler: %d\n", n)
		},
		func(err error) {
			fmt.Printf("  Error handler: %v\n", err)
		},
	)

	divide(100, 0).Tap(
		func(n int) {
			fmt.Printf("  Success handler: %d\n", n)
		},
		func(err error) {
			fmt.Printf("  Error handler: %v\n", err)
		},
	)

	fmt.Println()
}

// ============================================================================
// Main
// ============================================================================

func main() {
	fmt.Println("╔════════════════════════════════════════╗")
	fmt.Println("║  Result Package - Basic Examples      ║")
	fmt.Println("╚════════════════════════════════════════╝")
	fmt.Println()

	example1_Creating()
	example2_Extracting()
	example3_PatternMatching()
	example4_ErrorTypes()
	example5_Transforming()
	example6_GoIdioms()
	example7_RealWorld()
	example8_Inspection()

	fmt.Println("╔════════════════════════════════════════╗")
	fmt.Println("║  All examples completed!               ║")
	fmt.Println("╚════════════════════════════════════════╝")
}