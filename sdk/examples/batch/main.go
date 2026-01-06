// Copyright (c) Ultraviolet
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/sammyoina/vibe-cv/sdk"
)

func main() {
	// Get the server URL from environment or use default
	serverURL := os.Getenv("VIBE_CV_URL")
	if serverURL == "" {
		serverURL = "http://localhost:8080"
	}

	// Create a new SDK client
	client := sdk.NewClient(serverURL)

	// Example CV content
	cvContent := `
Jane Smith
Full-Stack Developer

EXPERIENCE:
- 4 years of web development
- React, Node.js, TypeScript
- PostgreSQL, MongoDB
- AWS deployment experience

SKILLS:
- Frontend: React, Vue.js, TypeScript
- Backend: Node.js, Express, Python
- Databases: PostgreSQL, MongoDB, Redis
- Cloud: AWS, Docker
`

	// Multiple job descriptions to apply to
	jobs := []struct {
		company     string
		description string
	}{
		{
			company:     "TechCorp",
			description: "Senior Frontend Engineer - React, TypeScript, 5+ years experience",
		},
		{
			company:     "CloudStart",
			description: "Full-Stack Developer - Node.js, React, AWS, microservices",
		},
		{
			company:     "DataSystems",
			description: "Backend Engineer - Python, PostgreSQL, API design",
		},
	}

	// Create batch items
	items := make([]sdk.BatchItem, len(jobs))
	for i, job := range jobs {
		items[i] = sdk.BatchItem{
			CV:             cvContent,
			JobDescription: job.description,
		}
	}

	// Submit batch job
	fmt.Printf("Submitting batch job for %d positions...\n", len(jobs))
	batchResp, err := client.BatchCustomize(context.Background(), items)
	if err != nil {
		log.Fatalf("Failed to submit batch job: %v", err)
	}

	fmt.Printf("✅ Batch job submitted! Job ID: %d\n", batchResp.JobID)
	fmt.Printf("Status: %s\n\n", batchResp.Status)

	// Wait for batch to complete with polling
	fmt.Println("Waiting for batch job to complete...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	results, err := client.WaitForBatch(ctx, batchResp.JobID, 3*time.Second)
	if err != nil {
		log.Fatalf("Failed to wait for batch: %v", err)
	}

	// Display results
	fmt.Printf("\n✅ Batch job completed!\n")
	fmt.Printf("Total items: %d\n", len(results.Items))
	fmt.Printf("Completed at: %s\n\n", results.CreatedAt.Format(time.RFC3339))

	fmt.Println("Results:")
	for i, item := range results.Items {
		fmt.Printf("\n%d. %s\n", i+1, jobs[i].company)
		fmt.Printf("   Status: %s\n", item.Status)
		if item.ErrorMessage != nil {
			fmt.Printf("   Error: %s\n", *item.ErrorMessage)
		} else {
			fmt.Printf("   ✅ CV customized successfully\n")
		}
	}

	fmt.Printf("\n💡 Download all results: %s/api/latest/batch/%d/download\n", serverURL, batchResp.JobID)
}
