package llm

import (
	"context"
	"fmt"

	"google.golang.org/genai"
)

// GeminiProvider implements the Provider interface using Google's Gemini API
type GeminiProvider struct {
	client *genai.Client
	model  string
}

// NewGeminiProvider creates a new Gemini provider
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

// Customize customizes a CV using Google Gemini
func (p *GeminiProvider) Customize(ctx context.Context, cv, jobDescription string, additionalContext []string) (*CustomizationResponse, error) {
	prompt := buildPrompt(cv, jobDescription, additionalContext)

	// Create request content
	var contents []*genai.Content

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
		return nil, fmt.Errorf("no response from Gemini")
	}

	// Extract content from response
	content := ""
	if candidate := resp.Candidates[0]; candidate != nil && len(candidate.Content.Parts) > 0 {
		if part := candidate.Content.Parts[0]; part != nil {
			content = part.Text
		}
	}

	modifiedCV, matchScore, modifications := parseResponse(content)

	return &CustomizationResponse{
		ModifiedCV:    modifiedCV,
		MatchScore:    matchScore,
		Modifications: modifications,
	}, nil
}

// GetName returns the provider name
func (p *GeminiProvider) GetName() string {
	return "gemini"
}
