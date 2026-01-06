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

	// Create a new SDK client
	client := sdk.NewClient(serverURL)

	// Example CV content
	cvContent := `
John Doe
Software Engineer

EXPERIENCE:
- 5 years of backend development with Go
- Built microservices using Docker and Kubernetes
- Worked with PostgreSQL and Redis

SKILLS:
- Go, Python, JavaScript
- Docker, Kubernetes
- PostgreSQL, MongoDB
- REST APIs, gRPC
`

	// Example job description
	jobDescription := `
Senior Backend Engineer

We are looking for a Senior Backend Engineer with:
- 5+ years of Go experience
- Strong knowledge of cloud technologies (AWS, GCP, or Azure)
- Experience with Kubernetes and containerization
- Database design and optimization skills
- Microservices architecture experience
`

	// Customize the CV
	fmt.Println("Customizing CV for job description...")
	req := &sdk.CustomizeCVRequest{
		CV:             cvContent,
		JobDescription: jobDescription,
		LLMConfig: &sdk.LLMConfig{
			Provider: sdk.ProviderOpenAI,
			Model:    sdk.ModelGPT4,
		},
	}

	resp, err := client.CustomizeCV(context.Background(), req, sdk.WithRequestAuthToken(authToken))
	if err != nil {
		log.Fatalf("Failed to customize CV: %v", err)
	}

	// Print the results
	fmt.Printf("\n✅ CV customized successfully!\n")
	fmt.Printf("Status: %s\n", resp.Status)
	fmt.Printf("Match Score: %.2f\n", resp.MatchScore)
	fmt.Printf("Customized CV URL: %s\n", resp.CustomizedCVURL)
	fmt.Printf("\nModifications made:\n")
	for i, mod := range resp.Modifications {
		fmt.Printf("  %d. %s\n", i+1, mod)
	}

	fmt.Printf("\n💡 Tip: You can download the PDF from: %s%s\n", serverURL, resp.CustomizedCVURL)
}
