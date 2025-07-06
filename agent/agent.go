package agent

import (
	"context"
	"log/slog"
	"slices"

	"github.com/aliphe/skipery/tool"
)

type Model interface {
	SendMessage(ctx context.Context, toolBelt tool.ToolBelt, conversation []*Message) (*Message, error)
}

type Author string

const (
	AuthorUser  Author = "user"
	AuthorModel Author = "model"
)

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

type conversationStore interface {
	Save(ctx context.Context, convID, title string) (string, error)
	GetMessages(ctx context.Context, convID string) ([]*Message, error)
	SaveMessages(ctx context.Context, convID string, messages []*Message) error
}

type Agent struct {
	toolBelt      tool.ToolBelt
	conversations conversationStore
	model         Model
}

func NewAgent(tools tool.ToolBelt, convStore conversationStore, model Model) *Agent {
	return &Agent{
		toolBelt:      tools,
		conversations: convStore,
		model:         model,
	}
}

// SendMessage is a basic function to send a message to the agent and receive a response.
func (a *Agent) SendMessage(ctx context.Context, convID string, msg string) ([]*Message, error) {
	messages, err := a.conversations.GetMessages(ctx, convID)
	if err != nil {
		return nil, err
	}
	newConv := false
	if len(messages) == 0 {
		newConv = true
	}

	conversation, err := a.sendMessage(ctx, append(messages, &Message{
		Author: AuthorUser,
		Text:   msg,
	}))
	if err != nil {
		return nil, err
	}

	if newConv {
		name, err := a.conversationName(ctx, conversation)
		if err != nil {
			return nil, err
		}

		convID, err = a.conversations.Save(ctx, convID, name)
		if err != nil {
			return nil, err
		}
	}

	newMsgs := conversation[len(messages):]

	err = a.conversations.SaveMessages(ctx, convID, newMsgs)
	if err != nil {
		return nil, err
	}

	return newMsgs, nil
}

func (a *Agent) conversationName(ctx context.Context, messages []*Message) (string, error) {
	res, err := a.model.SendMessage(ctx, nil, append(messages, &Message{
		Author: AuthorUser,
		Text:   "Sum up this conversation as one short nouns phrase, focusing on the user question.",
	}))
	if err != nil {
		return "", err
	}
	slog.Info("generated conversation name", "name", res.Text)

	return res.Text, nil
}

func (a *Agent) sendMessage(ctx context.Context, messages []*Message) ([]*Message, error) {
	rsp, err := a.model.SendMessage(ctx, a.toolBelt, messages)
	if err != nil {
		return nil, err
	}
	conv := slices.Clone(messages)
	conv = append(conv, rsp)

	if rsp.FunctionCalls != nil {
		message := &Message{
			Author:            AuthorUser,
			FunctionResponses: make(FunctionResponse),
		}
		for _, c := range rsp.FunctionCalls {
			toolRes, err := a.toolBelt.Call(ctx, c.Name, c.Args)
			if err != nil {
				return nil, err
			}
			message.FunctionResponses[c.Name] = toolRes
		}

		conv = append(conv, message)
		return a.sendMessage(ctx, conv)
	}

	return conv, nil
}
