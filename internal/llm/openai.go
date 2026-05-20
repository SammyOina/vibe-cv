// Copyright (c) Ultraviolet
// SPDX-License-Identifier: Apache-2.0

package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/sammyoina/vibe-cv/pkg/latex"
	"github.com/sashabaranov/go-openai"
)

// OpenAIProvider implements the Provider interface using OpenAI's API.
type OpenAIProvider struct {
	client *openai.Client
	model  string
}

// NewOpenAIProvider creates a new OpenAI provider.
func NewOpenAIProvider(apiKey, model string) *OpenAIProvider {
	return &OpenAIProvider{
		client: openai.NewClient(apiKey),
		model:  model,
	}
}

// Customize customizes a CV using OpenAI.
func (p *OpenAIProvider) Customize(ctx context.Context, cv, jobDescription string, additionalContext []string, latexTemplate string, isFullLatex bool) (*CustomizationResponse, error) {
	prompt := buildPrompt(cv, jobDescription, additionalContext)

	if isFullLatex && latexTemplate != "" {
		prompt += "\n\nYou must generate the complete CV conforming STRICTLY to the following LaTeX template format. Do NOT modify the layout/packages, only rewrite the textual content matching the candidate's specifics. Return the complete, compilable LaTeX code AFTER the ---LATEX--- delimiter."
		prompt += "\n\nLaTeX Template:\n" + latexTemplate
	}

	resp, err := p.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: p.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: systemPrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		Temperature: 0.7,
		TopP:        0.9,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to customize CV with OpenAI: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, errors.New("no response from OpenAI")
	}

	content := resp.Choices[0].Message.Content
	modifiedCV, matchScore, modifications := parseResponse(content, isFullLatex)

	return &CustomizationResponse{
		ModifiedCV:    modifiedCV,
		MatchScore:    matchScore,
		Modifications: modifications,
	}, nil
}

// GetName returns the provider name.
func (p *OpenAIProvider) GetName() string {
	return "openai"
}

const systemPrompt = `You are an expert CV consultant with deep knowledge of ATS (Applicant Tracking Systems) and job market trends.
Your task is to customize CVs to match job descriptions while maintaining authenticity and truthfulness.

When customizing a CV:
1. Analyze the job description to identify key requirements and keywords
2. Map the candidate's experience to the job requirements
3. Reorder and reword bullet points to emphasize relevant experience
4. Ensure all modifications are truthful and represent actual work done
5. Use industry-standard terminology and keywords from the job description
6. Maintain the CV's original structure and professionalism

CRITICAL LaTeX compilation rules (ALWAYS follow these):
- Only use packages from texlive-base, texlive-latex-extra, texlive-fonts-recommended
- SAFE packages: geometry, hyperref, xcolor, titlesec, enumitem, multicol, tabularx, booktabs, parskip, microtype, fancyhdr, array, paracol, helvet, mathptmx, lmodern, fontenc, inputenc
- DO NOT use: fontawesome5, lato, roboto, opensans, sourcesanspro, tikz-heavy libraries
- NEVER use \write18, \input{/...}, \include{/...}, \openin, \openout
- ALWAYS ensure every \begin{env} has a matching \end{env}
- ALWAYS include \documentclass, \begin{document}, and \end{document}
- Escape special characters: & → \&, % → \%, # → \#, _ → \_ (in text mode)
- Use \textbf{} not **bold**, use \textit{} not *italic*

Respond STRICTLY in the following format:

---METADATA---
{
  "match_score": 0.95,
  "modifications": ["List of changes..."]
}
---LATEX---
\documentclass{...}
... complete LaTeX code here ...`

func buildPrompt(cv, jobDescription string, additionalContext []string) string {
	var contextStr string
	if len(additionalContext) > 0 {
		contextStr = "\nAdditional Context:\n" + strings.Join(additionalContext, "\n")
	}

	return fmt.Sprintf(`Please customize the following CV to match this job description:

Job Description:
%s

Original CV:
%s
%s

Return your response in the format specified in your instructions.`, jobDescription, cv, contextStr)
}

