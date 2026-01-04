// Copyright (c) Ultraviolet
// SPDX-License-Identifier: Apache-2.0

package db

import (
	"database/sql"
	"fmt"
	"time"

	// Import postgres driver for database/sql.
	_ "github.com/lib/pq"
	migrate "github.com/rubenv/sql-migrate"
)

// InitDB initializes the database connection and runs migrations
// with optimized connection pooling configuration.
func InitDB(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pooling for production use
	// MaxOpenConns: Maximum number of open connections to the database
	// The appropriate value depends on your workload and database configuration
	// A good starting point is the number of CPU cores * 4
	db.SetMaxOpenConns(32)

	// MinOpenConns: Minimum number of connections to maintain
	db.SetMaxIdleConns(8)

	// ConnMaxLifetime: Maximum amount of time a connection may be reused
	// This helps prevent stale connections and database session timeouts
	db.SetConnMaxLifetime(5 * time.Minute)

	// ConnMaxIdleTime: Maximum amount of time a connection may be idle
	db.SetConnMaxIdleTime(2 * time.Minute)

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Run migrations
	migrations := Migrations()

	n, err := migrate.ExecMax(db, "postgres", migrations, migrate.Up, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	fmt.Printf("Applied %d migrations\n", n)

	// Create indexes for performance optimization
	if err := createOptimizationIndexes(db); err != nil {
		fmt.Printf("Warning: Failed to create optimization indexes: %v\n", err)
		// Don't fail on index creation; they may already exist
	}

	return db, nil
}

// createOptimizationIndexes creates indexes for better query performance.
func createOptimizationIndexes(db *sql.DB) error {
	indexes := []string{
		// Index on cv_versions.created_at for time-range queries
		`CREATE INDEX IF NOT EXISTS idx_cv_versions_created_at 
		 ON cv_versions(created_at DESC)`,

		// Index on batch_jobs.status for filtering
		`CREATE INDEX IF NOT EXISTS idx_batch_jobs_status 
		 ON batch_jobs(status)`,

		// Index on batch_jobs.created_at for recent jobs
		`CREATE INDEX IF NOT EXISTS idx_batch_jobs_created_at 
		 ON batch_jobs(created_at DESC)`,

		// Composite index for batch items
		`CREATE INDEX IF NOT EXISTS idx_batch_job_items_job_status 
		 ON batch_job_items(batch_job_id, status)`,

		// Index on analytics snapshots
		`CREATE INDEX IF NOT EXISTS idx_analytics_snapshots_timestamp 
		 ON analytics_snapshots(timestamp DESC)`,

		// Index on identities for lookups
		`CREATE INDEX IF NOT EXISTS idx_identities_kratos_id 
		 ON identities(kratos_id) WHERE kratos_id IS NOT NULL`,

		// Index on identities email
		`CREATE INDEX IF NOT EXISTS idx_identities_email 
		 ON identities(email) WHERE email IS NOT NULL`,

		// Composite index for efficient CV version queries
		`CREATE INDEX IF NOT EXISTS idx_cv_versions_cv_created 
		 ON cv_versions(cv_id, created_at DESC)`,
	}

	for _, indexSQL := range indexes {
		if _, err := db.Exec(indexSQL); err != nil {
			// Log but don't fail - indexes may already exist
			fmt.Printf("Index creation note: %v\n", err)
		}
	}

	return nil
}

// Close closes the database connection.
func Close(db *sql.DB) error {
	return db.Close()
}
