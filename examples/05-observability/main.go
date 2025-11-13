// Package main demonstrates observability patterns with the result package.
//
// This example covers:
// - Context propagation (request ID, trace ID, user ID)
// - Structured logging with correlation IDs
// - Metrics recording (success/error counts, latency)
// - Error tracking with rich metadata
// - Tap and Inspect patterns for observability
//
// Run:
//
//	go run main.go
package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/0xsj/result"
)

// ============================================================================
// Simple Logger (simulating structured logging)
// ============================================================================

type Logger struct {
	name string
}

func NewLogger(name string) *Logger {
	return &Logger{name: name}
}

func (l *Logger) Info(ctx context.Context, msg string, fields ...interface{}) {
	l.log(ctx, "INFO", msg, fields...)
}

func (l *Logger) Error(ctx context.Context, msg string, fields ...interface{}) {
	l.log(ctx, "ERROR", msg, fields...)
}

func (l *Logger) Debug(ctx context.Context, msg string, fields ...interface{}) {
	l.log(ctx, "DEBUG", msg, fields...)
}

func (l *Logger) log(ctx context.Context, level, msg string, fields ...interface{}) {
	timestamp := time.Now().Format("15:04:05.000")
	requestID := result.GetRequestID(ctx)
	traceID := result.GetTraceID(ctx)
	userID := result.GetUserID(ctx)

	fmt.Printf("[%s] %s [%s]", timestamp, level, l.name)

	if requestID != "" {
		fmt.Printf(" [req:%s]", requestID)
	}
	if traceID != "" {
		fmt.Printf(" [trace:%s]", traceID)
	}
	if userID != "" {
		fmt.Printf(" [user:%s]", userID)
	}

	fmt.Printf(" %s", msg)

	// Print fields
	for i := 0; i < len(fields); i += 2 {
		if i+1 < len(fields) {
			fmt.Printf(" %v=%v", fields[i], fields[i+1])
		}
	}

	fmt.Println()
}

// ============================================================================
// Simple Metrics (simulating metrics recording)
// ============================================================================

type Metrics struct {
	successCounts map[string]int
	errorCounts   map[string]int
	latencies     map[string][]time.Duration
}

func NewMetrics() *Metrics {
	return &Metrics{
		successCounts: make(map[string]int),
		errorCounts:   make(map[string]int),
		latencies:     make(map[string][]time.Duration),
	}
}

func (m *Metrics) RecordSuccess(operation string) {
	m.successCounts[operation]++
	fmt.Printf("  [Metrics] %s.success = %d\n", operation, m.successCounts[operation])
}

func (m *Metrics) RecordError(operation string) {
	m.errorCounts[operation]++
	fmt.Printf("  [Metrics] %s.error = %d\n", operation, m.errorCounts[operation])
}

func (m *Metrics) RecordLatency(operation string, duration time.Duration) {
	m.latencies[operation] = append(m.latencies[operation], duration)
	fmt.Printf("  [Metrics] %s.latency = %v\n", operation, duration)
}

func (m *Metrics) PrintSummary() {
	fmt.Println("\n=== Metrics Summary ===")
	for op, count := range m.successCounts {
		fmt.Printf("  %s: %d successes, %d errors\n", op, count, m.errorCounts[op])
		if latencies, ok := m.latencies[op]; ok && len(latencies) > 0 {
			avg := time.Duration(0)
			for _, l := range latencies {
				avg += l
			}
			avg = avg / time.Duration(len(latencies))
			fmt.Printf("    avg latency: %v\n", avg)
		}
	}
}

// ============================================================================
// Example 1: Basic Context Propagation
// ============================================================================

type User struct {
	ID   string
	Name string
}

var (
	log     = NewLogger("UserService")
	metrics = NewMetrics()
)

func fetchUser(ctx context.Context, id string) result.Result[*User] {
	log.Info(ctx, "fetching user", "user_id", id)

	// Simulate some work
	time.Sleep(10 * time.Millisecond)

	if id == "404" {
		return result.Err[*User](
			result.NotFound("fetchUser", "user"),
		).WithContext(ctx)
	}

	return result.Ok(&User{ID: id, Name: fmt.Sprintf("User-%s", id)}).
		WithContext(ctx)
}

