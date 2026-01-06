// Copyright (c) Ultraviolet
// SPDX-License-Identifier: Apache-2.0

package sdk

import (
	"context"
	"fmt"
)

// CustomizeCV customizes a CV for a specific job description.
func (c *Client) CustomizeCV(ctx context.Context, req *CustomizeCVRequest) (*CustomizeCVResponse, error) {
	if req == nil {
		return nil, &ValidationError{Field: "request", Message: "request cannot be nil"}
	}
	if req.CV == "" {
		return nil, &ValidationError{Field: "cv", Message: "CV content is required"}
	}
	if req.JobDescription == "" {
		return nil, &ValidationError{Field: "job_description", Message: "job description is required"}
	}

	var resp CustomizeCVResponse
	if err := c.doRequest(ctx, "POST", "/api/latest/customize-cv", req, &resp); err != nil {
		return nil, fmt.Errorf("failed to customize CV: %w", err)
	}

	return &resp, nil
}
