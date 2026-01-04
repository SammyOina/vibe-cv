// Copyright (c) Ultraviolet
// SPDX-License-Identifier: Apache-2.0

package agent

import (
	"context"
	"time"
)

// AgentType defines the type of agent.
type AgentType string

const (
	AgentTypeAnalyzer     AgentType = "analyzer"
	AgentTypeOptimizer    AgentType = "optimizer"
	AgentTypeValidator    AgentType = "validator"
	AgentTypeOrchestrator AgentType = "orchestrator"
)

// AgentState represents the internal state of an agent.
type AgentState struct {
	CV                  string
	JobDescription      string
	AdditionalContext   []string
	CurrentVersion      string
	IterationCount      int
	MaxIterations       int
	JobComplexity       float64
	RequiredSkills      []string
	PreferredSkills     []string
	KeyKeywords         []string
	MissingSkills       []string
	MatchScore          float64
	Modifications       []string
	ValidationErrors    []string
	IsValid             bool
	ConversationHistory []Message
	DecisionHistory     []Decision
	ToolCalls           []ToolCall
	LastUpdate          time.Time
}

// Message represents a message in conversation history.
type Message struct {
	Role      string
	Content   string
	Timestamp time.Time
}

// Decision tracks an agent's decision.
type Decision struct {
	Timestamp time.Time
	AgentType AgentType
	Decision  string
	Reasoning string
}

// ToolCall represents a tool invocation.
type ToolCall struct {
	ToolName  string
	Input     map[string]any
	Output    any
	Duration  time.Duration
	Timestamp time.Time
}

// Tool defines a function an agent can call.
type Tool struct {
	Name        string
	Description string
	Execute     func(ctx context.Context, input map[string]any) (any, error)
}

// Agent represents an agentic component.
type Agent interface {
	Execute(ctx context.Context, state *AgentState) (*AgentState, error)
	GetType() AgentType
	RegisterTool(tool *Tool)
	GetTools() []*Tool
}

// WorkflowResult contains the final output.
type WorkflowResult struct {
	Status              string
	CustomizedCV        string
	MatchScore          float64
	Modifications       []string
	RequiredSkills      []string
	PreferredSkills     []string
	JobComplexity       float64
	IterationsUsed      int
	ExecutionTime       time.Duration
	DecisionHistory     []Decision
	ConversationHistory []Message
	ValidationErrors    []string
	IsValid             bool
}

// WorkflowMetrics tracks performance.
type WorkflowMetrics struct {
	TotalExecutionTime  time.Duration
	AgentExecutionTimes map[AgentType]time.Duration
	ToolCallCounts      map[string]int
	TotalIterations     int
}

// OrchestratorConfig holds configuration.
type OrchestratorConfig struct {
	MaxIterations         int
	EnableMemory          bool
	EnableToolUse         bool
	JobAnalysisEnabled    bool
	CVOptimizationEnabled bool
	ValidationEnabled     bool
	Provider              string
	Model                 string
	APIKey                string
}

// DefaultOrchestratorConfig returns defaults.
func DefaultOrchestratorConfig() *OrchestratorConfig {
	return &OrchestratorConfig{
		MaxIterations:         5,
		EnableMemory:          true,
		EnableToolUse:         true,
		JobAnalysisEnabled:    true,
		CVOptimizationEnabled: true,
		ValidationEnabled:     true,
		Provider:              "openai",
		Model:                 "gpt-4",
	}
}

// AgentConfig holds agent configuration.
type AgentConfig struct {
	Type          AgentType
	Name          string
	Model         string
	Provider      string
	MaxIterations int
	Timeout       time.Duration
	Temperature   float64
	EnableMemory  bool
	EnableToolUse bool
}
