// Copyright (c) Ultraviolet
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/sammyoina/vibe-cv/internal/db"
	"github.com/sammyoina/vibe-cv/internal/input"
	"github.com/sammyoina/vibe-cv/pkg/auth"
)

// LinkedInHandler handles LinkedIn import endpoints.
type LinkedInHandler struct {
	repo   *db.Repository
	parser *input.LinkedInParser
}

// NewLinkedInHandler creates a new LinkedIn handler.
func NewLinkedInHandler(repo *db.Repository) *LinkedInHandler {
	return &LinkedInHandler{
		repo:   repo,
		parser: input.NewLinkedInParser(),
	}
}

// ImportRequest represents the LinkedIn import request.
type ImportRequest struct {
	LinkedInURL string `json:"linkedin_url"`
	ProfileText string `json:"profile_text"`
}

// ImportLinkedIn handles POST /api/latest/linkedin/import.
func (h *LinkedInHandler) ImportLinkedIn(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req ImportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request"}`, http.StatusBadRequest)

		return
	}

	// Validate LinkedIn URL if provided
	if err := input.ValidateLinkedInURL(req.LinkedInURL); err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusBadRequest)

		return
	}

	// Extract user if authenticated
	var identityID *int
	if user := auth.GetUser(r.Context()); user != nil {
		if user.KratosID != "" {
			identity, err := h.repo.GetOrCreateIdentity(user.KratosID, user.Email)
			if err != nil {
				fmt.Printf("Failed to get/create identity: %v\n", err)
			} else {
				identityID = &identity.ID
			}
		}
	}

	// Create import record
	linkedinImport, err := h.repo.CreateLinkedInImport(identityID, req.LinkedInURL)
	if err != nil {
		http.Error(w, `{"error": "failed to create import"}`, http.StatusInternalServerError)

		return
	}

	// Parse LinkedIn profile
	profile, err := h.parser.ParseProfile(req.ProfileText)
	if err != nil {
		// Update import with error
		errorMsg := err.Error()
		_ = h.repo.UpdateLinkedInImport(linkedinImport.ID, "failed", nil, nil, &errorMsg)
		http.Error(w, fmt.Sprintf(`{"error": "failed to parse profile: %s"}`, err.Error()), http.StatusBadRequest)

		return
	}

	// Convert profile to CV format
	cvText := input.ConvertToCV(profile)

	// Store profile data and extracted CV
	profileJSON, _ := json.Marshal(profile)
	err = h.repo.UpdateLinkedInImport(
		linkedinImport.ID,
		"completed",
		(*json.RawMessage)(&profileJSON),
		&cvText,
		nil,
	)
	if err != nil {
		fmt.Printf("Failed to update import: %v\n", err)
	}

	// Prepare response
	response := map[string]interface{}{
		"id":            linkedinImport.ID,
		"linkedin_url":  req.LinkedInURL,
		"import_status": "completed",
		"extracted_cv":  cvText,
		"profile_data": map[string]interface{}{
			"name":       profile.Name,
			"title":      profile.Title,
			"summary":    profile.Summary,
			"experience": profile.Experience,
			"skills":     profile.Skills,
			"education":  profile.Education,
		},
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, `{"error": "failed to encode response"}`, http.StatusInternalServerError)
	}
}

// GetLinkedInImports handles GET /api/latest/linkedin/imports.
func (h *LinkedInHandler) GetLinkedInImports(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract user if authenticated
	var identityID *int
	if user := auth.GetUser(r.Context()); user != nil {
		if user.KratosID != "" {
			identity, err := h.repo.GetOrCreateIdentity(user.KratosID, user.Email)
			if err != nil {
				http.Error(w, `{"error": "authentication required"}`, http.StatusUnauthorized)

				return
			}
			identityID = &identity.ID
		}
	}

	if identityID == nil {
		http.Error(w, `{"error": "authentication required"}`, http.StatusUnauthorized)

		return
	}

	// Get imports for user
	imports, err := h.repo.GetLinkedInImports(*identityID)
	if err != nil {
		http.Error(w, `{"error": "failed to retrieve imports"}`, http.StatusInternalServerError)

		return
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(imports); err != nil {
		http.Error(w, `{"error": "failed to encode response"}`, http.StatusInternalServerError)
	}
}

// GetLinkedInImport handles GET /api/latest/linkedin/{import_id}.
func (h *LinkedInHandler) GetLinkedInImport(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	importIDStr := r.PathValue("import_id")
	importID, err := strconv.Atoi(importIDStr)
	if err != nil {
		http.Error(w, `{"error": "invalid import_id"}`, http.StatusBadRequest)

		return
	}

	// Get import
	linkedinImport, err := h.repo.GetLinkedInImport(importID)
	if err != nil {
		http.Error(w, `{"error": "import not found"}`, http.StatusNotFound)

		return
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(linkedinImport); err != nil {
		http.Error(w, `{"error": "failed to encode response"}`, http.StatusInternalServerError)
	}
}
