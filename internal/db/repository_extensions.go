// Copyright (c) Ultraviolet
// SPDX-License-Identifier: Apache-2.0

package db

import (
	"encoding/json"
	"time"
)

// CreateATSAnalysis creates a new ATS analysis record.
func (r *Repository) CreateATSAnalysis(cvVersionID int, overallScore *float64, keywordMatches *json.RawMessage, formattingIssues *json.RawMessage, sectionCompleteness *json.RawMessage, recommendations *json.RawMessage) (*ATSAnalysis, error) {
	var id int

	err := r.db.QueryRow(
		"INSERT INTO ats_analysis (cv_version_id, overall_score, keyword_matches, formatting_issues, section_completeness, recommendations) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id",
		cvVersionID, overallScore, keywordMatches, formattingIssues, sectionCompleteness, recommendations,
	).Scan(&id)
	if err != nil {
		return nil, err
	}

	return &ATSAnalysis{
		ID:                  id,
		CVVersionID:         cvVersionID,
		OverallScore:        overallScore,
		KeywordMatches:      keywordMatches,
		FormattingIssues:    formattingIssues,
		SectionCompleteness: sectionCompleteness,
		Recommendations:     recommendations,
		CreatedAt:           time.Now(),
	}, nil
}

// GetATSAnalysis retrieves an ATS analysis by CV version ID.
func (r *Repository) GetATSAnalysis(cvVersionID int) (*ATSAnalysis, error) {
	var analysis ATSAnalysis

	err := r.db.QueryRow(
		"SELECT id, cv_version_id, overall_score, keyword_matches, formatting_issues, section_completeness, recommendations, created_at FROM ats_analysis WHERE cv_version_id = $1",
		cvVersionID,
	).Scan(&analysis.ID, &analysis.CVVersionID, &analysis.OverallScore, &analysis.KeywordMatches, &analysis.FormattingIssues, &analysis.SectionCompleteness, &analysis.Recommendations, &analysis.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &analysis, nil
}

// CreateLinkedInImport creates a new LinkedIn import record.
func (r *Repository) CreateLinkedInImport(identityID *int, linkedinURL string) (*LinkedInImport, error) {
	var id int

	err := r.db.QueryRow(
		"INSERT INTO linkedin_imports (identity_id, linkedin_url, import_status) VALUES ($1, $2, 'pending') RETURNING id",
		identityID, linkedinURL,
	).Scan(&id)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	return &LinkedInImport{
		ID:           id,
		IdentityID:   identityID,
		LinkedInURL:  linkedinURL,
		ImportStatus: "pending",
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

// UpdateLinkedInImport updates a LinkedIn import record.
func (r *Repository) UpdateLinkedInImport(id int, status string, rawData *json.RawMessage, extractedCV *string, errorMessage *string) error {
	_, err := r.db.Exec(
		"UPDATE linkedin_imports SET import_status = $1, raw_data = $2, extracted_cv = $3, error_message = $4, updated_at = CURRENT_TIMESTAMP WHERE id = $5",
		status, rawData, extractedCV, errorMessage, id,
	)

	return err
}

// GetLinkedInImports retrieves all LinkedIn imports for an identity.
func (r *Repository) GetLinkedInImports(identityID int) ([]*LinkedInImport, error) {
	rows, err := r.db.Query(
		"SELECT id, identity_id, linkedin_url, raw_data, extracted_cv, import_status, error_message, created_at, updated_at FROM linkedin_imports WHERE identity_id = $1 ORDER BY created_at DESC",
		identityID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var imports []*LinkedInImport

	for rows.Next() {
		var imp LinkedInImport
		if err := rows.Scan(&imp.ID, &imp.IdentityID, &imp.LinkedInURL, &imp.RawData, &imp.ExtractedCV, &imp.ImportStatus, &imp.ErrorMessage, &imp.CreatedAt, &imp.UpdatedAt); err != nil {
			return nil, err
		}

		imports = append(imports, &imp)
	}

	return imports, rows.Err()
}

// GetLinkedInImport retrieves a specific LinkedIn import by ID.
func (r *Repository) GetLinkedInImport(id int) (*LinkedInImport, error) {
	var imp LinkedInImport

	err := r.db.QueryRow(
		"SELECT id, identity_id, linkedin_url, raw_data, extracted_cv, import_status, error_message, created_at, updated_at FROM linkedin_imports WHERE id = $1",
		id,
	).Scan(&imp.ID, &imp.IdentityID, &imp.LinkedInURL, &imp.RawData, &imp.ExtractedCV, &imp.ImportStatus, &imp.ErrorMessage, &imp.CreatedAt, &imp.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &imp, nil
}
