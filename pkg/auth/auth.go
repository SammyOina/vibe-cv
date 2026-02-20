// Copyright (c) Ultraviolet
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

// UserContext is the context key for user information.
type contextKey string

const UserContextKey contextKey = "user"

// User represents an authenticated user.
type User struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	KratosID string `json:"kratos_id"`
}

// Config holds authentication configuration.
type Config struct {
	Enabled       bool
	PublicURL     string
	AdminURL      string
	SessionCookie string
}

// LoadConfig loads authentication configuration from environment.
func LoadConfig() *Config {
	enabled := strings.ToLower(os.Getenv("KRATOS_ENABLED")) == "true"

	return &Config{
		Enabled:       enabled,
		PublicURL:     os.Getenv("KRATOS_PUBLIC_URL"),
		AdminURL:      os.Getenv("KRATOS_ADMIN_URL"),
		SessionCookie: "ory_kratos_session",
	}
}

// Middleware returns an HTTP middleware that validates Kratos sessions (if enabled).
func Middleware(config *Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// If auth is disabled, pass through
			if !config.Enabled {
				next.ServeHTTP(w, r)

				return
			}

			// Try to extract session from cookie or Authorization header
			sessionToken, err := extractSessionToken(r, config.SessionCookie)
			if err != nil {
				// No valid session, continue as unauthenticated
				next.ServeHTTP(w, r)

				return
			}

			// Validate session with Kratos
			// Use Public URL for session validation (whoami endpoint)
			user, err := validateSessionWithKratos(config.PublicURL, sessionToken)
			if err != nil {
				// Invalid session, continue as unauthenticated
				next.ServeHTTP(w, r)

				return
			}

			// Inject user into context
			ctx := context.WithValue(r.Context(), UserContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// extractSessionToken extracts session token from cookie or Authorization header.
func extractSessionToken(r *http.Request, cookieName string) (string, error) {
	// Try cookies with prefix matching
	for _, cookie := range r.Cookies() {
		if strings.HasPrefix(cookie.Name, cookieName) && cookie.Value != "" {
			return cookie.Value, nil
		}
	}

	// Try Authorization header (Bearer token)
	authHeader := r.Header.Get("Authorization")
	if after, ok := strings.CutPrefix(authHeader, "Bearer "); ok {
		return after, nil
	}

	return "", errors.New("no session token found")
}

// validateSessionWithKratos validates a session token with Kratos public API.
func validateSessionWithKratos(kratosPublicURL, sessionToken string) (*User, error) {
	if sessionToken == "" {
		return nil, errors.New("empty session token")
	}

	// Make actual HTTP call to Kratos public API to validate session (whoami)
	client := &http.Client{Timeout: 10 * time.Second}
	requestURL := kratosPublicURL + "/sessions/whoami"

	req, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kratos request: %w", err)
	}

	// Add session token in X-Session-Token header (standard for Kratos API)
	req.Header.Set("X-Session-Token", sessionToken)

	// Also add as a cookie - this is often more reliable for Kratos when validating session cookies
	req.AddCookie(&http.Cookie{
		Name:  "ory_kratos_session",
		Value: sessionToken,
	})

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to validate session with Kratos: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("kratos returned status %d for /sessions/whoami", resp.StatusCode)
	}

	// Parse Kratos response (Session object)
	var session map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&session); err != nil {
		return nil, fmt.Errorf("failed to parse Kratos response: %w", err)
	}

	// Extract identity information from session
	identity, ok := session["identity"].(map[string]any)
	if !ok {
		return nil, errors.New("no identity found in session")
	}

	// Extract user information
	var userID, email string
	if id, ok := identity["id"].(string); ok {
		userID = id
	}

	if traits, ok := identity["traits"].(map[string]any); ok {
		if e, ok := traits["email"].(string); ok {
			email = e
		}
	}

	if userID == "" {
		return nil, errors.New("failed to extract user ID from Kratos response")
	}

	return &User{
		ID:       userID,
		Email:    email,
		KratosID: userID,
	}, nil
}

// GetUser extracts the user from context.
func GetUser(ctx context.Context) *User {
	user, ok := ctx.Value(UserContextKey).(*User)
	if !ok {
		return nil
	}

	return user
}

// RequireAuth is a middleware that requires authentication
// It wraps the optional middleware to enforce authentication when needed.
func RequireAuth(config *Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// If auth is not enabled, treat as always authenticated
			if !config.Enabled {
				next.ServeHTTP(w, r)

				return
			}

			// Check if user is in context
			user := GetUser(r.Context())
			if user == nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)

				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
