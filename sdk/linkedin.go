// Copyright (c) Ultraviolet
// SPDX-License-Identifier: Apache-2.0

package sdk

import (
	"context"
	"fmt"
)

// LinkedInImportResponse represents a LinkedIn import.
type LinkedInImportResponse struct {
	ID           int                    `json:"id"`
	LinkedInURL  string                 `json:"linkedin_url"`
	ImportStatus string                 `json:"import_status"`
	ExtractedCV  string                 `json:"extracted_cv,omitempty"`
	ProfileData  map[string]interface{} `json:"profile_data,omitempty"`
	ErrorMessage string                 `json:"error_message,omitempty"`
	CreatedAt    string                 `json:"created_at,omitempty"`
	UpdatedAt    string                 `json:"updated_at,omitempty"`
}

// ImportLinkedInRequest represents the request for LinkedIn import.
type ImportLinkedInRequest struct {
	LinkedInURL string `json:"linkedin_url"`
	ProfileText string `json:"profile_text"`
}

// ImportLinkedInText imports a LinkedIn profile from text.
func (c *Client) ImportLinkedInText(ctx context.Context, url string, profileText string, opts ...RequestOption) (*LinkedInImportResponse, error) {
	req := ImportLinkedInRequest{
		LinkedInURL: url,
		ProfileText: profileText,
	}

	var result LinkedInImportResponse
	err := c.doRequest(ctx, "POST", "/api/latest/linkedin/import", req, &result, opts...)
	if err != nil {
		return nil, fmt.Errorf("LinkedIn import failed: %w", err)
	}

	return &result, nil
}

// GetLinkedInImports retrieves all LinkedIn imports for the authenticated user.
func (c *Client) GetLinkedInImports(ctx context.Context, opts ...RequestOption) ([]LinkedInImportResponse, error) {
	var result []LinkedInImportResponse
	err := c.doRequest(ctx, "GET", "/api/latest/linkedin/imports", nil, &result, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to get LinkedIn imports: %w", err)
	}

	return result, nil
}

// GetLinkedInImport retrieves a specific LinkedIn import by ID.
func (c *Client) GetLinkedInImport(ctx context.Context, importID int, opts ...RequestOption) (*LinkedInImportResponse, error) {
	path := fmt.Sprintf("/api/latest/linkedin/%d", importID)

	var result LinkedInImportResponse
	err := c.doRequest(ctx, "GET", path, nil, &result, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to get LinkedIn import: %w", err)
	}

	return &result, nil
}
