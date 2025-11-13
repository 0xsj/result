package result

import (
	"context"
	"testing"
)

// ============================================================================
// Context Setter Tests
// ============================================================================

func TestWithRequestID(t *testing.T) {
	ctx := context.Background()
	requestID := "req_123"

	ctx = WithRequestID(ctx, requestID)

	got := GetRequestID(ctx)
	if got != requestID {
		t.Errorf("GetRequestID() = %v, want %v", got, requestID)
	}
}

func TestWithTraceID(t *testing.T) {
	ctx := context.Background()
	traceID := "trace_xyz"

	ctx = WithTraceID(ctx, traceID)

	got := GetTraceID(ctx)
	if got != traceID {
		t.Errorf("GetTraceID() = %v, want %v", got, traceID)
	}
}

func TestWithSpanID(t *testing.T) {
	ctx := context.Background()
	spanID := "span_abc"

	ctx = WithSpanID(ctx, spanID)

	got := GetSpanID(ctx)
	if got != spanID {
		t.Errorf("GetSpanID() = %v, want %v", got, spanID)
	}
}

func TestWithUserID(t *testing.T) {
	ctx := context.Background()
	userID := "user_456"

	ctx = WithUserID(ctx, userID)

	got := GetUserID(ctx)
	if got != userID {
		t.Errorf("GetUserID() = %v, want %v", got, userID)
	}
}

func TestWithTenantID(t *testing.T) {
	ctx := context.Background()
	tenantID := "tenant_789"

	ctx = WithTenantID(ctx, tenantID)

	got := GetTenantID(ctx)
	if got != tenantID {
		t.Errorf("GetTenantID() = %v, want %v", got, tenantID)
	}
}

// ============================================================================
// Context Getter Tests
// ============================================================================

func TestGetRequestID_NotPresent(t *testing.T) {
	ctx := context.Background()

	got := GetRequestID(ctx)
	if got != "" {
		t.Errorf("GetRequestID() = %v, want empty string", got)
	}
}

func TestGetTraceID_NotPresent(t *testing.T) {
	ctx := context.Background()

	got := GetTraceID(ctx)
	if got != "" {
		t.Errorf("GetTraceID() = %v, want empty string", got)
	}
}

func TestGetSpanID_NotPresent(t *testing.T) {
	ctx := context.Background()

	got := GetSpanID(ctx)
	if got != "" {
		t.Errorf("GetSpanID() = %v, want empty string", got)
	}
}

func TestGetUserID_NotPresent(t *testing.T) {
	ctx := context.Background()

	got := GetUserID(ctx)
	if got != "" {
		t.Errorf("GetUserID() = %v, want empty string", got)
	}
}

func TestGetTenantID_NotPresent(t *testing.T) {
	ctx := context.Background()

	got := GetTenantID(ctx)
	if got != "" {
		t.Errorf("GetTenantID() = %v, want empty string", got)
	}
}

func TestGetRequestID_WrongType(t *testing.T) {
	// Test that wrong type returns empty string (doesn't panic)
	ctx := context.WithValue(context.Background(), requestIDKey, 123)

	got := GetRequestID(ctx)
	if got != "" {
		t.Errorf("GetRequestID() with wrong type = %v, want empty string", got)
	}
}

// ============================================================================
// Multiple Context Values
// ============================================================================

func TestMultipleContextValues(t *testing.T) {
	ctx := context.Background()

	ctx = WithRequestID(ctx, "req_123")
	ctx = WithTraceID(ctx, "trace_xyz")
	ctx = WithUserID(ctx, "user_456")
	ctx = WithTenantID(ctx, "tenant_789")

	if GetRequestID(ctx) != "req_123" {
		t.Error("Failed to get request_id")
	}

	if GetTraceID(ctx) != "trace_xyz" {
		t.Error("Failed to get trace_id")
	}

	if GetUserID(ctx) != "user_456" {
		t.Error("Failed to get user_id")
	}

	if GetTenantID(ctx) != "tenant_789" {
		t.Error("Failed to get tenant_id")
	}
}

func TestContextChaining(t *testing.T) {
	ctx := WithRequestID(
		WithTraceID(
			WithUserID(context.Background(), "user_1"),
			"trace_1",
		),
		"req_1",
	)

	if GetRequestID(ctx) != "req_1" {
		t.Error("Chained context should have request_id")
	}

	if GetTraceID(ctx) != "trace_1" {
		t.Error("Chained context should have trace_id")
	}

	if GetUserID(ctx) != "user_1" {
		t.Error("Chained context should have user_id")
	}
}

