// Copyright (c) Ultraviolet
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/sammyoina/vibe-cv/sdk"
)

func main() {
	// Get the server URL from environment or use default
	serverURL := os.Getenv("VIBE_CV_URL")
	if serverURL == "" {
		serverURL = "http://localhost:8080"
	}

	// Get auth token from environment
	authToken := os.Getenv("VIBE_CV_TOKEN")
	if authToken == "" {
		log.Fatal("VIBE_CV_TOKEN environment variable is required")
	}

	// Create a new SDK client with custom options
	client := sdk.NewClient(
		serverURL,
		sdk.WithUserAgent("my-app/1.0.0"),
	)

	ctx := context.Background()

	// 1. Health Check (no auth required)
	fmt.Println("=== Health Check ===")
	if err := client.Ping(ctx); err != nil {
		log.Fatalf("Service is not healthy: %v", err)
	}
	fmt.Println("✅ Service is healthy")

	// 2. Customize CV with different LLM providers
	fmt.Println("=== CV Customization with Different LLMs ===")

	cvContent := "Software Engineer with 5 years of experience in Go and cloud technologies..."
	jobDesc := "Senior Backend Engineer - Go, Kubernetes, AWS"

	// Try with OpenAI (requires authentication)
	fmt.Println("Customizing with OpenAI GPT-4...")
	req := &sdk.CustomizeCVRequest{
		CV:             cvContent,
		JobDescription: jobDesc,
		LLMConfig: &sdk.LLMConfig{
			Provider: sdk.ProviderOpenAI,
			Model:    sdk.ModelGPT4,
		},
		AdditionalContext: []sdk.ContextItem{
			{Type: "text", Content: "Specialized in microservices architecture"},
			{Type: "text", Content: "Led team of 5 engineers"},
		},
	}

	resp, err := client.CustomizeCV(ctx, req, sdk.WithRequestAuthToken(authToken))
	if err != nil {
		// Demonstrate error handling
		if apiErr, ok := err.(*sdk.APIError); ok {
			fmt.Printf("API Error: %s (status: %d)\n", apiErr.Message, apiErr.StatusCode)
			if apiErr.IsUnauthorized() {
				fmt.Println("Authentication required!")
			}
		} else {
			fmt.Printf("Error: %v\n", err)
		}
	} else {
		fmt.Printf("✅ Match Score: %.2f\n", resp.MatchScore)
		fmt.Printf("Modifications: %d changes made\n", len(resp.Modifications))
	}

	// 3. Version Management
	fmt.Println("=== Version Management ===")

	// Assume we have a CV ID from previous customization
	cvID := 1

	versions, err := client.GetVersions(ctx, cvID, sdk.WithRequestAuthToken(authToken))
	if err != nil {
		fmt.Printf("Could not fetch versions: %v\n", err)
	} else {
		fmt.Printf("Found %d versions for CV %d\n", len(versions), cvID)

		if len(versions) >= 2 {
			// Compare two versions
			comparison, err := client.CompareVersions(ctx, versions[0].ID, versions[1].ID, sdk.WithRequestAuthToken(authToken))
			if err != nil {
				fmt.Printf("Could not compare versions: %v\n", err)
			} else {
				fmt.Printf("Comparing versions %d and %d\n", versions[0].ID, versions[1].ID)
				if comparison.MatchScoreDiff != nil {
					fmt.Printf("Match score difference: %.2f\n", *comparison.MatchScoreDiff)
				}
			}
		}
	}
	fmt.Println()

	// 4. Download CV
	fmt.Println("=== Download CV ===")
	versionID := 1
	pdfData, err := client.DownloadCV(ctx, versionID, sdk.WithRequestAuthToken(authToken))
	if err != nil {
		fmt.Printf("Could not download CV: %v\n", err)
	} else {
		filename := fmt.Sprintf("cv-version-%d.pdf", versionID)
		if err := os.WriteFile(filename, pdfData, 0644); err != nil {
			fmt.Printf("Could not save PDF: %v\n", err)
		} else {
			fmt.Printf("✅ Downloaded CV to %s (%d bytes)\n", filename, len(pdfData))
		}
	}
	fmt.Println()

	// 5. Analytics
	fmt.Println("=== Analytics ===")
	analytics, err := client.GetAnalytics(ctx, 10, sdk.WithRequestAuthToken(authToken))
	if err != nil {
		fmt.Printf("Could not fetch analytics: %v\n", err)
	} else {
		fmt.Printf("Retrieved %d analytics snapshots\n", len(analytics.Snapshots))
		for i, snapshot := range analytics.Snapshots {
			if snapshot.MatchScore != nil {
				fmt.Printf("  %d. Match Score: %.2f at %s\n",
					i+1, *snapshot.MatchScore, snapshot.Timestamp.Format("2006-01-02 15:04"))
			}
		}
	}
	fmt.Println()

	// 6. Dashboard
	fmt.Println("=== Dashboard Statistics ===")
	dashboard, err := client.GetDashboard(ctx, sdk.WithRequestAuthToken(authToken))
	if err != nil {
		fmt.Printf("Could not fetch dashboard: %v\n", err)
	} else {
		fmt.Printf("Total CVs: %d\n", dashboard.TotalCVs)
		fmt.Printf("Total Versions: %d\n", dashboard.TotalVersions)
		fmt.Printf("Total Batch Jobs: %d\n", dashboard.TotalBatchJobs)
		fmt.Printf("Average Match Score: %.2f\n", dashboard.AverageMatchScore)
		fmt.Printf("Active Users: %d\n", dashboard.ActiveUsers)
	}

	fmt.Println("\n✅ Advanced features demonstration complete!")
}
