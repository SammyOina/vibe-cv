// Copyright (c) Ultraviolet
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/sammyoina/vibe-cv/internal/analytics"
	"github.com/sammyoina/vibe-cv/internal/auth"
	"github.com/sammyoina/vibe-cv/internal/batch"
	"github.com/sammyoina/vibe-cv/internal/db"
	"github.com/sammyoina/vibe-cv/internal/input"
	"github.com/sammyoina/vibe-cv/internal/latex"
	"github.com/sammyoina/vibe-cv/internal/llm"
	"github.com/sammyoina/vibe-cv/internal/parser"
	"github.com/sammyoina/vibe-cv/internal/types"
)

// LatestHandler consolidates Phase 1-3 endpoints into /api/latest.
type LatestHandler struct {
	provider     llm.Provider
	repo         *db.Repository
	queue        *batch.JobQueue
	collector    *analytics.Collector
	authConfig   *auth.Config
	inputParser  *input.EnhancedParser
	cvParser     *parser.CVParser
	texGenerator *latex.LaTeXGenerator
	outputDir    string
}

// NewLatestHandler creates a new consolidated handler.
func NewLatestHandler(provider llm.Provider, repo *db.Repository, authConfig *auth.Config) *LatestHandler {
	outputDir := "./outputs"
	handler := &LatestHandler{
		provider:     provider,
		repo:         repo,
		queue:        batch.NewJobQueue(repo, 4), // 4 workers
		collector:    analytics.NewCollector(repo),
		authConfig:   authConfig,
		inputParser:  input.NewEnhancedParser(),
		cvParser:     parser.NewCVParser(),
		texGenerator: latex.NewLaTeXGenerator(outputDir, "pdflatex"),
		outputDir:    outputDir,
	}
	// Set the LLM provider on the batch queue
	handler.queue.SetProvider(provider)

	return handler
}

// StartQueue starts the batch job queue workers.
func (h *LatestHandler) StartQueue() {
	if h.queue != nil {
		h.queue.Start()
	}
}

// StopQueue stops the batch job queue workers.
func (h *LatestHandler) StopQueue() {
	if h.queue != nil {
		h.queue.Stop()
	}
}

// RegisterRoutes registers all latest API routes.
func (h *LatestHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/latest/customize-cv", h.CustomizeCV)
	mux.HandleFunc("POST /api/latest/batch-customize", h.BatchCustomize)
	mux.HandleFunc("GET /api/latest/versions/{cv_id}", h.GetVersions)
	mux.HandleFunc("GET /api/latest/versions/{version_id}/detail", h.GetVersionDetail)
	mux.HandleFunc("GET /api/latest/download/{version_id}", h.DownloadCV)
	mux.HandleFunc("POST /api/latest/compare-versions", h.CompareVersions)
	mux.HandleFunc("GET /api/latest/analytics", h.GetAnalytics)
	mux.HandleFunc("GET /api/latest/dashboard", h.GetDashboard)
	mux.HandleFunc("GET /api/latest/batch/{job_id}/status", h.GetBatchStatus)
	mux.HandleFunc("GET /api/latest/batch/{job_id}/download", h.DownloadBatch)
	mux.HandleFunc("GET /api/latest/health", h.Health)
}

