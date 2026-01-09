// Copyright (c) Ultraviolet
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/sammyoina/vibe-cv/internal/ats"
	"github.com/sammyoina/vibe-cv/internal/db"
	"github.com/sammyoina/vibe-cv/internal/llm"
)

// ATSHandler handles ATS analysis endpoints.
type ATSHandler struct {
	analyzer *ats.Analyzer
	repo     *db.Repository
}

// NewATSHandler creates a new ATS handler.
func NewATSHandler(provider llm.Provider, repo *db.Repository) *ATSHandler {
	return &ATSHandler{
		analyzer: ats.NewAnalyzer(provider),
		repo:     repo,
	}
}

// AnalyzeRequest represents the ATS analysis request.
type AnalyzeRequest struct {
	CVVersionID    int    `json:"cv_version_id"`
	JobDescription string `json:"job_description"`
}

// AnalyzeCV handles POST /api/latest/ats/analyze
func (h *ATSHandler) AnalyzeCV(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req AnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request"}`, http.StatusBadRequest)
		return
	}

	// Get CV version
	version, err := h.repo.GetCVVersion(req.CVVersionID)
	if err != nil {
		http.Error(w, `{"error": "CV version not found"}`, http.StatusNotFound)
		return
	}

	// Perform ATS analysis
	result, err := h.analyzer.AnalyzeCV(r.Context(), version.CustomizedCV, req.JobDescription)
	if err != nil {
		fmt.Printf("ATS analysis failed: %v\n", err)
		http.Error(w, `{"error": "analysis failed"}`, http.StatusInternalServerError)
		return
	}

	// Convert result to JSON for storage
	keywordMatchesJSON, _ := json.Marshal(result.KeywordMatches)
	formattingIssuesJSON, _ := json.Marshal(result.FormattingIssues)
	sectionCompletenessJSON, _ := json.Marshal(result.SectionCompleteness)
	recommendationsJSON, _ := json.Marshal(result.Recommendations)

	// Store analysis in database
	analysis, err := h.repo.CreateATSAnalysis(
		req.CVVersionID,
		&result.OverallScore,
		(*json.RawMessage)(&keywordMatchesJSON),
		(*json.RawMessage)(&formattingIssuesJSON),
		(*json.RawMessage)(&sectionCompletenessJSON),
		(*json.RawMessage)(&recommendationsJSON),
	)
	if err != nil {
		fmt.Printf("Failed to store ATS analysis: %v\n", err)
		// Continue anyway, return the result
	}

	// Prepare response
	response := map[string]interface{}{
		"id":                   analysis.ID,
		"cv_version_id":        req.CVVersionID,
		"overall_score":        result.OverallScore,
		"keyword_matches":      result.KeywordMatches,
		"formatting_issues":    result.FormattingIssues,
		"section_completeness": result.SectionCompleteness,
		"recommendations":      result.Recommendations,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetATSAnalysis handles GET /api/latest/ats/{cv_version_id}
func (h *ATSHandler) GetATSAnalysis(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	cvVersionIDStr := r.PathValue("cv_version_id")
	cvVersionID, err := strconv.Atoi(cvVersionIDStr)
	if err != nil {
		http.Error(w, `{"error": "invalid cv_version_id"}`, http.StatusBadRequest)
		return
	}

	// Get ATS analysis from database
	analysis, err := h.repo.GetATSAnalysis(cvVersionID)
	if err != nil {
		http.Error(w, `{"error": "analysis not found"}`, http.StatusNotFound)
		return
	}

	// Parse JSON fields
	var keywordMatches ats.KeywordMatches
	var formattingIssues []ats.FormattingIssue
	var sectionCompleteness ats.SectionCompleteness
	var recommendations []ats.Recommendation

	if analysis.KeywordMatches != nil {
		json.Unmarshal(*analysis.KeywordMatches, &keywordMatches)
	}
	if analysis.FormattingIssues != nil {
		json.Unmarshal(*analysis.FormattingIssues, &formattingIssues)
	}
	if analysis.SectionCompleteness != nil {
		json.Unmarshal(*analysis.SectionCompleteness, &sectionCompleteness)
	}
	if analysis.Recommendations != nil {
		json.Unmarshal(*analysis.Recommendations, &recommendations)
	}

	response := map[string]interface{}{
		"id":                   analysis.ID,
		"cv_version_id":        analysis.CVVersionID,
		"overall_score":        analysis.OverallScore,
		"keyword_matches":      keywordMatches,
		"formatting_issues":    formattingIssues,
		"section_completeness": sectionCompleteness,
		"recommendations":      recommendations,
		"created_at":           analysis.CreatedAt,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
