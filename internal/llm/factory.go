// Copyright (c) Ultraviolet
// SPDX-License-Identifier: Apache-2.0

package llm

import (
	"context"
	"errors"
	"fmt"
)

// Factory creates LLM providers based on configuration.
type Factory struct {
	providers map[string]Provider
}

// NewFactory creates a new LLM factory.
func NewFactory() *Factory {
	return &Factory{
		providers: make(map[string]Provider),
	}
}

// CreateProvider creates a provider based on the name and API key.
func (f *Factory) CreateProvider(ctx context.Context, name, apiKey, model string) (Provider, error) {
	switch name {
	case "openai":
		if apiKey == "" {
			return nil, errors.New("OpenAI API key is required")
		}

		return NewOpenAIProvider(apiKey, model), nil
	case "anthropic":
		if apiKey == "" {
			return nil, errors.New("anthropic API key is required")
		}

		return NewAnthropicProvider(apiKey, model), nil
	case "gemini":
		if apiKey == "" {
			return nil, errors.New("gemini API key is required")
		}

		return NewGeminiProvider(ctx, apiKey, model)
	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s", name)
	}
}
