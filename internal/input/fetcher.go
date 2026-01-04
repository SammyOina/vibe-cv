// Copyright (c) Ultraviolet
// SPDX-License-Identifier: Apache-2.0

package input

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Fetcher handles fetching content from URLs.
type Fetcher struct {
	httpClient *http.Client
	timeout    time.Duration
}

// NewFetcher creates a new URL fetcher.
func NewFetcher(timeout time.Duration) *Fetcher {
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &Fetcher{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		timeout: timeout,
	}
}

// FetchJobDescription fetches job description from a URL.
func (f *Fetcher) FetchJobDescription(url string) (string, error) {
	if url == "" {
		return "", errors.New("URL cannot be empty")
	}

	// Validate URL format
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "https://" + url
	}

	// Fetch the page
	resp, err := f.httpClient.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch URL: status code %d", resp.StatusCode)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	// Extract text from HTML
	text := extractTextFromHTML(string(body))

	if text == "" {
		return "", errors.New("no meaningful content extracted from URL")
	}

	return text, nil
}

// extractTextFromHTML is a simple HTML to text converter.
func extractTextFromHTML(html string) string {
	// Remove script and style elements
	html = removeHTMLTags("script", html)
	html = removeHTMLTags("style", html)

	// Remove HTML comments
	html = strings.NewReplacer(
		"<!--", "",
		"-->", "",
	).Replace(html)

	// Remove all HTML tags
	inTag := false

	var result strings.Builder

	for _, r := range html {
		if r == '<' {
			inTag = true
		} else if r == '>' {
			inTag = false

			result.WriteRune(' ')
		} else if !inTag {
			result.WriteRune(r)
		}
	}

	text := result.String()

	// Clean up whitespace
	lines := strings.Split(text, "\n")

	var cleanedLines []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			cleanedLines = append(cleanedLines, line)
		}
	}

	return strings.Join(cleanedLines, "\n")
}

// removeHTMLTags removes specific HTML tags and their content.
func removeHTMLTags(tagName, html string) string {
	opening := "<" + tagName
	closing := "</" + tagName + ">"

	for {
		startIdx := strings.Index(strings.ToLower(html), strings.ToLower(opening))
		if startIdx == -1 {
			break
		}

		endIdx := strings.Index(strings.ToLower(html), strings.ToLower(closing))
		if endIdx == -1 {
			html = html[:startIdx]

			break
		}

		html = html[:startIdx] + html[endIdx+len(closing):]
	}

	return html
}

// IsValidJobURL checks if a URL is likely a job listing.
func IsValidJobURL(url string) bool {
	lowercaseURL := strings.ToLower(url)

	jobDomains := []string{
		"linkedin.com",
		"indeed.com",
		"glassdoor.com",
		"dice.com",
		"builtin.com",
		"techcrunch.com",
		"careers.",
		"/careers",
		"/jobs",
	}

	for _, domain := range jobDomains {
		if strings.Contains(lowercaseURL, domain) {
			return true
		}
	}

	return false
}
