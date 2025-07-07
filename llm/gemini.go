package llm

import (
	"context"
	"log/slog"

	"github.com/aliphe/skipery/agent/chat"
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

	// Convert type string to genai type
	var schemaType genai.Type
	switch sch.Type {
	case "object":
		schemaType = genai.TypeObject
	case "array":
		schemaType = genai.TypeArray
	case "string":
		schemaType = genai.TypeString
	case "number":
		schemaType = genai.TypeNumber
	case "integer":
		schemaType = genai.TypeInteger
	case "boolean":
		schemaType = genai.TypeBoolean
	default:
		schemaType = genai.TypeObject
	}

	schema := &genai.Schema{
		Type:             schemaType,
		Description:      sch.Description,
		Properties:       props,
		Required:         sch.Required,
		PropertyOrdering: sch.PropertyOrdering,
	}

	if sch.Items != nil {
		schema.Items = fromJSONSchema(*sch.Items)
	}

	return schema
}

func fromToolBelt(tb tool.ToolBelt) ([]*genai.Tool, error) {
	toolMap := make(map[tool.Tool]bool)
	var tools []*genai.Tool

	for _, t := range tb {
		if toolMap[t] {
			continue // Skip duplicate tools
		}
		toolMap[t] = true

		var functionDeclarations []*genai.FunctionDeclaration
		for _, fct := range t.Functions() {
			functionDeclarations = append(functionDeclarations, &genai.FunctionDeclaration{
				Name:        fct.ID,
				Description: fct.Description,
				Parameters:  fromJSONSchema(fct.Parameters),
				Response:    fromJSONSchema(fct.Response),
			})
		}
		if len(functionDeclarations) > 0 {
			tools = append(tools, &genai.Tool{
				FunctionDeclarations: functionDeclarations,
			})
		}
	}
	return tools, nil
}

func fromChat(messages []*chat.Message) []*genai.Content {
	var parts []*genai.Content
	for _, msg := range messages {
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
			// Handle system messages by using user role (Gemini doesn't have system role)
			role := string(msg.Author)
			if msg.Author == chat.AuthorSystem {
				role = "user"
			}
			parts = append(parts, &genai.Content{
				Parts: []*genai.Part{
					{
						Text: msg.Text,
					},
				},
				Role: role,
			})
		}
	}
	return parts
}

func (g *Gemini) SendMessage(ctx context.Context, tb tool.ToolBelt, messages []*chat.Message) (*chat.Message, error) {
	slog.Debug("generating content", "chat", messages)
	history := fromChat(messages)
	tools, err := fromToolBelt(tb)
	if err != nil {
		return nil, err
	}
	content, err := g.cli.Models.GenerateContent(
		ctx,
		"gemini-2.0-flash",
		history,
		&genai.GenerateContentConfig{Tools: tools},
	)
	if err != nil {
		return nil, err
	}
	slog.Debug("received response from gemini", "content", content)

	res := &chat.Message{
		Author: chat.AuthorModel,
		Text:   content.Text(),
	}

	for _, fc := range content.FunctionCalls() {
		res.FunctionCalls = append(res.FunctionCalls, chat.FunctionCall{
			Name: fc.Name,
			Args: fc.Args,
		})
	}

	return res, nil
}