// CustomizeCV handles the main CV customization endpoint.
func (h *LatestHandler) CustomizeCV(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req types.CustomizeCVRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request"}`, http.StatusBadRequest)

		return
	}

	// Extract user if authenticated
	var identityID *int

	if user := auth.GetUser(r.Context()); user != nil {
		// In a real implementation, map user ID to database identity
		// For now, create a placeholder identity ID from user email hash
		// This should be replaced with proper user->identity mapping from DB
		if user.Email != "" {
			// Skip analytics if no proper identity mapping
			_ = user
		}
	}

	// Get CV text (simplified - just use raw text for now)
	cvText := req.CV

	// Parse CV to database
	cvRecord, err := h.repo.CreateCV(identityID, cvText)
	if err != nil {
		http.Error(w, `{"error": "failed to store CV"}`, http.StatusInternalServerError)

		return
	}

	// Get job description
	jobDesc := req.JobDescription

	// Combine additional context strings
	contextStrings := make([]string, 0)

	for _, ctx := range req.AdditionalContext {
		if ctx.Type == "text" {
			contextStrings = append(contextStrings, ctx.Content)
		}
	}

	// Customize CV using LLM
	result, err := h.provider.Customize(r.Context(), cvText, jobDesc, contextStrings)
	if err != nil {
		http.Error(w, `{"error": "customization failed"}`, http.StatusInternalServerError)

		return
	}

	// Store version
	resultJSON, _ := json.Marshal(result.Modifications)

	_, err = h.repo.CreateCVVersion(cvRecord.ID, jobDesc, result.ModifiedCV, &result.MatchScore, (*json.RawMessage)(&resultJSON), nil)
	if err != nil {
		fmt.Printf("Failed to store version: %v\n", err)
	}

	// Generate PDF from the customized CV
	pdfFilename := fmt.Sprintf("cv-%d", cvRecord.ID)

	_, err = h.texGenerator.GeneratePDF(result.ModifiedCV, pdfFilename)
	if err != nil {
		// Log the detailed error for debugging
		fmt.Printf("Failed to generate PDF for CV %d: %v\n", cvRecord.ID, err)
	}

	// Prepare response
	customizeResp := &types.CustomizeCVResponse{
		Status:          "success",
		CustomizedCVURL: fmt.Sprintf("/outputs/cv-%d.pdf", cvRecord.ID),
		MatchScore:      result.MatchScore,
		Modifications:   result.Modifications,
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(customizeResp)
}

// BatchCustomize handles batch CV customization.
func (h *LatestHandler) BatchCustomize(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var batchReq map[string]any
	if err := json.NewDecoder(r.Body).Decode(&batchReq); err != nil {
		http.Error(w, `{"error": "invalid request"}`, http.StatusBadRequest)

		return
	}

	items, ok := batchReq["items"].([]any)
	if !ok || len(items) == 0 {
		http.Error(w, `{"error": "items required"}`, http.StatusBadRequest)

		return
	}

	// Convert items to map format
	itemMaps := make([]map[string]any, len(items))
	for i, item := range items {
		itemMaps[i] = item.(map[string]any)
	}

	var identityID *int

	// Create batch job with number of items
	jobID, err := h.queue.CreateJob(identityID, len(items))
	if err != nil {
		http.Error(w, `{"error": "failed to create batch job"}`, http.StatusInternalServerError)

		return
	}

	response := map[string]any{
		"job_id": jobID,
		"status": "processing",
	}

	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(response)
}

// GetVersions retrieves all versions for a CV.
func (h *LatestHandler) GetVersions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	cvIDStr := r.PathValue("cv_id")

	cvID, err := strconv.Atoi(cvIDStr)
	if err != nil {
		http.Error(w, `{"error": "invalid cv_id"}`, http.StatusBadRequest)

		return
	}

	versions, err := h.repo.GetCVVersions(cvID)
	if err != nil {
		http.Error(w, `{"error": "failed to retrieve versions"}`, http.StatusInternalServerError)

		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(versions)
}

// GetVersionDetail retrieves a specific version with full details.
func (h *LatestHandler) GetVersionDetail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	versionIDStr := r.PathValue("version_id")

	versionID, err := strconv.Atoi(versionIDStr)
	if err != nil {
		http.Error(w, `{"error": "invalid version_id"}`, http.StatusBadRequest)

		return
	}

	version, err := h.repo.GetCVVersion(versionID)
	if err != nil {
		http.Error(w, `{"error": "version not found"}`, http.StatusNotFound)

		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(version)
}

// CompareVersions compares two CV versions.
func (h *LatestHandler) CompareVersions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var compareReq map[string]int
	if err := json.NewDecoder(r.Body).Decode(&compareReq); err != nil {
		http.Error(w, `{"error": "invalid request"}`, http.StatusBadRequest)

		return
	}

	version1ID := compareReq["version_id_1"]
	version2ID := compareReq["version_id_2"]

	v1, err := h.repo.GetCVVersion(version1ID)
	if err != nil {
		http.Error(w, `{"error": "version 1 not found"}`, http.StatusNotFound)

		return
	}

	v2, err := h.repo.GetCVVersion(version2ID)
	if err != nil {
		http.Error(w, `{"error": "version 2 not found"}`, http.StatusNotFound)

		return
	}

	comparison := map[string]any{
		"version_1":  v1,
		"version_2":  v2,
		"job_desc_1": v1.JobDescription,
		"job_desc_2": v2.JobDescription,
	}

	// Calculate match score difference if available
	if v2.MatchScore != nil && v1.MatchScore != nil {
		comparison["match_score_diff"] = *v2.MatchScore - *v1.MatchScore
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(comparison)
}

