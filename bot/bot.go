package bot

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/NagaseYami/saucenao-telegram-bot/service/ascii2d"
	"github.com/NagaseYami/saucenao-telegram-bot/service/dice"
	"github.com/NagaseYami/saucenao-telegram-bot/service/saucenao"
	log "github.com/sirupsen/logrus"
	tb "gopkg.in/tucnak/telebot.v2"
)

type Bot struct {
	*Config
	TelegramBot *tb.Bot
}

var Instance *Bot

func NewBot(config *Config) *Bot {
	bot := &Bot{
		Config: config,
	}

	bot.Init()
	return bot
}

func (bot *Bot) Init() {
	var err error

	// TelegramBot初始化
	bot.TelegramBot, err = tb.NewBot(tb.Settings{
		// You can also set custom API URL.
		// If field is empty it equals to "https://api.telegram.org".
		URL:    "",
		Token:  bot.TelegramBotToken,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		log.Fatal(err)
		return
	}

	//Handle Service
	if bot.SaucenaoConfig.Enable {
		saucenao.Instance = &saucenao.Service{Config: bot.SaucenaoConfig}

	}

	if bot.Ascii2dConfig.Enable {
		ascii2d.Instance = &ascii2d.Service{Config: bot.Ascii2dConfig}
	}

	if bot.DiceConfig.Enable {
		dice.Instance = &dice.Service{Config: bot.DiceConfig}

	}

	bot.TelegramBot.Handle(tb.OnPhoto, func(m *tb.Message) {
		if bot.SaucenaoConfig.Enable {
			go bot.saucenao(m)
		} else {
			go bot.featureDisabled(m)
		}
	})
	bot.TelegramBot.Handle("/sauce", func(m *tb.Message) {
		if bot.SaucenaoConfig.Enable {
			go bot.saucenao(m)
		} else {
			go bot.featureDisabled(m)
		}
	})
	bot.TelegramBot.Handle("/dice", func(m *tb.Message) {
		if bot.DiceConfig.Enable {
			go bot.dice(m)
		} else {
			go bot.featureDisabled(m)
		}
	})

	bot.TelegramBot.Start()
}

func (bot *Bot) saucenao(requestMessage *tb.Message) {
	var msg *tb.Message
	var err error

	// Get photo file ID
	var fileID string
	if requestMessage.Photo != nil {
		fileID = requestMessage.Photo.FileID
	} else if requestMessage.IsReply() && requestMessage.ReplyTo.Photo != nil {
		fileID = requestMessage.ReplyTo.Photo.FileID
	}

	if fileID == "" {
		msg, err = bot.TelegramBot.Reply(requestMessage, "需要图片")
		if err != nil {
			log.Warn(err)
			return
		}
		go func() {
			time.Sleep(bot.DeleteMessageInterval)
			bot.TelegramBot.Delete(msg)
		}()
		return
	}

	msg, err = bot.TelegramBot.Reply(requestMessage, "SauceNAO搜索中...")
	if err != nil {
		log.Warn(err)
		return
	}

	// Get photo file URL
	var fileURL string
	fileURL, err = bot.TelegramBot.FileURLByID(fileID)
	if err != nil {
		log.Error(err)
		return
	}

	// Search on SauceNAO
	var result *saucenao.Result
	result, err = saucenao.Instance.Search(fileURL)

	if err != nil {
		log.Error(err)
		return
	}

	msg, err = bot.TelegramBot.Edit(msg, result.Text, result.URLSelector)
	if err != nil {
		log.Warn(err)
		return
	}

	if !result.Success {
		go func() {
			time.Sleep(bot.DeleteMessageInterval)
			bot.TelegramBot.Delete(msg)
		}()

		if bot.Ascii2dConfig.Enable {
			go bot.ascii2d(requestMessage, fileURL)
		}
	}
}

func (bot *Bot) ascii2d(requestMessage *tb.Message, fileURL string) {

	msg, err := bot.TelegramBot.Reply(requestMessage, "ascii2d搜索中...")
	if err != nil {
		log.Warn(err)
		return
	}

	var result *ascii2d.Result
	result, err = ascii2d.Instance.Search(fileURL)
	if err != nil {
		log.Error(err)
		return
	}

	// ascii2d搜索无结果，通常这是你的访问IP在ascii2d的黑名单里所导致的
	if result == nil {
		msg, err = bot.TelegramBot.Edit(msg, "ascii2d搜索失败\n这很有可能是因为网络问题导致的，请联系管理员")
		if err != nil {
			log.Warn(err)
			return
		}
		go func() {
			time.Sleep(bot.DeleteMessageInterval)
			bot.TelegramBot.Delete(msg)
		}()
	} else {
		_, err = bot.TelegramBot.Reply(requestMessage, result.Photo, result.URLSelector)
		if err != nil {
			log.Warn(err)
			return
		}
	}
}

func (bot *Bot) dice(m *tb.Message) {
	var err error
	cmd := strings.Split(strings.ToLower(m.Payload), " ")[0]

	if cmd == "" {
		cmd = "1d6"
	}

	s := strings.Split(cmd, "d")
	if len(s) == 2 {
		if s[0] == "" {
			s[0] = "1"
		}
		num, err := strconv.ParseInt(s[0], 10, 64)
		if num > 100 {
			_, err = bot.TelegramBot.Reply(m, "为了保证机器人不会炸掉，请控制投掷次数≤100次")
			if err != nil {
				log.Error(err)
			}
			return
		}
		if err == nil {
			face, err := strconv.ParseInt(s[1], 10, 64)
			if face > 10000 {
				_, err = bot.TelegramBot.Reply(m, "为了保证机器人不会炸掉，请控制骰子面数≤10000")
				if err != nil {
					log.Error(err)
				}
				return
			}
			if err == nil {
				rand.Seed(time.Now().UnixNano())
				var results []int64
				var sum int64
				for i := num; i != 0; i-- {
					n := rand.Int63n(face)
					results = append(results, n)
					sum += n
				}
				_, err = bot.TelegramBot.Reply(m, fmt.Sprintf("投掷D%d骰子%d次的结果为%d\n最终合计值为%d", face, num, results, sum))
				if err != nil {
					log.Error(err)
				}
				return
			}
		}
	}

	_, err = bot.TelegramBot.Reply(m, "格式不正确，正确用法例：「/dice 1d6」")
	if err != nil {
		log.Error(err)
	}
}

func (bot *Bot) featureDisabled(requestMessage *tb.Message) {
	msg, err := bot.TelegramBot.Reply(requestMessage, "该功能未启动，请联系管理员")
	if err != nil {
		log.Warn(err)
	}
	go func() {
		time.Sleep(bot.DeleteMessageInterval)
		bot.TelegramBot.Delete(requestMessage)
		bot.TelegramBot.Delete(msg)
	}()
}
