package tool

import (
	"context"
	"encoding/json"
	"fmt"
)

type Tool interface {
	Functions() []Function
	Call(ctx context.Context, function string, args map[string]any) (map[string]any, error)
}

// Function represents a tool that can be used by the agent.
type Function struct {
	// ID uniquely identiies a tool
	ID string

	// DisplayName returns the display name of the tool
	DisplayName string

	// Description provides a brief description of the tool
	Description string

	// A JSON schema describing the tool's parameters
	Parameters []Parameter
}

type Parameter struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
}

type ToolBelt map[string]Tool

func (tb *ToolBelt) Call(ctx context.Context, name string, args map[string]any) (map[string]any, error) {
	if _, ok := (*tb)[name]; !ok {
		return nil, fmt.Errorf("tool %s not found", name)
	}
	return (*tb)[name].Call(ctx, name, args)
}

func NewToolBelt(tools ...Tool) ToolBelt {
	belt := make(ToolBelt)

	for _, t := range tools {
		fs := t.Functions()
		for _, f := range fs {
			belt[f.ID] = t
		}
	}
	return belt
}

func (tb ToolBelt) Describe() string {
	// json marshal
	jsonBytes, err := json.Marshal(tb)
	if err != nil {
		return fmt.Sprintf("Error marshaling tool belt: %v", err)
	}
	return string(jsonBytes)
}
