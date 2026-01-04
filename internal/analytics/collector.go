// Copyright (c) Ultraviolet
// SPDX-License-Identifier: Apache-2.0

package analytics

import (
	"sync"
	"time"

	"github.com/sammyoina/vibe-cv/internal/db"
)

// Collector collects analytics metrics.
type Collector struct {
	mu   sync.RWMutex
	repo *db.Repository
}

// NewCollector creates a new analytics collector.
func NewCollector(repo *db.Repository) *Collector {
	return &Collector{
		repo: repo,
	}
}

// AnalyticsData represents analytics information.
type AnalyticsData struct {
	AverageMatchScore      float64                 `json:"average_match_score"`
	AverageKeywordCoverage float64                 `json:"average_keyword_coverage"`
	TotalCustomizations    int                     `json:"total_customizations"`
	TimeRange              string                  `json:"time_range"`
	MatchScoreDistribution map[string]int          `json:"match_score_distribution"`
	RecentSnapshots        []*db.AnalyticsSnapshot `json:"recent_snapshots"`
}

// GetAnalytics retrieves analytics data for an identity.
func (c *Collector) GetAnalytics(identityID *int, limit int) (*AnalyticsData, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	snapshots, err := c.repo.GetAnalyticsStats(identityID, limit)
	if err != nil {
		return nil, err
	}

	analytics := &AnalyticsData{
		TimeRange:              "last_entries",
		MatchScoreDistribution: make(map[string]int),
		RecentSnapshots:        snapshots,
	}

	if len(snapshots) == 0 {
		return analytics, nil
	}

	totalScore := 0.0
	totalCoverage := 0.0
	scoreCount := 0
	coverageCount := 0

	for _, snapshot := range snapshots {
		if snapshot.MatchScore != nil {
			totalScore += *snapshot.MatchScore
			scoreCount++

			// Distribute into buckets
			bucket := distributeBucket(*snapshot.MatchScore)
			analytics.MatchScoreDistribution[bucket]++
		}

		if snapshot.KeywordCoverage != nil {
			totalCoverage += *snapshot.KeywordCoverage
			coverageCount++
		}
	}

	if scoreCount > 0 {
		analytics.AverageMatchScore = totalScore / float64(scoreCount)
	}

	if coverageCount > 0 {
		analytics.AverageKeywordCoverage = totalCoverage / float64(coverageCount)
	}

	analytics.TotalCustomizations = len(snapshots)

	return analytics, nil
}

// distributeBucket places a match score into a bucket.
func distributeBucket(score float64) string {
	if score >= 0.9 {
		return "0.9-1.0"
	} else if score >= 0.8 {
		return "0.8-0.9"
	} else if score >= 0.7 {
		return "0.7-0.8"
	} else if score >= 0.6 {
		return "0.6-0.7"
	}

	return "0.0-0.6"
}

// AnalyticsDashboard represents dashboard statistics.
type AnalyticsDashboard struct {
	TotalUsers          int                     `json:"total_users"`
	TotalCustomizations int                     `json:"total_customizations"`
	AverageMatchScore   float64                 `json:"average_match_score"`
	MatchScoreTrend     []DailyMetric           `json:"match_score_trend"`
	TopKeywords         []KeywordMetric         `json:"top_keywords"`
	RecentActivities    []*db.AnalyticsSnapshot `json:"recent_activities"`
}

// DailyMetric represents a daily metric.
type DailyMetric struct {
	Date       time.Time `json:"date"`
	MatchScore float64   `json:"match_score"`
	Count      int       `json:"count"`
}

// KeywordMetric represents keyword frequency.
type KeywordMetric struct {
	Keyword      string  `json:"keyword"`
	Count        int     `json:"count"`
	AverageScore float64 `json:"average_score"`
}

// GetDashboard retrieves dashboard statistics.
func (c *Collector) GetDashboard() (*AnalyticsDashboard, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Get recent snapshots (global, not user-scoped)
	snapshots, err := c.repo.GetAnalyticsStats(nil, 100)
	if err != nil {
		return nil, err
	}

	dashboard := &AnalyticsDashboard{
		MatchScoreTrend:  make([]DailyMetric, 0),
		TopKeywords:      make([]KeywordMetric, 0),
		RecentActivities: snapshots,
	}

	// Calculate aggregate statistics
	totalScore := 0.0
	scoreCount := 0

	for _, snapshot := range snapshots {
		if snapshot.MatchScore != nil {
			totalScore += *snapshot.MatchScore
			scoreCount++
		}
	}

	if scoreCount > 0 {
		dashboard.AverageMatchScore = totalScore / float64(scoreCount)
	}

	dashboard.TotalCustomizations = len(snapshots)

	return dashboard, nil
}
