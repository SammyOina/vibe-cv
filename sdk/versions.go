// Copyright (c) Ultraviolet
// SPDX-License-Identifier: Apache-2.0

package sdk

import (
	"context"
	"fmt"
)

// GetVersions retrieves all versions for a specific CV.
func (c *Client) GetVersions(ctx context.Context, cvID int) ([]*CVVersion, error) {
	if cvID <= 0 {
		return nil, &ValidationError{Field: "cvID", Message: "CV ID must be positive"}
	}

	path := fmt.Sprintf("/api/latest/versions/%d", cvID)
	var versions []*CVVersion
	if err := c.doRequest(ctx, "GET", path, nil, &versions); err != nil {
		return nil, fmt.Errorf("failed to get versions: %w", err)
	}

	return versions, nil
}

// GetVersionDetail retrieves detailed information about a specific CV version.
func (c *Client) GetVersionDetail(ctx context.Context, versionID int) (*CVVersion, error) {
	if versionID <= 0 {
		return nil, &ValidationError{Field: "versionID", Message: "version ID must be positive"}
	}

	path := fmt.Sprintf("/api/latest/versions/%d/detail", versionID)
	var version CVVersion
	if err := c.doRequest(ctx, "GET", path, nil, &version); err != nil {
		return nil, fmt.Errorf("failed to get version detail: %w", err)
	}

	return &version, nil
}

// CompareVersions compares two CV versions.
func (c *Client) CompareVersions(ctx context.Context, versionID1, versionID2 int) (*VersionComparison, error) {
	if versionID1 <= 0 {
		return nil, &ValidationError{Field: "versionID1", Message: "version ID 1 must be positive"}
	}
	if versionID2 <= 0 {
		return nil, &ValidationError{Field: "versionID2", Message: "version ID 2 must be positive"}
	}

	req := CompareVersionsRequest{
		VersionID1: versionID1,
		VersionID2: versionID2,
	}

	var comparison VersionComparison
	if err := c.doRequest(ctx, "POST", "/api/latest/compare-versions", req, &comparison); err != nil {
		return nil, fmt.Errorf("failed to compare versions: %w", err)
	}

	return &comparison, nil
}

// DownloadCV downloads a CV version as a PDF or text file.
func (c *Client) DownloadCV(ctx context.Context, versionID int) ([]byte, error) {
	if versionID <= 0 {
		return nil, &ValidationError{Field: "versionID", Message: "version ID must be positive"}
	}

	path := fmt.Sprintf("/api/latest/download/%d", versionID)
	data, err := c.doRequestRaw(ctx, "GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to download CV: %w", err)
	}

	return data, nil
}
