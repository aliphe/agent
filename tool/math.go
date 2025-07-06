package tool

import (
	"context"
	"fmt"

	"github.com/aliphe/skipery/pkg/jsonschema"
)

type Math struct {
}

func NewMath() *Math {
	return &Math{}
}

func (h *Math) Functions() []Function {
	return []Function{
		{
			ID:          "sum",
			DisplayName: "Sum",
			Description: "Returns the sum of the two provided numbers",
			Parameters: jsonschema.JSONSchema{
				Type: "object",
				Properties: map[string]jsonschema.JSONSchema{
					"a": jsonschema.JSONSchema{
						Type: "number",
					},
					"b": jsonschema.JSONSchema{
						Type: "number",
					},
				},
				Required: []string{"a", "b"},
			},
		},
		{
			ID:          "subtract",
			DisplayName: "Subtract",
			Description: "Returns the difference of the two provided numbers",
			Parameters: jsonschema.JSONSchema{
				Type: "object",
				Properties: map[string]jsonschema.JSONSchema{
					"a": jsonschema.JSONSchema{
						Type: "number",
					},
					"b": jsonschema.JSONSchema{
						Type: "number",
					},
				},
				Required:         []string{"a", "b"},
				PropertyOrdering: []string{"a", "b"},
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
