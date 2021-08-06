package bot

import (
	"fmt"
	"math"
	"os"
	"time"

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

	bot.Handle(tb.OnPhoto, ReverseImageSearch)
	bot.Handle("/sauce", ReverseImageSearch)

	bot.Start()
}

func ReverseImageSearch(m *tb.Message) {
	msg, err := bot.Reply(m, "搜索中...")
	if err != nil {
		log.Fatalln(err)
	}

	// Get photo file ID
	var fileID string
	if m.Photo != nil {
		fileID = m.Photo.FileID
	} else if m.IsReply() && m.ReplyTo.Photo != nil {
		fileID = m.ReplyTo.Photo.FileID
	}

	if fileID == "" {
		bot.Reply(m, "需要图片")
		return
	}

	// Get photo file URL
	fileURL, err := bot.FileURLByID(fileID)
	if err != nil {
		log.Fatal(err)
		return
	}

	// Search on SauceNAO
	header, results := saucenao.Search(fileURL)

	text := fmt.Sprintf("API 30s 搜索次数限制 : %s/%s\nAPI 24h 搜索次数限制 : %s/%s", header.ShortRemain, header.ShortLimit, header.LongRemain, header.LongLimit)

	var selector = &tb.ReplyMarkup{}
	if len(results) != 0 {

		var btns []tb.Btn
		for _, result := range results {
			for i := 0; i < len(result.Databases); i++ {
				btns = append(btns, tb.Btn{
					Text: result.Databases[i],
					URL:  result.URLs[i],
				})
			}
		}

		var rows []tb.Row
		for i := 0; i < int(math.Ceil(float64(len(btns)/3.0))); i++ {
			if len(btns)-(i+1)*3 < 0 {
				rows = append(rows, selector.Row(btns[i*3:]...))
			} else {
				rows = append(rows, selector.Row(btns[i*3:i*3+3]...))
			}
		}

		selector.Inline(rows...)
	} else {
		text += "\n无结果（搜索结果相似度均低于80）"
	}
	_, err = bot.Edit(msg, text, selector)
	if err != nil {
		log.Fatal(err)
	}
}
