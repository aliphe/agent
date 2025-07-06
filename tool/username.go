package tool

import (
	"context"
	"fmt"
	"os/user"

	"github.com/aliphe/skipery/pkg/jsonschema"
)

type UserName struct {
}

func NewUserName() *Math {
	return &Math{}
}

func (un *UserName) Functions() []Function {
	return []Function{
		{
			ID:          "user_name",
			DisplayName: "User Name",
			Description: "Returns the current user's name",
			Parameters:  jsonschema.JSONSchema{},
		},
	}
}

func (un *UserName) Call(ctx context.Context, _ string, _ map[string]any) (map[string]any, error) {
	currentUser, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}
	return map[string]any{"name": currentUser.Username}, nil
}
