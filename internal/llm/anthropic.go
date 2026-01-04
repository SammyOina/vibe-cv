// Copyright (c) Ultraviolet
// SPDX-License-Identifier: Apache-2.0

package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// AnthropicProvider implements the Provider interface using Anthropic's API.
type AnthropicProvider struct {
	apiKey string
	model  string
}

// NewAnthropicProvider creates a new Anthropic provider.
func NewAnthropicProvider(apiKey, model string) Provider {
	return &AnthropicProvider{
		apiKey: apiKey,
		model:  model,
	}
}

// Customize customizes a CV using Anthropic via REST API.
func (p *AnthropicProvider) Customize(ctx context.Context, cv, jobDescription string, additionalContext []string) (*CustomizationResponse, error) {
	prompt := buildPrompt(cv, jobDescription, additionalContext)

	// Create request body
	reqBody := map[string]any{
		"model":      p.model,
		"max_tokens": 2000,
		"system":     systemPrompt,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.anthropic.com/v1/messages", strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-Api-Key", p.apiKey)
	req.Header.Set("Anthropic-Version", "2023-06-01")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Anthropic API: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("anthropic API error: %s", string(respBody))
	}

	// Parse response
	var respData map[string]any

	err = json.Unmarshal(respBody, &respData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	content := ""

	if contentData, ok := respData["content"].([]any); ok && len(contentData) > 0 {
		if textBlock, ok := contentData[0].(map[string]any); ok {
			if text, ok := textBlock["text"].(string); ok {
				content = text
			}
		}
	}

	modifiedCV, matchScore, modifications := parseResponse(content)

	return &CustomizationResponse{
		ModifiedCV:    modifiedCV,
		MatchScore:    matchScore,
		Modifications: modifications,
	}, nil
}

// GetName returns the provider name.
func (p *AnthropicProvider) GetName() string {
	return "anthropic"
}