func example1_BasicContext() {
	fmt.Println("=== Example 1: Basic Context Propagation ===\n")

	// Create context with request ID and user ID
	ctx := context.Background()
	ctx = result.WithRequestID(ctx, "req_001")
	ctx = result.WithUserID(ctx, "alice")

	log.Info(ctx, "starting request")

	fetchUser(ctx, "123").Match(
		func(user *User) {
			log.Info(ctx, "user fetched successfully", "user_name", user.Name)
		},
		func(err error) {
			log.Error(ctx, "failed to fetch user", "error", err)
		},
	)

	fmt.Println()
}

// ============================================================================
// Example 2: Inspect and Tap for Observability
// ============================================================================

func getUserWithObservability(ctx context.Context, id string) result.Result[*User] {
	start := time.Now()

	return fetchUser(ctx, id).
		Tap(
			func(user *User) {
				// Success path
				duration := time.Since(start)
				log.Info(ctx, "operation succeeded", 
					"user_id", user.ID,
					"duration_ms", duration.Milliseconds(),
				)
				metrics.RecordSuccess("get_user")
				metrics.RecordLatency("get_user", duration)
			},
			func(err error) {
				// Error path
				duration := time.Since(start)
				log.Error(ctx, "operation failed",
					"error", err,
					"error_kind", result.KindOf(err).String(),
					"duration_ms", duration.Milliseconds(),
				)
				metrics.RecordError("get_user")
				metrics.RecordLatency("get_user", duration)
			},
		)
}

func example2_InspectAndTap() {
	fmt.Println("=== Example 2: Inspect and Tap for Observability ===\n")

	ctx := result.WithRequestID(context.Background(), "req_002")

	// Success case
	log.Info(ctx, "attempting to get user 456")
	getUserWithObservability(ctx, "456").Match(
		func(user *User) {
			log.Info(ctx, "user retrieved", "user_name", user.Name)
		},
		func(err error) {
			log.Error(ctx, "retrieval failed", "error", err)
		},
	)

	// Error case
	log.Info(ctx, "attempting to get user 404")
	getUserWithObservability(ctx, "404").Match(
		func(user *User) {
			log.Info(ctx, "user retrieved", "user_name", user.Name)
		},
		func(err error) {
			log.Error(ctx, "retrieval failed", "error", err)
		},
	)

	fmt.Println()
}

// ============================================================================
// Example 3: Pipeline with Full Observability
// ============================================================================

type Order struct {
	ID       string
	UserID   string
	Total    float64
	Status   string
}

func validateOrder(ctx context.Context, order *Order) result.Result[*Order] {
	start := time.Now()
	log.Debug(ctx, "validating order", "order_id", order.ID)

	// Simulate validation
	time.Sleep(5 * time.Millisecond)

	if order.Total <= 0 {
		return result.Err[*Order](
			result.Validation("validateOrder", "invalid total", map[string]interface{}{
				"order_id": order.ID,
				"total":    order.Total,
			}),
		).
			WithContext(ctx).
			Tap(nil, func(err error) {
				log.Error(ctx, "validation failed",
					"order_id", order.ID,
					"duration_ms", time.Since(start).Milliseconds(),
				)
				metrics.RecordError("validate_order")
			})
	}

	order.Status = "validated"
	
	return result.Ok(order).
		Inspect(func(o *Order) {
			log.Debug(ctx, "validation succeeded",
				"order_id", o.ID,
				"duration_ms", time.Since(start).Milliseconds(),
			)
			metrics.RecordSuccess("validate_order")
		})
}

