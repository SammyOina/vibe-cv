// Copyright (c) Ultraviolet
// SPDX-License-Identifier: Apache-2.0

package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Client is the vibe-cv SDK client.
type Client struct {
	baseURL    string
	httpClient *http.Client
	userAgent  string
}

// ClientOption is a functional option for configuring the Client.
type ClientOption func(*Client)

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(client *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = client
	}
}

// WithTimeout sets the HTTP client timeout.
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.httpClient.Timeout = timeout
	}
}

// RequestOption is a functional option for configuring individual requests.
type RequestOption func(*requestConfig)

// requestConfig holds per-request configuration.
type requestConfig struct {
	authToken string
}

// WithRequestAuthToken sets the authentication token for a single request.
// This is required for authenticated endpoints in multi-user SaaS scenarios.
func WithRequestAuthToken(token string) RequestOption {
	return func(rc *requestConfig) {
		rc.authToken = token
	}
}

// buildRequestConfig creates a requestConfig from the provided options.
func buildRequestConfig(opts ...RequestOption) *requestConfig {
	config := &requestConfig{}
	for _, opt := range opts {
		opt(config)
	}
	return config
}

// WithUserAgent sets a custom User-Agent header.
func WithUserAgent(userAgent string) ClientOption {
	return func(c *Client) {
		c.userAgent = userAgent
	}
}

// NewClient creates a new vibe-cv SDK client.
func NewClient(baseURL string, options ...ClientOption) *Client {
	client := &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		userAgent: "vibe-cv-sdk-go/1.0.0",
	}

	for _, option := range options {
		option(client)
	}

	return client
}

// doRequest performs an HTTP request with proper error handling.
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}, result interface{}, opts ...RequestOption) error {
	reqConfig := buildRequestConfig(opts...)
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonData)
	}

	fullURL, err := url.JoinPath(c.baseURL, path)
	if err != nil {
		return fmt.Errorf("failed to build URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	if reqConfig.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+reqConfig.authToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check for error status codes
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return parseAPIError(resp)
	}

	// Parse response if result is provided
	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

// doRequestRaw performs an HTTP request and returns the raw response body.
func (c *Client) doRequestRaw(ctx context.Context, method, path string, body interface{}, opts ...RequestOption) ([]byte, error) {
	reqConfig := buildRequestConfig(opts...)
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonData)
	}

	fullURL, err := url.JoinPath(c.baseURL, path)
	if err != nil {
		return nil, fmt.Errorf("failed to build URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("User-Agent", c.userAgent)
	if reqConfig.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+reqConfig.authToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check for error status codes
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, parseAPIError(resp)
	}

	// Read raw response body
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return data, nil
}
