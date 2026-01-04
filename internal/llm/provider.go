package llm

import "context"

// Provider is the interface that all LLM providers must implement
type Provider interface {
	// Customize takes a CV and job description and returns a customized CV with metadata
	Customize(ctx context.Context, cv, jobDescription string, additionalContext []string) (*CustomizationResponse, error)
	// GetName returns the provider name
	GetName() string
}

// CustomizationResponse contains the result of CV customization
type CustomizationResponse struct {
	ModifiedCV    string
	MatchScore    float64
	Modifications []string
}

// ProviderConfig holds configuration for a specific provider
type ProviderConfig struct {
	Name   string
	APIKey string
	Model  string
}
