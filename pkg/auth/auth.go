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
			user, err := validateSessionWithKratos(config.AdminURL, sessionToken)
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
	// Try cookie first
	cookie, err := r.Cookie(cookieName)
	if err == nil && cookie.Value != "" {
		return cookie.Value, nil
	}

	// Try Authorization header (Bearer token)
	authHeader := r.Header.Get("Authorization")
	if after, ok := strings.CutPrefix(authHeader, "Bearer "); ok {
		return after, nil
	}

	return "", errors.New("no session token found")
}

// validateSessionWithKratos validates a session token with Kratos admin API.
func validateSessionWithKratos(kratosAdminURL, sessionToken string) (*User, error) {
	if sessionToken == "" {
		return nil, errors.New("empty session token")
	}

	// Make actual HTTP call to Kratos admin API to validate session
	client := &http.Client{Timeout: 10 * time.Second}
	requestURL := kratosAdminURL + "/admin/identities"

	req, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kratos request: %w", err)
	}

	// Add session token in Authorization header
	req.Header.Set("Authorization", "Bearer "+sessionToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to validate session with Kratos: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("kratos returned status %d", resp.StatusCode)
	}

	// Parse Kratos response
	var kratosIdentity map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&kratosIdentity); err != nil {
		return nil, fmt.Errorf("failed to parse Kratos response: %w", err)
	}

	// Extract user information from response
	var userID, email string
	if id, ok := kratosIdentity["id"].(string); ok {
		userID = id
	}

	if traits, ok := kratosIdentity["traits"].(map[string]any); ok {
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
