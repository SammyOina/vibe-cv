// Copyright (c) Ultraviolet
// SPDX-License-Identifier: Apache-2.0

package ats

// KeywordMatches represents matched and missing keywords.
type KeywordMatches struct {
	Matched []string `json:"matched"`
	Missing []string `json:"missing"`
}

// FormattingIssue represents a detected formatting problem.
type FormattingIssue struct {
	Type     string `json:"type"`     // e.g., "missing_section", "poor_formatting", "length_issue"
	Severity string `json:"severity"` // "high", "medium", "low"
	Message  string `json:"message"`
}

// Recommendation represents an actionable suggestion.
type Recommendation struct {
	Category   string `json:"category"` // e.g., "keywords", "formatting", "content"
	Priority   string `json:"priority"` // "high", "medium", "low"
	Suggestion string `json:"suggestion"`
}

// SectionCompleteness represents quality scores for CV sections.
type SectionCompleteness struct {
	Experience float64 `json:"experience"` // 0.0 to 1.0
	Education  float64 `json:"education"`  // 0.0 to 1.0
	Skills     float64 `json:"skills"`     // 0.0 to 1.0
	Summary    float64 `json:"summary"`    // 0.0 to 1.0
}

// ATSAnalysisResult represents the complete ATS analysis result.
type ATSAnalysisResult struct {
	OverallScore        float64             `json:"overall_score"`
	KeywordMatches      KeywordMatches      `json:"keyword_matches"`
	FormattingIssues    []FormattingIssue   `json:"formatting_issues"`
	SectionCompleteness SectionCompleteness `json:"section_completeness"`
	Recommendations     []Recommendation    `json:"recommendations"`
}