// parseResponse parses the LLM response.
func parseResponse(content string, isFullLatex bool) (string, float64, []string) {
	var modifiedCV string
	score := 0.5
	mods := []string{"Failed to parse properly"}
	parsed := false

	// Try the new robust split format first
	if strings.Contains(content, "---LATEX---") {
		parts := strings.Split(content, "---LATEX---")
		modifiedCV = strings.TrimSpace(parts[1])

		// Parse metadata from the first part
		metaStart := strings.Index(parts[0], "{")
		metaEnd := strings.LastIndex(parts[0], "}")
		if metaStart != -1 && metaEnd != -1 {
			var metadata struct {
				MatchScore    float64  `json:"match_score"`
				Modifications []string `json:"modifications"`
			}
			if err := json.Unmarshal([]byte(parts[0][metaStart:metaEnd+1]), &metadata); err == nil {
				score = metadata.MatchScore
				mods = metadata.Modifications
			}
		}
		parsed = true
	}

	// Fallback for old JSON format or tagged formats
	if !parsed {
		startIdx := strings.Index(content, "{")
		endIdx := strings.LastIndex(content, "}")

		if startIdx != -1 && endIdx != -1 {
			jsonStr := content[startIdx : endIdx+1]
			var resp struct {
				CustomizedCV  string   `json:"customized_cv"`
				MatchScore    float64  `json:"match_score"`
				Modifications []string `json:"modifications"`
			}

			// If it's valid JSON, let's try to unmarshal it
			if err := json.Unmarshal([]byte(jsonStr), &resp); err == nil && resp.CustomizedCV != "" {
				modifiedCV = resp.CustomizedCV
				score = resp.MatchScore
				mods = resp.Modifications
				parsed = true
			} else {
				// Regex-based robust fallback for malformed JSON to extract customized_cv
				// Match "customized_cv"\s*:\s*"([^"\\]*(?:\\.[^"\\]*)*)"
				re := regexp.MustCompile(`"customized_cv"\s*:\s*"([^"\\]*(?:\\.[^"\\]*)*)"`)
				if matches := re.FindStringSubmatch(jsonStr); len(matches) > 1 {
					escapedVal := matches[1]
					unescaped := strings.ReplaceAll(escapedVal, `\"`, `"`)
					unescaped = strings.ReplaceAll(unescaped, `\\`, `\`)
					unescaped = strings.ReplaceAll(unescaped, `\n`, "\n")
					unescaped = strings.ReplaceAll(unescaped, `\t`, "\t")
					unescaped = strings.ReplaceAll(unescaped, `\r`, "\r")

					modifiedCV = unescaped
					parsed = true

					reScore := regexp.MustCompile(`"match_score"\s*:\s*([0-9.]+)`)
					if scoreMatches := reScore.FindStringSubmatch(jsonStr); len(scoreMatches) > 1 {
						var parsedScore float64
						if _, err := fmt.Sscanf(scoreMatches[1], "%f", &parsedScore); err == nil {
							score = parsedScore
						}
					}

					reMods := regexp.MustCompile(`"modifications"\s*:\s*\[([^\]]*)\]`)
					if modsMatches := reMods.FindStringSubmatch(jsonStr); len(modsMatches) > 1 {
						modsListStr := modsMatches[1]
						reModItem := regexp.MustCompile(`"([^"\\]*(?:\\.[^"\\]*)*)"`)
						items := reModItem.FindAllStringSubmatch(modsListStr, -1)
						if len(items) > 0 {
							mods = nil
							for _, item := range items {
								itemEscaped := item[1]
								itemUnescaped := strings.ReplaceAll(itemEscaped, `\"`, `"`)
								itemUnescaped = strings.ReplaceAll(itemUnescaped, `\\`, `\`)
								mods = append(mods, itemUnescaped)
							}
						}
					}
				}
			}
		}
	}

	// ULTIMATE FALLBACK: If we see LaTeX markers but everything else failed
	if !parsed {
		if strings.Contains(content, "\\documentclass") {
			docStart := strings.Index(content, "\\documentclass")
			modifiedCV = content[docStart:]
			score = 0.5
			mods = []string{"Extracted via fallback"}
		} else {
			modifiedCV = content
		}
	}

	// Apply LaTeX sanitization if this is a full LaTeX compilation request
	if isFullLatex && modifiedCV != "" {
		if sanitized, err := latex.SanitizeLatex(modifiedCV); err == nil {
			modifiedCV = sanitized
		} else if sanitized != "" {
			modifiedCV = sanitized // Use partially sanitized content even on validation error
		}
	}

	return modifiedCV, score, mods
}
