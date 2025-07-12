package agent

import (
	"context"
	"encoding/json"
	"os"

	"github.com/aliphe/skipery/mcp"
	"github.com/aliphe/skipery/tool"
)

type Config struct {
	MCP *mcp.Client
}

func (c *Config) Tools() []tool.Tool {
	if c.MCP == nil {
		return nil
	}
	return c.MCP.Tools()
}

func ParseConfig(ctx context.Context, path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var fileConfig struct {
		MCPServers map[string]*mcp.Config `json:"mcpServers"`
	}

	err = json.Unmarshal(data, &fileConfig)
	if err != nil {
		return nil, err
	}

	cli := mcp.NewClient()

	for name, cfg := range fileConfig.MCPServers {
		cfg.Name = name
		cli.Connect(ctx, cfg)

	}

	return &Config{
		MCP: cli,
	}, nil
}
