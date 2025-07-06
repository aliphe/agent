package jsonschema

type JSONSchema struct {
	Type             string                `json:"type"`
	Properties       map[string]JSONSchema `json:"properties"`
	Required         []string              `json:"required"`
	PropertyOrdering []string              `json:"propertyOrdering"`
}
