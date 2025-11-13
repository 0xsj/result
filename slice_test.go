package result

import (
	"context"
	"errors"
	"testing"
)

// ============================================================================
// Collect Tests
// ============================================================================

func TestCollect_AllOk(t *testing.T) {
	results := []Result[int]{
		Ok(1),
		Ok(2),
		Ok(3),
	}

	r := Collect(results)

	if !r.IsOk() {
		t.Error("Collect should return Ok when all results are Ok")
	}

	if len(r.value) != 3 {
		t.Errorf("Collect value length = %d, want 3", len(r.value))
	}

	expected := []int{1, 2, 3}
	for i, v := range r.value {
		if v != expected[i] {
			t.Errorf("Collect value[%d] = %v, want %v", i, v, expected[i])
		}
	}
}

func TestCollect_OneErr(t *testing.T) {
	testErr := errors.New("test error")
	results := []Result[int]{
		Ok(1),
		Err[int](testErr),
		Ok(3),
	}

	r := Collect(results)

	if !r.IsErr() {
		t.Error("Collect should return Err when any result is Err")
	}

	if r.err != testErr {
		t.Errorf("Collect err = %v, want %v", r.err, testErr)
	}
}

func TestCollect_FirstErr(t *testing.T) {
	err1 := errors.New("error 1")
	err2 := errors.New("error 2")

	results := []Result[int]{
		Ok(1),
		Err[int](err1),
		Err[int](err2),
	}

	r := Collect(results)

	if !r.IsErr() {
		t.Error("Collect should return Err")
	}

	// Should return first error
	if r.err != err1 {
		t.Errorf("Collect should return first error, got %v", r.err)
	}
}

func TestCollect_Empty(t *testing.T) {
	results := []Result[int]{}

	r := Collect(results)

	if !r.IsOk() {
		t.Error("Collect should return Ok for empty slice")
	}

	if len(r.value) != 0 {
		t.Errorf("Collect value length = %d, want 0", len(r.value))
	}
}

func TestCollect_AddsIndex(t *testing.T) {
	results := []Result[int]{
		Ok(1),
		Ok(2),
		Err[int](errors.New("error at 2")),
		Ok(4),
	}

	r := Collect(results)

	if !r.IsErr() {
		t.Error("Collect should return Err")
	}

	if r.meta == nil {
		t.Fatal("Collect should add metadata")
	}

	if r.meta["index"] != 2 {
		t.Errorf("Collect meta[index] = %v, want 2", r.meta["index"])
	}
}

func TestCollect_PreservesOp(t *testing.T) {
	results := []Result[int]{
		Ok(1),
		Err[int](NotFound("test", "resource")).WithOp("Service.Get"),
	}

	r := Collect(results)

	if !r.IsErr() {
		t.Error("Collect should return Err")
	}

	if r.op != "Service.Get" {
		t.Errorf("Collect should preserve op, got %v", r.op)
	}
}

func TestCollect_PreservesExistingMeta(t *testing.T) {
	results := []Result[int]{
		Ok(1),
		Err[int](NotFound("test", "resource")).
			WithMeta("custom", "value"),
	}

	r := Collect(results)

	if r.meta["custom"] != "value" {
		t.Error("Collect should preserve existing metadata")
	}

	if r.meta["index"] != 1 {
		t.Error("Collect should add index metadata")
	}
}

// ============================================================================
// Partition Tests
// ============================================================================

func TestPartition_AllOk(t *testing.T) {
	results := []Result[int]{
		Ok(1),
		Ok(2),
		Ok(3),
	}

	values, errs := Partition(results)

	if len(values) != 3 {
		t.Errorf("Partition values length = %d, want 3", len(values))
	}

	if len(errs) != 0 {
		t.Errorf("Partition errors length = %d, want 0", len(errs))
	}

	expected := []int{1, 2, 3}
	for i, v := range values {
		if v != expected[i] {
			t.Errorf("Partition value[%d] = %v, want %v", i, v, expected[i])
		}
	}
}

func TestPartition_AllErr(t *testing.T) {
	err1 := errors.New("error 1")
	err2 := errors.New("error 2")

	results := []Result[int]{
		Err[int](err1),
		Err[int](err2),
	}

	values, errs := Partition(results)

	if len(values) != 0 {
		t.Errorf("Partition values length = %d, want 0", len(values))
	}

	if len(errs) != 2 {
		t.Errorf("Partition errors length = %d, want 2", len(errs))
	}

	if errs[0] != err1 {
		t.Error("Partition should preserve error order")
	}

	if errs[1] != err2 {
		t.Error("Partition should preserve error order")
	}
}

func TestPartition_Mixed(t *testing.T) {
	err1 := errors.New("error 1")
	err2 := errors.New("error 2")

	results := []Result[int]{
		Ok(1),
		Err[int](err1),
		Ok(3),
		Err[int](err2),
		Ok(5),
	}

	values, errs := Partition(results)

	if len(values) != 3 {
		t.Errorf("Partition values length = %d, want 3", len(values))
	}

	if len(errs) != 2 {
		t.Errorf("Partition errors length = %d, want 2", len(errs))
	}

	expectedValues := []int{1, 3, 5}
	for i, v := range values {
		if v != expectedValues[i] {
			t.Errorf("Partition value[%d] = %v, want %v", i, v, expectedValues[i])
		}
	}

	if errs[0] != err1 || errs[1] != err2 {
		t.Error("Partition should preserve error order")
	}
}

