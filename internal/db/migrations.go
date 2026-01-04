package db

import migrate "github.com/rubenv/sql-migrate"

// Migrations returns the in-code migration source
func Migrations() *migrate.MemoryMigrationSource {
	return &migrate.MemoryMigrationSource{
		Migrations: []*migrate.Migration{
			{
				Id: "001_init_schema",
				Up: []string{`
					CREATE TABLE IF NOT EXISTS identities (
						id SERIAL PRIMARY KEY,
						kratos_id VARCHAR(255) UNIQUE,
						email VARCHAR(255),
						created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
						updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
					);

					CREATE TABLE IF NOT EXISTS cvs (
						id SERIAL PRIMARY KEY,
						identity_id INTEGER REFERENCES identities(id) ON DELETE SET NULL,
						original_text TEXT NOT NULL,
						created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
						updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
					);

					CREATE TABLE IF NOT EXISTS cv_versions (
						id SERIAL PRIMARY KEY,
						cv_id INTEGER NOT NULL REFERENCES cvs(id) ON DELETE CASCADE,
						job_description TEXT NOT NULL,
						customized_cv TEXT NOT NULL,
						match_score FLOAT,
						agent_metrics_json JSONB,
						workflow_history_json JSONB,
						created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
					);

					CREATE TABLE IF NOT EXISTS batch_jobs (
						id SERIAL PRIMARY KEY,
						identity_id INTEGER REFERENCES identities(id) ON DELETE SET NULL,
						status VARCHAR(50) DEFAULT 'pending',
						total_items INTEGER DEFAULT 0,
						completed_items INTEGER DEFAULT 0,
						created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
						completed_at TIMESTAMP,
						updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
					);

					CREATE TABLE IF NOT EXISTS batch_job_items (
						id SERIAL PRIMARY KEY,
						batch_job_id INTEGER NOT NULL REFERENCES batch_jobs(id) ON DELETE CASCADE,
						cv_id INTEGER REFERENCES cvs(id) ON DELETE SET NULL,
						job_description TEXT NOT NULL,
						status VARCHAR(50) DEFAULT 'pending',
						result_json JSONB,
						error_message TEXT,
						created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
						updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
					);

					CREATE TABLE IF NOT EXISTS analytics_snapshots (
						id SERIAL PRIMARY KEY,
						identity_id INTEGER REFERENCES identities(id) ON DELETE SET NULL,
						match_score FLOAT,
						keyword_coverage FLOAT,
						timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
						metadata_json JSONB
					);

					CREATE INDEX IF NOT EXISTS idx_cvs_identity ON cvs(identity_id);
					CREATE INDEX IF NOT EXISTS idx_cv_versions_cv ON cv_versions(cv_id);
					CREATE INDEX IF NOT EXISTS idx_batch_jobs_identity ON batch_jobs(identity_id);
					CREATE INDEX IF NOT EXISTS idx_batch_job_items_batch ON batch_job_items(batch_job_id);
					CREATE INDEX IF NOT EXISTS idx_analytics_identity ON analytics_snapshots(identity_id);
				`},
				Down: []string{`
					DROP TABLE IF EXISTS analytics_snapshots;
					DROP TABLE IF EXISTS batch_job_items;
					DROP TABLE IF EXISTS batch_jobs;
					DROP TABLE IF EXISTS cv_versions;
					DROP TABLE IF EXISTS cvs;
					DROP TABLE IF EXISTS identities;
				`},
			},
			{
				Id: "002_phase5_performance_indexes",
				Up: []string{`
					-- Phase 5: Performance optimization indexes

					-- Index on cv_versions.created_at for time-range analytics queries
					CREATE INDEX IF NOT EXISTS idx_cv_versions_created_at 
					ON cv_versions(created_at DESC);

					-- Index on batch_jobs.status for filtering by job status
					CREATE INDEX IF NOT EXISTS idx_batch_jobs_status 
					ON batch_jobs(status);

					-- Index on batch_jobs.created_at for recent jobs queries
					CREATE INDEX IF NOT EXISTS idx_batch_jobs_created_at 
					ON batch_jobs(created_at DESC);

					-- Composite index for batch job items by job and status
					CREATE INDEX IF NOT EXISTS idx_batch_job_items_job_status 
					ON batch_job_items(batch_job_id, status);

					-- Index on analytics_snapshots.timestamp for time-series queries
					CREATE INDEX IF NOT EXISTS idx_analytics_snapshots_timestamp 
					ON analytics_snapshots(timestamp DESC);

					-- Index on identities for quick kratos_id lookups
					CREATE INDEX IF NOT EXISTS idx_identities_kratos_id 
					ON identities(kratos_id) WHERE kratos_id IS NOT NULL;

					-- Index on identities email for user lookups
					CREATE INDEX IF NOT EXISTS idx_identities_email 
					ON identities(email) WHERE email IS NOT NULL;

					-- Composite index for efficient CV version queries
					CREATE INDEX IF NOT EXISTS idx_cv_versions_cv_created 
					ON cv_versions(cv_id, created_at DESC);
				`},
				Down: []string{`
					DROP INDEX IF EXISTS idx_cv_versions_created_at;
					DROP INDEX IF EXISTS idx_batch_jobs_status;
					DROP INDEX IF EXISTS idx_batch_jobs_created_at;
					DROP INDEX IF EXISTS idx_batch_job_items_job_status;
					DROP INDEX IF EXISTS idx_analytics_snapshots_timestamp;
					DROP INDEX IF EXISTS idx_identities_kratos_id;
					DROP INDEX IF EXISTS idx_identities_email;
					DROP INDEX IF EXISTS idx_cv_versions_cv_created;
				`},
			},
		},
	}
}
