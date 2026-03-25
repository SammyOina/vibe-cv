// Copyright (c) Ultraviolet
// SPDX-License-Identifier: Apache-2.0

package agent

import (
	"context"
	"fmt"

	"github.com/sammyoina/vibe-cv/internal/llm"
)

// CompilerAgent fixes LaTeX compilation errors.
type CompilerAgent struct {
	*BaseAgent
}

// NewCompilerAgent creates a new compiler agent.
func NewCompilerAgent(config *AgentConfig, provider llm.Provider) *CompilerAgent {
	config.Type = AgentTypeCompiler

	return &CompilerAgent{BaseAgent: NewBaseAgent(config, provider)}
}

// Execute attempts to fix LaTeX based on compilation errors.
// It expects state.CurrentVersion to contain the broken LaTeX,
// and state.CompilationError to contain the pdflatex error log.
func (ca *CompilerAgent) Execute(ctx context.Context, state *AgentState) (*AgentState, error) {
	newState, _ := ca.BaseAgent.Execute(ctx, state)

	if state.CompilationError == "" {
		// Nothing to fix
		newState.FixedLaTeX = state.CurrentVersion

		return newState, nil
	}

	prompt := fmt.Sprintf(`The following LaTeX code failed to compile. 

LaTeX Code:
%s

Compilation Error Log:
%s

Please analyze the error log and fix the LaTeX code. 
Return ONLY the corrected, full LaTeX code that will compile successfully. Do not include any markdown formatting, explanations, or JSON wrappers. Just the plain LaTeX string.`, state.CurrentVersion, state.CompilationError)

	// We misuse Customize here to just get a raw string response from the LLM based on our prompt.
	// Since the LLMs currently return JSON by default due to systemPrompt,
	// we will leverage the fact that if json parsing fails, it returns the raw content as CustomizedCV.

	resp, err := ca.llmProvider.Customize(ctx, state.CurrentVersion, prompt, state.AdditionalContext, "", false)
	if err == nil {
		newState.FixedLaTeX = resp.ModifiedCV
		newState.Modifications = append(newState.Modifications, "Fixed LaTeX compilation error")
	}

	return newState, nil
}