func TestPartition_Empty(t *testing.T) {
	results := []Result[int]{}

	values, errs := Partition(results)

	if len(values) != 0 {
		t.Errorf("Partition values length = %d, want 0", len(values))
	}

	if len(errs) != 0 {
		t.Errorf("Partition errors length = %d, want 0", len(errs))
	}
}

// ============================================================================
// CollectWithContext Tests
// ============================================================================

func TestCollectWithContext_AllOk(t *testing.T) {
	ctx := WithRequestID(context.Background(), "req_123")

	results := []Result[int]{
		Ok(1),
		Ok(2),
		Ok(3),
	}

	r := CollectWithContext(ctx, results)

	if !r.IsOk() {
		t.Error("CollectWithContext should return Ok when all results are Ok")
	}

	if len(r.value) != 3 {
		t.Errorf("CollectWithContext value length = %d, want 3", len(r.value))
	}
}

func TestCollectWithContext_OneErr(t *testing.T) {
	ctx := WithRequestID(context.Background(), "req_123")

	results := []Result[int]{
		Ok(1),
		Err[int](errors.New("test error")),
		Ok(3),
	}

	r := CollectWithContext(ctx, results)

	if !r.IsErr() {
		t.Error("CollectWithContext should return Err")
	}

	if r.meta == nil {
		t.Fatal("CollectWithContext should add metadata")
	}

	if r.meta["request_id"] != "req_123" {
		t.Errorf("CollectWithContext should add request_id, got %v", r.meta["request_id"])
	}

	if r.meta["index"] != 1 {
		t.Errorf("CollectWithContext should add index, got %v", r.meta["index"])
	}
}

func TestCollectWithContext_MultipleContextValues(t *testing.T) {
	ctx := context.Background()
	ctx = WithRequestID(ctx, "req_123")
	ctx = WithTraceID(ctx, "trace_xyz")
	ctx = WithUserID(ctx, "user_456")

	results := []Result[int]{
		Err[int](errors.New("test error")),
	}

	r := CollectWithContext(ctx, results)

	if r.meta["request_id"] != "req_123" {
		t.Error("Should add request_id")
	}

	if r.meta["trace_id"] != "trace_xyz" {
		t.Error("Should add trace_id")
	}

	if r.meta["user_id"] != "user_456" {
		t.Error("Should add user_id")
	}

	if r.meta["index"] != 0 {
		t.Error("Should add index")
	}
}

func TestCollectWithContext_EmptyContext(t *testing.T) {
	ctx := context.Background()

	results := []Result[int]{
		Err[int](errors.New("test error")),
	}

	r := CollectWithContext(ctx, results)

	if !r.IsErr() {
		t.Error("CollectWithContext should return Err")
	}

	// Should have index but not context values
	if r.meta["index"] != 0 {
		t.Error("Should add index")
	}

	// Should not add empty context values
	if _, exists := r.meta["request_id"]; exists {
		t.Error("Should not add empty request_id")
	}
}

func TestCollectWithContext_PreservesExistingMeta(t *testing.T) {
	ctx := WithRequestID(context.Background(), "req_from_context")

	results := []Result[int]{
		Err[int](errors.New("test error")).
			WithMeta("request_id", "req_explicit").
			WithMeta("custom", "value"),
	}

	r := CollectWithContext(ctx, results)

	// Should NOT override explicit request_id
	if r.meta["request_id"] != "req_explicit" {
		t.Error("CollectWithContext should not override existing metadata")
	}

	// Should preserve other metadata
	if r.meta["custom"] != "value" {
		t.Error("CollectWithContext should preserve existing metadata")
	}

	// Should add index
	if r.meta["index"] != 0 {
		t.Error("CollectWithContext should add index")
	}
}

// ============================================================================
// PartitionWithContext Tests
// ============================================================================

func TestPartitionWithContext_AllOk(t *testing.T) {
	ctx := WithRequestID(context.Background(), "req_123")

	results := []Result[int]{
		Ok(1),
		Ok(2),
		Ok(3),
	}

	values, errs := PartitionWithContext(ctx, results)

	if len(values) != 3 {
		t.Errorf("PartitionWithContext values length = %d, want 3", len(values))
	}

	if len(errs) != 0 {
		t.Errorf("PartitionWithContext errors length = %d, want 0", len(errs))
	}
}

