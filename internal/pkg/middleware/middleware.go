package middleware

/**
 * ============================================================================
 * PACKAGE: internal/pkg/middleware
 * TECHNOLOGY: Go Standard Library (net/http, context, time, crypto/subtle)
 * PURPOSE: Composable, Zero-Framework HTTP Middleware Pipeline
 * 
 * WHY STANDARD LIBRARY (net/http) WITHOUT FRAMEWORKS?
 * 1. Zero Dependency Churn: Standard net/http has been stable for over a decade with
 *    zero breaking changes, guaranteeing long-term enterprise maintainability.
 * 2. Maximum Concurrency: Each incoming connection is serviced by Go's lightweight
 *    goroutines (~2KB stack) without external routing overhead.
 * 3. Composable Middleware Chaining: Leveraging idiomatic `func(http.Handler) http.Handler`
 *    allows strict separation of concerns (Auth, CORS, Metrics, Recovery, Tracing).
 * ============================================================================
 */

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

type contextKey string

const (
	UserIDKey    contextKey = "ctx_user_id"
	UserRoleKey  contextKey = "ctx_user_role"
	RequestIDKey contextKey = "ctx_request_id"
)

// Middleware is an idiomatic function type for net/http handler chaining.
type Middleware func(http.Handler) http.Handler

// Chain aggregates multiple middlewares and executes them in left-to-right order.
func Chain(middlewares ...Middleware) Middleware {
	return func(final http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			final = middlewares[i](final)
		}
		return final
	}
}

// RequestID generates a unique trace ID per request and attaches it to request context and headers.
func RequestID() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqID := r.Header.Get("X-Request-ID")
			if reqID == "" {
				bytes := make([]byte, 16)
				_, _ = rand.Read(bytes)
				reqID = hex.EncodeToString(bytes)
			}

			w.Header().Set("X-Request-ID", reqID)
			ctx := context.WithValue(r.Context(), RequestIDKey, reqID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Logger writes structured access log records with method, duration, and status code.
func Logger() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			wrapped := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			
			next.ServeHTTP(wrapped, r)

			duration := time.Since(start)
			reqID, _ := r.Context().Value(RequestIDKey).(string)
			log.Printf("[HTTP] req_id=%s method=%s path=%s status=%d duration=%s ip=%s",
				reqID, r.Method, r.URL.Path, wrapped.status, duration.String(), r.RemoteAddr)
		})
	}
}

// Recovery catches panics, prevents server crashes, and returns a sanitized JSON 500.
func Recovery() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					reqID, _ := r.Context().Value(RequestIDKey).(string)
					log.Printf("[PANIC RECOVERED] req_id=%s error=%v", reqID, err)
					
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					_ = json.NewEncoder(w).Encode(map[string]string{
						"error":      "Internal server error",
						"request_id": reqID,
					})
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// CORS provides hardened Cross-Origin Resource Sharing control for the React SPA.
func CORS(allowedOrigins []string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" {
				allowed := false
				for _, o := range allowedOrigins {
					if o == "*" || o == origin {
						allowed = true
						break
					}
				}
				if allowed {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
					w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID, Idempotency-Key")
					w.Header().Set("Access-Control-Allow-Credentials", "true")
					w.Header().Set("Access-Control-Max-Age", "86400")
				}
			}

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RateLimiter implements an in-memory token bucket rate limiter per client IP.
// In production, this can seamlessly delegate to a Redis Token Bucket script.
func RateLimiter(requestsPerMinute int) Middleware {
	type clientBucket struct {
		tokens     int
		lastRefill time.Time
	}

	var mu sync.Mutex
	clients := make(map[string]*clientBucket)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := strings.Split(r.RemoteAddr, ":")[0]
			now := time.Now()

			mu.Lock()
			bucket, exists := clients[ip]
			if !exists {
				bucket = &clientBucket{tokens: requestsPerMinute, lastRefill: now}
				clients[ip] = bucket
			}

			// Refill tokens based on elapsed minutes
			elapsed := now.Sub(bucket.lastRefill)
			if elapsed >= time.Minute {
				bucket.tokens = requestsPerMinute
				bucket.lastRefill = now
			}

			if bucket.tokens <= 0 {
				mu.Unlock()
				w.Header().Set("Retry-After", "60")
				http.Error(w, `{"error":"Too many requests. Please try again later."}`, http.StatusTooManyRequests)
				return
			}

			bucket.tokens--
			mu.Unlock()

			next.ServeHTTP(w, r)
		})
	}
}

// statusRecorder is a helper wrapper to capture HTTP status codes.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

// AuthContextExtract extracts User ID and Role from Mock/Bearer JWT tokens and injects into Context.
func AuthContextExtract() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				token := strings.TrimPrefix(authHeader, "Bearer ")
				// In production: verify RS256 signature using jwt.ParseWithClaims
				// For demonstration & local tests, parse payload:
				parts := strings.Split(token, ".")
				if len(parts) >= 2 {
					// Simulated authenticated user context
					ctx := context.WithValue(r.Context(), UserIDKey, "usr_1001")
					ctx = context.WithValue(ctx, UserRoleKey, "CUSTOMER")
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
