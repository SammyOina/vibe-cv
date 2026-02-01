// Copyright (c) Ultraviolet
// SPDX-License-Identifier: Apache-2.0

package sdk

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	client := NewClient("http://localhost:8080")
	if client == nil {
		t.Fatal("expected client to be created")
	}
	if client.baseURL != "http://localhost:8080" {
		t.Errorf("expected baseURL to be http://localhost:8080, got %s", client.baseURL)
	}
	if client.userAgent == "" {
		t.Error("expected userAgent to be set")
	}
}

func TestClientOptions(t *testing.T) {
	timeout := 60 * time.Second
	userAgent := "test-agent/1.0"

	client := NewClient(
		"http://localhost:8080",
		WithTimeout(timeout),
		WithUserAgent(userAgent),
	)

	if client.httpClient.Timeout != timeout {
		t.Errorf("expected timeout %v, got %v", timeout, client.httpClient.Timeout)
	}
	if client.userAgent != userAgent {
		t.Errorf("expected user agent %s, got %s", userAgent, client.userAgent)
	}
}

func TestCustomizeCV_Validation(t *testing.T) {
	client := NewClient("http://localhost:8080")

	tests := []struct {
		name    string
		req     *CustomizeCVRequest
		wantErr bool
	}{
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
		},
		{
			name: "empty CV",
			req: &CustomizeCVRequest{
				JobDescription: "test",
			},
			wantErr: true,
		},
		{
			name: "empty job description",
			req: &CustomizeCVRequest{
				CV: "test",
			},
			wantErr: true,
		},
		{
			name: "valid request",
			req: &CustomizeCVRequest{
				CV:             "test cv",
				JobDescription: "test job",
			},
			wantErr: false, // Will fail on network, but validation passes
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := client.CustomizeCV(context.Background(), tt.req)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if !IsValidationError(err) {
					t.Errorf("expected validation error, got %T", err)
				}
			}
		})
	}
}

func TestAPIError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		message    string
		checkFunc  func(*APIError) bool
	}{
		{
			name:       "not found",
			statusCode: 404,
			message:    "not found",
			checkFunc:  func(e *APIError) bool { return e.IsNotFound() },
		},
		{
			name:       "bad request",
			statusCode: 400,
			message:    "bad request",
			checkFunc:  func(e *APIError) bool { return e.IsBadRequest() },
		},
		{
			name:       "unauthorized",
			statusCode: 401,
			message:    "unauthorized",
			checkFunc:  func(e *APIError) bool { return e.IsUnauthorized() },
		},
		{
			name:       "server error",
			statusCode: 500,
			message:    "internal server error",
			checkFunc:  func(e *APIError) bool { return e.IsServerError() },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &APIError{
				StatusCode: tt.statusCode,
				Message:    tt.message,
			}

			if !tt.checkFunc(err) {
				t.Errorf("check function failed for status code %d", tt.statusCode)
			}

			if !IsAPIError(err) {
				t.Error("IsAPIError should return true")
			}
		})
	}
}

func TestHealthCheck(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/latest/health" {
			t.Errorf("expected path /api/latest/health, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok","version":"test","database":"connected","timestamp":"2024-01-01T00:00:00Z"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	health, err := client.Health(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if health.Status != "ok" {
		t.Errorf("expected status ok, got %s", health.Status)
	}

	// Test Ping
	if err := client.Ping(context.Background()); err != nil {
		t.Errorf("expected ping to succeed, got %v", err)
	}
}

func TestBatchValidation(t *testing.T) {
	client := NewClient("http://localhost:8080")

	// Test empty items
	_, err := client.BatchCustomize(context.Background(), []BatchItem{})
	if err == nil {
		t.Error("expected error for empty items")
	}
	if !IsValidationError(err) {
		t.Errorf("expected validation error, got %T", err)
	}
}

func TestVersionValidation(t *testing.T) {
	client := NewClient("http://localhost:8080")

	// Test invalid CV ID
	_, err := client.GetVersions(context.Background(), 0)
	if err == nil {
		t.Error("expected error for invalid CV ID")
	}
	if !IsValidationError(err) {
		t.Errorf("expected validation error, got %T", err)
	}

	// Test invalid version ID
	_, err = client.GetVersionDetail(context.Background(), -1)
	if err == nil {
		t.Error("expected error for invalid version ID")
	}
	if !IsValidationError(err) {
		t.Errorf("expected validation error, got %T", err)
	}

	// Test invalid comparison
	_, err = client.CompareVersions(context.Background(), 0, 1)
	if err == nil {
		t.Error("expected error for invalid version ID")
	}
	if !IsValidationError(err) {
		t.Errorf("expected validation error, got %T", err)
	}
}
