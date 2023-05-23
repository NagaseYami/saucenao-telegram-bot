package bot

import (
	"strings"
	"time"

	"github.com/sashabaranov/go-openai"
	log "github.com/sirupsen/logrus"
	tele "gopkg.in/telebot.v3"
	"telegram-bot/service"
)

type Bot struct {
	*Config
	TelegramBot *tele.Bot
}

func NewBot(config *Config) *Bot {
	bot := &Bot{
		Config: config,
	}
	return bot
}

func (bot *Bot) Init() {
	var err error

	if bot.TelegramBotToken == "" {
		log.Fatal("缺少Telegram Bot Token，启动失败")
	}

	// TelegramBot初始化
	bot.TelegramBot, err = tele.NewBot(tele.Settings{
		// You can also set custom API URL.
		// If field is empty it equals to "https://api.telegram.org".
		URL:    "",
		Token:  bot.TelegramBotToken,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		log.Fatalf("Telegram Bot 初始化时发生错误：%s", err)
	}

	if bot.OpenAIConfig.Enable {
		service.OpenAIInstance.Init(bot.OpenAIConfig)
	}

	bot.TelegramBot.Handle(tele.OnText, bot.feature(bot.startChatGPTByReply, bot.OpenAIConfig.Enable))
	bot.TelegramBot.Handle("/chatgpt", bot.feature(bot.createTalk, bot.OpenAIConfig.Enable))
}

func (bot *Bot) Start() {
	bot.TelegramBot.Start()
}

func (bot *Bot) feature(f tele.HandlerFunc, enable bool) tele.HandlerFunc {
	if enable {
		return f
	} else {
		return bot.featureDisabled
	}
}

func (bot *Bot) startChatGPTByReply(c tele.Context) error {
	if c.Message().IsReply() {
		talk := service.OpenAIInstance.GetTalkByMessageID(c.Message().ReplyTo.ID)
		if talk != nil {
			text := c.Message().Text
			if strings.ReplaceAll(strings.ReplaceAll(text, " ", ""), "　", "") == "" {
				text = "你好。你是谁？你能做些什么？"
			}
			talk.Messages = append(talk.Messages, struct {
				IsUser    bool
				MessageID int
				Message   string
			}{
				IsUser:    true,
				MessageID: c.Message().ID,
				Message:   text,
			})
			talk.LastUsedAt = time.Now().Unix()
			return bot.chat(c, talk)
		}
		bot.TelegramBot.Reply(c.Message(), "抱歉，出于技术原因，我不记得这段对话了，请开始一段新的对话")
	}
	return nil
}

func (bot *Bot) createTalk(c tele.Context) error {
	text := strings.Replace(c.Message().Text, "/chatgpt", "", 1)

	if strings.ReplaceAll(strings.ReplaceAll(text, " ", ""), "　", "") == "" {
		text = "你好。你是谁？你能做些什么？"
	}

	talk := &service.OpenAIChatGPTTalk{
		Messages: []struct {
			IsUser    bool
			MessageID int
			Message   string
		}{
			{IsUser: true, MessageID: c.Message().ID, Message: text},
		},
		LastUsedAt: time.Now().Unix(),
	}

	service.OpenAIInstance.AddTalk(talk)

	return bot.chat(c, talk)
}

func (bot *Bot) chat(c tele.Context, talk *service.OpenAIChatGPTTalk) error {
	r, err := bot.TelegramBot.Reply(c.Message(), "请等待...")
	var chatCompletionMessages []openai.ChatCompletionMessage
	for _, msg := range talk.Messages {
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
	resp, err := service.OpenAIInstance.ChatCompletion(chatCompletionMessages)
	if err != nil {
		bot.TelegramBot.Send(c.Recipient(), err)
		return err
	}
	r, err = bot.TelegramBot.Edit(r, resp)
	if err != nil {
		bot.TelegramBot.Send(c.Recipient(), err)
		return err
	}
	talk.Messages = append(talk.Messages, struct {
		IsUser    bool
		MessageID int
		Message   string
	}{
		IsUser:    false,
		MessageID: r.ID,
		Message:   resp,
	})

	return err
}

func (bot *Bot) featureDisabled(c tele.Context) error {
	_, err := bot.TelegramBot.Reply(c.Message(), "该功能未启动，请联系管理员")
	return err
}
