package provider

import (
	"context"
	"errors"
	"fmt"
	"github.com/sashabaranov/go-openai"
	log "github.com/sirupsen/logrus"
	"io"
)

type ChatGPTService interface {
	CreateChatCompletion(ctx context.Context, req ChatCompletionRequest) (*ChatCompletionResponse, error)
	CreateChatCompletionStream(ctx context.Context, req ChatCompletionRequest, send chan<- ChatCompletionStreamResponse) error
}

type chatGPTServiceImpl struct {
	apiKey string
	model  string
}

func NewChatGPTService(apiKey, model string) ChatGPTService {
	return &chatGPTServiceImpl{apiKey: apiKey, model: model}
}

func (s *chatGPTServiceImpl) CreateChatCompletionStream(ctx context.Context,
	req ChatCompletionRequest, send chan<- ChatCompletionStreamResponse) error {
	log.Trace("in CreateChatCompletionStream")
	defer close(send)

	c := openai.NewClient(s.apiKey)

	chatReq := openai.ChatCompletionRequest{
		Model:    s.model,
		Stream:   true,
		Messages: getChatCompletionMessages(req.Messages),
	}
	stream, err := c.CreateChatCompletionStream(ctx, chatReq)
	if err != nil {
		return errors.New(fmt.Sprintf("chat completion stream error: %v\n", err))
	}
	defer stream.Close()

	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			log.Trace("stream closed")
			return nil
		}
		if err != nil {
			return errors.New(fmt.Sprintf("chat completion stream error: %v\n", err))
		}
		choices := make([]ChatCompletionStreamChoice, 0, len(response.Choices))
		for _, choice := range response.Choices {
			choices = append(choices, ChatCompletionStreamChoice{
				Index: choice.Index,
				Delta: ChatCompletionStreamChoiceDelta{
					Content: choice.Delta.Content,
				},
				FinishReason: string(choice.FinishReason),
			})
		}
		if len(choices) > 0 {
			send <- ChatCompletionStreamResponse{
				ID:      response.ID,
				Object:  response.Object,
				Created: response.Created,
				Model:   response.Model,
				Choices: choices,
			}
		}
	}
}

func (s *chatGPTServiceImpl) CreateChatCompletion(ctx context.Context, req ChatCompletionRequest) (*ChatCompletionResponse, error) {
	c := openai.NewClient(s.apiKey)

	resp, err := c.CreateChatCompletion(
		ctx,
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
