# result

**Functional error handling for Go with Result types and pattern matching.**

Inspired by Rust's `Result<T, E>`, this package brings railway-oriented programming and composable error handling to Go, making complex error flows elegant and type-safe.

---

## Why Result?

Traditional Go error handling can become verbose and error-prone with complex workflows:
```go
// Traditional Go - lots of if err != nil
user, err := repo.FindByID(ctx, id)
if err != nil {
    return nil, err
}

validated, err := validate(user)
if err != nil {
    return nil, err
}

enriched, err := enrich(validated)
if err != nil {
    return nil, err
}

saved, err := save(enriched)
if err != nil {
    return nil, err
}

return saved, nil
```

With `result`, error handling becomes composable and expressive:
```go
// With result - railway-oriented programming
return repo.FindByID(ctx, id).
    AndThen(validate).
    AndThen(enrich).
    AndThen(save)
```

---

## Features

- ✅ **Type-safe error handling** - `Result[T]` makes success types explicit
- ✅ **Composable operations** - Chain operations with `AndThen`, `Map`, `OrElse`
- ✅ **Pattern matching** - Handle errors elegantly with `Match`, `MatchKind`
- ✅ **Graceful recovery** - Built-in fallback mechanisms with `Recover`, `RecoverWith`
- ✅ **Zero reflection** - Pure compile-time generics, no runtime overhead
- ✅ **Observability-ready** - Context tracking with `WithOp`, `WithMeta`
- ✅ **Go-compatible** - Easy interop with stdlib via `.Value()` → `(T, error)`

---

## Installation
```bash
go get github.com/0xsj/result
```

**Requirements:** Go 1.18+ (generics support)

---

## Quick Start

### Basic Usage
```go
import "github.com/0xsj/result"

// Create results
success := result.Ok(42)
failure := result.Err[int](errors.New("oops"))

// Check status
if success.IsOk() {
    fmt.Println("Got:", success.Unwrap())
}

// Extract values
value, err := success.Value()  // Compatible with Go idioms
```

### HTTP Handler Example
```go
func GetUser(w http.ResponseWriter, r *http.Request) {
    userID := chi.URLParam(r, "id")
    
    userService.GetUser(r.Context(), userID).Match(
        func(user *User) {
            json.NewEncoder(w).Encode(user)
        },
        func(err error) {
            http.Error(w, err.Error(), result.HTTPStatus(err))
        },
    )
}
```

### Railway-Oriented Pipeline
```go
func ProcessOrder(ctx context.Context, orderID string) result.Result[*Receipt] {
    return findOrder(ctx, orderID).
        AndThen(validateInventory).
        AndThen(chargePayment).
        AndThen(shipOrder).
        AndThen(generateReceipt).
        Inspect(func(receipt *Receipt) {
            log.Info("order processed", "id", receipt.ID)
        }).
        InspectErr(func(err error) {
            log.Error("order failed", "error", err)
        })
}
```

### Graceful Fallbacks
```go
// Try cache, fallback to DB, fallback to default
func GetUserWithFallbacks(id string) result.Result[*User] {
    return cache.Get(id).
        RecoverWith(result.KindNotFound, func(err error) result.Result[*User] {
            return db.FindByID(ctx, id)
        }).
        Recover(result.KindNotFound, func(err error) *User {
            return NewGuestUser()
        })
}
```

### Pattern Matching on Error Types
```go
userResult.MatchKind(
    map[result.Kind]func(error){
        result.KindNotFound: func(e error) {
            http.Error(w, "User not found", http.StatusNotFound)
        },
        result.KindValidation: func(e error) {
            http.Error(w, "Invalid input", http.StatusBadRequest)
        },
        result.KindUnauthorized: func(e error) {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
        },
    },
    func(e error) {
        http.Error(w, "Internal error", http.StatusInternalServerError)
    },
)
```

---

## Core Concepts

### Result[T]

A `Result[T]` represents either:
- **Ok** - Contains a value of type `T`
- **Err** - Contains an error

This makes error handling explicit at the type level.
```go
func FindUser(id string) result.Result[*User] {
    // Returns Result[*User], not (*User, error)
}
```

### Error Kinds

Structured errors with categories for better error handling:
```go
result.NotFound("Repository.FindUser", "user")        // 404
result.Validation("Service.CreateUser", "invalid email", meta)  // 400
result.Conflict("Service.CreateUser", "email")        // 409
result.Unauthorized("Auth.Login", "invalid token")    // 401
result.Forbidden("Auth.Access", "insufficient permissions")     // 403
result.Domain("User.ChangeEmail", "email unchanged")  // 422
result.Infrastructure("DB.Query", dbErr)              // 503
result.Internal("Service.Process", unexpectedErr)     // 500
```

### Railway-Oriented Programming

Operations short-circuit on first error, like a train switching to an error track:
```go
// If any step fails, remaining steps are skipped
result.Ok(input).
    AndThen(step1).  // ✓ succeeds, continues
    AndThen(step2).  // ✗ fails, switches to error track
    AndThen(step3).  // skipped
    AndThen(step4)   // skipped
// Returns the error from step2
```

