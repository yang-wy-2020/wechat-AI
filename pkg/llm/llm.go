package llm

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Provider interface {
	Name() string
	Chat(messages []ChatMessage) (string, error)
}
