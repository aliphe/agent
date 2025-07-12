package mcp

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/aliphe/skipery/pkg/jsonschema"
	"github.com/aliphe/skipery/tool"
	mcpjson "github.com/modelcontextprotocol/go-sdk/jsonschema"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

type Config struct {
	Name    string
	Command string   `json:"command"`
	Args    []string `json:"args"`
	url     string
}

type Client struct {
	cli      *mcpsdk.Client
	sessions []*Session
}

func NewClient() *Client {
	cli := mcpsdk.NewClient(&mcpsdk.Implementation{
		Name:    "agent leaf",
		Version: "0.0.1",
	}, nil)

	return &Client{
		cli: cli,
	}
}

func (c *Client) Connect(ctx context.Context, cfg *Config) error {
	var t mcpsdk.Transport
	switch {
	case cfg.Command != "":
		t = mcpsdk.NewCommandTransport(exec.Command(cfg.Command, cfg.Args...))
	case cfg.url != "":
		t = mcpsdk.NewSSEClientTransport(cfg.url, nil)
	default:
		return fmt.Errorf("invalid command for MCP server %s", cfg.Name)
	}
	s, err := c.cli.Connect(ctx, t)
	if err != nil {
		return err
	}
	c.sessions = append(c.sessions, &Session{session: s})
	return nil
}

func (c *Client) Tools() []tool.Tool {
	if c == nil {
		return nil
	}
	// Manual conversion is required because []ConcreteType and []Interface have different
	// memory layouts in Go.
	out := make([]tool.Tool, 0)
	for _, s := range c.sessions {
		out = append(out, s)
	}
	return out
}

type Session struct {
	session *mcpsdk.ClientSession
}

// Call executes a function on the MCP session.
func (s *Session) Call(ctx context.Context, function string, args map[string]any) (map[string]any, error) {
	res, err := s.session.CallTool(ctx, &mcpsdk.CallToolParams{
		Name:      function,
		Arguments: args,
	})
	if err != nil {
		return nil, err
	}

	cs := make([]string, 0, len(res.Content))
	for _, content := range res.Content {
		b, _ := content.MarshalJSON()
		cs = append(cs, string(b))
	}

	return map[string]any{
		"result": strings.Join(cs, ", "),
	}, nil
}

// Functions returns a list of functions that the MCP session can perform.
func (s *Session) Functions(ctx context.Context) []tool.Function {
	res, err := s.session.ListTools(ctx, nil)
	if err != nil {
		return nil
	}

	tools := make([]tool.Function, 0, len(res.Tools))
	for _, t := range res.Tools {
		tools = append(tools, tool.Function{
			ID:          t.Name,
			DisplayName: t.Title,
			Description: t.Description,
			Parameters:  toJSONSchema(t.InputSchema),
			Response:    toJSONSchema(t.OutputSchema),
		})
	}
	return tools
}

func toJSONSchema(sch *mcpjson.Schema) jsonschema.JSONSchema {
	if sch == nil {
		return jsonschema.JSONSchema{}
	}

	result := jsonschema.JSONSchema{
		Type:        sch.Type,
		Description: sch.Description,
		Required:    sch.Required,
		Examples:    sch.Examples,
	}

	// Convert properties if they exist
	if sch.Properties != nil {
		result.Properties = make(map[string]jsonschema.JSONSchema)
		for name, prop := range sch.Properties {
			result.Properties[name] = toJSONSchema(prop)
		}
	}

	// Convert array items if they exist
	if sch.Items != nil {
		converted := toJSONSchema(sch.Items)
		result.Items = &converted
	}

	return result
}
