// Package main demonstrates railway-oriented programming with the result package.
//
// This example covers:
// - Complex AndThen chains (railway pattern)
// - AndThenMap for type transformations
// - OrElse for fallback chains
// - Map and MapErr transformations
// - Recover and RecoverWith patterns
// - Real-world order processing pipeline
//
// Run:
//
//	go run main.go
package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/0xsj/result"
)

// ============================================================================
// Example 1: Basic Railway Pattern
// ============================================================================

// Step represents a processing step
type Step struct {
	Name   string
	Result string
}

func step1(input string) result.Result[string] {
	fmt.Printf("  Step 1: Processing '%s'\n", input)
	if input == "" {
		return result.Err[string](errors.New("step1: empty input"))
	}
	return result.Ok(input + " -> step1")
}

func step2(input string) result.Result[string] {
	fmt.Printf("  Step 2: Processing '%s'\n", input)
	return result.Ok(input + " -> step2")
}

func step3(input string) result.Result[string] {
	fmt.Printf("  Step 3: Processing '%s'\n", input)
	return result.Ok(input + " -> step3")
}

func example1_BasicRailway() {
	fmt.Println("=== Example 1: Basic Railway Pattern ===")

	// Success path - all steps execute
	fmt.Println("\nSuccess path:")
	result1 := result.Ok("input").
		AndThen(step1).
		AndThen(step2).
		AndThen(step3)

	result1.Match(
		func(output string) {
			fmt.Printf("✓ Success: %s\n", output)
		},
		func(err error) {
			fmt.Printf("✗ Error: %v\n", err)
		},
	)

	// Error path - short circuits after step1
	fmt.Println("\nError path:")
	result2 := result.Ok("").
		AndThen(step1). // Fails here
		AndThen(step2). // Skipped
		AndThen(step3)  // Skipped

	result2.Match(
		func(output string) {
			fmt.Printf("✓ Success: %s\n", output)
		},
		func(err error) {
			fmt.Printf("✗ Error: %v (short-circuited)\n", err)
		},
	)

	fmt.Println()
}

// ============================================================================
// Example 2: Type Transformations with AndThenMap
// ============================================================================

type RawData struct {
	Value string
}

type ParsedData struct {
	Number int
}

type ProcessedData struct {
	Number int
	Double int
}

func parseRawData(raw RawData) result.Result[ParsedData] {
	fmt.Printf("  Parsing: %s\n", raw.Value)

	// Simplified parsing
	if raw.Value == "42" {
		return result.Ok(ParsedData{Number: 42})
	}

	return result.Err[ParsedData](
		result.Validation("parseRawData", "invalid number", nil),
	)
}

func processData(parsed ParsedData) result.Result[ProcessedData] {
	fmt.Printf("  Processing: %d\n", parsed.Number)

	return result.Ok(ProcessedData{
		Number: parsed.Number,
		Double: parsed.Number * 2,
	})
}

func example2_TypeTransformations() {
	fmt.Println("=== Example 2: Type Transformations ===")

	// RawData -> ParsedData -> ProcessedData
	fmt.Println("\nTransforming types through pipeline:")

	rawResult := result.Ok(RawData{Value: "42"})

	// Use AndThenMap to transform types
	processedResult := result.AndThenMap(
		rawResult,
		parseRawData,
	)

	finalResult := result.AndThenMap(
		processedResult,
		processData,
	)

	finalResult.Match(
		func(data ProcessedData) {
			fmt.Printf("✓ Final result: Number=%d, Double=%d\n", data.Number, data.Double)
		},
		func(err error) {
			fmt.Printf("✗ Error: %v\n", err)
		},
	)

	// Error case
	fmt.Println("\nError in transformation:")
	errorResult := result.AndThenMap(
		result.Ok(RawData{Value: "invalid"}),
		parseRawData,
	)

	errorResult.Match(
		func(data ParsedData) {
			fmt.Printf("✓ Success: %+v\n", data)
		},
		func(err error) {
			fmt.Printf("✗ Error: %v\n", err)
		},
	)

	fmt.Println()
}

