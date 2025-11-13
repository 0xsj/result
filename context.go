package result

import "context"

// Context keys for observability data
type contextKey string

const (
	requestIDKey contextKey = "result:request_id"
	traceIDKey   contextKey = "result:trace_id"
	spanIDKey    contextKey = "result:span_id"
	userIDKey    contextKey = "result:user_id"
	tenantIDKey  contextKey = "result:tenant_id"
)

// ============================================================================
// Context Setters
// ============================================================================

// WithRequestID adds a request ID to the context.
// This is typically called by middleware to inject a correlation ID.
//
// Example:
//
//	ctx := result.WithRequestID(r.Context(), uuid.New().String())
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// WithTraceID adds a trace ID to the context.
// This is useful for distributed tracing integration (e.g., OpenTelemetry).
//
// Example:
//
//	ctx := result.WithTraceID(ctx, span.SpanContext().TraceID().String())
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey, traceID)
}

// WithSpanID adds a span ID to the context.
// This is useful for distributed tracing integration.
//
// Example:
//
//	ctx := result.WithSpanID(ctx, span.SpanContext().SpanID().String())
func WithSpanID(ctx context.Context, spanID string) context.Context {
	return context.WithValue(ctx, spanIDKey, spanID)
}

// WithUserID adds a user ID to the context.
// This is typically called by authentication middleware.
//
// Example:
//
//	ctx := result.WithUserID(ctx, user.ID)
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// WithTenantID adds a tenant ID to the context.
// This is useful for multi-tenant applications.
//
// Example:
//
//	ctx := result.WithTenantID(ctx, tenant.ID)
func WithTenantID(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, tenantIDKey, tenantID)
}

// ============================================================================
// Context Getters
// ============================================================================

// GetRequestID extracts the request ID from the context.
// Returns empty string if not present.
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return ""
}

// GetTraceID extracts the trace ID from the context.
// Returns empty string if not present.
func GetTraceID(ctx context.Context) string {
	if id, ok := ctx.Value(traceIDKey).(string); ok {
		return id
	}
	return ""
}

// GetSpanID extracts the span ID from the context.
// Returns empty string if not present.
func GetSpanID(ctx context.Context) string {
	if id, ok := ctx.Value(spanIDKey).(string); ok {
		return id
	}
	return ""
}

// GetUserID extracts the user ID from the context.
// Returns empty string if not present.
func GetUserID(ctx context.Context) string {
	if id, ok := ctx.Value(userIDKey).(string); ok {
		return id
	}
	return ""
}

// GetTenantID extracts the tenant ID from the context.
// Returns empty string if not present.
func GetTenantID(ctx context.Context) string {
	if id, ok := ctx.Value(tenantIDKey).(string); ok {
		return id
	}
	return ""
}

// ============================================================================
// Result Integration
// ============================================================================

// WithContext automatically extracts observability data from context
// and attaches it to Result metadata. This is useful for error tracking
// in async jobs or when errors need to be persisted with correlation IDs.
//
// This method only attaches metadata if the Result is Err.
// Metadata is only added if not already present (doesn't override).
//
// Example:
//
//	return repo.FindByID(ctx, id).
//	    WithContext(ctx).
//	    AndThen(validate)
func (r Result[T]) WithContext(ctx context.Context) Result[T] {
	// Only attach context to errors
	if r.IsErr() {
		if r.meta == nil {
			r.meta = make(map[string]any)
		}

		// Only attach if not already present (don't override explicit values)
		if _, exists := r.meta["request_id"]; !exists {
			if requestID := GetRequestID(ctx); requestID != "" {
				r.meta["request_id"] = requestID
			}
		}

		if _, exists := r.meta["trace_id"]; !exists {
			if traceID := GetTraceID(ctx); traceID != "" {
				r.meta["trace_id"] = traceID
			}
		}

		if _, exists := r.meta["span_id"]; !exists {
			if spanID := GetSpanID(ctx); spanID != "" {
				r.meta["span_id"] = spanID
			}
		}

		if _, exists := r.meta["user_id"]; !exists {
			if userID := GetUserID(ctx); userID != "" {
				r.meta["user_id"] = userID
			}
		}

		if _, exists := r.meta["tenant_id"]; !exists {
			if tenantID := GetTenantID(ctx); tenantID != "" {
				r.meta["tenant_id"] = tenantID
			}
		}
	}

	return r
}

// WithContextOp is a convenience that combines WithContext and WithOp.
// This is useful for adding both operation context and correlation IDs in one call.
//
// Example:
//
//	return repo.FindByID(ctx, id).
//	    WithContextOp(ctx, "UserRepository.FindByID").
//	    AndThen(validate)
func (r Result[T]) WithContextOp(ctx context.Context, op string) Result[T] {
	return r.WithOp(op).WithContext(ctx)
}
