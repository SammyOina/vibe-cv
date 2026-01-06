// Copyright (c) Ultraviolet
// SPDX-License-Identifier: Apache-2.0

package sdk

import (
	"context"
	"fmt"
)

// GetAnalytics retrieves analytics data for the current user.
func (c *Client) GetAnalytics(ctx context.Context, limit int, opts ...RequestOption) (*Analytics, error) {
	if limit < 0 {
		return nil, &ValidationError{Field: "limit", Message: "limit cannot be negative"}
	}

	path := "/api/latest/analytics"
	if limit > 0 {
		path = fmt.Sprintf("%s?limit=%d", path, limit)
	}

	var analytics Analytics
	if err := c.doRequest(ctx, "GET", path, nil, &analytics, opts...); err != nil {
		return nil, fmt.Errorf("failed to get analytics: %w", err)
	}

	return &analytics, nil
}

// GetDashboard retrieves global dashboard statistics.
func (c *Client) GetDashboard(ctx context.Context, opts ...RequestOption) (*Dashboard, error) {
	var dashboard Dashboard
	if err := c.doRequest(ctx, "GET", "/api/latest/dashboard", nil, &dashboard, opts...); err != nil {
		return nil, fmt.Errorf("failed to get dashboard: %w", err)
	}

	return &dashboard, nil
}
