package bot

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/NagaseYami/telegram-bot/service"
	log "github.com/sirupsen/logrus"
	tb "gopkg.in/tucnak/telebot.v2"
)

type Bot struct {
	*Config
	TelegramBot     *tb.Bot
	saucenaoService *service.SaucenaoService
	ascii2dService  *service.Ascii2dService
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
	bot.TelegramBot, err = tb.NewBot(tb.Settings{
		// You can also set custom API URL.
		// If field is empty it equals to "https://api.telegram.org".
		URL:    "",
		Token:  bot.TelegramBotToken,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		log.Fatalf("Telegram Bot 初始化时发生错误：%s", err)
	}

	// Handle SaucenaoService
	if bot.SaucenaoConfig.Enable {
		bot.saucenaoService = &service.SaucenaoService{SaucenaoConfig: bot.SaucenaoConfig}
	}

	if bot.Ascii2dConfig.Enable {
		bot.ascii2dService = &service.Ascii2dService{Ascii2dConfig: bot.Ascii2dConfig}
		bot.ascii2dService.Init()
	}

	bot.TelegramBot.Handle(tb.OnPhoto, bot.feature(bot.saucenao, bot.SaucenaoConfig.Enable))
	bot.TelegramBot.Handle("/sauce", bot.feature(bot.saucenao, bot.SaucenaoConfig.Enable))
	bot.TelegramBot.Handle("/ascii2d", bot.feature(bot.ascii2d, bot.Ascii2dConfig.Enable))
	bot.TelegramBot.Handle("/dice", bot.feature(bot.dice, bot.DiceConfig.Enable))
}

func (bot *Bot) Start() {
	bot.TelegramBot.Start()
}

func (bot *Bot) feature(f func(*tb.Message), enable bool) func(message *tb.Message) {
	if enable {
		return f
	} else {
		return bot.featureDisabled
	}
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
		_, err := bot.TelegramBot.Reply(requestMessage, "需要图片")
		if err != nil {
			log.Error(err)
		}
		return ""
	}

	// Get photo file URL
	url, err := bot.TelegramBot.FileURLByID(fileID)
	if err != nil {
		log.Error(err)
		return ""
	}
	log.Debugf("成功获取文件ID%s的URL：%s", fileID, url)
	return url
}

func (bot *Bot) saucenao(m *tb.Message) {
	url := bot.getPhotoFileURL(m)
	if url == "" {
		return
	}

	msg, err := bot.TelegramBot.Reply(m, "SauceNAO搜索中...")
	if err != nil {
		log.Error(err)
		return
	}

	// Search on SauceNAO
	var result *service.SaucenaoResult
	result = bot.saucenaoService.Search(url)

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
		text = "搜索过于频繁，已达到30秒内搜索次数上限\nSauceNAO搜索失败\n将启动ascii2d搜索"
		go bot.ascii2d(m)
	} else if result.LongRemain <= 0 {
		text = "搜索过于频繁，已达到24小时内搜索次数上限\nSauceNAO搜索失败\n将启动ascii2d搜索"
		go bot.ascii2d(m)
	} else if len(result.SearchResult) != 0 {
		text = "SauceNAO搜索完毕"
		if result.Similarity <= bot.SaucenaoConfig.LowSimilarityWarningLevel {
			text += fmt.Sprintf("\n请注意，搜索结果的最低相似度为%g，低于阈值%g\n这说明搜索结果有可能不准确\n如果搜索结果不理想，请使用其他引擎搜索",
				result.Similarity, bot.SaucenaoConfig.LowSimilarityWarningLevel)
		}
	} else {
		text = fmt.Sprintf("SauceNAO搜索失败（搜索结果相似度过低）\n将启动ascii2d搜索")
		go bot.ascii2d(m)
	}

	msg, err = bot.TelegramBot.Edit(msg, text, selector)
	if err != nil {
		log.Error(err)
		return
	}
}

func (bot *Bot) ascii2d(m *tb.Message) {
	url := bot.getPhotoFileURL(m)
	if url == "" {
		return
	}

	msg, err := bot.TelegramBot.Reply(m, "ascii2d搜索中...")

	result := bot.ascii2dService.Search(url)

	colorBtn := tb.Btn{
		Text: "ascii2d色合搜索结果",
		URL:  result.ColorURL,
	}

	bovwBtn := tb.Btn{
		Text: "ascii2d特征搜索结果",
		URL:  result.BovwURL,
	}

	colorSelector := &tb.ReplyMarkup{}
	colorSelector.Inline(colorSelector.Row(colorBtn))

	bovwSelector := &tb.ReplyMarkup{}
	bovwSelector.Inline(bovwSelector.Row(bovwBtn))

	bot.TelegramBot.Delete(msg)

	_, err = bot.TelegramBot.Reply(m, &tb.Photo{File: tb.FromURL(result.ColorThumbnail)}, colorSelector)
	if err != nil {
		log.Error(err)
	}

	_, err = bot.TelegramBot.Reply(m, &tb.Photo{File: tb.FromURL(result.BovwThumbnail)}, bovwSelector)
	if err != nil {
		log.Error(err)
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
	_, err := bot.TelegramBot.Reply(requestMessage, "该功能未启动，请联系管理员")
	if err != nil {
		log.Error(err)
	}
}