// ============================================================================
// Example 3: Fallback Chains with OrElse
// ============================================================================

func fetchFromCache(id string) result.Result[string] {
	fmt.Printf("  Trying cache for %s...\n", id)
	return result.Err[string](
		result.NotFound("fetchFromCache", "data"),
	)
}

func fetchFromRedis(id string) result.Result[string] {
	fmt.Printf("  Trying Redis for %s...\n", id)
	return result.Err[string](
		result.NotFound("fetchFromRedis", "data"),
	)
}

func fetchFromDatabase(id string) result.Result[string] {
	fmt.Printf("  Trying database for %s...\n", id)
	if id == "exists" {
		return result.Ok(fmt.Sprintf("data_%s", id))
	}
	return result.Err[string](
		result.NotFound("fetchFromDatabase", "data"),
	)
}

func getDefaultData() string {
	fmt.Println("  Using default data")
	return "default_data"
}

func example3_FallbackChains() {
	fmt.Println("=== Example 3: Fallback Chains ===")

	// Fallback chain: Cache -> Redis -> Database -> Default
	fmt.Println("\nFallback chain (eventually succeeds at database):")

	result1 := fetchFromCache("exists").
		OrElse(func(err error) result.Result[string] {
			return fetchFromRedis("exists")
		}).
		OrElse(func(err error) result.Result[string] {
			return fetchFromDatabase("exists")
		}).
		OrElse(func(err error) result.Result[string] {
			return result.Ok(getDefaultData())
		})

	result1.Match(
		func(data string) {
			fmt.Printf("✓ Got data: %s\n", data)
		},
		func(err error) {
			fmt.Printf("✗ Error: %v\n", err)
		},
	)

	// All fail - use default
	fmt.Println("\nFallback chain (all fail, use default):")

	result2 := fetchFromCache("missing").
		OrElse(func(err error) result.Result[string] {
			return fetchFromRedis("missing")
		}).
		OrElse(func(err error) result.Result[string] {
			return fetchFromDatabase("missing")
		}).
		OrElse(func(err error) result.Result[string] {
			return result.Ok(getDefaultData())
		})

	result2.Match(
		func(data string) {
			fmt.Printf("✓ Got data: %s\n", data)
		},
		func(err error) {
			fmt.Printf("✗ Error: %v\n", err)
		},
	)

	fmt.Println()
}

// ============================================================================
// Example 4: Recovery Patterns
// ============================================================================

func riskyOperation(input string) result.Result[string] {
	fmt.Printf("  Risky operation with: %s\n", input)

	if input == "fail" {
		return result.Err[string](
			result.Infrastructure("riskyOperation", errors.New("network timeout")),
		)
	}

	if input == "invalid" {
		return result.Err[string](
			result.Validation("riskyOperation", "invalid input", nil),
		)
	}

	return result.Ok(input + "_processed")
}

func example4_RecoveryPatterns() {
	fmt.Println("=== Example 4: Recovery Patterns ===")

	// Recover from specific error kinds
	fmt.Println("\nRecover from infrastructure errors:")

	result1 := riskyOperation("fail").
		Recover(result.KindInfrastructure, func(err error) string {
			fmt.Printf("  Recovering from infrastructure error: %v\n", err)
			return "recovered_value"
		}).
		Recover(result.KindValidation, func(err error) string {
			fmt.Printf("  Recovering from validation error: %v\n", err)
			return "fallback_value"
		})

	result1.Match(
		func(output string) {
			fmt.Printf("✓ Result: %s\n", output)
		},
		func(err error) {
			fmt.Printf("✗ Error: %v\n", err)
		},
	)

	// Don't recover from validation errors
	fmt.Println("\nValidation errors not recovered:")

	result2 := riskyOperation("invalid").
		Recover(result.KindInfrastructure, func(err error) string {
			return "recovered"
		})

	result2.Match(
		func(output string) {
			fmt.Printf("✓ Result: %s\n", output)
		},
		func(err error) {
			fmt.Printf("✗ Error: %v (not recovered)\n", err)
		},
	)

	// RecoverWith - recovery can also fail
	fmt.Println("\nRecoverWith - recovery itself can fail:")

	result3 := riskyOperation("fail").
		RecoverWith(result.KindInfrastructure, func(err error) result.Result[string] {
			fmt.Println("  Attempting recovery...")
			return riskyOperation("success")
		})

	result3.Match(
		func(output string) {
			fmt.Printf("✓ Result: %s\n", output)
		},
		func(err error) {
			fmt.Printf("✗ Error: %v\n", err)
		},
	)

	fmt.Println()
}

