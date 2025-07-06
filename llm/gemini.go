package llm

import (
	"context"
	"log/slog"

	"github.com/aliphe/skipery/agent"
	"github.com/aliphe/skipery/pkg/jsonschema"
	"github.com/aliphe/skipery/tool"
	"google.golang.org/genai"
)

type Gemini struct {
	cli *genai.Client
}

func NewGemini(cli *genai.Client) *Gemini {
	return &Gemini{
		cli: cli,
	}
}

func fromJSONSchema(sch jsonschema.JSONSchema) *genai.Schema {
	props := make(map[string]*genai.Schema)
	for k, prop := range sch.Properties {
		props[k] = fromJSONSchema(prop)
	}

	return &genai.Schema{
		Type:             genai.TypeObject,
		Properties:       props,
		Required:         sch.Required,
		PropertyOrdering: sch.PropertyOrdering,
	}
}

func fromToolBelt(tb tool.ToolBelt) ([]*genai.Tool, error) {
	var tools []*genai.Tool
	for _, tool := range tb {
		for _, fct := range tool.Functions() {
			tools = append(tools, &genai.Tool{
				FunctionDeclarations: []*genai.FunctionDeclaration{
					{
						Name:        fct.ID,
						Description: fct.Description,
						Parameters:  fromJSONSchema(fct.Parameters),
					},
				},
			})
		}
	}
	return tools, nil
}

func fromConversation(conversation []*agent.Message) []*genai.Content {
	var parts []*genai.Content
	for _, msg := range conversation {
		if len(msg.FunctionCalls) != 0 {
			for _, call := range msg.FunctionCalls {
				parts = append(parts, &genai.Content{
					Parts: []*genai.Part{
						{
							Text: msg.Text,
							FunctionCall: &genai.FunctionCall{
								Name: call.Name,
								Args: call.Args,
							},
						},
					},
				})
			}
		} else if len(msg.FunctionResponses) != 0 {
			for name, response := range msg.FunctionResponses {
				parts = append(parts, &genai.Content{
					Parts: []*genai.Part{
						{
							Text: msg.Text,
							FunctionResponse: &genai.FunctionResponse{
								Name:     name,
								Response: response,
							},
						},
					},
				})
			}
		} else {
			parts = append(parts, &genai.Content{
				Parts: []*genai.Part{
					{
						Text: msg.Text,
					},
				},
				Role: string(msg.Author),
			})
		}
	}
	return parts
}

func (g *Gemini) SendMessage(ctx context.Context, tb tool.ToolBelt, conversation []*agent.Message) (*agent.Message, error) {
	slog.Debug("generating content", "conversation", conversation)
	history := fromConversation(conversation)
	tools, err := fromToolBelt(tb)
	if err != nil {
		return nil, err
	}
	content, err := g.cli.Models.GenerateContent(
		ctx,
		"gemini-2.5-flash-lite-preview-06-17",
		history,
		&genai.GenerateContentConfig{Tools: tools},
	)
	if err != nil {
		return nil, err
	}
	slog.Debug("received response from gemini", "content", content)

	res := &agent.Message{
		Author: agent.AuthorModel,
		Text:   content.Text(),
	}

	for _, fc := range content.FunctionCalls() {
		res.FunctionCalls = append(res.FunctionCalls, agent.FunctionCall{
			Name: fc.Name,
			Args: fc.Args,
		})
	}

	return res, nil
}
