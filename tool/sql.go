package tool

import (
	"context"
	"fmt"

	"github.com/aliphe/skipery/pkg/jsonschema"
	"github.com/jmoiron/sqlx"
)

type SQL struct {
	db *sqlx.DB
}

func NewSQL(db *sqlx.DB) *SQL {
	return &SQL{db: db}
}

func (s *SQL) Functions() []Function {
	return []Function{
		{
			ID:          "sql_query",
			DisplayName: "SQL Query",
			Description: "Execute a SQL query against the database and return the results",
			Parameters: jsonschema.JSONSchema{
				Type: "object",
				Properties: map[string]jsonschema.JSONSchema{
					"query": {
						Type: "string",
					},
				},
				Required: []string{"query"},
			},
			Response: jsonschema.JSONSchema{
				Type: "object",
				Properties: map[string]jsonschema.JSONSchema{
					"results": {
						Type: "array",
						Items: &jsonschema.JSONSchema{
							Type: "object",
						},
					},
				},
			},
		},
	}
}

func (s *SQL) Call(ctx context.Context, fn string, params map[string]any) (map[string]any, error) {
	switch fn {
	case "sql_query":
		query, ok := params["query"].(string)
		if !ok {
			return nil, fmt.Errorf("query parameter must be a string")
		}

		rows, err := s.db.QueryxContext(ctx, query)
		if err != nil {
			return nil, fmt.Errorf("failed to execute query: %w", err)
		}
		defer rows.Close()

		var results []map[string]any
		for rows.Next() {
			row := make(map[string]any)
			err := rows.MapScan(row)
			if err != nil {
				return nil, fmt.Errorf("failed to scan row: %w", err)
			}
			results = append(results, row)
		}

		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("error iterating rows: %w", err)
		}

		return map[string]any{"results": results}, nil
	default:
		return nil, fmt.Errorf("unknown function: %s", fn)
	}
}