// ============================================================================
// Example 5: Real-World Order Processing Pipeline
// ============================================================================

type Order struct {
	ID     string
	Items  []string
	Total  float64
	Status string
}

type Payment struct {
	OrderID string
	Amount  float64
	Status  string
}

type Shipment struct {
	OrderID      string
	TrackingCode string
}

type Receipt struct {
	OrderID      string
	Amount       float64
	TrackingCode string
	Timestamp    time.Time
}

func validateOrder(order *Order) result.Result[*Order] {
	fmt.Printf("  Validating order %s\n", order.ID)

	if len(order.Items) == 0 {
		return result.Err[*Order](
			result.Validation("validateOrder", "order has no items", map[string]interface{}{
				"order_id": order.ID,
			}),
		)
	}

	if order.Total <= 0 {
		return result.Err[*Order](
			result.Validation("validateOrder", "invalid total", map[string]interface{}{
				"order_id": order.ID,
				"total":    order.Total,
			}),
		)
	}

	order.Status = "validated"
	return result.Ok(order)
}

func checkInventory(order *Order) result.Result[*Order] {
	fmt.Printf("  Checking inventory for order %s\n", order.ID)

	// Simulate inventory check
	order.Status = "inventory_checked"
	return result.Ok(order)
}

func processPayment(order *Order) result.Result[*Payment] {
	fmt.Printf("  Processing payment for order %s (amount: $%.2f)\n", order.ID, order.Total)

	// Simulate payment processing
	payment := &Payment{
		OrderID: order.ID,
		Amount:  order.Total,
		Status:  "completed",
	}

	return result.Ok(payment)
}

func createShipment(payment *Payment) result.Result[*Shipment] {
	fmt.Printf("  Creating shipment for order %s\n", payment.OrderID)

	shipment := &Shipment{
		OrderID:      payment.OrderID,
		TrackingCode: fmt.Sprintf("TRACK_%s", payment.OrderID),
	}

	return result.Ok(shipment)
}

func generateReceipt(shipment *Shipment, payment *Payment) result.Result[*Receipt] {
	fmt.Printf("  Generating receipt for order %s\n", shipment.OrderID)

	receipt := &Receipt{
		OrderID:      shipment.OrderID,
		Amount:       payment.Amount,
		TrackingCode: shipment.TrackingCode,
		Timestamp:    time.Now(),
	}

	return result.Ok(receipt)
}

type ProcessingContext struct {
	Order    *Order
	Payment  *Payment
	Shipment *Shipment
}

func processOrder(order *Order) result.Result[*Receipt] {
	validatedOrder := result.Ok(order).
		AndThen(validateOrder).
		AndThen(checkInventory)

	if validatedOrder.IsErr() {
		return result.Err[*Receipt](validatedOrder.Error())
	}

	paymentResult := processPayment(validatedOrder.Unwrap()).
		RecoverWith(result.KindInfrastructure, func(err error) result.Result[*Payment] {
			fmt.Println("  Payment failed, retrying...")
			return processPayment(order)
		})

	if paymentResult.IsErr() {
		return result.Err[*Receipt](paymentResult.Error())
	}

	payment := paymentResult.Unwrap()

	return result.AndThenMap(
		createShipment(payment),
		func(shipment *Shipment) result.Result[*Receipt] {
			return generateReceipt(shipment, payment)
		},
	)
}

