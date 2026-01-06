// Copyright (c) Ultraviolet
// SPDX-License-Identifier: Apache-2.0

package sdk

import (
	"context"
	"fmt"
)

// Health performs a health check on the vibe-cv service.
func (c *Client) Health(ctx context.Context) (*HealthResponse, error) {
	var health HealthResponse
	if err := c.doRequest(ctx, "GET", "/api/latest/health", nil, &health); err != nil {
		return nil, fmt.Errorf("health check failed: %w", err)
	}

	return &health, nil
}

// Ping performs a simple connectivity check to the vibe-cv service.
// It returns nil if the service is reachable and healthy.
func (c *Client) Ping(ctx context.Context) error {
	health, err := c.Health(ctx)
	if err != nil {
		return err
	}

	if health.Status != "ok" {
		return fmt.Errorf("service is unhealthy: status=%s", health.Status)
	}

	return nil
}
