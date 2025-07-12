package tool

import (
	"context"
	"fmt"

	"github.com/aliphe/skipery/pkg/jsonschema"
	"github.com/jmoiron/sqlx"
)

var _ Tool = (*SQL)(nil)

type SQL struct {
	db *sqlx.DB
}

func NewSQL(db *sqlx.DB) *SQL {
	return &SQL{db: db}
}

func (s *SQL) Functions(ctx context.Context) []Function {
	return []Function{
		{
			ID:          "sql_query",
			DisplayName: "SQL Query",
			Description: "Execute a SQL query against the database and return the results. Use this tool to query the chat database, retrieve conversation history, analyze chat patterns, or perform any database operations. The database contains chat and message data.",
			Parameters: jsonschema.JSONSchema{
				Type:        "object",
				Description: "Parameters for executing a SQL query against the database",
				Properties: map[string]jsonschema.JSONSchema{
					"query": {
						Type:        "string",
						Description: "A valid SQL query to execute against the database. Must be a properly formatted SQL statement using standard SQL syntax. The database schema includes: 'chats' table (id TEXT PRIMARY KEY, title TEXT, created_at TIMESTAMP) and 'messages' table (id TEXT PRIMARY KEY, chat_id TEXT, author TEXT, content TEXT, function_calls BLOB, function_responses BLOB, created_at TIMESTAMP). Use standard SQL functions and avoid database-specific syntax.",
						Examples: []any{
							"SELECT * FROM chats ORDER BY created_at DESC LIMIT 10",
							"SELECT COUNT(*) as total_messages FROM messages",
							"SELECT title, COUNT(messages.id) as message_count FROM chats LEFT JOIN messages ON chats.id = messages.chat_id GROUP BY chats.id, chats.title ORDER BY message_count DESC",
							"SELECT author, COUNT(*) as message_count FROM messages GROUP BY author",
							"SELECT * FROM messages WHERE chat_id = 'abc123' ORDER BY created_at",
							"SELECT author, content FROM messages WHERE content LIKE '%error%' ORDER BY created_at DESC",
							"SELECT chats.title, COUNT(messages.id) as msg_count FROM chats LEFT JOIN messages ON chats.id = messages.chat_id GROUP BY chats.id, chats.title HAVING msg_count > 5",
						},
					},
				},
				Required:         []string{"query"},
				PropertyOrdering: []string{"query"},
			},
			Response: jsonschema.JSONSchema{
				Type:        "object",
				Description: "The results of the SQL query execution",
				Properties: map[string]jsonschema.JSONSchema{
					"results": {
						Type:        "array",
						Description: "Array of result rows from the SQL query. Each row is an object with column names as keys and their values. Empty array if no results found.",
						Items: &jsonschema.JSONSchema{
							Type:        "object",
							Description: "A single row result with column names as keys and their corresponding values",
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