func TestPartitionWithContext_Mixed(t *testing.T) {
	ctx := WithRequestID(context.Background(), "req_123")

	results := []Result[int]{
		Ok(1),
		Err[int](NotFound("test", "resource")),
		Ok(3),
		Err[int](Validation("test", "invalid", nil)),
	}

	values, errs := PartitionWithContext(ctx, results)

	if len(values) != 2 {
		t.Errorf("PartitionWithContext values length = %d, want 2", len(values))
	}

	if len(errs) != 2 {
		t.Errorf("PartitionWithContext errors length = %d, want 2", len(errs))
	}

	// Check first error metadata
	meta1 := MetaOf(errs[0])
	if meta1["request_id"] != "req_123" {
		t.Error("First error should have request_id")
	}

	if meta1["index"] != 1 {
		t.Errorf("First error should have index 1, got %v", meta1["index"])
	}

	// Check second error metadata
	meta2 := MetaOf(errs[1])
	if meta2["request_id"] != "req_123" {
		t.Error("Second error should have request_id")
	}

	if meta2["index"] != 3 {
		t.Errorf("Second error should have index 3, got %v", meta2["index"])
	}
}

func TestPartitionWithContext_MultipleContextValues(t *testing.T) {
	ctx := context.Background()
	ctx = WithRequestID(ctx, "req_123")
	ctx = WithUserID(ctx, "user_456")
	ctx = WithTenantID(ctx, "tenant_789")

	results := []Result[int]{
		Err[int](errors.New("error 1")),
		Err[int](errors.New("error 2")),
	}

	_, errs := PartitionWithContext(ctx, results)

	if len(errs) != 2 {
		t.Fatalf("Expected 2 errors, got %d", len(errs))
	}

	for i, err := range errs {
		meta := MetaOf(err)

		if meta["request_id"] != "req_123" {
			t.Errorf("Error %d should have request_id", i)
		}

		if meta["user_id"] != "user_456" {
			t.Errorf("Error %d should have user_id", i)
		}

		if meta["tenant_id"] != "tenant_789" {
			t.Errorf("Error %d should have tenant_id", i)
		}

		if meta["index"] != i {
			t.Errorf("Error %d should have correct index", i)
		}
	}
}

func TestPartitionWithContext_PreservesErrorKind(t *testing.T) {
	ctx := WithRequestID(context.Background(), "req_123")

	results := []Result[int]{
		Err[int](NotFound("test", "resource")),
		Err[int](Validation("test", "invalid", nil)),
	}

	_, errs := PartitionWithContext(ctx, results)

	if len(errs) != 2 {
		t.Fatalf("Expected 2 errors, got %d", len(errs))
	}

	if KindOf(errs[0]) != KindNotFound {
		t.Errorf("First error kind = %v, want KindNotFound", KindOf(errs[0]))
	}

	if KindOf(errs[1]) != KindValidation {
		t.Errorf("Second error kind = %v, want KindValidation", KindOf(errs[1]))
	}
}

// ============================================================================
// Integration Tests
// ============================================================================

func TestCollect_WithMap(t *testing.T) {
	results := []Result[int]{
		Ok(1),
		Ok(2),
		Ok(3),
	}

	r := Collect(results).Map(func(values []int) []int {
		// Double all values
		doubled := make([]int, len(values))
		for i, v := range values {
			doubled[i] = v * 2
		}
		return doubled
	})

	if !r.IsOk() {
		t.Error("Map after Collect should preserve Ok")
	}

	expected := []int{2, 4, 6}
	for i, v := range r.value {
		if v != expected[i] {
			t.Errorf("value[%d] = %v, want %v", i, v, expected[i])
		}
	}
}

func TestPartition_RealWorldScenario(t *testing.T) {
	// Simulate batch user fetches
	type User struct {
		ID   string
		Name string
	}

	results := []Result[*User]{
		Ok(&User{ID: "1", Name: "Alice"}),
		Err[*User](NotFound("repo", "user")),
		Ok(&User{ID: "3", Name: "Charlie"}),
		Err[*User](Infrastructure("db", errors.New("timeout"))),
		Ok(&User{ID: "5", Name: "Eve"}),
	}

	users, errs := Partition(results)

	if len(users) != 3 {
		t.Errorf("Expected 3 successful users, got %d", len(users))
	}

	if len(errs) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(errs))
	}

	// Verify successful users
	if users[0].Name != "Alice" {
		t.Error("First user should be Alice")
	}

	if users[1].Name != "Charlie" {
		t.Error("Second user should be Charlie")
	}

	if users[2].Name != "Eve" {
		t.Error("Third user should be Eve")
	}

	// Verify error kinds
	if KindOf(errs[0]) != KindNotFound {
		t.Error("First error should be NotFound")
	}

	if KindOf(errs[1]) != KindInfrastructure {
		t.Error("Second error should be Infrastructure")
	}
}

func TestCollectWithContext_Pipeline(t *testing.T) {
	ctx := WithRequestID(context.Background(), "req_pipeline")

	results := []Result[int]{
		Ok(1),
		Ok(2),
		Ok(3),
	}

	r := CollectWithContext(ctx, results).
		Map(func(values []int) []int {
			// Sum all values
			sum := 0
			for _, v := range values {
				sum += v
			}
			return []int{sum}
		}).
		Inspect(func(values []int) {
			if values[0] != 6 {
				t.Errorf("Sum should be 6, got %d", values[0])
			}
		})

	if !r.IsOk() {
		t.Error("Pipeline should succeed")
	}
}