func example5_OrderProcessing() {
	fmt.Println("=== Example 5: Order Processing Pipeline ===")

	// Valid order
	fmt.Println("\nProcessing valid order:")
	order1 := &Order{
		ID:     "ORD001",
		Items:  []string{"item1", "item2"},
		Total:  99.99,
		Status: "pending",
	}

	processOrder(order1).
		Inspect(func(receipt *Receipt) {
			fmt.Printf("✓ Order processed successfully!\n")
			fmt.Printf("  Order ID: %s\n", receipt.OrderID)
			fmt.Printf("  Amount: $%.2f\n", receipt.Amount)
			fmt.Printf("  Tracking: %s\n", receipt.TrackingCode)
		}).
		InspectErr(func(err error) {
			fmt.Printf("✗ Order processing failed: %v\n", err)
		})

	// Invalid order - no items
	fmt.Println("\nProcessing invalid order (no items):")
	order2 := &Order{
		ID:     "ORD002",
		Items:  []string{},
		Total:  50.00,
		Status: "pending",
	}

	processOrder(order2).
		InspectErr(func(err error) {
			fmt.Printf("✗ Order failed validation: %v\n", err)
			if meta := result.MetaOf(err); meta != nil {
				fmt.Printf("  Metadata: %+v\n", meta)
			}
		})

	// Invalid order - zero total
	fmt.Println("\nProcessing invalid order (zero total):")
	order3 := &Order{
		ID:     "ORD003",
		Items:  []string{"item1"},
		Total:  0,
		Status: "pending",
	}

	processOrder(order3).
		InspectErr(func(err error) {
			fmt.Printf("✗ Order failed validation: %v\n", err)
		})

	fmt.Println()
}

// ============================================================================
// Example 6: Complex Pipeline with Map and MapErr
// ============================================================================

func example6_MapOperations() {
	fmt.Println("=== Example 6: Map and MapErr Operations ===")

	// Map transforms success values
	fmt.Println("\nMap transformation:")
	result.Ok(10).
		Map(func(n int) int {
			fmt.Printf("  Doubling: %d -> %d\n", n, n*2)
			return n * 2
		}).
		Map(func(n int) int {
			fmt.Printf("  Adding 5: %d -> %d\n", n, n+5)
			return n + 5
		}).
		Inspect(func(n int) {
			fmt.Printf("✓ Final value: %d\n", n)
		})

	// MapErr transforms errors
	fmt.Println("\nMapErr transformation:")
	result.Err[int](errors.New("base error")).
		MapErr(func(err error) error {
			fmt.Printf("  Wrapping error: %v\n", err)
			return fmt.Errorf("layer 1: %w", err)
		}).
		MapErr(func(err error) error {
			fmt.Printf("  Wrapping again: %v\n", err)
			return fmt.Errorf("layer 2: %w", err)
		}).
		InspectErr(func(err error) {
			fmt.Printf("✗ Final error: %v\n", err)
		})

	fmt.Println()
}

// ============================================================================
// Main
// ============================================================================

func main() {
	fmt.Println("╔════════════════════════════════════════╗")
	fmt.Println("║  Pipeline Examples - Railway Pattern  ║")
	fmt.Println("╚════════════════════════════════════════╝")
	fmt.Println()

	example1_BasicRailway()
	example2_TypeTransformations()
	example3_FallbackChains()
	example4_RecoveryPatterns()
	example5_OrderProcessing()
	example6_MapOperations()

	fmt.Println("╔════════════════════════════════════════╗")
	fmt.Println("║  All pipeline examples completed!      ║")
	fmt.Println("╚════════════════════════════════════════╝")
}
