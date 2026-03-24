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
		prompt += "\n\nYou must generate the complete CV conforming STRICTLY to the following LaTeX template format. Do NOT modify the layout/packages, only rewrite the textual content matching the candidate's specifics. Return the complete, compilable LaTeX code in the 'customized_cv' JSON field."
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
	modifiedCV, matchScore, modifications := parseResponse(content)

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

Respond with a JSON object containing:
- "customized_cv": the modified CV text
- "match_score": a number between 0 and 1 indicating how well the CV matches the job
- "modifications": an array of strings describing the changes made

CRITICAL JSON REQUIREMENT:
Because the customized_cv field contains LaTeX code, you MUST properly escape all backslashes according to JSON format rules.
For example, you must write "\\documentclass" instead of "\documentclass" and "\\textbf" instead of "\textbf".
If you fail to escape backslashes, the JSON parser will crash.`

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

Return your response as a valid JSON object with the structure specified in your instructions.`, jobDescription, cv, contextStr)
}

// responseJSON represents the expected JSON response from the LLM.
type responseJSON struct {
	CustomizedCV  string   `json:"customized_cv"`
	MatchScore    float64  `json:"match_score"`
	Modifications []string `json:"modifications"`
}

// parseResponse parses the LLM response.
func parseResponse(content string) (string, float64, []string) {
	// Extract JSON from response (might have additional text)
	startIdx := strings.Index(content, "{")
	endIdx := strings.LastIndex(content, "}")

	if startIdx == -1 || endIdx == -1 {
		// Fallback if JSON parsing fails
		return content, 0.5, []string{"CV customized"}
	}

	jsonStr := content[startIdx : endIdx+1]

	// FIX: Escape backslashes that the LLM forgot to escape.
	// We replace single backslashes with double backslashes, but avoid 
	// double-escaping things the LLM already properly escaped like \n, \t, or \\
	re := regexp.MustCompile(`\\([^"\\/bfnrt])`)
	jsonStr = re.ReplaceAllString(jsonStr, `\\$1`)

	// Try to parse as JSON
	var resp responseJSON

	err := json.Unmarshal([]byte(jsonStr), &resp)
	if err != nil {
		// Print the error and the JSON string to logs for debugging
		fmt.Printf("JSON unmarshal error: %v\nJSON string: %s\n", err, jsonStr)
		// Fallback to basic string extraction
		return content, 0.5, []string{"CV customized"}
	}

	if resp.CustomizedCV == "" {
		resp.CustomizedCV = content
	}

	return resp.CustomizedCV, resp.MatchScore, resp.Modifications
}
