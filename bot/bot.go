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
	tb        *tele.Bot
	whiteList []string
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
	bot.tb, err = tele.NewBot(tele.Settings{
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

	bot.tb.Handle(tele.OnText, bot.feature(bot.continueTalk, bot.OpenAIConfig.Enable))
	bot.tb.Handle("/"+bot.OpenAIConfig.Endpoint, bot.feature(bot.createTalk, bot.OpenAIConfig.Enable))
}

func (bot *Bot) Start() {
	bot.tb.Start()
}

func (bot *Bot) feature(f func(tele.Context, chan error), enable bool) tele.HandlerFunc {
	if enable {
		return func(context tele.Context) error {
			var ch = make(chan error)
			go f(context, ch)
			return <-ch
		}
	} else {
		return bot.featureDisabled
	}
}

func (bot *Bot) continueTalk(c tele.Context, ch chan error) {
	if c.Message().IsReply() {
		talk := service.OpenAIInstance.GetTalkByMessageID(c.Message().ReplyTo.ID)
		if talk != nil {
			text := c.Message().Text
			if strings.ReplaceAll(strings.ReplaceAll(text, " ", ""), "　", "") == "" {
				text = "你好"
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
			ch <- bot.chat(c, talk)
			return
		}
		bot.tb.Reply(c.Message(), "抱歉，我无法继续这段对话，建议您重新开启一段新的对话。\n"+
			"以下原因会导致该错误：\n"+
			"1. 重启Bot会清空所有对话\n"+
			"2. 您Reply的消息还在输出结果，请等到🔚出现后再回复。")
	}
	ch <- nil
}

func (bot *Bot) createTalk(c tele.Context, ch chan error) {
	text := c.Message().Text
	for _, e := range c.Message().Entities {
		if e.Type == tele.EntityCommand {
			entityText := c.Message().EntityText(e)
			text = strings.Replace(text, entityText, "", 1)
			break
		}
	}

	if strings.ReplaceAll(strings.ReplaceAll(text, " ", ""), "　", "") == "" {
		text = "你好"
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

	ch <- bot.chat(c, talk)
}

func (bot *Bot) chat(c tele.Context, talk *service.OpenAIChatGPTTalk) error {
	reply, err := bot.tb.Reply(c.Message(), "请等待...")
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

	service.OpenAIInstance.ChatCompletion(chatCompletionMessages, func(resp string) {
		replyText := resp + "🔚"
		bot.tb.Edit(reply, replyText)
		if resp != "" {
			talk.Messages = append(talk.Messages, struct {
				IsUser    bool
				MessageID int
				Message   string
			}{
				IsUser:    false,
				MessageID: reply.ID,
				Message:   resp,
			})
		}
	}, func(err error) {
		bot.tb.Edit(reply, err.Error())
	}, 0)

	// service.OpenAIInstance.ChatStreamCompletion(chatCompletionMessages, func(resp string, finished bool) {
	// 	result += resp
	// 	var replyText string
	// 	if finished {
	// 		replyText = result + "🔚"
	// 	} else {
	// 		replyText = result + "⏳"
	// 	}
	// 	bot.tb.Edit(reply, replyText)
	// 	if result != "" && finished {
	// 		talk.Messages = append(talk.Messages, struct {
	// 			IsUser    bool
	// 			MessageID int
	// 			Message   string
	// 		}{
	// 			IsUser:    false,
	// 			MessageID: reply.ID,
	// 			Message:   result,
	// 		})
	// 	}
	// }, func(err error) {
	// 	bot.tb.Edit(reply, err.Error())
	// }, 0)

	return err
}

func (bot *Bot) featureDisabled(c tele.Context) error {
	_, err := bot.tb.Reply(c.Message(), "该功能未启动，请联系管理员")
	return err
}
