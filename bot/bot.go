package bot

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/NagaseYami/saucenao-telegram-bot/service/ascii2d"
	"github.com/NagaseYami/saucenao-telegram-bot/service/saucenao"
	log "github.com/sirupsen/logrus"
	tb "gopkg.in/tucnak/telebot.v2"
)

type Bot struct {
	*Config
	TelegramBot     *tb.Bot
	saucenaoService *saucenao.Service
	ascii2dService  *ascii2d.Service
}

func NewBot(config *Config) *Bot {
	bot := &Bot{
		Config: config,
	}

	bot.Init()
	return bot
}

func (bot *Bot) Init() {
	var err error

	if bot.TelegramBotToken == "" {
		log.Fatal("缺少Telegram Bot Token，启动失败")
	}

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

	// Handle Service
	if bot.SaucenaoConfig.Enable {
		bot.saucenaoService = &saucenao.Service{Config: bot.SaucenaoConfig}
	}

	if bot.Ascii2dConfig.Enable {
		bot.ascii2dService = &ascii2d.Service{Config: bot.Ascii2dConfig}
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
	bot.TelegramBot.Handle("/ascii2d", func(m *tb.Message) {
		if bot.SaucenaoConfig.Enable {
			go bot.ascii2d(m)
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
}

func (bot *Bot) Start() {
	log.Info("幾重にも辛酸を舐め、七難八苦を超え、艱難辛苦の果て、満願成就に至る。")
	bot.TelegramBot.Start()
}

func (bot *Bot) getPhotoFileURL(requestMessage *tb.Message) string {
	// Get photo file ID
	var fileID string
	if requestMessage.Photo != nil {
		fileID = requestMessage.Photo.FileID
	} else if requestMessage.IsReply() && requestMessage.ReplyTo.Photo != nil {
		fileID = requestMessage.ReplyTo.Photo.FileID
	}

	if fileID == "" {
		msg, err := bot.TelegramBot.Reply(requestMessage, "需要图片")
		if err != nil {
			log.Warn(err)
		}
		go func() {
			time.Sleep(bot.DeleteMessageInterval)
			err := bot.TelegramBot.Delete(msg)
			if err != nil {
				log.Warn(err)
			}
		}()

		return ""
	}

	// Get photo file URL
	url, err := bot.TelegramBot.FileURLByID(fileID)
	if err != nil {
		log.Warn(err)
	}
	return url
}

func (bot *Bot) saucenao(requestMessage *tb.Message) {
	var msg *tb.Message
	var err error

	url := bot.getPhotoFileURL(requestMessage)

	msg, err = bot.TelegramBot.Reply(requestMessage, "SauceNAO搜索中...")
	if err != nil {
		log.Warn(err)
		return
	}

	// Search on SauceNAO
	var result *saucenao.Result
	result, err = bot.saucenaoService.Search(url)

	if err != nil {
		log.Error(err)
		return
	}

	selector := &tb.ReplyMarkup{}
	var buttons []tb.Btn
	for key, value := range result.SearchResult {
		buttons = append(buttons, tb.Btn{
			Text: key,
			URL:  value,
		})
	}
	var rows []tb.Row
	for i := 0; i < int(math.Ceil(float64(len(buttons))/3.0)); i++ {
		if len(buttons)-(i+1)*3 < 0 {
			rows = append(rows, selector.Row(buttons[i*3:]...))
		} else {
			rows = append(rows, selector.Row(buttons[i*3:i*3+3]...))
		}
	}
	selector.Inline(rows...)

	var text string

	if result.ShortRemain <= 0 {
		text = "搜索过于频繁，已达到30秒内搜索次数上限\nSauceNAO搜索失败"
	} else if result.LongRemain <= 0 {
		text = "搜索过于频繁，已达到24小时内搜索次数上限\nSauceNAO搜索失败"
	} else if len(result.SearchResult) != 0 {
		text = "SauceNAO搜索完毕"
	} else {
		text = fmt.Sprintf("SauceNAO搜索失败（搜索结果相似度均低于%g）", bot.SaucenaoConfig.Similarity)

		go func() {
			time.Sleep(bot.DeleteMessageInterval)
			err := bot.TelegramBot.Delete(msg)
			if err != nil {
				log.Warn(err)
			}
		}()

		if bot.Ascii2dConfig.Enable {
			go bot.ascii2dWithFileURL(requestMessage, url)
		}
	}

	msg, err = bot.TelegramBot.Edit(msg, text, selector)
	if err != nil {
		log.Warn(err)
		return
	}
}

func (bot *Bot) ascii2d(requestMessage *tb.Message) {
	url := bot.getPhotoFileURL(requestMessage)
	bot.ascii2dWithFileURL(requestMessage, url)
}

func (bot *Bot) ascii2dWithFileURL(requestMessage *tb.Message, fileURL string) {
	msg, err := bot.TelegramBot.Reply(requestMessage, "ascii2d搜索中...")
	if err != nil {
		log.Error(err)
		return
	}

	var result *ascii2d.Result
	result, err = bot.ascii2dService.Search(fileURL)
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
			err := bot.TelegramBot.Delete(msg)
			if err != nil {
			}
		}()
	} else {
		photo := &tb.Photo{File: tb.FromURL(result.ThumbnailURL)}
		selector := &tb.ReplyMarkup{}
		selector.Inline(tb.Row{
			tb.Btn{
				Text: "ascii2d搜索结果",
				URL:  result.ImageURL,
			},
		})
		err = bot.TelegramBot.Delete(msg)
		if err != nil {
			log.Warn(err)
		}
		_, err = bot.TelegramBot.Reply(requestMessage, photo, selector)
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
		err := bot.TelegramBot.Delete(requestMessage)
		if err != nil {
			log.Warn(err)
		}
		err = bot.TelegramBot.Delete(msg)
		if err != nil {
			log.Warn()
		}
	}()
}
