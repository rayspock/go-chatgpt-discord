package provider

type ChatMessageRole string

// The Chat message role defined by the OpenAI API refers to the role that a message plays in a conversation between
// a user and an AI language model. The API allows developers to define the role of each message in a conversation,
// which can help the language model understand the context and intent of the conversation.
const (
	ChatMessageRoleSystem    ChatMessageRole = "system"
	ChatMessageRoleUser      ChatMessageRole = "user"
	ChatMessageRoleAssistant ChatMessageRole = "assistant"
)

type ChatCompletionMessage struct {
	Role    ChatMessageRole
	Content string
}

type ChatCompletionRequest struct {
	Messages []ChatCompletionMessage
}

type ChatCompletionResponse struct {
	Content string `json:"content"`
}

type ChatCompletionStreamChoiceDelta struct {
	Content string `json:"content"`
}

type ChatCompletionStreamChoice struct {
	Index        int                             `json:"index"`
	Delta        ChatCompletionStreamChoiceDelta `json:"delta"`
	FinishReason string                          `json:"finish_reason"`
}

type ChatCompletionStreamResponse struct {
	ID      string                       `json:"id"`
	Object  string                       `json:"object"`
	Created int64                        `json:"created"`
	Model   string                       `json:"model"`
	Choices []ChatCompletionStreamChoice `json:"choices"`
}
