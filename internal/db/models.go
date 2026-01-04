// Copyright (c) Ultraviolet
// SPDX-License-Identifier: Apache-2.0

package db

import (
	"database/sql"
	"encoding/json"
	"time"
)

// Identity represents a user identity (optional, linked to Kratos).
type Identity struct {
	ID        int       `json:"id"`
	KratosID  *string   `json:"kratos_id"`
	Email     *string   `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CV represents an original CV document.
type CV struct {
	ID           int       `json:"id"`
	IdentityID   *int      `json:"identity_id"`
	OriginalText string    `json:"original_text"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// CVVersion represents a generated CV version.
type CVVersion struct {
	ID              int              `json:"id"`
	CVID            int              `json:"cv_id"`
	JobDescription  string           `json:"job_description"`
	CustomizedCV    string           `json:"customized_cv"`
	MatchScore      *float64         `json:"match_score"`
	AgentMetrics    *json.RawMessage `json:"agent_metrics"`
	WorkflowHistory *json.RawMessage `json:"workflow_history"`
	CreatedAt       time.Time        `json:"created_at"`
}

// BatchJob represents an async batch processing job.
type BatchJob struct {
	ID             int        `json:"id"`
	IdentityID     *int       `json:"identity_id"`
	Status         string     `json:"status"` // pending, processing, completed, failed
	TotalItems     int        `json:"total_items"`
	CompletedItems int        `json:"completed_items"`
	CreatedAt      time.Time  `json:"created_at"`
	CompletedAt    *time.Time `json:"completed_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// BatchJobItem represents an individual item in a batch job.
type BatchJobItem struct {
	ID             int              `json:"id"`
	BatchJobID     int              `json:"batch_job_id"`
	CVID           *int             `json:"cv_id"`
	JobDescription string           `json:"job_description"`
	Status         string           `json:"status"` // pending, processing, completed, failed
	Result         *json.RawMessage `json:"result"`
	ErrorMessage   *string          `json:"error_message"`
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`
}

// AnalyticsSnapshot represents a metrics snapshot.
type AnalyticsSnapshot struct {
	ID              int              `json:"id"`
	IdentityID      *int             `json:"identity_id"`
	MatchScore      *float64         `json:"match_score"`
	KeywordCoverage *float64         `json:"keyword_coverage"`
	Timestamp       time.Time        `json:"timestamp"`
	Metadata        *json.RawMessage `json:"metadata"`
}

// Repository defines database operations.
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// GetDB returns the underlying database connection.
func (r *Repository) GetDB() *sql.DB {
	return r.db
}

// CreateCV creates a new CV record.
func (r *Repository) CreateCV(identityID *int, originalText string) (*CV, error) {
	var id int

	err := r.db.QueryRow(
		"INSERT INTO cvs (identity_id, original_text) VALUES ($1, $2) RETURNING id",
		identityID, originalText,
	).Scan(&id)
	if err != nil {
		return nil, err
	}

	return &CV{
		ID:           id,
		IdentityID:   identityID,
		OriginalText: originalText,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}, nil
}

// GetCV retrieves a CV by ID.
func (r *Repository) GetCV(id int) (*CV, error) {
	var cv CV

	err := r.db.QueryRow(
		"SELECT id, identity_id, original_text, created_at, updated_at FROM cvs WHERE id = $1",
		id,
	).Scan(&cv.ID, &cv.IdentityID, &cv.OriginalText, &cv.CreatedAt, &cv.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &cv, nil
}

// CreateCVVersion creates a new CV version.
func (r *Repository) CreateCVVersion(cvID int, jobDescription string, customizedCV string, matchScore *float64, agentMetrics *json.RawMessage, workflowHistory *json.RawMessage) (*CVVersion, error) {
	var id int

	err := r.db.QueryRow(
		"INSERT INTO cv_versions (cv_id, job_description, customized_cv, match_score, agent_metrics_json, workflow_history_json) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id",
		cvID, jobDescription, customizedCV, matchScore, agentMetrics, workflowHistory,
	).Scan(&id)
	if err != nil {
		return nil, err
	}

	return &CVVersion{
		ID:              id,
		CVID:            cvID,
		JobDescription:  jobDescription,
		CustomizedCV:    customizedCV,
		MatchScore:      matchScore,
		AgentMetrics:    agentMetrics,
		WorkflowHistory: workflowHistory,
		CreatedAt:       time.Now(),
	}, nil
}

// GetCVVersions retrieves all versions for a CV.
func (r *Repository) GetCVVersions(cvID int) ([]*CVVersion, error) {
	rows, err := r.db.Query(
		"SELECT id, cv_id, job_description, customized_cv, match_score, agent_metrics_json, workflow_history_json, created_at FROM cv_versions WHERE cv_id = $1 ORDER BY created_at DESC",
		cvID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []*CVVersion

	for rows.Next() {
		var v CVVersion
		if err := rows.Scan(&v.ID, &v.CVID, &v.JobDescription, &v.CustomizedCV, &v.MatchScore, &v.AgentMetrics, &v.WorkflowHistory, &v.CreatedAt); err != nil {
			return nil, err
		}

		versions = append(versions, &v)
	}

	return versions, rows.Err()
}

// GetCVVersion retrieves a specific CV version.
func (r *Repository) GetCVVersion(id int) (*CVVersion, error) {
	var v CVVersion

	err := r.db.QueryRow(
		"SELECT id, cv_id, job_description, customized_cv, match_score, agent_metrics_json, workflow_history_json, created_at FROM cv_versions WHERE id = $1",
		id,
	).Scan(&v.ID, &v.CVID, &v.JobDescription, &v.CustomizedCV, &v.MatchScore, &v.AgentMetrics, &v.WorkflowHistory, &v.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &v, nil
}

// CreateBatchJob creates a new batch job.
func (r *Repository) CreateBatchJob(identityID *int, totalItems int) (*BatchJob, error) {
	var id int

	err := r.db.QueryRow(
		"INSERT INTO batch_jobs (identity_id, total_items, status) VALUES ($1, $2, 'pending') RETURNING id",
		identityID, totalItems,
	).Scan(&id)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	return &BatchJob{
		ID:             id,
		IdentityID:     identityID,
		Status:         "pending",
		TotalItems:     totalItems,
		CompletedItems: 0,
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}

// GetBatchJob retrieves a batch job.
func (r *Repository) GetBatchJob(id int) (*BatchJob, error) {
	var job BatchJob

	err := r.db.QueryRow(
		"SELECT id, identity_id, status, total_items, completed_items, created_at, completed_at, updated_at FROM batch_jobs WHERE id = $1",
		id,
	).Scan(&job.ID, &job.IdentityID, &job.Status, &job.TotalItems, &job.CompletedItems, &job.CreatedAt, &job.CompletedAt, &job.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &job, nil
}

// UpdateBatchJobStatus updates batch job status.
func (r *Repository) UpdateBatchJobStatus(id int, status string, completedItems int) error {
	query := "UPDATE batch_jobs SET status = $1, completed_items = $2, updated_at = CURRENT_TIMESTAMP"
	args := []any{status, completedItems}

	if status == "completed" || status == "failed" {
		query += ", completed_at = CURRENT_TIMESTAMP"
	}

	query += " WHERE id = $3"

	args = append(args, id)

	_, err := r.db.Exec(query, args...)

	return err
}

// CreateBatchJobItem creates a batch job item.
func (r *Repository) CreateBatchJobItem(batchJobID int, cvID *int, jobDescription string) (*BatchJobItem, error) {
	var id int

	err := r.db.QueryRow(
		"INSERT INTO batch_job_items (batch_job_id, cv_id, job_description, status) VALUES ($1, $2, $3, 'pending') RETURNING id",
		batchJobID, cvID, jobDescription,
	).Scan(&id)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	return &BatchJobItem{
		ID:             id,
		BatchJobID:     batchJobID,
		CVID:           cvID,
		JobDescription: jobDescription,
		Status:         "pending",
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}

// GetBatchJobItems retrieves all items for a batch job.
func (r *Repository) GetBatchJobItems(batchJobID int) ([]*BatchJobItem, error) {
	rows, err := r.db.Query(
		"SELECT id, batch_job_id, cv_id, job_description, status, result, error_message, created_at, updated_at FROM batch_job_items WHERE batch_job_id = $1",
		batchJobID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*BatchJobItem

	for rows.Next() {
		var item BatchJobItem
		if err := rows.Scan(&item.ID, &item.BatchJobID, &item.CVID, &item.JobDescription, &item.Status, &item.Result, &item.ErrorMessage, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}

		items = append(items, &item)
	}

	return items, rows.Err()
}

// UpdateBatchJobItem updates a batch job item.
func (r *Repository) UpdateBatchJobItem(id int, status string, result *json.RawMessage, errorMessage *string) error {
	_, err := r.db.Exec(
		"UPDATE batch_job_items SET status = $1, result = $2, error_message = $3, updated_at = CURRENT_TIMESTAMP WHERE id = $4",
		status, result, errorMessage, id,
	)

	return err
}

// RecordAnalyticsSnapshot records an analytics metric.
func (r *Repository) RecordAnalyticsSnapshot(identityID *int, matchScore *float64, keywordCoverage *float64, metadata *json.RawMessage) error {
	_, err := r.db.Exec(
		"INSERT INTO analytics_snapshots (identity_id, match_score, keyword_coverage, metadata) VALUES ($1, $2, $3, $4)",
		identityID, matchScore, keywordCoverage, metadata,
	)

	return err
}

// GetAnalyticsStats retrieves analytics statistics.
func (r *Repository) GetAnalyticsStats(identityID *int, limit int) ([]*AnalyticsSnapshot, error) {
	query := "SELECT id, identity_id, match_score, keyword_coverage, timestamp, metadata FROM analytics_snapshots"

	var args []any

	if identityID != nil {
		query += " WHERE identity_id = $1"

		args = append(args, identityID)

		if limit > 0 {
			query += " ORDER BY timestamp DESC LIMIT $2"

			args = append(args, limit)
		} else {
			query += " ORDER BY timestamp DESC"
		}
	} else {
		if limit > 0 {
			query += " ORDER BY timestamp DESC LIMIT $1"

			args = append(args, limit)
		} else {
			query += " ORDER BY timestamp DESC"
		}
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var snapshots []*AnalyticsSnapshot

	for rows.Next() {
		var s AnalyticsSnapshot
		if err := rows.Scan(&s.ID, &s.IdentityID, &s.MatchScore, &s.KeywordCoverage, &s.Timestamp, &s.Metadata); err != nil {
			return nil, err
		}

		snapshots = append(snapshots, &s)
	}

	return snapshots, rows.Err()
}