---

## API Overview

### Constructors
- `Ok[T](value T)` - Create successful result
- `Err[T](err error)` - Create failed result
- `From[T](value T, err error)` - Convert `(T, error)` to `Result[T]`

### Predicates
- `IsOk()` - Check if result is successful
- `IsErr()` - Check if result is an error
- `IsErrKind(kind Kind)` - Check specific error type

### Extractors
- `Unwrap()` - Get value or panic (use in tests)
- `UnwrapOr(default T)` - Get value or default
- `Value()` - Get `(T, error)` tuple (Go idiom)
- `Expect(msg string)` - Unwrap with custom panic message

### Transformers
- `Map(f func(T) T)` - Transform success value
- `MapErr(f func(error) error)` - Transform error
- `MapValue[U](f func(T) U)` - Transform to different type

### Combinators
- `AndThen(f func(T) Result[T])` - Chain operations (monadic bind)
- `AndThenMap[U](f func(T) Result[U])` - Chain with type change
- `OrElse(f func(error) Result[T])` - Provide alternative on error

### Pattern Matching
- `Match(onOk, onErr)` - Execute side effects based on result
- `MatchValue[U](onOk, onErr)` - Transform result to single value
- `MatchKind(cases, default)` - Match on error kinds
- `MatchKindValue[U](onOk, cases, default)` - Match with value return

### Recovery
- `Recover(kind, f)` - Replace specific error with value
- `RecoverWith(kind, f)` - Replace specific error with Result
- `RecoverAll(f)` - Recover from any error
- `RecoverAllWith(f)` - Recover from any error with Result

### Inspection (Observability)
- `Inspect(f)` - Side effect on success (logging, metrics)
- `InspectErr(f)` - Side effect on error
- `Tap(onOk, onErr)` - Side effects on both

### Context
- `WithOp(op string)` - Add operation name for tracing
- `WithMeta(key, value)` - Add metadata for observability

---

## Use Cases

### ✅ Perfect For:
- **DDD/CQRS** architectures with complex command chains
- **Microservices** with multiple external dependencies
- **Data pipelines** with multi-step transformations
- **Background jobs** with retry/fallback logic
- **CLI tools** with chained operations
- **API clients** with automatic retries
- **Any workflow** with 3+ sequential operations

### ⚠️ Not Ideal For:
- Simple CRUD with no business logic
- Performance-critical hot paths (millions of ops/sec)
- Pure stdlib wrappers with no domain logic
- Public libraries (use idiomatic Go for wider adoption)

---

## Design Philosophy

1. **Type Safety** - Compiler catches unhandled errors
2. **Composability** - Operations chain naturally
3. **Explicit over Implicit** - Clear success/error paths
4. **Zero Magic** - Pure compile-time generics, no reflection
5. **Go Compatible** - Works alongside traditional Go error handling

---

## Performance

- **No reflection** - All generics are compile-time
- **Minimal overhead** - ~32 bytes per Result (vs 16 for `(T, error)`)
- **Zero CPU cost** - Same branching as traditional Go
- **Stack allocated** - Most Results don't escape to heap

---

## Comparison with Other Patterns

### vs Traditional Go
```go
// Traditional: Lots of if err != nil
user, err := getUser(id)
if err != nil { return nil, err }
validated, err := validate(user)
if err != nil { return nil, err }

// result: Railway-oriented
return getUser(id).AndThen(validate)
```

### vs samber/mo
- `result` is focused purely on error handling
- `mo` provides many monad types (Option, Either, Task, etc.)
- `result` has DDD-specific error kinds
- `result` has built-in observability (WithOp, WithMeta)

### vs Rust's Result<T, E>
- Rust: `Result<User, DatabaseError>`
- Go: `Result[*User]` (error is always `error` interface)
- Otherwise very similar API

---

## Examples

See the [`examples/`](./examples) directory for complete runnable examples:
- [`01-basic/`](./examples/01-basic) - Basic usage patterns
- [`02-http-handler/`](./examples/02-http-handler) - HTTP handlers
- [`03-railway/`](./examples/03-railway) - Railway-oriented programming
- [`04-pattern-matching/`](./examples/04-pattern-matching) - Error matching
- [`05-ddd-example/`](./examples/05-ddd-example) - DDD/CQRS example
- [`06-comparison/`](./examples/06-comparison) - Traditional vs Result

---

## Contributing

Contributions welcome! Please open an issue or PR.

---

## License

MIT License - see [LICENSE](./LICENSE) for details

---

## Inspiration

This package is inspired by:
- **Rust** - `Result<T, E>` and error handling
- **Scala** - `Either[L, R]` and functional composition
- **Railway-Oriented Programming** - Scott Wlaschin's pattern
- **Go proverbs** - Simplicity, clarity, composition

---

## Resources

- [Railway-Oriented Programming](https://fsharpforfunandprofit.com/rop/) by Scott Wlaschin
- [Rust's Result Type](https://doc.rust-lang.org/std/result/)
- [Go Error Handling](https://go.dev/blog/error-handling-and-go)

---

**Made with ❤️ for Gophers who want better error handling**