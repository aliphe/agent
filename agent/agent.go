package agent

import (
	"context"
	"log/slog"
	"slices"

	"github.com/aliphe/skipery/agent/chat"
	"github.com/aliphe/skipery/tool"
)

type Model interface {
	SendMessage(ctx context.Context, toolBelt tool.ToolBelt, chat []*chat.Message) (*chat.Message, error)
}

type Agent struct {
	config    *Config
	toolBelt  tool.ToolBelt
	chatStore chat.Store
	model     Model
}

func NewAgent(config *Config, tools tool.ToolBelt, chatStore chat.Store, model Model) *Agent {
	return &Agent{
		config:    config,
		toolBelt:  tools,
		chatStore: chatStore,
		model:     model,
	}
}

// SendMessage is a basic function to send a message to the agent and receive a response.
func (a *Agent) SendMessage(ctx context.Context, chatID string, msg string) ([]*chat.Message, error) {
	chatSession, err := chat.LoadChat(ctx, chatID, a.chatStore)
	if err != nil {
		return nil, err
	}

	chatSession.AddUserMessage(msg)

	msgs, err := a.sendMessage(ctx, chatSession.Messages())
	if err != nil {
		return nil, err
	}

	// Update the chat with new messages
	for _, newMsg := range msgs[len(chatSession.Messages()):] {
		chatSession.AddMessage(newMsg)
	}

	if chatSession.IsNew() {
		name, err := a.chatName(ctx, msgs)
		if err != nil {
			return nil, err
		}

		err = chatSession.SaveWithTitle(ctx, name)
		if err != nil {
			return nil, err
		}
	}

	newMsgs := chatSession.NewMessages()

	err = chatSession.SaveNewMessages(ctx)
	if err != nil {
		return nil, err
	}

	return newMsgs, nil
}

func (a *Agent) chatName(ctx context.Context, messages []*chat.Message) (string, error) {
	res, err := a.model.SendMessage(ctx, nil, append(messages, &chat.Message{
		Author: chat.AuthorUser,
		Text:   "Sum up this chat as one short nouns phrase, focusing on the user question.",
	}))
	if err != nil {
		return "", err
	}
	slog.Info("generated chat name", "name", res.Text)

	return res.Text, nil
}

func (a *Agent) sendMessage(ctx context.Context, messages []*chat.Message) ([]*chat.Message, error) {
	rsp, err := a.model.SendMessage(ctx, a.toolBelt, messages)
	if err != nil {
		return nil, err
	}
	msgs := slices.Clone(messages)
	msgs = append(msgs, rsp)

	if rsp.FunctionCalls != nil {
		message := &chat.Message{
			Author:            chat.AuthorUser,
			FunctionResponses: make(chat.FunctionResponse),
		}
		for _, c := range rsp.FunctionCalls {
			toolRes, err := a.toolBelt.Call(ctx, c.Name, c.Args)
			if err != nil {
				message.FunctionResponses[c.Name] = map[string]any{
					"error": err.Error(),
				}
			} else {
				message.FunctionResponses[c.Name] = toolRes
			}
		}
		msgs = append(msgs, message)

		return a.sendMessage(ctx, msgs)
	}

	return msgs, nil
}
