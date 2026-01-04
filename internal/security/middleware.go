// Copyright (c) Ultraviolet
// SPDX-License-Identifier: Apache-2.0

package security

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sammyoina/vibe-cv/internal/auth"
)

// RequestValidator validates HTTP request structure and content.
type RequestValidator struct {
	maxBodySize int64
}

// NewRequestValidator creates a new request validator.
func NewRequestValidator(maxBodySize int64) *RequestValidator {
	return &RequestValidator{
		maxBodySize: maxBodySize,
	}
}

// ValidateContentType validates the Content-Type header.
func (rv *RequestValidator) ValidateContentType(contentType string, expected string) error {
	if contentType == "" && expected != "" {
		return errors.New("missing Content-Type header")
	}
	// Parse and compare base content type (ignore charset etc)
	actual := strings.Split(contentType, ";")[0]
	if actual != expected && expected != "" {
		return fmt.Errorf("invalid Content-Type: expected %s, got %s", expected, actual)
	}

	return nil
}

// ValidateJSONBody validates and unmarshals JSON request body.
func (rv *RequestValidator) ValidateJSONBody(r *http.Request, v any) error {
	// Check body size
	if r.ContentLength > rv.maxBodySize {
		return fmt.Errorf("request body too large: %d bytes (max %d)", r.ContentLength, rv.maxBodySize)
	}

	// Check Content-Type
	if err := rv.ValidateContentType(r.Header.Get("Content-Type"), "application/json"); err != nil {
		return err
	}

	// Decode JSON
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields() // Prevent accepting unknown fields

	if err := decoder.Decode(v); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	return nil
}

// RateLimiter implements token bucket rate limiting.
type RateLimiter struct {
	mu      sync.RWMutex
	buckets map[string]*TokenBucket
	rps     float64 // requests per second
	burst   int     // burst capacity
	cleanup time.Duration
	lastGC  time.Time
}

// TokenBucket represents a rate limiting bucket for a client.
type TokenBucket struct {
	tokens     float64
	lastRefill time.Time
	ip         string
}

// NewRateLimiter creates a new rate limiter.
func NewRateLimiter(requestsPerSecond float64, burst int) *RateLimiter {
	return &RateLimiter{
		buckets: make(map[string]*TokenBucket),
		rps:     requestsPerSecond,
		burst:   burst,
		cleanup: 1 * time.Hour,
		lastGC:  time.Now(),
	}
}

// Allow checks if a request from the given IP is allowed.
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Periodic cleanup of old buckets
	if time.Since(rl.lastGC) > rl.cleanup {
		rl.cleanupOldBuckets()
		rl.lastGC = time.Now()
	}

	bucket, exists := rl.buckets[ip]
	if !exists {
		bucket = &TokenBucket{
			tokens:     float64(rl.burst),
			lastRefill: time.Now(),
			ip:         ip,
		}
		rl.buckets[ip] = bucket
	}

	// Refill tokens
	now := time.Now()
	elapsed := now.Sub(bucket.lastRefill).Seconds()
	bucket.tokens = minFloat(float64(rl.burst), bucket.tokens+elapsed*rl.rps)
	bucket.lastRefill = now

	// Check if token available
	if bucket.tokens >= 1.0 {
		bucket.tokens--

		return true
	}

	return false
}

// cleanupOldBuckets removes buckets that haven't been used recently.
func (rl *RateLimiter) cleanupOldBuckets() {
	cutoff := time.Now().Add(-rl.cleanup)
	for ip, bucket := range rl.buckets {
		if bucket.lastRefill.Before(cutoff) {
			delete(rl.buckets, ip)
		}
	}
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}

	return b
}

// CORSConfig holds CORS configuration.
type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

// DefaultCORSConfig returns sensible CORS defaults.
func DefaultCORSConfig() *CORSConfig {
	return &CORSConfig{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: false,
		MaxAge:           3600,
	}
}

