package main

import (
	"testing"

	"github.com/sammyoina/vibe-cv/internal/config"
)

// TestConfigLoad tests configuration loading
func TestConfigLoad(t *testing.T) {
	// Set environment variables for testing
	t.Setenv("LLM_PROVIDER", "openai")
	t.Setenv("LLM_API_KEY", "test-key")
	t.Setenv("LLM_MODEL", "gpt-4")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.LLMProvider != "openai" {
		t.Errorf("Expected LLM provider openai, got %s", cfg.LLMProvider)
	}

	if cfg.LLMAPIKey != "test-key" {
		t.Errorf("Expected API key 'test-key', got %s", cfg.LLMAPIKey)
	}
}

// TestConfigDefaults tests default configuration values
func TestConfigDefaults(t *testing.T) {
	t.Setenv("LLM_API_KEY", "test-key")
	// Don't set other values to test defaults

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.ServerPort != "8080" {
		t.Errorf("Expected default port 8080, got %s", cfg.ServerPort)
	}

	if cfg.ServerHost != "localhost" {
		t.Errorf("Expected default host localhost, got %s", cfg.ServerHost)
	}
}

// TestConfigValidation tests configuration validation
func TestConfigValidation(t *testing.T) {
	t.Setenv("LLM_PROVIDER", "openai")
	// Don't set LLM_API_KEY to test validation

	_, err := config.Load()
	if err == nil {
		t.Error("Expected error when LLM_API_KEY is missing")
	}
}