func processPayment(ctx context.Context, order *Order) result.Result[*Order] {
	start := time.Now()
	log.Debug(ctx, "processing payment", "order_id", order.ID, "amount", order.Total)

	// Simulate payment processing
	time.Sleep(20 * time.Millisecond)

	// Simulate random failures (10% chance)
	if rand.Float32() < 0.1 {
		return result.Err[*Order](
			result.Infrastructure("processPayment", fmt.Errorf("payment gateway timeout")),
		).
			WithContext(ctx).
			Tap(nil, func(err error) {
				log.Error(ctx, "payment failed",
					"order_id", order.ID,
					"duration_ms", time.Since(start).Milliseconds(),
				)
				metrics.RecordError("process_payment")
			})
	}

	order.Status = "paid"
	
	return result.Ok(order).
		Inspect(func(o *Order) {
			log.Info(ctx, "payment processed",
				"order_id", o.ID,
				"amount", o.Total,
				"duration_ms", time.Since(start).Milliseconds(),
			)
			metrics.RecordSuccess("process_payment")
		})
}

func shipOrder(ctx context.Context, order *Order) result.Result[*Order] {
	start := time.Now()
	log.Debug(ctx, "shipping order", "order_id", order.ID)

	// Simulate shipping
	time.Sleep(15 * time.Millisecond)

	order.Status = "shipped"
	
	return result.Ok(order).
		Inspect(func(o *Order) {
			log.Info(ctx, "order shipped",
				"order_id", o.ID,
				"duration_ms", time.Since(start).Milliseconds(),
			)
			metrics.RecordSuccess("ship_order")
		})
}

func processOrderPipeline(ctx context.Context, order *Order) result.Result[*Order] {
	overallStart := time.Now()
	
	log.Info(ctx, "starting order processing", "order_id", order.ID)

	result := validateOrder(ctx, order).
		AndThen(func(o *Order) result.Result[*Order] {
			return processPayment(ctx, o)
		}).
		RecoverWith(result.KindInfrastructure, func(err error) result.Result[*Order] {
			log.Info(ctx, "retrying payment after infrastructure error")
			return processPayment(ctx, order)
		}).
		AndThen(func(o *Order) result.Result[*Order] {
			return shipOrder(ctx, o)
		}).
		Tap(
			func(o *Order) {
				duration := time.Since(overallStart)
				log.Info(ctx, "order processing completed",
					"order_id", o.ID,
					"status", o.Status,
					"total_duration_ms", duration.Milliseconds(),
				)
				metrics.RecordLatency("process_order", duration)
			},
			func(err error) {
				duration := time.Since(overallStart)
				log.Error(ctx, "order processing failed",
					"order_id", order.ID,
					"error", err,
					"error_kind", result.KindOf(err).String(),
					"total_duration_ms", duration.Milliseconds(),
				)
			},
		)

	return result
}

func example3_PipelineObservability() {
	fmt.Println("=== Example 3: Pipeline with Full Observability ===\n")

	// Process valid order
	ctx1 := context.Background()
	ctx1 = result.WithRequestID(ctx1, "req_003")
	ctx1 = result.WithTraceID(ctx1, "trace_abc")
	ctx1 = result.WithUserID(ctx1, "alice")

	order1 := &Order{
		ID:     "ORD-001",
		UserID: "alice",
		Total:  99.99,
		Status: "pending",
	}

	processOrderPipeline(ctx1, order1)

	// Process invalid order
	ctx2 := result.WithRequestID(context.Background(), "req_004")
	order2 := &Order{
		ID:     "ORD-002",
		UserID: "bob",
		Total:  0, // Invalid
		Status: "pending",
	}

	processOrderPipeline(ctx2, order2)

	fmt.Println()
}

// ============================================================================
// Example 4: Error Metadata and Tracking
// ============================================================================

func complexOperation(ctx context.Context, id string) result.Result[string] {
	log.Info(ctx, "starting complex operation", "resource_id", id)

	if id == "error" {
		// Create error with rich metadata
		return result.Err[string](
			result.Domain("complexOperation", "business rule violated"),
		).
			WithContext(ctx).
			WithMeta("resource_id", id).
			WithMeta("violation_type", "insufficient_funds").
			WithMeta("required_amount", 100.0).
			WithMeta("available_amount", 50.0).
			InspectErr(func(err error) {
				meta := result.MetaOf(err)
				log.Error(ctx, "operation failed with rich context",
					"error", err,
					"metadata", fmt.Sprintf("%+v", meta),
				)
			})
	}

	return result.Ok("success")
}