// IsOriginAllowed checks if an origin is allowed.
func (cc *CORSConfig) IsOriginAllowed(origin string) bool {
	for _, allowed := range cc.AllowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
	}

	return false
}

// AuditLogger logs API calls for security and compliance.
type AuditLogger struct {
	mu   sync.Mutex
	logs []AuditLogEntry
}

// AuditLogEntry represents a single audit log entry.
type AuditLogEntry struct {
	Timestamp  time.Time
	Method     string
	Path       string
	StatusCode int
	IP         string
	UserID     string
	Error      string
}

// NewAuditLogger creates a new audit logger.
func NewAuditLogger() *AuditLogger {
	return &AuditLogger{
		logs: make([]AuditLogEntry, 0, 10000),
	}
}

// LogRequest logs an API request.
func (al *AuditLogger) LogRequest(method string, path string, statusCode int, ip string, ctx any, err error) {
	al.mu.Lock()
	defer al.mu.Unlock()

	entry := AuditLogEntry{
		Timestamp:  time.Now().UTC(),
		Method:     method,
		Path:       path,
		StatusCode: statusCode,
		IP:         ip,
	}

	// Extract user from context if available
	if user, ok := ctx.(*auth.User); ok && user != nil {
		entry.UserID = user.ID
	}

	// Log error if present
	if err != nil {
		entry.Error = err.Error()
	}

	// Keep a rolling buffer of recent logs (10k entries)
	if len(al.logs) >= 10000 {
		al.logs = al.logs[1:]
	}

	al.logs = append(al.logs, entry)
}

// GetRecentLogs returns recent audit logs.
func (al *AuditLogger) GetRecentLogs(count int) []AuditLogEntry {
	al.mu.Lock()
	defer al.mu.Unlock()

	if count > len(al.logs) {
		count = len(al.logs)
	}

	if count == 0 {
		return []AuditLogEntry{}
	}

	// Return the most recent entries
	return al.logs[len(al.logs)-count:]
}

// Middleware functions

// ValidationMiddleware validates request structure.
func ValidationMiddleware(validator *RequestValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// For POST/PUT/PATCH, validate Content-Type
			if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
				if r.Body != nil && r.ContentLength > 0 {
					if err := validator.ValidateContentType(r.Header.Get("Content-Type"), "application/json"); err != nil {
						http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusBadRequest)

						return
					}
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RateLimitMiddleware enforces rate limiting.
func RateLimitMiddleware(limiter *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract client IP
			ip := r.Header.Get("X-Forwarded-For")
			if ip == "" {
				ip = strings.Split(r.RemoteAddr, ":")[0]
			}

			// Check rate limit
			if !limiter.Allow(ip) {
				w.Header().Set("Retry-After", "1")
				http.Error(w, `{"error": "rate limit exceeded"}`, http.StatusTooManyRequests)

				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// CORSMiddleware handles CORS headers.
func CORSMiddleware(config *CORSConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Check if origin is allowed
			if origin != "" && config.IsOriginAllowed(origin) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", strings.Join(config.AllowedMethods, ", "))
				w.Header().Set("Access-Control-Allow-Headers", strings.Join(config.AllowedHeaders, ", "))

				if config.AllowCredentials {
					w.Header().Set("Access-Control-Allow-Credentials", "true")
				}

				w.Header().Set("Access-Control-Max-Age", strconv.Itoa(config.MaxAge))
			}

			// Handle preflight requests
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)

				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// AuditMiddleware logs all requests.
func AuditMiddleware(logger *AuditLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Wrap response writer to capture status code
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// Extract IP
			ip := r.Header.Get("X-Forwarded-For")
			if ip == "" {
				ip = strings.Split(r.RemoteAddr, ":")[0]
			}

			// Call next handler
			next.ServeHTTP(wrapped, r)

			// Log after request is complete
			logger.LogRequest(r.Method, r.RequestURI, wrapped.statusCode, ip, auth.GetUser(r.Context()), nil)
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code.
type responseWriter struct {
	http.ResponseWriter

	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	return rw.ResponseWriter.Write(b)
}
