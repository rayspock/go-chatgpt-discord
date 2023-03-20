package provider

import (
	"context"
	"errors"
	"fmt"
	"github.com/sashabaranov/go-openai"
)

type ChatGPTService interface {
	CreateChatCompletion(ctx context.Context, req ChatCompletionRequest) (resp *ChatCompletionResponse, err error)
}

type chatGPTServiceImpl struct {
	apiKey string
	model  string
}

func NewChatGPTService(apiKey, model string) ChatGPTService {
	return &chatGPTServiceImpl{apiKey: apiKey, model: model}
}

func (s *chatGPTServiceImpl) CreateChatCompletion(ctx context.Context, req ChatCompletionRequest) (*ChatCompletionResponse, error) {
	c := openai.NewClient(s.apiKey)

	resp, err := c.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    s.model,
			Messages: getChatCompletionMessages(req.Messages),
		},
	)

	if err != nil {
		return nil, errors.New(fmt.Sprintf("chat completion error: %v\n", err))
	}

	if len(resp.Choices) <= 0 {
		return nil, errors.New(fmt.Sprintf("empty choices for chat completion"))
	}

	return &ChatCompletionResponse{Content: resp.Choices[0].Message.Content}, nil
}

func getChatCompletionMessages(messages []ChatCompletionMessage) []openai.ChatCompletionMessage {
	result := make([]openai.ChatCompletionMessage, 0)
	for _, msg := range messages {
		result = append(result, openai.ChatCompletionMessage{
			Role:    string(msg.Role),
			Content: msg.Content,
		})
	}
	return result
}
