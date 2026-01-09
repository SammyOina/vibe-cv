// Copyright (c) Ultraviolet
// SPDX-License-Identifier: Apache-2.0

package ats

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/sammyoina/vibe-cv/internal/llm"
)

// Analyzer handles ATS compatibility analysis.
type Analyzer struct {
	provider llm.Provider
}

// NewAnalyzer creates a new ATS analyzer.
func NewAnalyzer(provider llm.Provider) *Analyzer {
	return &Analyzer{provider: provider}
}

// AnalyzeCV performs a complete ATS analysis on a CV.
func (a *Analyzer) AnalyzeCV(ctx context.Context, cvContent, jobDescription string) (*ATSAnalysisResult, error) {
	// Extract keywords from job description
	keywords, err := a.ExtractKeywords(ctx, jobDescription)
	if err != nil {
		return nil, fmt.Errorf("failed to extract keywords: %w", err)
	}

	// Calculate keyword match
	matchScore, matchedKeywords := a.CalculateKeywordMatch(cvContent, keywords)

	// Determine missing keywords
	missingKeywords := []string{}
	for _, kw := range keywords {
		if !matchedKeywords[kw] {
			missingKeywords = append(missingKeywords, kw)
		}
	}

	// Check formatting
	formattingIssues := a.CheckFormatting(cvContent)

	// Analyze section completeness
	sectionScores := a.AnalyzeSectionCompleteness(cvContent)

	// Calculate overall score
	overallScore := a.calculateOverallScore(matchScore, formattingIssues, sectionScores)

	result := &ATSAnalysisResult{
		OverallScore: overallScore,
		KeywordMatches: KeywordMatches{
			Matched: getMatchedKeywordsList(matchedKeywords),
			Missing: missingKeywords,
		},
		FormattingIssues:    formattingIssues,
		SectionCompleteness: sectionScores,
	}

	// Generate recommendations
	result.Recommendations = a.GenerateRecommendations(result)

	return result, nil
}

// ExtractKeywords uses LLM to extract important keywords from job description.
func (a *Analyzer) ExtractKeywords(ctx context.Context, jobDescription string) ([]string, error) {
	prompt := fmt.Sprintf(`Extract the most important technical skills, qualifications, and keywords from this job description. 
Return ONLY a comma-separated list of keywords, no explanations.

Job Description:
%s

Keywords:`, jobDescription)

	// Use Customize with empty CV to get a simple completion
	result, err := a.provider.Customize(ctx, "", prompt, []string{})
	if err != nil {
		return nil, err
	}

	response := result.ModifiedCV

	// Parse comma-separated keywords
	keywords := []string{}
	parts := strings.Split(response, ",")
	for _, part := range parts {
		kw := strings.TrimSpace(part)
		if kw != "" {
			keywords = append(keywords, strings.ToLower(kw))
		}
	}

	return keywords, nil
}

// CalculateKeywordMatch calculates how many keywords are present in the CV.
func (a *Analyzer) CalculateKeywordMatch(cvContent string, keywords []string) (float64, map[string]bool) {
	cvLower := strings.ToLower(cvContent)
	matched := make(map[string]bool)
	matchCount := 0

	for _, keyword := range keywords {
		kwLower := strings.ToLower(keyword)
		// Use word boundary regex for more accurate matching
		pattern := regexp.MustCompile(`\b` + regexp.QuoteMeta(kwLower) + `\b`)
		if pattern.MatchString(cvLower) {
			matched[keyword] = true
			matchCount++
		}
	}

	if len(keywords) == 0 {
		return 0.0, matched
	}

	return float64(matchCount) / float64(len(keywords)), matched
}

