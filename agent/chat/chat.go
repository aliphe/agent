package chat

import (
	"context"
)

type Author string

const (
	AuthorUser   Author = "user"
	AuthorModel  Author = "model"
	AuthorSystem Author = "system"
)

type Store interface {
	Save(ctx context.Context, id, title string) (string, error)
	GetMessages(ctx context.Context, id string) ([]*Message, error)
	SaveMessages(ctx context.Context, id string, messages []*Message) error
}

type Chat struct {
	ID       string
	Name     string
	messages []*Message
	isNew    bool
	// savedIndex represents the index of the last saved message
	savedIndex int
	store      Store
}

type Message struct {
	Author        Author
	Text          string         `json:"text"`
	FunctionCalls []FunctionCall `json:"function_calls"`
	// Responses objects by function name
	FunctionResponses FunctionResponse `json:"function_responses"`
}

func (m *Message) String() string {
	return m.Text
}

type FunctionCall struct {
	Name string
	Args map[string]any
}
type FunctionResponse map[string]map[string]any

// NewChat creates a new chat instance
func NewChat(id string, store Store) *Chat {
	systemPrompt := &Message{
		Author: AuthorSystem,
		Text:   "You are a helpful AI assistant with access to various tools. You should actively use the available tools to help users accomplish their tasks. When a user asks for something that could benefit from using a tool, always prefer using the appropriate tool rather than just providing a text response. Be proactive in suggesting and using tools that can provide more accurate, up-to-date, or comprehensive information. Your goal is to leverage your tools effectively to give users the best possible assistance.",
	}
	return &Chat{
		ID:       id,
		messages: []*Message{systemPrompt},
		isNew:    true,
		store:    store,
	}
}

// LoadChat creates a chat instance and loads existing messages from store
func LoadChat(ctx context.Context, id string, store Store) (*Chat, error) {
	messages, err := store.GetMessages(ctx, id)
	if err != nil {
		return nil, err
	}

	isNew := len(messages) == 0
	return &Chat{
		ID:         id,
		messages:   messages,
		isNew:      isNew,
		savedIndex: len(messages),
		store:      store,
	}, nil
}

// AddMessage adds a message to the chat
func (c *Chat) AddMessage(msg *Message) {
	c.messages = append(c.messages, msg)
}

// AddUserMessage adds a user message to the chat
func (c *Chat) AddUserMessage(text string) {
	c.AddMessage(&Message{
		Author: AuthorUser,
		Text:   text,
	})
}

// Messages returns all messages in the chat
func (c *Chat) Messages() []*Message {
	return c.messages
}

// // IsNew returns true if this is a new chat
func (c *Chat) IsNew() bool {
	return c.isNew
}

// GetNewMessages returns messages added since chat initialization
func (c *Chat) GetNewMessages() []*Message {
	if c.savedIndex >= len(c.messages) {
		return []*Message{}
	}
	return c.messages[c.savedIndex:]
}

// SaveNewMessages saves only the new messages added since initialization
func (c *Chat) SaveNewMessages(ctx context.Context) error {
	newMessages := c.GetNewMessages()
	if len(newMessages) == 0 {
		return nil
	}

	return c.store.SaveMessages(ctx, c.ID, newMessages)
}

// SaveWithTitle saves the chat with a title (for new chats)
func (c *Chat) SaveWithTitle(ctx context.Context, title string) error {
	c.Name = title
	if c.isNew {
		_, err := c.store.Save(ctx, c.ID, title)
		if err != nil {
			return err
		}
		c.isNew = false
	}
	return nil
}
