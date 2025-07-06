package tool

import (
	"context"
	"fmt"
	"os/user"
)

type UserName struct {
}

func NewUserName() *UserName {
	return &UserName{}
}

func (h *UserName) Functions() []Function {
	return []Function{
		{
			ID:          "user_name",
			DisplayName: "User Name",
			Description: "Returns the current user's name",
			Parameters:  []Parameter{},
		},
	}
}

func (h *UserName) Call(ctx context.Context, _ string, _ map[string]any) (map[string]any, error) {
	currentUser, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}
	return map[string]any{"name": currentUser.Username}, nil
}
