// Copyright (c) Ultraviolet
// SPDX-License-Identifier: Apache-2.0

package observability

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"
)

// Metrics tracks application metrics.
type Metrics struct {
	mu sync.RWMutex

	// Request metrics
	RequestCount    int64
	RequestErrors   int64
	RequestDuration int64 // total milliseconds
	AverageDuration float64
	TotalRequests   int64

	// LLM metrics
	LLMCallCount      int64
	LLMErrorCount     int64
	LLMTotalDuration  int64 // milliseconds
	AverageLLMLatency float64

	// Database metrics
	DBQueryCount    int64
	DBErrorCount    int64
	DBTotalDuration int64 // milliseconds

	// Resource metrics
	MemoryAllocated  uint64
	MemoryTotalAlloc uint64
	Goroutines       int
	StartTime        time.Time

	// Batch processing
	BatchJobsProcessed  int64
	BatchItemsProcessed int64
	BatchErrors         int64
}

// NewMetrics creates a new metrics tracker.
func NewMetrics() *Metrics {
	return &Metrics{
		StartTime: time.Now(),
	}
}

// RecordRequest records a completed request.
func (m *Metrics) RecordRequest(durationMs int64, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.RequestCount++
	m.TotalRequests++
	m.RequestDuration += durationMs

	if err != nil {
		m.RequestErrors++
	}

	if m.RequestCount > 0 {
		m.AverageDuration = float64(m.RequestDuration) / float64(m.RequestCount)
	}
}

// RecordLLMCall records an LLM API call.
func (m *Metrics) RecordLLMCall(durationMs int64, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.LLMCallCount++
	m.LLMTotalDuration += durationMs

	if err != nil {
		m.LLMErrorCount++
	}

	if m.LLMCallCount > 0 {
		m.AverageLLMLatency = float64(m.LLMTotalDuration) / float64(m.LLMCallCount)
	}
}

// RecordDBQuery records a database query.
func (m *Metrics) RecordDBQuery(durationMs int64, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.DBQueryCount++
	m.DBTotalDuration += durationMs

	if err != nil {
		m.DBErrorCount++
	}
}

// RecordBatchJob records a completed batch job.
func (m *Metrics) RecordBatchJob(itemsProcessed int64, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.BatchJobsProcessed++
	m.BatchItemsProcessed += itemsProcessed

	if err != nil {
		m.BatchErrors++
	}
}

// GetMetrics returns a snapshot of current metrics.
func (m *Metrics) GetMetrics() map[string]any {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	uptime := time.Since(m.StartTime).Seconds()

	return map[string]any{
		"uptime_seconds": uptime,
		"requests": map[string]any{
			"total":               m.TotalRequests,
			"errors":              m.RequestErrors,
			"average_duration_ms": m.AverageDuration,
		},
		"llm": map[string]any{
			"calls":              m.LLMCallCount,
			"errors":             m.LLMErrorCount,
			"average_latency_ms": m.AverageLLMLatency,
		},
		"database": map[string]any{
			"queries":           m.DBQueryCount,
			"errors":            m.DBErrorCount,
			"total_duration_ms": m.DBTotalDuration,
		},
		"batch": map[string]any{
			"jobs_processed":  m.BatchJobsProcessed,
			"items_processed": m.BatchItemsProcessed,
			"errors":          m.BatchErrors,
		},
		"memory": map[string]any{
			"allocated_mb":   float64(ms.Alloc) / 1024 / 1024,
			"total_alloc_mb": float64(ms.TotalAlloc) / 1024 / 1024,
			"sys_mb":         float64(ms.Sys) / 1024 / 1024,
		},
		"goroutines": runtime.NumGoroutine(),
	}
}

// HealthCheck provides detailed health status.
type HealthCheck struct {
	mu      sync.RWMutex
	checks  map[string]ComponentHealth
	metrics *Metrics
}

// ComponentHealth represents the health of a component.
type ComponentHealth struct {
	Status    string    `json:"status"` // "healthy", "degraded", "unhealthy"
	Message   string    `json:"message"`
	LastCheck time.Time `json:"last_check"`
	Details   any       `json:"details,omitempty"`
}

// NewHealthCheck creates a new health checker.
func NewHealthCheck(metrics *Metrics) *HealthCheck {
	return &HealthCheck{
		checks:  make(map[string]ComponentHealth),
		metrics: metrics,
	}
}

// RegisterCheck registers a health check for a component.
func (hc *HealthCheck) RegisterCheck(name string, status string, message string) {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	hc.checks[name] = ComponentHealth{
		Status:    status,
		Message:   message,
		LastCheck: time.Now(),
	}
}

