package tool

import (
	"context"
)

type Name struct {
}

func NewName() *Name {
	return &Name{}
}

func (h *Name) Functions() []Function {
	return []Function{
		{
			ID:          "name",
			DisplayName: "User Name",
			Description: "Returns the current user's name",
			Parameters:  []Parameter{},
		},
	}
}

func (h *Name) Call(ctx context.Context, _ string, _ map[string]any) (map[string]any, error) {
	return map[string]any{"name": "Matthias Alif"}, nil
}