// ============================================================================
// Result.WithContext Tests
// ============================================================================

func TestResultWithContext_Ok(t *testing.T) {
	ctx := WithRequestID(context.Background(), "req_123")

	r := Ok(42).WithContext(ctx)

	if !r.IsOk() {
		t.Error("WithContext should preserve Ok result")
	}

	// Ok result should not have metadata attached
	if r.meta != nil {
		t.Error("WithContext should not attach metadata to Ok result")
	}
}

func TestResultWithContext_Err(t *testing.T) {
	ctx := WithRequestID(context.Background(), "req_123")

	r := Err[int](NotFound("test", "resource")).
		WithContext(ctx)

	if !r.IsErr() {
		t.Error("WithContext should preserve Err result")
	}

	if r.meta == nil {
		t.Fatal("WithContext should attach metadata to Err result")
	}

	if r.meta["request_id"] != "req_123" {
		t.Errorf("metadata[request_id] = %v, want req_123", r.meta["request_id"])
	}
}

func TestResultWithContext_MultipleIDs(t *testing.T) {
	ctx := context.Background()
	ctx = WithRequestID(ctx, "req_123")
	ctx = WithTraceID(ctx, "trace_xyz")
	ctx = WithUserID(ctx, "user_456")
	ctx = WithTenantID(ctx, "tenant_789")

	r := Err[int](NotFound("test", "resource")).
		WithContext(ctx)

	if r.meta["request_id"] != "req_123" {
		t.Error("Should attach request_id")
	}

	if r.meta["trace_id"] != "trace_xyz" {
		t.Error("Should attach trace_id")
	}

	if r.meta["user_id"] != "user_456" {
		t.Error("Should attach user_id")
	}

	if r.meta["tenant_id"] != "tenant_789" {
		t.Error("Should attach tenant_id")
	}
}

func TestResultWithContext_EmptyContext(t *testing.T) {
	ctx := context.Background()

	r := Err[int](NotFound("test", "resource")).
		WithContext(ctx)

	// Should not panic, but also shouldn't add empty values
	if r.meta != nil && len(r.meta) > 0 {
		t.Error("WithContext should not add empty values")
	}
}

func TestResultWithContext_PartialContext(t *testing.T) {
	ctx := WithRequestID(context.Background(), "req_123")
	// Only request_id set, others empty

	r := Err[int](NotFound("test", "resource")).
		WithContext(ctx)

	if r.meta["request_id"] != "req_123" {
		t.Error("Should attach present request_id")
	}

	// Should not add empty keys
	if _, exists := r.meta["trace_id"]; exists {
		t.Error("Should not add empty trace_id")
	}

	if _, exists := r.meta["user_id"]; exists {
		t.Error("Should not add empty user_id")
	}
}

func TestResultWithContext_PreservesExistingMeta(t *testing.T) {
	ctx := WithRequestID(context.Background(), "req_from_context")

	r := Err[int](NotFound("test", "resource")).
		WithMeta("request_id", "req_explicit").
		WithMeta("custom", "value").
		WithContext(ctx)

	// Should NOT override explicit request_id
	if r.meta["request_id"] != "req_explicit" {
		t.Error("WithContext should not override existing metadata")
	}

	// Should preserve other metadata
	if r.meta["custom"] != "value" {
		t.Error("WithContext should preserve existing metadata")
	}
}

func TestResultWithContext_PreservesOp(t *testing.T) {
	ctx := WithRequestID(context.Background(), "req_123")

	r := Err[int](NotFound("test", "resource")).
		WithOp("Service.GetUser").
		WithContext(ctx)

	if r.op != "Service.GetUser" {
		t.Error("WithContext should preserve op")
	}

	if r.meta["request_id"] != "req_123" {
		t.Error("WithContext should attach request_id")
	}
}

// ============================================================================
// Result.WithContextOp Tests
// ============================================================================

func TestResultWithContextOp_Ok(t *testing.T) {
	ctx := WithRequestID(context.Background(), "req_123")

	r := Ok(42).WithContextOp(ctx, "Service.GetUser")

	if !r.IsOk() {
		t.Error("WithContextOp should preserve Ok result")
	}

	if r.op != "Service.GetUser" {
		t.Error("WithContextOp should set op")
	}

	// Ok result should not have metadata
	if r.meta != nil {
		t.Error("WithContextOp should not attach metadata to Ok result")
	}
}

