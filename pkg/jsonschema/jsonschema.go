package jsonschema

type JSONSchema struct {
	Type             string                `json:"type"`
	Description      string                `json:"description,omitempty"`
	Properties       map[string]JSONSchema `json:"properties"`
	Required         []string              `json:"required"`
	PropertyOrdering []string              `json:"propertyOrdering"`
	Items            *JSONSchema           `json:"items"`
	Examples         []any                 `json:"examples,omitempty"`
}
