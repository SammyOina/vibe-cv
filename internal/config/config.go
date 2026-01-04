package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config holds application configuration
type Config struct {
	LLMProvider     string
	LLMAPIKey       string
	LLMModel        string
	ServerPort      string
	ServerHost      string
	OutputDir       string
	LaTeXPath       string
	DatabaseURL     string
	KratosEnabled   bool
	KratosPublicURL string
	KratosAdminURL  string
}

// Load loads configuration from environment variables and .env file
func Load() (*Config, error) {
	// Try loading from .env file (won't fail if not present)
	_ = godotenv.Load()

	cfg := &Config{
		LLMProvider:     getEnv("LLM_PROVIDER", "openai"),
		LLMAPIKey:       getEnv("LLM_API_KEY", ""),
		LLMModel:        getEnv("LLM_MODEL", "gpt-4"),
		ServerPort:      getEnv("SERVER_PORT", "8080"),
		ServerHost:      getEnv("SERVER_HOST", "localhost"),
		OutputDir:       getEnv("OUTPUT_DIR", "./outputs"),
		LaTeXPath:       getEnv("LATEX_PATH", "pdflatex"),
		DatabaseURL:     getEnv("DATABASE_URL", ""),
		KratosEnabled:   getEnv("KRATOS_ENABLED", "false") == "true",
		KratosPublicURL: getEnv("KRATOS_PUBLIC_URL", "http://localhost:4433"),
		KratosAdminURL:  getEnv("KRATOS_ADMIN_URL", "http://localhost:4434"),
	}

	// Validate required fields
	if cfg.LLMAPIKey == "" {
		return nil, fmt.Errorf("LLM_API_KEY environment variable not set")
	}

	return cfg, nil
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
