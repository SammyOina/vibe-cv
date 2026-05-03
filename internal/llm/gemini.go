// Copyright (c) Ultraviolet
// SPDX-License-Identifier: Apache-2.0

package llm

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/genai"
)

// GeminiProvider implements the Provider interface using Google's Gemini API.
type GeminiProvider struct {
	client *genai.Client
	model  string
}

// NewGeminiProvider creates a new Gemini provider.
func NewGeminiProvider(ctx context.Context, apiKey, model string) (*GeminiProvider, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	return &GeminiProvider{
		client: client,
		model:  model,
	}, nil
}

// Customize customizes a CV using Google Gemini.
func (p *GeminiProvider) Customize(ctx context.Context, cv, jobDescription string, additionalContext []string, latexTemplate string, isFullLatex bool) (*CustomizationResponse, error) {
	prompt := buildPrompt(cv, jobDescription, additionalContext)

	if isFullLatex && latexTemplate != "" {
		prompt += "\n\nYou must generate the complete CV conforming STRICTLY to the following LaTeX template format. Do NOT modify the layout/packages, only rewrite the textual content matching the candidate's specifics. Return the complete, compilable LaTeX code AFTER the ---LATEX--- delimiter."
		prompt += "\n\nLaTeX Template:\n" + latexTemplate
	}

	// Create request content
	contents := make([]*genai.Content, 0, 2)

	// Add system instruction as first message
	contents = append(contents, &genai.Content{
		Role:  "user",
		Parts: []*genai.Part{genai.NewPartFromText(systemPrompt)},
	})

	// Add the actual prompt
	contents = append(contents, &genai.Content{
		Role:  "user",
		Parts: []*genai.Part{genai.NewPartFromText(prompt)},
	})

	// Set up generation config with temperature and topP
	temp := float32(0.7)
	topP := float32(0.9)

	config := &genai.GenerateContentConfig{
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{genai.NewPartFromText(systemPrompt)},
		},
		Temperature: &temp,
		TopP:        &topP,
	}

	// Generate response
	resp, err := p.client.Models.GenerateContent(ctx, p.model, contents, config)
	if err != nil {
		return nil, fmt.Errorf("failed to customize CV with Gemini: %w", err)
	}

	if resp == nil || len(resp.Candidates) == 0 {
		return nil, errors.New("no response from Gemini")
	}

	// Extract content from response
	content := ""

	if candidate := resp.Candidates[0]; candidate != nil && len(candidate.Content.Parts) > 0 {
		if part := candidate.Content.Parts[0]; part != nil {
			content = part.Text
		}
	}

	modifiedCV, matchScore, modifications := parseResponse(content, isFullLatex)

	return &CustomizationResponse{
		ModifiedCV:    modifiedCV,
		MatchScore:    matchScore,
		Modifications: modifications,
	}, nil
}

// GetName returns the provider name.
func (p *GeminiProvider) GetName() string {
	return "gemini"
}
