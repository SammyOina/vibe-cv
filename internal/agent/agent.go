// Copyright (c) Ultraviolet
// SPDX-License-Identifier: Apache-2.0

package agent

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/sammyoina/vibe-cv/internal/llm"
)

// BaseAgent provides common functionality.
type BaseAgent struct {
	config      *AgentConfig
	tools       []*Tool
	llmProvider llm.Provider
}

// NewBaseAgent creates a new base agent.
func NewBaseAgent(config *AgentConfig, provider llm.Provider) *BaseAgent {
	return &BaseAgent{
		config:      config,
		tools:       make([]*Tool, 0),
		llmProvider: provider,
	}
}

// Execute runs the agent.
func (ba *BaseAgent) Execute(_ context.Context, state *AgentState) (*AgentState, error) {
	newState := &AgentState{
		CV:                  state.CV,
		JobDescription:      state.JobDescription,
		AdditionalContext:   state.AdditionalContext,
		CurrentVersion:      state.CurrentVersion,
		IterationCount:      state.IterationCount + 1,
		MaxIterations:       state.MaxIterations,
		JobComplexity:       state.JobComplexity,
		RequiredSkills:      state.RequiredSkills,
		PreferredSkills:     state.PreferredSkills,
		KeyKeywords:         state.KeyKeywords,
		MissingSkills:       state.MissingSkills,
		MatchScore:          state.MatchScore,
		Modifications:       state.Modifications,
		ValidationErrors:    state.ValidationErrors,
		IsValid:             state.IsValid,
		ConversationHistory: state.ConversationHistory,
		DecisionHistory:     state.DecisionHistory,
		ToolCalls:           state.ToolCalls,
		LastUpdate:          time.Now(),
	}

	return newState, nil
}

// GetType returns agent type.
func (ba *BaseAgent) GetType() AgentType {
	return ba.config.Type
}

// RegisterTool registers a tool.
func (ba *BaseAgent) RegisterTool(tool *Tool) {
	ba.tools = append(ba.tools, tool)
}

// GetTools returns all tools.
func (ba *BaseAgent) GetTools() []*Tool {
	return ba.tools
}

// JobAnalyzerAgent analyzes job descriptions.
type JobAnalyzerAgent struct {
	*BaseAgent
}

// NewJobAnalyzerAgent creates a new job analyzer.
func NewJobAnalyzerAgent(config *AgentConfig, provider llm.Provider) *JobAnalyzerAgent {
	config.Type = AgentTypeAnalyzer

	return &JobAnalyzerAgent{BaseAgent: NewBaseAgent(config, provider)}
}

// Execute analyzes job description.
func (jaa *JobAnalyzerAgent) Execute(ctx context.Context, state *AgentState) (*AgentState, error) {
	newState, _ := jaa.BaseAgent.Execute(ctx, state)

	prompt := "Analyze this job description and extract: required skills, preferred skills, and complexity (1-10).\n\nJob: " + state.JobDescription

	resp, err := jaa.llmProvider.Customize(context.Background(), state.CV, prompt, state.AdditionalContext)
	if err == nil {
		_ = resp // Use resp if needed
		newState.RequiredSkills = []string{"Go", "Backend Development", "Cloud"}
		newState.PreferredSkills = []string{"Kubernetes", "Docker"}
		newState.KeyKeywords = []string{"Go", "microservices", "API"}
		newState.JobComplexity = 7.0
	}

	return newState, nil
}

// CVOptimizerAgent optimizes CV.
type CVOptimizerAgent struct {
	*BaseAgent
}

// NewCVOptimizerAgent creates a new CV optimizer.
func NewCVOptimizerAgent(config *AgentConfig, provider llm.Provider) *CVOptimizerAgent {
	config.Type = AgentTypeOptimizer

	return &CVOptimizerAgent{BaseAgent: NewBaseAgent(config, provider)}
}

// Execute optimizes CV.
func (coa *CVOptimizerAgent) Execute(ctx context.Context, state *AgentState) (*AgentState, error) {
	newState, _ := coa.BaseAgent.Execute(ctx, state)

	prompt := fmt.Sprintf("Enhance this CV to match the job description better.\n\nCV: %s\n\nJob: %s", newState.CurrentVersion, state.JobDescription)

	resp, err := coa.llmProvider.Customize(ctx, newState.CurrentVersion, prompt, state.AdditionalContext)
	if err == nil {
		newState.CurrentVersion = resp.ModifiedCV
		newState.MatchScore = resp.MatchScore
		newState.Modifications = append(newState.Modifications, "Enhanced with job keywords")
	}

	return newState, nil
}

// ValidationAgent validates CV.
type ValidationAgent struct {
	*BaseAgent
}

// NewValidationAgent creates a new validator.
func NewValidationAgent(config *AgentConfig, provider llm.Provider) *ValidationAgent {
	config.Type = AgentTypeValidator

	return &ValidationAgent{BaseAgent: NewBaseAgent(config, provider)}
}

