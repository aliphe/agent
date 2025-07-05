package agent

import (
	"context"

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
	FunctionResponses map[string]map[string]any `json:"function_responses"`
}

func (m *Message) String() string {
	return m.Text
}

type FunctionCall struct {
	Name string
	Args map[string]any
}

type Agent struct {
	toolBelt     tool.ToolBelt
	model        Model
	conversation []*Message
}

func NewAgent(tools tool.ToolBelt, model Model) *Agent {
	return &Agent{
		toolBelt: tools,
		model:    model,
	}
}

// SendMessage is a basic function to send a message to the agent and receive a response.
func (a *Agent) SendMessage(ctx context.Context, msg string) (string, error) {
	a.conversation = append(a.conversation, &Message{
		Author: AuthorUser,
		Text:   msg,
	})

	conversation, err := a.sendMessage(ctx)
	if err != nil {
		return "", err
	}

	return conversation.String(), nil
}

func (a *Agent) sendMessage(ctx context.Context) (*Message, error) {
	rsp, err := a.model.SendMessage(ctx, a.toolBelt, a.conversation)
	if err != nil {
		return nil, err
	}

	a.conversation = append(a.conversation, rsp)

	if rsp.FunctionCalls != nil {
		message := &Message{
			Author:            AuthorUser,
			FunctionResponses: make(map[string]map[string]any),
		}
		for _, c := range rsp.FunctionCalls {
			toolRes, err := a.toolBelt.Call(ctx, c.Name, c.Args)
			if err != nil {
				return nil, err
			}
			message.FunctionResponses[c.Name] = toolRes
		}

		a.conversation = append(a.conversation, message)
		return a.sendMessage(ctx)
	}

	return rsp, nil
}