// CheckFormatting detects common formatting issues.
func (a *Analyzer) CheckFormatting(cvContent string) []FormattingIssue {
	issues := []FormattingIssue{}

	// Check length
	wordCount := len(strings.Fields(cvContent))
	if wordCount < 200 {
		issues = append(issues, FormattingIssue{
			Type:     "length_issue",
			Severity: "high",
			Message:  "CV is too short (less than 200 words). ATS systems may flag this.",
		})
	} else if wordCount > 1500 {
		issues = append(issues, FormattingIssue{
			Type:     "length_issue",
			Severity: "medium",
			Message:  "CV is very long (over 1500 words). Consider condensing for better ATS compatibility.",
		})
	}

	// Check for essential sections
	cvLower := strings.ToLower(cvContent)
	if !strings.Contains(cvLower, "experience") && !strings.Contains(cvLower, "work history") {
		issues = append(issues, FormattingIssue{
			Type:     "missing_section",
			Severity: "high",
			Message:  "Missing 'Experience' or 'Work History' section.",
		})
	}

	if !strings.Contains(cvLower, "education") {
		issues = append(issues, FormattingIssue{
			Type:     "missing_section",
			Severity: "medium",
			Message:  "Missing 'Education' section.",
		})
	}

	if !strings.Contains(cvLower, "skill") {
		issues = append(issues, FormattingIssue{
			Type:     "missing_section",
			Severity: "medium",
			Message:  "Missing 'Skills' section.",
		})
	}

	// Check for contact information
	emailPattern := regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	if !emailPattern.MatchString(cvContent) {
		issues = append(issues, FormattingIssue{
			Type:     "missing_contact",
			Severity: "high",
			Message:  "No email address found. Contact information is essential.",
		})
	}

	return issues
}

// AnalyzeSectionCompleteness analyzes the quality of each CV section.
func (a *Analyzer) AnalyzeSectionCompleteness(cvContent string) SectionCompleteness {
	cvLower := strings.ToLower(cvContent)

	// Simple heuristic-based scoring
	scores := SectionCompleteness{
		Experience: 0.5, // Default medium score
		Education:  0.5,
		Skills:     0.5,
		Summary:    0.5,
	}

	// Experience scoring
	if strings.Contains(cvLower, "experience") || strings.Contains(cvLower, "work history") {
		// Count bullet points or job entries (simple heuristic)
		bulletCount := strings.Count(cvContent, "•") + strings.Count(cvContent, "-") + strings.Count(cvContent, "*")
		if bulletCount > 10 {
			scores.Experience = 1.0
		} else if bulletCount > 5 {
			scores.Experience = 0.8
		} else {
			scores.Experience = 0.6
		}
	} else {
		scores.Experience = 0.0
	}

	// Education scoring
	if strings.Contains(cvLower, "education") {
		if strings.Contains(cvLower, "university") || strings.Contains(cvLower, "college") || strings.Contains(cvLower, "degree") {
			scores.Education = 1.0
		} else {
			scores.Education = 0.7
		}
	} else {
		scores.Education = 0.0
	}

	// Skills scoring
	if strings.Contains(cvLower, "skill") {
		// Count number of skills mentioned (simple heuristic)
		skillSection := extractSection(cvContent, "skill")
		commaCount := strings.Count(skillSection, ",")
		if commaCount > 10 {
			scores.Skills = 1.0
		} else if commaCount > 5 {
			scores.Skills = 0.8
		} else {
			scores.Skills = 0.6
		}
	} else {
		scores.Skills = 0.0
	}

	// Summary scoring
	if strings.Contains(cvLower, "summary") || strings.Contains(cvLower, "about") || strings.Contains(cvLower, "profile") {
		summarySection := extractSection(cvContent, "summary")
		wordCount := len(strings.Fields(summarySection))
		if wordCount > 50 && wordCount < 200 {
			scores.Summary = 1.0
		} else if wordCount > 20 {
			scores.Summary = 0.7
		} else {
			scores.Summary = 0.5
		}
	} else {
		scores.Summary = 0.0
	}

	return scores
}

