// Copyright (c) Ultraviolet
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/sammyoina/vibe-cv/api"
	"github.com/sammyoina/vibe-cv/internal/config"
	"github.com/sammyoina/vibe-cv/internal/db"
	"github.com/sammyoina/vibe-cv/internal/llm"
	"github.com/sammyoina/vibe-cv/internal/observability"
	"github.com/sammyoina/vibe-cv/internal/security"
	"github.com/sammyoina/vibe-cv/pkg/auth"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	var (
		repo       *db.Repository
		authConfig *auth.Config
	)

	if cfg.DatabaseURL != "" {
		dbConn, err := db.InitDB(cfg.DatabaseURL)
		if err != nil {
			log.Fatalf("Failed to initialize database: %v", err)
		}
		defer db.Close(dbConn)

		repo = db.NewRepository(dbConn)

		log.Println("Database initialized successfully")
	}

	authConfig = auth.LoadConfig()
	if authConfig.Enabled {
		log.Println("Authentication enabled via Kratos")
	} else {
		log.Println("Authentication disabled - all endpoints are public")
	}

	factory := llm.NewFactory()
	mux := http.NewServeMux()

	provider, err := factory.CreateProvider(context.TODO(), cfg.LLMProvider, cfg.LLMAPIKey, cfg.LLMModel)
	if err != nil {
		log.Fatalf("Failed to create LLM provider: %v", err)
	}

	validator := security.NewRequestValidator(10 * 1024 * 1024) // 10MB max request size
	rateLimiter := security.NewRateLimiter(100.0, 200)          // 100 req/s with burst of 200
	corsConfig := security.DefaultCORSConfig()
	auditLogger := security.NewAuditLogger()

	metrics := observability.NewMetrics()
	healthCheck := observability.NewHealthCheck(metrics)
	healthCheck.UpdateCheck("api", "healthy", "API is operational", nil)

	if repo != nil {
		healthCheck.UpdateCheck("database", "healthy", "Database connection established", nil)
	} else {
		healthCheck.UpdateCheck("database", "degraded", "Database not configured", nil)
	}

	outputDir := cfg.OutputDir
	if outputDir == "" {
		outputDir = "./outputs"
	}

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		log.Fatalf("Failed to create output directory %s: %v", outputDir, err)
	}

	testFile := outputDir + "/.write-test"
	if err := os.WriteFile(testFile, []byte("test"), 0o644); err != nil {
		log.Fatalf("Output directory %s is not writable: %v", outputDir, err)
	}

	os.Remove(testFile)
	log.Printf("Output directory validated: %s", outputDir)

	latestHandler := api.NewLatestHandler(provider, repo, authConfig)

	if repo != nil {
		latestHandler.StartQueue()
		log.Println("Batch job queue started with 4 workers")
	}

	authMiddleware := auth.Middleware(authConfig)
	latestMux := http.NewServeMux()
	latestHandler.RegisterRoutes(latestMux)

	var apiHandler http.Handler = latestMux

	// Apply middlewares in reverse order (they wrap each other)
	apiHandler = security.AuditMiddleware(auditLogger)(apiHandler)
	apiHandler = security.CORSMiddleware(corsConfig)(apiHandler)
	apiHandler = security.RateLimitMiddleware(rateLimiter)(apiHandler)
	apiHandler = security.ValidationMiddleware(validator)(apiHandler)
	apiHandler = authMiddleware(apiHandler)

	mux.Handle("/api/latest/", apiHandler)

	// Observability endpoints (no auth required for monitoring)
	mux.HandleFunc("GET /api/health", observability.HealthCheckHandler(healthCheck))
	mux.HandleFunc("GET /api/metrics", observability.MetricsHandler(metrics))
	mux.HandleFunc("GET /metrics", observability.PrometheusMetricsHandler(metrics)) // Prometheus scrape endpoint
	mux.HandleFunc("GET /api/audit-logs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		count := 100
		if countParam := r.URL.Query().Get("count"); countParam != "" {
			if _, err := fmt.Sscanf(countParam, "%d", &count); err != nil {
				count = 100
			}
		}

		logs := auditLogger.GetRecentLogs(count)

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(logs)
	})

	addr := fmt.Sprintf("%s:%s", cfg.ServerHost, cfg.ServerPort)
	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		log.Printf("Starting vibe-cv server on %s", addr)
		log.Println("Available endpoints at /api/latest/*:")
		log.Println("  POST   /api/latest/customize-cv          - Customize CV for job description")
		log.Println("  POST   /api/latest/batch-customize        - Submit batch customization jobs")
		log.Println("  GET    /api/latest/versions/{cv_id}       - List CV versions")
		log.Println("  GET    /api/latest/versions/{id}/detail   - Get version details")
		log.Println("  POST   /api/latest/compare-versions       - Compare two versions")
		log.Println("  GET    /api/latest/analytics              - Get user analytics")
		log.Println("  GET    /api/latest/dashboard              - Get global dashboard")
		log.Println("  GET    /api/latest/batch/{job_id}/status  - Check batch job status")
		log.Println("  GET    /api/latest/batch/{job_id}/download- Download batch results")
		log.Println("  GET    /api/latest/health                 - Health check")

		if repo == nil {
			log.Println("\nWARNING: Database not configured - persistence features unavailable")
			log.Println("Set DATABASE_URL environment variable to enable full Phase 4 features")
		}

		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server error: %v", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down server...")

	if repo != nil {
		latestHandler.StopQueue()
		log.Println("Batch job queue stopped")
	}

	if err := server.Close(); err != nil {
		log.Fatalf("Server shutdown error: %v", err)
	}
}