func example4_ErrorMetadata() {
	fmt.Println("=== Example 4: Error Metadata and Tracking ===\n")

	ctx := context.Background()
	ctx = result.WithRequestID(ctx, "req_005")
	ctx = result.WithUserID(ctx, "charlie")

	complexOperation(ctx, "error").Match(
		func(s string) {
			log.Info(ctx, "operation succeeded", "result", s)
		},
		func(err error) {
			// Extract and log all metadata
			meta := result.MetaOf(err)
			
			log.Error(ctx, "extracted error details",
				"error", err.Error(),
				"kind", result.KindOf(err).String(),
				"op", result.OpOf(err),
			)

			if meta != nil {
				fmt.Println("\n  Error Metadata:")
				for k, v := range meta {
					fmt.Printf("    %s: %v\n", k, v)
				}
			}
		},
	)

	fmt.Println()
}

// ============================================================================
// Example 5: Distributed Tracing Simulation
// ============================================================================

func serviceA(ctx context.Context) result.Result[string] {
	log.Info(ctx, "Service A: processing request")
	time.Sleep(10 * time.Millisecond)
	
	return result.Ok("data_from_A").
		Inspect(func(s string) {
			log.Info(ctx, "Service A: completed", "output", s)
		})
}

func serviceB(ctx context.Context, input string) result.Result[string] {
	log.Info(ctx, "Service B: processing request", "input", input)
	time.Sleep(15 * time.Millisecond)
	
	return result.Ok(input + "_processed_by_B").
		Inspect(func(s string) {
			log.Info(ctx, "Service B: completed", "output", s)
		})
}

func serviceC(ctx context.Context, input string) result.Result[string] {
	log.Info(ctx, "Service C: processing request", "input", input)
	time.Sleep(12 * time.Millisecond)
	
	return result.Ok(input + "_finalized_by_C").
		Inspect(func(s string) {
			log.Info(ctx, "Service C: completed", "output", s)
		})
}

func distributedOperation(ctx context.Context) result.Result[string] {
	start := time.Now()
	log.Info(ctx, "starting distributed operation")

	result := serviceA(ctx).
		AndThen(func(data string) result.Result[string] {
			return serviceB(ctx, data)
		}).
		AndThen(func(data string) result.Result[string] {
			return serviceC(ctx, data)
		}).
		Tap(
			func(final string) {
				duration := time.Since(start)
				log.Info(ctx, "distributed operation completed",
					"final_output", final,
					"total_duration_ms", duration.Milliseconds(),
				)
			},
			func(err error) {
				duration := time.Since(start)
				log.Error(ctx, "distributed operation failed",
					"error", err,
					"total_duration_ms", duration.Milliseconds(),
				)
			},
		)

	return result
}

func example5_DistributedTracing() {
	fmt.Println("=== Example 5: Distributed Tracing Simulation ===\n")

	ctx := context.Background()
	ctx = result.WithRequestID(ctx, "req_006")
	ctx = result.WithTraceID(ctx, "trace_distributed_123")

	distributedOperation(ctx)

	fmt.Println()
}

// ============================================================================
// Main
// ============================================================================

func main() {
	fmt.Println("╔════════════════════════════════════════╗")
	fmt.Println("║  Observability Examples                ║")
	fmt.Println("╚════════════════════════════════════════╝")
	fmt.Println()

	// Seed random for payment failures
	rand.Seed(time.Now().UnixNano())

	example1_BasicContext()
	example2_InspectAndTap()
	example3_PipelineObservability()
	example4_ErrorMetadata()
	example5_DistributedTracing()

	// Print metrics summary
	metrics.PrintSummary()

	fmt.Println("\n╔════════════════════════════════════════╗")
	fmt.Println("║  All observability examples completed! ║")
	fmt.Println("╚════════════════════════════════════════╝")
}