package chat

type Author string

const (
	AuthorUser   Author = "user"
	AuthorModel  Author = "model"
	AuthorSystem Author = "system"
)

type Chat struct {
	ID   string
	Name string
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