// GetAnalytics retrieves analytics for the current user.
func (h *LatestHandler) GetAnalytics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var identityID *int

	limit := 50

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	analytics, err := h.collector.GetAnalytics(identityID, limit)
	if err != nil {
		http.Error(w, `{"error": "failed to retrieve analytics"}`, http.StatusInternalServerError)

		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(analytics)
}

// GetDashboard retrieves global dashboard statistics.
func (h *LatestHandler) GetDashboard(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	dashboard, err := h.collector.GetDashboard()
	if err != nil {
		http.Error(w, `{"error": "failed to retrieve dashboard"}`, http.StatusInternalServerError)

		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(dashboard)
}

// GetBatchStatus retrieves the status of a batch job.
func (h *LatestHandler) GetBatchStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	jobIDStr := r.PathValue("job_id")

	jobID, err := strconv.Atoi(jobIDStr)
	if err != nil {
		http.Error(w, `{"error": "invalid job_id"}`, http.StatusBadRequest)

		return
	}

	jobStatus, err := h.queue.GetBatchJobStatus(jobID)
	if err != nil {
		http.Error(w, `{"error": "job not found"}`, http.StatusNotFound)

		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(jobStatus)
}

// DownloadBatch downloads results for a batch job.
func (h *LatestHandler) DownloadBatch(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	jobIDStr := r.PathValue("job_id")

	jobID, err := strconv.Atoi(jobIDStr)
	if err != nil {
		http.Error(w, `{"error": "invalid job_id"}`, http.StatusBadRequest)

		return
	}

	// Get batch job and items
	job, err := h.repo.GetBatchJob(jobID)
	if err != nil {
		http.Error(w, `{"error": "job not found"}`, http.StatusInternalServerError)

		return
	}

	items, err := h.repo.GetBatchJobItems(jobID)
	if err != nil {
		http.Error(w, `{"error": "failed to retrieve batch items"}`, http.StatusInternalServerError)

		return
	}

	result := map[string]any{
		"job_id":     job.ID,
		"status":     job.Status,
		"created_at": job.CreatedAt,
		"items":      items,
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(result)
}

// DownloadCV retrieves a customized CV version from the database and serves it as a PDF.
func (h *LatestHandler) DownloadCV(w http.ResponseWriter, r *http.Request) {
	versionIDStr := r.PathValue("version_id")

	versionID, err := strconv.Atoi(versionIDStr)
	if err != nil {
		http.Error(w, `{"error": "invalid version_id"}`, http.StatusBadRequest)

		return
	}

	// Get the CV version from database
	version, err := h.repo.GetCVVersion(versionID)
	if err != nil {
		http.Error(w, `{"error": "version not found"}`, http.StatusNotFound)

		return
	}

	// Try to generate PDF from the customized CV content
	pdfFilename := fmt.Sprintf("cv-version-%d", versionID)
	pdfPath, pdfErr := h.texGenerator.GeneratePDF(version.CustomizedCV, pdfFilename)

	// If PDF generation succeeds, serve the PDF
	if pdfErr == nil {
		pdfContent, err := os.ReadFile(pdfPath)
		if err == nil {
			w.Header().Set("Content-Type", "application/pdf")
			w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"cv-version-%d.pdf\"", versionID))
			w.Header().Set("Content-Length", strconv.Itoa(len(pdfContent)))
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(pdfContent)

			return
		}
		// Log file read error
		fmt.Printf("Failed to read generated PDF file: %v\n", err)
	} else {
		// Log detailed PDF generation error
		fmt.Printf("Failed to generate PDF for version %d: %v\n", versionID, pdfErr)
	}

	// Return error instead of silently falling back to text
	w.Header().Set("Content-Type", "application/json")

	errorMsg := "PDF generation failed"
	if pdfErr != nil {
		errorMsg = fmt.Sprintf("PDF generation failed: %v", pdfErr)
	}

	http.Error(w, fmt.Sprintf(`{"error": "%s"}`, errorMsg), http.StatusInternalServerError)
}

// Health checks the health of the service.
func (h *LatestHandler) Health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	health := map[string]any{
		"status":    "ok",
		"version":   "phase-4",
		"database":  "connected",
		"timestamp": time.Now().Format(time.RFC3339),
		"auth":      map[string]any{"enabled": h.authConfig.Enabled},
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(health)
}
