package db

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/aliphe/skipery/agent/chat"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type chatRecord struct {
	ID        string
	Title     string
	CreatedAt time.Time `db:"created_at"`
}

type message struct {
	ID                string
	ChatID            string `db:"chat_id"`
	Author            string
	FunctionCalls     string `db:"function_calls"`
	FunctionResponses string `db:"function_responses"`
	Content           string
	CreatedAt         time.Time `db:"created_at"`
}

func (m *message) chat() (*chat.Message, error) {
	var functionCalls []chat.FunctionCall
	if m.FunctionCalls != "" {
		if err := json.Unmarshal([]byte(m.FunctionCalls), &functionCalls); err != nil {
			return nil, fmt.Errorf("unmarshal function calls: %w", err)
		}
	}
	var functionResponses chat.FunctionResponse
	if m.FunctionResponses != "" {
		if err := json.Unmarshal([]byte(m.FunctionResponses), &functionResponses); err != nil {
			return nil, fmt.Errorf("unmarshal function responses: %w", err)
		}
	}
	return &chat.Message{
		Author:            chat.Author(m.Author),
		Text:              m.Content,
		FunctionCalls:     functionCalls,
		FunctionResponses: functionResponses,
	}, nil
}

type ChatStore struct {
	db *sqlx.DB
}

func NewChatStore(db *sqlx.DB) *ChatStore {
	return &ChatStore{db: db}
}

func (s *ChatStore) Save(ctx context.Context, id, title string) (string, error) {
	if _, err := s.db.Exec("INSERT INTO chats (id, title, created_at) VALUES (?, ?, ?)",
		id, title, time.Now()); err != nil {
		return "", fmt.Errorf("save chat: %w", err)
	}
	return id, nil
}

// GetMessages retrieves messages from the chat store.
func (s *ChatStore) GetMessages(ctx context.Context, id string) ([]*chat.Message, error) {
	var m []*message
	err := s.db.Select(&m, "SELECT * FROM messages WHERE chat_id = ?", id)
	if err != nil {
		return nil, fmt.Errorf("fetch messages: %w", err)
	}

	out := make([]*chat.Message, len(m))
	for i, msg := range m {
		out[i], err = msg.chat()
		if err != nil {
			return nil, fmt.Errorf("chat message: %w", err)
		}
	}
	return out, nil
}

// SaveMessages saves messages to the chat store.
func (s *ChatStore) SaveMessages(ctx context.Context, id string, messages []*chat.Message) error {
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
			ChatID:            id,
			Author:            string(msg.Author),
			FunctionCalls:     string(fc),
			FunctionResponses: string(fr),
			Content:           msg.Text,
			CreatedAt:         time.Now(),
		}
		if _, err := s.db.Exec("INSERT INTO messages (id, chat_id, author, function_calls, function_responses, content, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
			m.ID, m.ChatID, m.Author, m.FunctionCalls, m.FunctionResponses, m.Content, m.CreatedAt); err != nil {
			return fmt.Errorf("insert message: %w", err)
		}
	}
	return nil
}
