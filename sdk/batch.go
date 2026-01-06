// Copyright (c) Ultraviolet
// SPDX-License-Identifier: Apache-2.0

package sdk

import (
	"context"
	"fmt"
	"time"
)

// BatchCustomize submits a batch customization job.
func (c *Client) BatchCustomize(ctx context.Context, items []BatchItem, opts ...RequestOption) (*BatchJobResponse, error) {
	if len(items) == 0 {
		return nil, &ValidationError{Field: "items", Message: "at least one item is required"}
	}

	req := BatchCustomizeRequest{Items: items}
	var resp BatchJobResponse
	if err := c.doRequest(ctx, "POST", "/api/latest/batch-customize", req, &resp, opts...); err != nil {
		return nil, fmt.Errorf("failed to submit batch job: %w", err)
	}

	return &resp, nil
}

// GetBatchStatus retrieves the status of a batch job.
func (c *Client) GetBatchStatus(ctx context.Context, jobID int, opts ...RequestOption) (*BatchJobStatus, error) {
	if jobID <= 0 {
		return nil, &ValidationError{Field: "jobID", Message: "job ID must be positive"}
	}

	path := fmt.Sprintf("/api/latest/batch/%d/status", jobID)
	var status BatchJobStatus
	if err := c.doRequest(ctx, "GET", path, nil, &status, opts...); err != nil {
		return nil, fmt.Errorf("failed to get batch status: %w", err)
	}

	return &status, nil
}

// DownloadBatchResults downloads the results of a completed batch job.
func (c *Client) DownloadBatchResults(ctx context.Context, jobID int, opts ...RequestOption) (*BatchResults, error) {
	if jobID <= 0 {
		return nil, &ValidationError{Field: "jobID", Message: "job ID must be positive"}
	}

	path := fmt.Sprintf("/api/latest/batch/%d/download", jobID)
	var results BatchResults
	if err := c.doRequest(ctx, "GET", path, nil, &results, opts...); err != nil {
		return nil, fmt.Errorf("failed to download batch results: %w", err)
	}

	return &results, nil
}

// WaitForBatch polls a batch job until it completes or the context is cancelled.
// It returns the final results when the job is complete.
func (c *Client) WaitForBatch(ctx context.Context, jobID int, pollInterval time.Duration, opts ...RequestOption) (*BatchResults, error) {
	if pollInterval <= 0 {
		pollInterval = 5 * time.Second
	}

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			status, err := c.GetBatchStatus(ctx, jobID, opts...)
			if err != nil {
				return nil, err
			}

			switch status.Status {
			case "completed":
				return c.DownloadBatchResults(ctx, jobID, opts...)
			case "failed":
				return nil, fmt.Errorf("batch job %d failed", jobID)
			case "pending", "processing":
				// Continue polling
				continue
			default:
				return nil, fmt.Errorf("unknown batch job status: %s", status.Status)
			}
		}
	}
}