// GenerateRecommendations creates actionable suggestions based on analysis.
func (a *Analyzer) GenerateRecommendations(result *ATSAnalysisResult) []Recommendation {
	recommendations := []Recommendation{}

	// Keyword recommendations
	if len(result.KeywordMatches.Missing) > 0 {
		missingList := strings.Join(result.KeywordMatches.Missing[:min(5, len(result.KeywordMatches.Missing))], ", ")
		recommendations = append(recommendations, Recommendation{
			Category:   "keywords",
			Priority:   "high",
			Suggestion: fmt.Sprintf("Add these missing keywords to improve ATS match: %s", missingList),
		})
	}

	// Formatting recommendations
	for _, issue := range result.FormattingIssues {
		if issue.Severity == "high" {
			recommendations = append(recommendations, Recommendation{
				Category:   "formatting",
				Priority:   "high",
				Suggestion: issue.Message,
			})
		}
	}

	// Section completeness recommendations
	if result.SectionCompleteness.Experience < 0.7 {
		recommendations = append(recommendations, Recommendation{
			Category:   "content",
			Priority:   "high",
			Suggestion: "Expand your experience section with more detailed accomplishments and bullet points.",
		})
	}

	if result.SectionCompleteness.Skills < 0.7 {
		recommendations = append(recommendations, Recommendation{
			Category:   "content",
			Priority:   "medium",
			Suggestion: "Add more relevant skills to your skills section.",
		})
	}

	if result.SectionCompleteness.Summary < 0.5 {
		recommendations = append(recommendations, Recommendation{
			Category:   "content",
			Priority:   "medium",
			Suggestion: "Add a professional summary or profile section at the top of your CV.",
		})
	}

	// Overall score recommendation
	if result.OverallScore < 0.6 {
		recommendations = append(recommendations, Recommendation{
			Category:   "general",
			Priority:   "high",
			Suggestion: "Your CV needs significant improvements for ATS compatibility. Focus on adding relevant keywords and improving formatting.",
		})
	}

	return recommendations
}

// calculateOverallScore computes the final ATS compatibility score.
func (a *Analyzer) calculateOverallScore(keywordMatch float64, issues []FormattingIssue, sections SectionCompleteness) float64 {
	// Weighted scoring
	keywordWeight := 0.4
	formattingWeight := 0.3
	sectionWeight := 0.3

	// Keyword score
	keywordScore := keywordMatch

	// Formatting score (penalize for issues)
	formattingScore := 1.0
	for _, issue := range issues {
		switch issue.Severity {
		case "high":
			formattingScore -= 0.2
		case "medium":
			formattingScore -= 0.1
		case "low":
			formattingScore -= 0.05
		}
	}
	if formattingScore < 0 {
		formattingScore = 0
	}

	// Section score (average of all sections)
	sectionScore := (sections.Experience + sections.Education + sections.Skills + sections.Summary) / 4.0

	// Calculate weighted overall score
	overall := (keywordScore * keywordWeight) + (formattingScore * formattingWeight) + (sectionScore * sectionWeight)

	// Ensure score is between 0 and 1
	if overall > 1.0 {
		overall = 1.0
	}
	if overall < 0.0 {
		overall = 0.0
	}

	return overall
}

// Helper functions

func getMatchedKeywordsList(matched map[string]bool) []string {
	list := []string{}
	for kw, isMatched := range matched {
		if isMatched {
			list = append(list, kw)
		}
	}
	return list
}

func extractSection(content, sectionName string) string {
	lines := strings.Split(content, "\n")
	inSection := false
	sectionContent := ""

	for _, line := range lines {
		lineLower := strings.ToLower(line)
		if strings.Contains(lineLower, strings.ToLower(sectionName)) {
			inSection = true
			continue
		}

		// Stop at next section header (simple heuristic)
		if inSection && len(strings.TrimSpace(line)) > 0 && strings.TrimSpace(line) == strings.ToUpper(strings.TrimSpace(line)) {
			break
		}

		if inSection {
			sectionContent += line + "\n"
		}
	}

	return sectionContent
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ToJSON converts the analysis result to JSON.
func (r *ATSAnalysisResult) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}