func TestResultWithContextOp_Err(t *testing.T) {
	ctx := WithRequestID(context.Background(), "req_123")

	r := Err[int](NotFound("test", "resource")).
		WithContextOp(ctx, "Service.GetUser")

	if !r.IsErr() {
		t.Error("WithContextOp should preserve Err result")
	}

	if r.op != "Service.GetUser" {
		t.Error("WithContextOp should set op")
	}

	if r.meta["request_id"] != "req_123" {
		t.Error("WithContextOp should attach request_id")
	}
}

func TestResultWithContextOp_CombinesOpAndContext(t *testing.T) {
	ctx := context.Background()
	ctx = WithRequestID(ctx, "req_123")
	ctx = WithUserID(ctx, "user_456")

	r := Err[int](NotFound("test", "resource")).
		WithContextOp(ctx, "UserService.FindByID")

	if r.op != "UserService.FindByID" {
		t.Errorf("op = %v, want UserService.FindByID", r.op)
	}

	if r.meta["request_id"] != "req_123" {
		t.Error("Should have request_id from context")
	}

	if r.meta["user_id"] != "user_456" {
		t.Error("Should have user_id from context")
	}
}

// ============================================================================
// Integration Tests
// ============================================================================

func TestContextPipeline(t *testing.T) {
	ctx := context.Background()
	ctx = WithRequestID(ctx, "req_pipeline")
	ctx = WithUserID(ctx, "user_pipeline")

	step1 := func(n int) Result[int] {
		if n < 0 {
			return Err[int](Validation("step1", "negative", nil)).
				WithContext(ctx)
		}
		return Ok(n * 2)
	}

	r := Ok(10).
		AndThen(step1).
		AndThen(func(n int) Result[int] {
			if n > 100 {
				return Err[int](Domain("step2", "too large")).
					WithContext(ctx)
			}
			return Ok(n + 5)
		})

	if !r.IsOk() {
		t.Error("Pipeline should succeed")
	}

	if r.value != 25 {
		t.Errorf("value = %v, want 25", r.value)
	}

	// Now test error path
	r2 := Ok(-1).
		AndThen(step1)

	if !r2.IsErr() {
		t.Error("Pipeline should fail on negative")
	}

	if r2.meta["request_id"] != "req_pipeline" {
		t.Error("Error should have request_id from context")
	}

	if r2.meta["user_id"] != "user_pipeline" {
		t.Error("Error should have user_id from context")
	}
}

func TestContextWithRecovery(t *testing.T) {
	ctx := WithRequestID(context.Background(), "req_recovery")

	metaChecked := false

	r := Err[int](NotFound("test", "resource")).
		WithContext(ctx).
		InspectErr(func(err error) {
			// We can check the Result's metadata before recovery
			metaChecked = true
		}).
		Recover(KindNotFound, func(err error) int {
			return 99
		})

	if !metaChecked {
		t.Error("Should have inspected error")
	}

	if !r.IsOk() {
		t.Error("Should recover from NotFound")
	}

	if r.value != 99 {
		t.Errorf("value = %v, want 99", r.value)
	}

}

func TestContextChainPreservation(t *testing.T) {
	ctx := context.Background()
	ctx = WithRequestID(ctx, "req_chain")
	ctx = WithTraceID(ctx, "trace_chain")

	r := Err[int](NotFound("test", "resource")).
		WithContext(ctx).
		Map(func(n int) int { return n * 2 }).
		MapErr(func(err error) error {
			return Wrap(err, "wrapper")
		})

	// Metadata should be preserved through transformations
	if r.meta["request_id"] != "req_chain" {
		t.Error("Metadata should be preserved through Map/MapErr")
	}

	if r.meta["trace_id"] != "trace_chain" {
		t.Error("Metadata should be preserved through Map/MapErr")
	}
}

func TestContextWithInspect(t *testing.T) {
	ctx := WithRequestID(context.Background(), "req_inspect")

	inspected := false

	r := Err[int](NotFound("test", "resource")).
		WithContext(ctx).
		InspectErr(func(err error) {
			inspected = true
		})

	if !inspected {
		t.Error("InspectErr should be called")
	}

	// Check metadata on Result (not on the error)
	if r.meta["request_id"] != "req_inspect" {
		t.Error("Result should have context metadata")
	}
}
