// Copyright (c) Ultraviolet
// SPDX-License-Identifier: Apache-2.0

package types

// CustomizeCVRequest represents the request to customize a CV.
type CustomizeCVRequest struct {
	CV                string        `json:"cv"`
	CVFile            string        `json:"cv_file,omitempty"` // Base64 encoded file for Phase 2
	JobDescription    string        `json:"job_description"`
	JobDescriptionURL string        `json:"job_description_url,omitempty"` // Phase 2
	LinkedInProfile   string        `json:"linkedin_profile,omitempty"`    // Phase 2
	AdditionalContext []ContextItem `json:"additional_context,omitempty"`
	LLMConfig         *LLMConfig    `json:"llm_config,omitempty"`
	InputSources      []InputSource `json:"input_sources,omitempty"` // Phase 2
}

// InputSource represents a source of input (Phase 2).
type InputSource struct {
	Type     string `json:"type"`     // "text", "url", "pdf", "docx", "linkedin"
	Content  string `json:"content"`  // For text input or base64 for files
	URL      string `json:"url"`      // For URL input
	FileName string `json:"filename"` // Original filename for files
}

// ContextItem represents additional context (text or URL).
type ContextItem struct {
	Type    string `json:"type"` // "text" or "url"
	Content string `json:"content"`
}

// LLMConfig allows per-request override of LLM provider settings.
type LLMConfig struct {
	Provider string `json:"provider"` // "openai", "anthropic", etc.
	Model    string `json:"model"`
	APIKey   string `json:"api_key,omitempty"`
}

// CustomizeCVResponse represents the response from CV customization.
type CustomizeCVResponse struct {
	Status          string   `json:"status"`
	CustomizedCVURL string   `json:"customized_cv_url"`
	MatchScore      float64  `json:"match_score"`
	Modifications   []string `json:"modifications"`
	Error           string   `json:"error,omitempty"`
}

// CVContent represents parsed CV content.
type CVContent struct {
	RawText    string
	Parsed     map[string]any
	SourceType string // "text", "pdf"
}

// JobDescription represents parsed job description.
type JobDescription struct {
	RawText      string
	Title        string
	Company      string
	Requirements []string
	Keywords     []string
	SourceType   string // "text", "url"
}

// CustomizationResult represents the result of CV customization.
type CustomizationResult struct {
	MatchScore    float64
	ModifiedCV    string
	Modifications []string
}
