# vibe-cv Go SDK

A comprehensive Go SDK for the vibe-cv API, providing type-safe access to CV customization, batch processing, version management, and analytics features.

## Installation

```bash
go get github.com/sammyoina/vibe-cv/sdk
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/sammyoina/vibe-cv/sdk"
)

func main() {
    // Create a new client
    client := sdk.NewClient("http://localhost:8080")
    
    // Customize a CV with per-request authentication
    req := &sdk.CustomizeCVRequest{
        CV:             "Your CV content here...",
        JobDescription: "Job description here...",
        LLMConfig: &sdk.LLMConfig{
            Provider: sdk.ProviderOpenAI,
            Model:    sdk.ModelGPT4,
        },
    }
    
    // Pass authentication token with the request
    resp, err := client.CustomizeCV(
        context.Background(),
        req,
        sdk.WithRequestAuthToken("user-session-token"),
    )
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Match Score: %.2f\n", resp.MatchScore)
    fmt.Printf("Download URL: %s\n", resp.CustomizedCVURL)
}
```

## Features

- **CV Customization**: Customize CVs for specific job descriptions using various LLM providers
- **Batch Processing**: Submit and track batch customization jobs
- **Version Management**: List, compare, and download CV versions
- **Analytics**: Track customization metrics and performance
- **Health Checks**: Monitor service availability
- **Error Handling**: Comprehensive error types with helper methods
- **Type Safety**: Full type definitions for all API operations
- **Per-Request Authentication**: Each request can have its own authentication token

## Configuration

### Client Options

```go
client := sdk.NewClient(
    "http://localhost:8080",
    sdk.WithTimeout(60 * time.Second),
    sdk.WithUserAgent("my-app/1.0.0"),
)
```

### Per-Request Authentication

**All authenticated endpoints require a token to be passed with each request:**

```go
// Single user scenario
client := sdk.NewClient("http://localhost:8080")
userToken := "user-session-token-123"

resp, err := client.CustomizeCV(
    ctx,
    req,
    sdk.WithRequestAuthToken(userToken),
)
```

**Multi-user SaaS scenario:**

```go
// Single client instance serving multiple users
client := sdk.NewClient("http://localhost:8080")

// User 1's request
resp1, err := client.CustomizeCV(
    ctx,
    req1,
    sdk.WithRequestAuthToken(user1Token),
)

// User 2's request  
resp2, err := client.CustomizeCV(
    ctx,
    req2,
    sdk.WithRequestAuthToken(user2Token),
)

### LLM Providers

The SDK supports multiple LLM providers:

```go
// OpenAI
llmConfig := &sdk.LLMConfig{
    Provider: sdk.ProviderOpenAI,
    Model:    sdk.ModelGPT4,
}

// Anthropic Claude
llmConfig := &sdk.LLMConfig{
    Provider: sdk.ProviderAnthropic,
    Model:    sdk.ModelClaude3Opus,
}

// Google Gemini
llmConfig := &sdk.LLMConfig{
    Provider: sdk.ProviderGemini,
    Model:    sdk.ModelGeminiPro,
}
```

## API Reference

### CV Customization

```go
// Customize a CV (requires authentication)
resp, err := client.CustomizeCV(
    ctx,
    &sdk.CustomizeCVRequest{
        CV:             cvContent,
        JobDescription: jobDesc,
        AdditionalContext: []sdk.ContextItem{
            {Type: "text", Content: "Additional skills..."},
        },
        LLMConfig: &sdk.LLMConfig{
            Provider: sdk.ProviderOpenAI,
            Model:    sdk.ModelGPT4,
        },
    },
    sdk.WithRequestAuthToken(userToken),
)
```

### Batch Processing

```go
// Submit batch job
items := []sdk.BatchItem{
    {CV: cvContent, JobDescription: "Job 1"},
    {CV: cvContent, JobDescription: "Job 2"},
}
batchResp, err := client.BatchCustomize(
    ctx,
    items,
    sdk.WithRequestAuthToken(userToken),
)

// Check status
status, err := client.GetBatchStatus(
    ctx,
    batchResp.JobID,
    sdk.WithRequestAuthToken(userToken),
)

// Wait for completion (with polling)
results, err := client.WaitForBatch(
    ctx,
    batchResp.JobID,
    5*time.Second,
    sdk.WithRequestAuthToken(userToken),
)

// Download results
results, err := client.DownloadBatchResults(
    ctx,
    batchResp.JobID,
    sdk.WithRequestAuthToken(userToken),
)
```

### Version Management

```go
// List all versions for a CV
versions, err := client.GetVersions(
    ctx,
    cvID,
    sdk.WithRequestAuthToken(userToken),
)

// Get version details
version, err := client.GetVersionDetail(
    ctx,
    versionID,
    sdk.WithRequestAuthToken(userToken),
)

// Compare two versions
comparison, err := client.CompareVersions(
    ctx,
    versionID1,
    versionID2,
    sdk.WithRequestAuthToken(userToken),
)

// Download CV as PDF
pdfData, err := client.DownloadCV(
    ctx,
    versionID,
    sdk.WithRequestAuthToken(userToken),
)
```

### Analytics

```go
// Get user analytics
analytics, err := client.GetAnalytics(
    ctx,
    50, // limit to 50 snapshots
    sdk.WithRequestAuthToken(userToken),
)

// Get dashboard statistics
dashboard, err := client.GetDashboard(
    ctx,
    sdk.WithRequestAuthToken(userToken),
)
```

### Health Checks

```go
// Full health check (no auth required)
health, err := client.Health(ctx)

// Simple ping (no auth required)
err := client.Ping(ctx)
```

## Error Handling

The SDK provides custom error types for better error handling:

```go
resp, err := client.CustomizeCV(ctx, req)
if err != nil {
    if apiErr, ok := err.(*sdk.APIError); ok {
        // Handle API errors
        fmt.Printf("API Error: %s (status: %d)\n", apiErr.Message, apiErr.StatusCode)
        
        if apiErr.IsNotFound() {
            // Handle 404
        } else if apiErr.IsUnauthorized() {
            // Handle 401
        } else if apiErr.IsServerError() {
            // Handle 5xx
        }
    } else if valErr, ok := err.(*sdk.ValidationError); ok {
        // Handle validation errors
        fmt.Printf("Validation Error on %s: %s\n", valErr.Field, valErr.Message)
    } else {
        // Handle other errors
        fmt.Printf("Error: %v\n", err)
    }
}
```

## Examples

See the [examples](./examples) directory for complete working examples:

- [basic](./examples/basic/main.go) - Simple CV customization
- [batch](./examples/batch/main.go) - Batch processing with polling
- [advanced](./examples/advanced/main.go) - All features with error handling

### Running Examples

```bash
# Set the server URL (optional, defaults to http://localhost:8080)
export VIBE_CV_URL=http://localhost:8080

# Run basic example
cd examples/basic
go run main.go

# Run batch example
cd examples/batch
go run main.go

# Run advanced example
cd examples/advanced
go run main.go
```

## Best Practices

1. **Use Context**: Always pass a context with timeout for production use:
   ```go
   ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
   defer cancel()
   ```

2. **Handle Errors**: Check for specific error types to handle different scenarios appropriately

3. **Batch Processing**: Use `WaitForBatch` for automatic polling instead of manual status checks

4. **Reuse Clients**: Create one client instance and reuse it across requests

5. **Configure Timeouts**: Set appropriate timeouts based on your use case (CV customization can take time)

## License

Apache-2.0 - See LICENSE file for details
