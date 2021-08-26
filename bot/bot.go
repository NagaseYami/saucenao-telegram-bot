package bot

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/NagaseYami/saucenao-telegram-bot/ascii2d"
	"github.com/NagaseYami/saucenao-telegram-bot/saucenao"
	log "github.com/sirupsen/logrus"
	tb "gopkg.in/tucnak/telebot.v2"
)

var bot *tb.Bot
var botToken string

func Init() {
	var err error

	botToken = os.Getenv("BOT_TOKEN")
	if botToken == "" {
		log.Fatal("环境变量「BOT_TOKEN」缺失")
	}

	bot, err = tb.NewBot(tb.Settings{
		// You can also set custom API URL.
		// If field is empty it equals to "https://api.telegram.org".
		URL:    "",
		Token:  botToken,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})

	if err != nil {
		log.Fatal(err)
		return
	}

	bot.Handle(tb.OnPhoto, func(m *tb.Message) { go Saucenao(m) })
	bot.Handle("/sauce", func(m *tb.Message) { go Saucenao(m) })
	bot.Handle("/dice", func(m *tb.Message) { go Dice(m) })

	bot.Start()
}

func Dice(m *tb.Message) {
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
			_, err = bot.Reply(m, "为了保证机器人不会炸掉，请控制投掷次数≤100次")
			if err != nil {
				log.Error(err)
			}
			return
		}
		if err == nil {
			face, err := strconv.ParseInt(s[1], 10, 64)
			if face > 10000 {
				_, err = bot.Reply(m, "为了保证机器人不会炸掉，请控制骰子面数≤10000")
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
				_, err = bot.Reply(m, fmt.Sprintf("投掷D%d骰子%d次的结果为%d\n最终合计值为%d", face, num, results, sum))
				if err != nil {
					log.Error(err)
				}
				return
			}
		}
	}

	_, err = bot.Reply(m, "格式不正确，正确用法例：「/dice 1d6」")
	if err != nil {
		log.Error(err)
	}
}

func Saucenao(m *tb.Message) {
	var msg *tb.Message
	var err error

	// Get photo file ID
	var fileID string
	if m.Photo != nil {
		fileID = m.Photo.FileID
	} else if m.IsReply() && m.ReplyTo.Photo != nil {
		fileID = m.ReplyTo.Photo.FileID
	}

	if fileID == "" {
		_, err = bot.Reply(m, "需要图片")
		if err != nil {
			log.Error(err)
			return
		}
		return
	}

	msg, err = bot.Reply(m, "SauceNAO搜索中...")
	if err != nil {
		log.Error(err)
		return
	}

	// Get photo file URL
	var fileURL string
	fileURL, err = bot.FileURLByID(fileID)
	if err != nil {
		log.Error(err)
		return
	}

	// Search on SauceNAO
	var header saucenao.Header
	var results []saucenao.Result
	header, results, err = saucenao.Search(fileURL)

	if err != nil {
		log.Error(err)
		return
	}

	var text string
	var selector = &tb.ReplyMarkup{}
	var needAscii2d = false

	if header.ShortRemain <= 0 {
		text = "搜索过于频繁，已达到30秒内搜索次数上限\nSauceNAO搜索失败，将启用ascii2d搜索"
		needAscii2d = true
	} else if header.ShortRemain <= 0 {
		text = "搜索过于频繁，已达到24小时内搜索次数上限\nSauceNAO搜索失败，将启用ascii2d搜索"
		needAscii2d = true
	} else if len(results) != 0 {
		text = "SauceNAO搜索完毕"
		var buttons []tb.Btn
		for _, result := range results {
			buttons = append(buttons, tb.Btn{
				Text: result.Database,
				URL:  result.URL,
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
	} else {
		text += "\nSauceNAO搜索失败（搜索结果相似度均低于80）\n将自动启动ascii2d搜索"
		needAscii2d = true
	}
	_, err = bot.Edit(msg, text, selector)
	if err != nil {
		log.Error(err)
		return
	}

	if needAscii2d {
		go ascii2dSearch(m, fileURL)
	}
}

func ascii2dSearch(m *tb.Message, fileURL string) {

	msg, err := bot.Reply(m, "ascii2d搜索中...\nascii2d没有相似度检测，搜索结果不一定正确")
	if err != nil {
		log.Error(err)
		return
	}

	result, err := ascii2d.Search(fileURL)
	if err != nil {
		log.Error(err)
		return
	}
	if !result.Exist {
		_, err = bot.Edit(msg, "ascii2d搜索失败")
		if err != nil {
			log.Error(err)
			return
		}
	} else {
		var selector = &tb.ReplyMarkup{}
		selector.Inline(tb.Row{
			tb.Btn{
				Text: "ascii2d搜索结果",
				URL:  result.URL,
			},
		})
		err = bot.Delete(msg)
		if err != nil {
			log.Warn(err)
		}

		_, err = bot.Reply(m, &tb.Photo{File: tb.FromURL(result.ThumbnailURL)}, selector)
		if err != nil {
			log.Error(err)
			return
		}
	}
}