// UpdateCheck updates a health check result.
func (hc *HealthCheck) UpdateCheck(name string, status string, message string, details any) {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	hc.checks[name] = ComponentHealth{
		Status:    status,
		Message:   message,
		LastCheck: time.Now(),
		Details:   details,
	}
}

// GetHealth returns the overall health status.
func (hc *HealthCheck) GetHealth() map[string]any {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	// Determine overall status
	overallStatus := "healthy"

	for _, check := range hc.checks {
		if check.Status == "unhealthy" {
			overallStatus = "unhealthy"

			break
		}

		if check.Status == "degraded" && overallStatus != "unhealthy" {
			overallStatus = "degraded"
		}
	}

	return map[string]any{
		"status":         overallStatus,
		"timestamp":      time.Now().UTC().Format(time.RFC3339),
		"components":     hc.checks,
		"uptime_seconds": time.Since(hc.metrics.StartTime).Seconds(),
	}
}

// HTTPHandlers for observability endpoints

// HealthCheckHandler returns health status.
func HealthCheckHandler(hc *HealthCheck) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		health := hc.GetHealth()

		w.Header().Set("Content-Type", "application/json")

		// Set status code based on health
		status := http.StatusOK
		switch health["status"] {
		case "degraded":
			status = http.StatusPartialContent
		case "unhealthy":
			status = http.StatusServiceUnavailable
		}

		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(health)
	}
}

// MetricsHandler returns current metrics.
func MetricsHandler(metrics *Metrics) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(metrics.GetMetrics())
	}
}

// PrometheusMetricsHandler returns metrics in Prometheus format.
func PrometheusMetricsHandler(metrics *Metrics) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		m := metrics.GetMetrics()

		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		w.WriteHeader(http.StatusOK)

		// Write Prometheus format metrics
		requestMetrics := m["requests"].(map[string]any)
		llmMetrics := m["llm"].(map[string]any)
		batchMetrics := m["batch"].(map[string]any)
		memoryMetrics := m["memory"].(map[string]any)

		fmt.Fprintf(w, "# HELP vibe_cv_requests_total Total number of API requests\n")
		fmt.Fprintf(w, "# TYPE vibe_cv_requests_total counter\n")
		fmt.Fprintf(w, "vibe_cv_requests_total %v\n\n", requestMetrics["total"])

		fmt.Fprintf(w, "# HELP vibe_cv_request_errors_total Total number of request errors\n")
		fmt.Fprintf(w, "# TYPE vibe_cv_request_errors_total counter\n")
		fmt.Fprintf(w, "vibe_cv_request_errors_total %v\n\n", requestMetrics["errors"])

		fmt.Fprintf(w, "# HELP vibe_cv_request_duration_ms Average request duration in milliseconds\n")
		fmt.Fprintf(w, "# TYPE vibe_cv_request_duration_ms gauge\n")
		fmt.Fprintf(w, "vibe_cv_request_duration_ms %v\n\n", requestMetrics["average_duration_ms"])

		fmt.Fprintf(w, "# HELP vibe_cv_llm_calls_total Total LLM API calls\n")
		fmt.Fprintf(w, "# TYPE vibe_cv_llm_calls_total counter\n")
		fmt.Fprintf(w, "vibe_cv_llm_calls_total %v\n\n", llmMetrics["calls"])

		fmt.Fprintf(w, "# HELP vibe_cv_llm_latency_ms Average LLM latency in milliseconds\n")
		fmt.Fprintf(w, "# TYPE vibe_cv_llm_latency_ms gauge\n")
		fmt.Fprintf(w, "vibe_cv_llm_latency_ms %v\n\n", llmMetrics["average_latency_ms"])

		fmt.Fprintf(w, "# HELP vibe_cv_batch_items_processed Total batch items processed\n")
		fmt.Fprintf(w, "# TYPE vibe_cv_batch_items_processed counter\n")
		fmt.Fprintf(w, "vibe_cv_batch_items_processed %v\n\n", batchMetrics["items_processed"])

		fmt.Fprintf(w, "# HELP vibe_cv_memory_allocated_mb Current memory allocation in MB\n")
		fmt.Fprintf(w, "# TYPE vibe_cv_memory_allocated_mb gauge\n")
		fmt.Fprintf(w, "vibe_cv_memory_allocated_mb %v\n\n", memoryMetrics["allocated_mb"])

		fmt.Fprintf(w, "# HELP vibe_cv_goroutines Current number of goroutines\n")
		fmt.Fprintf(w, "# TYPE vibe_cv_goroutines gauge\n")
		fmt.Fprintf(w, "vibe_cv_goroutines %v\n", m["goroutines"])
	}
}
