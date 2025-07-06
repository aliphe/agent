package db

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/aliphe/skipery/agent"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type conversation struct {
	ID        string
	Title     string
	CreatedAt time.Time `db:"created_at"`
}

type message struct {
	ID                string
	ConversationID    string `db:"conversation_id"`
	Author            string
	FunctionCalls     string `db:"function_calls"`
	FunctionResponses string `db:"function_responses"`
	Content           string
	CreatedAt         time.Time `db:"created_at"`
}

func (m *message) Domain() (*agent.Message, error) {
	var functionCalls []agent.FunctionCall
	if m.FunctionCalls != "" {
		if err := json.Unmarshal([]byte(m.FunctionCalls), &functionCalls); err != nil {
			return nil, fmt.Errorf("unmarshal function calls: %w", err)
		}
	}
	var functionResponses agent.FunctionResponse
	if m.FunctionResponses != "" {
		if err := json.Unmarshal([]byte(m.FunctionResponses), &functionResponses); err != nil {
			return nil, fmt.Errorf("unmarshal function responses: %w", err)
		}
	}
	return &agent.Message{
		Author:            agent.Author(m.Author),
		Text:              m.Content,
		FunctionCalls:     functionCalls,
		FunctionResponses: functionResponses,
	}, nil
}

type ConversationStore struct {
	db *sqlx.DB
}

func NewConversationStore(db *sqlx.DB) *ConversationStore {
	return &ConversationStore{db: db}
}

func (s *ConversationStore) Save(ctx context.Context, id, title string) (string, error) {
	if _, err := s.db.Exec("INSERT INTO conversations (id, title, created_at) VALUES (?, ?, ?)",
		id, title, time.Now()); err != nil {
		return "", fmt.Errorf("save conversation: %w", err)
	}
	return id, nil
}

// GetMessages retrieves messages from the conversation store.
func (s *ConversationStore) GetMessages(ctx context.Context, id string) ([]*agent.Message, error) {
	var m []*message
	err := s.db.Select(&m, "SELECT * FROM messages WHERE conversation_id = ?", id)
	if err != nil {
		return nil, fmt.Errorf("fetch messages: %w", err)
	}

	out := make([]*agent.Message, len(m))
	for i, msg := range m {
		out[i], err = msg.Domain()
		if err != nil {
			return nil, fmt.Errorf("domain message: %w", err)
		}
	}
	return out, nil
}

// SaveMessages saves messages to the conversation store.
func (s *ConversationStore) SaveMessages(ctx context.Context, id string, messages []*agent.Message) error {
	for _, msg := range messages {
		fc, err := json.Marshal(msg.FunctionCalls)
		if err != nil {
			return fmt.Errorf("marshal function calls: %w", err)
		}
		fr, err := json.Marshal(msg.FunctionResponses)
		if err != nil {
			return fmt.Errorf("marshal function responses: %w", err)
		}
		slog.Info("marshalled complex fields", "fc", fc, "fr", fr)
		m := &message{
			ID:                uuid.New().String(),
			ConversationID:    id,
			Author:            string(msg.Author),
			FunctionCalls:     string(fc),
			FunctionResponses: string(fr),
			Content:           msg.Text,
			CreatedAt:         time.Now(),
		}
		if _, err := s.db.Exec("INSERT INTO messages (id, conversation_id, author, function_calls, function_responses, content, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
			m.ID, m.ConversationID, m.Author, m.FunctionCalls, m.FunctionResponses, m.Content, m.CreatedAt); err != nil {
			return fmt.Errorf("insert message: %w", err)
		}
	}
	return nil
}
