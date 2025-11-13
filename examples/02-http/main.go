package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/0xsj/result"
	"github.com/0xsj/result/examples/02-http/handler"
	"github.com/0xsj/result/examples/02-http/repository"
	"github.com/0xsj/result/examples/02-http/service"
	"github.com/google/uuid"
)

// Middleware: Request ID
func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		ctx := result.WithRequestID(r.Context(), requestID)
		w.Header().Set("X-Request-ID", requestID)

		fmt.Printf("[Middleware] Request ID: %s\n", requestID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Middleware: Logging
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		fmt.Printf("[Middleware] --> %s %s\n", r.Method, r.URL.Path)

		next.ServeHTTP(w, r)

		fmt.Printf("[Middleware] <-- %s %s (took %v)\n",
			r.Method, r.URL.Path, time.Since(start))
	})
}

// Middleware: CORS
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Request-ID")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	fmt.Println("╔════════════════════════════════════════╗")
	fmt.Println("║  HTTP API Example with Result Package ║")
	fmt.Println("╚════════════════════════════════════════╝")
	fmt.Println()

	// Initialize layers
	repo := repository.NewUserRepository()
	svc := service.NewUserService(repo)
	userHandler := handler.NewUserHandler(svc)

	// Setup routes
	mux := http.NewServeMux()
	mux.Handle("/users", userHandler)
	mux.Handle("/users/", userHandler)

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Apply middleware
	var handler http.Handler = mux
	handler = corsMiddleware(handler)
	handler = requestIDMiddleware(handler)
	handler = loggingMiddleware(handler)

	// Start server
	addr := ":8080"
	fmt.Printf("🚀 Server starting on http://localhost%s\n\n", addr)
	fmt.Println("Try these commands:")
	fmt.Println("  # Health check")
	fmt.Println("  curl http://localhost:8080/health")
	fmt.Println()
	fmt.Println("  # Create user")
	fmt.Println(`  curl -X POST http://localhost:8080/users -d '{"name":"Alice","email":"alice@example.com"}'`)
	fmt.Println()
	fmt.Println("  # List users")
	fmt.Println("  curl http://localhost:8080/users")
	fmt.Println()

	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatal(err)
	}
}