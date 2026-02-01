// Copyright (c) Ultraviolet
// SPDX-License-Identifier: Apache-2.0

package sdk

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// APIError represents an error returned by the API.
type APIError struct {
	StatusCode int
	Message    string
	Details    map[string]interface{}
}

// Error implements the error interface.
func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("API error (status %d): %s", e.StatusCode, e.Message)
	}

	return fmt.Sprintf("API error (status %d)", e.StatusCode)
}

// IsNotFound returns true if the error is a 404 Not Found error.
func (e *APIError) IsNotFound() bool {
	return e.StatusCode == http.StatusNotFound
}

// IsBadRequest returns true if the error is a 400 Bad Request error.
func (e *APIError) IsBadRequest() bool {
	return e.StatusCode == http.StatusBadRequest
}

// IsUnauthorized returns true if the error is a 401 Unauthorized error.
func (e *APIError) IsUnauthorized() bool {
	return e.StatusCode == http.StatusUnauthorized
}

// IsServerError returns true if the error is a 5xx server error.
func (e *APIError) IsServerError() bool {
	return e.StatusCode >= 500 && e.StatusCode < 600
}

// ValidationError represents a client-side validation error.
type ValidationError struct {
	Field   string
	Message string
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error on field '%s': %s", e.Field, e.Message)
}

// parseAPIError attempts to parse an error response from the API.
func parseAPIError(resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &APIError{
			StatusCode: resp.StatusCode,
			Message:    "failed to read error response",
		}
	}

	// Try to parse as JSON error response
	var errorResp struct {
		Error   string                 `json:"error"`
		Message string                 `json:"message"`
		Details map[string]interface{} `json:"details"`
	}

	if err := json.Unmarshal(body, &errorResp); err == nil {
		message := errorResp.Error
		if message == "" {
			message = errorResp.Message
		}
		if message == "" {
			message = string(body)
		}

		return &APIError{
			StatusCode: resp.StatusCode,
			Message:    message,
			Details:    errorResp.Details,
		}
	}

	// Fallback to plain text error
	return &APIError{
		StatusCode: resp.StatusCode,
		Message:    string(body),
	}
}

// IsAPIError returns true if the error is an APIError.
func IsAPIError(err error) bool {
	_, ok := err.(*APIError)

	return ok
}

// IsValidationError returns true if the error is a ValidationError.
func IsValidationError(err error) bool {
	_, ok := err.(*ValidationError)

	return ok
}
