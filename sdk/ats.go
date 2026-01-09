// Copyright (c) Ultraviolet
// SPDX-License-Identifier: Apache-2.0

package sdk

import (
	"context"
	"fmt"
)

// ATSAnalysisResponse represents the response from ATS analysis.
type ATSAnalysisResponse struct {
	ID                  int                    `json:"id"`
	CVVersionID         int                    `json:"cv_version_id"`
	OverallScore        float64                `json:"overall_score"`
	KeywordMatches      map[string]interface{} `json:"keyword_matches"`
	FormattingIssues    []interface{}          `json:"formatting_issues"`
	SectionCompleteness map[string]float64     `json:"section_completeness"`
	Recommendations     []interface{}          `json:"recommendations"`
	CreatedAt           string                 `json:"created_at,omitempty"`
}

// AnalyzeATSRequest represents the request for ATS analysis.
type AnalyzeATSRequest struct {
	CVVersionID    int    `json:"cv_version_id"`
	JobDescription string `json:"job_description"`
}

// AnalyzeATS analyzes a CV version for ATS compatibility.
func (c *Client) AnalyzeATS(ctx context.Context, cvVersionID int, jobDescription string, opts ...RequestOption) (*ATSAnalysisResponse, error) {
	req := AnalyzeATSRequest{
		CVVersionID:    cvVersionID,
		JobDescription: jobDescription,
	}

	var result ATSAnalysisResponse
	err := c.doRequest(ctx, "POST", "/api/latest/ats/analyze", req, &result, opts...)
	if err != nil {
		return nil, fmt.Errorf("ATS analysis failed: %w", err)
	}

	return &result, nil
}

// GetATSAnalysis retrieves an existing ATS analysis for a CV version.
func (c *Client) GetATSAnalysis(ctx context.Context, cvVersionID int, opts ...RequestOption) (*ATSAnalysisResponse, error) {
	path := fmt.Sprintf("/api/latest/ats/%d", cvVersionID)

	var result ATSAnalysisResponse
	err := c.doRequest(ctx, "GET", path, nil, &result, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to get ATS analysis: %w", err)
	}

	return &result, nil
}
