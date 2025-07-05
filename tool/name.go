package tool

import (
	"context"
	"fmt"
	"os/user"
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
	currentUser, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}
	return map[string]any{"name": currentUser.Username}, nil
}
