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

type chatStore interface {
	Save(ctx context.Context, chatID, title string) (string, error)
	GetMessages(ctx context.Context, chatID string) ([]*chat.Message, error)
	SaveMessages(ctx context.Context, chatID string, messages []*chat.Message) error
}

type Agent struct {
	toolBelt tool.ToolBelt
	chats    chatStore
	model    Model
}

func NewAgent(tools tool.ToolBelt, chatStore chatStore, model Model) *Agent {
	return &Agent{
		toolBelt: tools,
		chats:    chatStore,
		model:    model,
	}
}

// SendMessage is a basic function to send a message to the agent and receive a response.
func (a *Agent) SendMessage(ctx context.Context, chatID string, msg string) ([]*chat.Message, error) {
	messages, err := a.chats.GetMessages(ctx, chatID)
	if err != nil {
		return nil, err
	}
	newChat := false
	if len(messages) == 0 {
		newChat = true
		// Add system prompt for new chats to make the agent eager to use tools
		systemPrompt := &chat.Message{
			Author: chat.AuthorSystem,
			Text:   "You are a helpful AI assistant with access to various tools. You should actively use the available tools to help users accomplish their tasks. When a user asks for something that could benefit from using a tool, always prefer using the appropriate tool rather than just providing a text response. Be proactive in suggesting and using tools that can provide more accurate, up-to-date, or comprehensive information. Your goal is to leverage your tools effectively to give users the best possible assistance.",
		}
		messages = append([]*chat.Message{systemPrompt}, messages...)
	}

	msgs, err := a.sendMessage(ctx, append(messages, &chat.Message{
		Author: chat.AuthorUser,
		Text:   msg,
	}))
	if err != nil {
		return nil, err
	}

	if newChat {
		name, err := a.chatName(ctx, msgs)
		if err != nil {
			return nil, err
		}

		chatID, err = a.chats.Save(ctx, chatID, name)
		if err != nil {
			return nil, err
		}
	}

	newMsgs := msgs[len(messages):]

	err = a.chats.SaveMessages(ctx, chatID, newMsgs)
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
				return nil, err
			}
			message.FunctionResponses[c.Name] = toolRes
		}

		msgs = append(msgs, message)
		return a.sendMessage(ctx, msgs)
	}

	return msgs, nil
}
