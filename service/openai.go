package service

import (
	"context"
	"errors"
	"io"
	"strconv"
	"time"

	"github.com/sashabaranov/go-openai"
	log "github.com/sirupsen/logrus"
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

func (service *OpenAIService) ChatStreamCompletion(messages []openai.ChatCompletionMessage, onResp func(string, bool),
	onFail func(error), retry int) {

	if retry > 5 {
		onFail(errors.New("失败重试次数过多，请稍后重试或联系管理员检查Log"))
		return
	}

	req := openai.ChatCompletionRequest{
		Model:    openai.GPT3Dot5Turbo0301,
		Messages: messages,
		Stream:   true,
	}
	stream, err := service.client.CreateChatCompletionStream(service.clientCtx, req)
	e := &openai.APIError{}
	if errors.As(err, &e) {
		onFail(errors.New("遇到API错误，正在重试。重试次数：" + strconv.Itoa(retry)))
		service.ChatStreamCompletion(messages, onResp, onFail, retry+1)
		return
	}
	defer stream.Close()

	startTime := time.Now().Unix()
	stackResp := ""
	finished := false
	for {
		response, err := stream.Recv()
		log.Debug(response)

		if err != nil {
			if errors.As(err, &e) {
				onFail(errors.New("遇到API错误，正在重试。重试次数：" + strconv.Itoa(retry)))
				service.ChatStreamCompletion(messages, onResp, onFail, retry+1)
				return
			}

			if errors.Is(err, io.EOF) {
				finished = true
			}
		}

		if finished {
			onResp(stackResp, true)
			break
		} else {
			stackResp += response.Choices[0].Delta.Content
		}

		if time.Now().Unix()-startTime >= 3 && stackResp != "" {
			startTime = time.Now().Unix()
			onResp(stackResp, finished)
			stackResp = ""
		}
	}
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
