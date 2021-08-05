package bot

import (
	"fmt"
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
		log.Fatal("Bot token not found.")
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
	msg, err := bot.Reply(m, "Searching...")
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
		bot.Reply(m, "Photo PLZ.")
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
	var selector = &tb.ReplyMarkup{}
	var btns []tb.Btn
	for _, result := range results {
		btns = append(btns, tb.Btn{
			Text: result.DataBaseName,
			URL:  result.URL,
		})
	}

	var rows []tb.Row
	for i := 0; i < len(btns)/3+1; i++ {
		if len(btns)-(i+1)*3 < 0 {
			rows = append(rows, selector.Row(btns[i*3:]...))
		} else {
			rows = append(rows, selector.Row(btns[i*3:i*3+3]...))
		}
	}

	selector.Inline(rows...)

	text := fmt.Sprintf("API 30s limit remianing : %s/%s\nAPI 24h limit remaining : %s/%s\nMinimum Similarity : %s", header.ShortRemain, header.ShortLimit, header.LongRemain, header.LongLimit, header.MinimumSimilarity)
	bot.Edit(msg, text, selector)
}
