package tool

import (
	"context"
	"fmt"

	"github.com/aliphe/skipery/pkg/jsonschema"
)

var _ Tool = (*Math)(nil)

type Math struct{}

func NewMath() *Math {
	return &Math{}
}

func (h *Math) Functions(ctx context.Context) []Function {
	return []Function{
		{
			ID:          "sum",
			DisplayName: "Sum",
			Description: "Calculates the sum of two numbers. Use this function when you need to add two numerical values together.",
			Parameters: jsonschema.JSONSchema{
				Type:        "object",
				Description: "Parameters for the sum operation",
				Properties: map[string]jsonschema.JSONSchema{
					"a": {
						Type:        "number",
						Description: "The first number to add. Can be any positive or negative number, including decimals.",
						Examples:    []any{5, 10.5, -3.2, 0},
					},
					"b": {
						Type:        "number",
						Description: "The second number to add. Can be any positive or negative number, including decimals.",
						Examples:    []any{3, 7.8, -1.5, 0},
					},
				},
				Required:         []string{"a", "b"},
				PropertyOrdering: []string{"a", "b"},
			},
			Response: jsonschema.JSONSchema{
				Type:        "object",
				Description: "The result of the sum operation",
				Properties: map[string]jsonschema.JSONSchema{
					"result": {
						Type:        "number",
						Description: "The sum of the two input numbers",
					},
				},
			},
		},
		{
			ID:          "subtract",
			DisplayName: "Subtract",
			Description: "Calculates the difference between two numbers (a - b). Use this function when you need to subtract one number from another.",
			Parameters: jsonschema.JSONSchema{
				Type:        "object",
				Description: "Parameters for the subtraction operation",
				Properties: map[string]jsonschema.JSONSchema{
					"a": {
						Type:        "number",
						Description: "The minuend (number to subtract from). Can be any positive or negative number, including decimals.",
						Examples:    []any{10, 5.5, -2.3, 0},
					},
					"b": {
						Type:        "number",
						Description: "The subtrahend (number to subtract). Can be any positive or negative number, including decimals.",
						Examples:    []any{3, 2.2, -1.8, 0},
					},
				},
				Required:         []string{"a", "b"},
				PropertyOrdering: []string{"a", "b"},
			},
			Response: jsonschema.JSONSchema{
				Type:        "object",
				Description: "The result of the subtraction operation",
				Properties: map[string]jsonschema.JSONSchema{
					"result": {
						Type:        "number",
						Description: "The difference between the two input numbers (a - b)",
					},
				},
			},
		},
	}
}

func (h *Math) Call(ctx context.Context, fn string, params map[string]any) (map[string]any, error) {
	switch fn {
	case "sum":
		a, _ := params["a"].(float64)
		b, _ := params["b"].(float64)
		return map[string]any{"result": a + b}, nil
	case "subtract":
		a, _ := params["a"].(float64)
		b, _ := params["b"].(float64)
		return map[string]any{"result": a - b}, nil
	default:
		return nil, fmt.Errorf("unknown function: %s", fn)
	}
}
