// Copyright (c) Ultraviolet
// SPDX-License-Identifier: Apache-2.0

package sdk

import "time"

// CustomizeCVRequest represents a request to customize a CV.
type CustomizeCVRequest struct {
	CV                string        `json:"cv"`
	CVFile            string        `json:"cv_file,omitempty"`
	JobDescription    string        `json:"job_description"`
	JobDescriptionURL string        `json:"job_description_url,omitempty"`
	LinkedInProfile   string        `json:"linkedin_profile,omitempty"`
	AdditionalContext []ContextItem `json:"additional_context,omitempty"`
	LLMConfig         *LLMConfig    `json:"llm_config,omitempty"`
	InputSources      []InputSource `json:"input_sources,omitempty"`
}

// ContextItem represents additional context (text or URL).
type ContextItem struct {
	Type    string `json:"type"` // "text" or "url"
	Content string `json:"content"`
}

// InputSource represents a source of input.
type InputSource struct {
	Type     string `json:"type"`     // "text", "url", "pdf", "docx", "linkedin"
	Content  string `json:"content"`  // For text input or base64 for files
	URL      string `json:"url"`      // For URL input
	FileName string `json:"filename"` // Original filename for files
}

// LLMConfig allows per-request override of LLM provider settings.
type LLMConfig struct {
	Provider string `json:"provider"` // "openai", "anthropic", "gemini"
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

// BatchItem represents a single item in a batch customization request.
type BatchItem struct {
	CV             string `json:"cv"`
	JobDescription string `json:"job_description"`
}

// BatchCustomizeRequest represents a batch customization request.
type BatchCustomizeRequest struct {
	Items []BatchItem `json:"items"`
}

// BatchJobResponse represents the response from submitting a batch job.
type BatchJobResponse struct {
	JobID  int    `json:"job_id"`
	Status string `json:"status"`
}

// BatchJobStatus represents the status of a batch job.
type BatchJobStatus struct {
	ID             int        `json:"id"`
	Status         string     `json:"status"`
	TotalItems     int        `json:"total_items"`
	CompletedItems int        `json:"completed_items"`
	CreatedAt      time.Time  `json:"created_at"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// BatchJobItem represents a single item in a batch job.
type BatchJobItem struct {
	ID             int         `json:"id"`
	BatchJobID     int         `json:"batch_job_id"`
	CVID           *int        `json:"cv_id,omitempty"`
	JobDescription string      `json:"job_description"`
	Status         string      `json:"status"`
	Result         interface{} `json:"result,omitempty"`
	ErrorMessage   *string     `json:"error_message,omitempty"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
}

// BatchResults represents the complete results of a batch job.
type BatchResults struct {
	JobID     int            `json:"job_id"`
	Status    string         `json:"status"`
	CreatedAt time.Time      `json:"created_at"`
	Items     []BatchJobItem `json:"items"`
}

// CVVersion represents a CV version.
type CVVersion struct {
	ID              int         `json:"id"`
	CVID            int         `json:"cv_id"`
	JobDescription  string      `json:"job_description"`
	CustomizedCV    string      `json:"customized_cv"`
	MatchScore      *float64    `json:"match_score,omitempty"`
	AgentMetrics    interface{} `json:"agent_metrics,omitempty"`
	WorkflowHistory interface{} `json:"workflow_history,omitempty"`
	CreatedAt       time.Time   `json:"created_at"`
}

// VersionComparison represents a comparison between two CV versions.
type VersionComparison struct {
	Version1       *CVVersion `json:"version_1"`
	Version2       *CVVersion `json:"version_2"`
	JobDesc1       string     `json:"job_desc_1"`
	JobDesc2       string     `json:"job_desc_2"`
	MatchScoreDiff *float64   `json:"match_score_diff,omitempty"`
}

// CompareVersionsRequest represents a request to compare two versions.
type CompareVersionsRequest struct {
	VersionID1 int `json:"version_id_1"`
	VersionID2 int `json:"version_id_2"`
}

// Analytics represents user analytics data.
type Analytics struct {
	Snapshots []AnalyticsSnapshot `json:"snapshots"`
}

// AnalyticsSnapshot represents a single analytics snapshot.
type AnalyticsSnapshot struct {
	ID              int         `json:"id"`
	IdentityID      *int        `json:"identity_id,omitempty"`
	MatchScore      *float64    `json:"match_score,omitempty"`
	KeywordCoverage *float64    `json:"keyword_coverage,omitempty"`
	Timestamp       time.Time   `json:"timestamp"`
	Metadata        interface{} `json:"metadata,omitempty"`
}

// Dashboard represents global dashboard statistics.
type Dashboard struct {
	TotalCVs          int     `json:"total_cvs"`
	TotalVersions     int     `json:"total_versions"`
	TotalBatchJobs    int     `json:"total_batch_jobs"`
	AverageMatchScore float64 `json:"average_match_score"`
	RecentActivity    int     `json:"recent_activity"`
	ActiveUsers       int     `json:"active_users"`
}

// HealthResponse represents the health check response.
type HealthResponse struct {
	Status    string                 `json:"status"`
	Version   string                 `json:"version"`
	Database  string                 `json:"database"`
	Timestamp string                 `json:"timestamp"`
	Auth      map[string]interface{} `json:"auth"`
}

// LLM Provider constants.
const (
	ProviderOpenAI    = "openai"
	ProviderAnthropic = "anthropic"
	ProviderGemini    = "gemini"
)

// Common LLM models.
const (
	ModelGPT4          = "gpt-4"
	ModelGPT35Turbo    = "gpt-3.5-turbo"
	ModelClaude3Opus   = "claude-3-opus"
	ModelClaude3Sonnet = "claude-3-sonnet"
	ModelGeminiPro     = "gemini-pro"
)
