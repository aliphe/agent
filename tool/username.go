package tool

import (
	"context"
	"fmt"
	"os/user"

	"github.com/aliphe/skipery/pkg/jsonschema"
)

type UserName struct {
}

func NewUserName() *UserName {
	return &UserName{}
}

func (un *UserName) Functions() []Function {
	return []Function{
		{
			ID:          "user_name",
			DisplayName: "User Name",
			Description: "Retrieves the current system user's username. Use this function when you need to know who is currently logged in or when personalizing responses.",
			Parameters: jsonschema.JSONSchema{
				Type:        "object",
				Description: "No parameters required for this function",
				Properties:  map[string]jsonschema.JSONSchema{},
			},
			Response: jsonschema.JSONSchema{
				Type:        "object",
				Description: "The current user's information",
				Properties: map[string]jsonschema.JSONSchema{
					"name": {
						Type:        "string",
						Description: "The username of the currently logged-in user",
						Examples:    []any{"john", "admin", "user123"},
					},
				},
			},
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
