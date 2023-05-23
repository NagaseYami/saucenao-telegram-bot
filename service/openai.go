package service

import (
	"context"

	"github.com/sashabaranov/go-openai"
)

type OpenAIConfig struct {
	Enable bool   `yaml:"Enable"`
	Token  string `yaml:"Token"`
}

type OpenAIService struct {
	*OpenAIConfig
	client    *openai.Client
	clientCtx context.Context
	talks     []*OpenAIChatGPTTalk
}

type OpenAIChatGPTTalk struct {
	Messages []struct {
		IsUser    bool
		MessageID int
		Message   string
	}
	LastUsedAt int64
}

var OpenAIInstance = NewOpenAIService()

func NewOpenAIService() *OpenAIService {
	return &OpenAIService{
		OpenAIConfig: nil,
		client:       nil,
		clientCtx:    nil,
	}
}

func (service *OpenAIService) Init(conf *OpenAIConfig) {
	service.OpenAIConfig = conf
	service.client = openai.NewClient(service.OpenAIConfig.Token)
	service.clientCtx = context.Background()
}

func (service *OpenAIService) AddTalk(talk *OpenAIChatGPTTalk) {
	service.talks = append(service.talks, talk)
}

func (service *OpenAIService) GetTalkByMessageID(messageId int) *OpenAIChatGPTTalk {
	for _, t := range service.talks {
		for _, m := range t.Messages {
			if m.MessageID == messageId {
				return t
			}
		}
	}
	return nil
}

func (service *OpenAIService) ChatCompletion(messages []openai.ChatCompletionMessage) (string, error) {
	resp, err := service.client.CreateChatCompletion(
		service.clientCtx,
		openai.ChatCompletionRequest{
			Model:    openai.GPT3Dot5Turbo,
			Messages: messages,
		},
	)

	if err != nil {
		return "", err
	}

	return resp.Choices[0].Message.Content, err
}

func (service *OpenAIService) GenerateChatCompletionMessage(messages []struct {
	IsUser  bool
	Message string
}) ([]openai.ChatCompletionMessage, error) {
	var chatCompletionMessages []openai.ChatCompletionMessage
	for _, msg := range messages {
		if msg.IsUser {
			chatCompletionMessages = append(chatCompletionMessages, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleUser,
				Content: msg.Message,
			})
		} else {
			chatCompletionMessages = append(chatCompletionMessages, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleAssistant,
				Content: msg.Message,
			})
		}

	}
	return chatCompletionMessages, nil
}
