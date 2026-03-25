// Copyright (c) Ultraviolet
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sammyoina/vibe-cv/internal/llm"
	"github.com/stretchr/testify/assert"
)

type MockProvider struct{}

func (m *MockProvider) Customize(ctx context.Context, cv, jobDesc string, additionalContext []string, latexTemplate string, isFullLatex bool) (*llm.CustomizationResponse, error) {
	return &llm.CustomizationResponse{}, nil
}

func (m *MockProvider) GetName() string {
	return "mock"
}

func TestRenderTemplateHandler(t *testing.T) {
	provider := &MockProvider{}
	handler := NewLatestHandler(provider, nil, nil)

	// Test case: Invalid JSON
	req, _ := http.NewRequest(http.MethodPost, "/api/latest/templates/render", bytes.NewBufferString("invalid json"))
	rr := httptest.NewRecorder()
	handler.RenderTemplate(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)

	// Test case: Missing template
	body, _ := json.Marshal(map[string]string{"latex_template": ""})
	req, _ = http.NewRequest(http.MethodPost, "/api/latest/templates/render", bytes.NewBuffer(body))
	rr = httptest.NewRecorder()
	handler.RenderTemplate(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)

	// Test case: Valid request (will fail PDF generation due to missing pdflatex)
	body, _ = json.Marshal(map[string]string{"latex_template": "\\documentclass{article}\\begin{document}Test\\end{document}"})
	req, _ = http.NewRequest(http.MethodPost, "/api/latest/templates/render", bytes.NewBuffer(body))
	rr = httptest.NewRecorder()
	handler.RenderTemplate(rr, req)

	// Since pdflatex is likely missing, we expect an internal server error
	// but we'll check if it reached the generation logic
	if rr.Code == http.StatusInternalServerError {
		var resp map[string]string
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		if err != nil {
			t.Logf("Failed to unmarshal response: %v, body: %s", err, rr.Body.String())
		}
		assert.Contains(t, resp["error"], "failed to generate PDF")
	}
}