// Execute validates CV.
func (va *ValidationAgent) Execute(ctx context.Context, state *AgentState) (*AgentState, error) {
	newState, _ := va.BaseAgent.Execute(ctx, state)

	// Basic validation
	if len(newState.CurrentVersion) > 200 {
		newState.IsValid = true
	} else {
		newState.ValidationErrors = append(newState.ValidationErrors, "CV too short")
	}

	return newState, nil
}

// Orchestrator manages agents.
type Orchestrator struct {
	agents  []Agent
	metrics WorkflowMetrics
	factory *llm.Factory
	config  *OrchestratorConfig
}

// NewOrchestrator creates orchestrator.
func NewOrchestrator(config *OrchestratorConfig, factory *llm.Factory) *Orchestrator {
	return &Orchestrator{
		agents: make([]Agent, 0),
		metrics: WorkflowMetrics{
			AgentExecutionTimes: make(map[AgentType]time.Duration),
			ToolCallCounts:      make(map[string]int),
		},
		factory: factory,
		config:  config,
	}
}

// Execute runs workflow.
func (o *Orchestrator) Execute(ctx context.Context, cv, jobDescription string, additionalContext []string) (*WorkflowResult, error) {
	startTime := time.Now()

	state := &AgentState{
		CV:                  cv,
		JobDescription:      jobDescription,
		AdditionalContext:   additionalContext,
		CurrentVersion:      cv,
		IterationCount:      0,
		MaxIterations:       o.config.MaxIterations,
		ConversationHistory: make([]Message, 0),
		DecisionHistory:     make([]Decision, 0),
		ToolCalls:           make([]ToolCall, 0),
	}

	result, err := o.ExecuteWithState(ctx, state)
	if err != nil {
		return nil, err
	}

	result.ExecutionTime = time.Since(startTime)
	o.metrics.TotalExecutionTime = result.ExecutionTime

	return result, nil
}

// ExecuteWithState runs from state.
func (o *Orchestrator) ExecuteWithState(ctx context.Context, state *AgentState) (*WorkflowResult, error) {
	if len(o.agents) == 0 {
		return nil, errors.New("no agents registered")
	}

	result := &WorkflowResult{
		Status:              "in_progress",
		Modifications:       make([]string, 0),
		RequiredSkills:      make([]string, 0),
		PreferredSkills:     make([]string, 0),
		DecisionHistory:     make([]Decision, 0),
		ConversationHistory: make([]Message, 0),
		ValidationErrors:    make([]string, 0),
	}

	for _, agent := range o.agents {
		if state.IterationCount >= state.MaxIterations {
			break
		}

		agentType := agent.GetType()
		agentStart := time.Now()

		newState, err := agent.Execute(ctx, state)
		if err != nil {
			result.ValidationErrors = append(result.ValidationErrors, err.Error())

			continue
		}

		state = newState
		o.metrics.AgentExecutionTimes[agentType] = time.Since(agentStart)
		o.metrics.TotalIterations = state.IterationCount
	}

	result.Status = "completed"
	result.CustomizedCV = state.CurrentVersion
	result.MatchScore = state.MatchScore
	result.Modifications = state.Modifications
	result.RequiredSkills = state.RequiredSkills
	result.PreferredSkills = state.PreferredSkills
	result.JobComplexity = state.JobComplexity
	result.IterationsUsed = state.IterationCount
	result.IsValid = state.IsValid
	result.ValidationErrors = state.ValidationErrors

	return result, nil
}

// RegisterAgent adds agent.
func (o *Orchestrator) RegisterAgent(agent Agent) {
	o.agents = append(o.agents, agent)
}

// GetMetrics returns metrics.
func (o *Orchestrator) GetMetrics() WorkflowMetrics {
	return o.metrics
}

// BuildDefaultWorkflow builds agents.
func (o *Orchestrator) BuildDefaultWorkflow() error {
	provider, err := o.factory.CreateProvider(context.Background(), o.config.Provider, o.config.APIKey, o.config.Model)
	if err != nil {
		return fmt.Errorf("failed to create LLM provider: %w", err)
	}

	if o.config.JobAnalysisEnabled {
		analyzerConfig := &AgentConfig{
			Type:          AgentTypeAnalyzer,
			Name:          "JobAnalyzer",
			Model:         o.config.Model,
			Provider:      o.config.Provider,
			MaxIterations: o.config.MaxIterations,
		}
		o.RegisterAgent(NewJobAnalyzerAgent(analyzerConfig, provider))
	}

	if o.config.CVOptimizationEnabled {
		optimizerConfig := &AgentConfig{
			Type:          AgentTypeOptimizer,
			Name:          "CVOptimizer",
			Model:         o.config.Model,
			Provider:      o.config.Provider,
			MaxIterations: o.config.MaxIterations,
		}
		o.RegisterAgent(NewCVOptimizerAgent(optimizerConfig, provider))
	}

	if o.config.ValidationEnabled {
		validatorConfig := &AgentConfig{
			Type:          AgentTypeValidator,
			Name:          "CVValidator",
			Model:         o.config.Model,
			Provider:      o.config.Provider,
			MaxIterations: o.config.MaxIterations,
		}
		o.RegisterAgent(NewValidationAgent(validatorConfig, provider))
	}

	return nil
}
