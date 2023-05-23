package bot

import (
	"fmt"
	"image"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/imroc/req"
	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
	log "github.com/sirupsen/logrus"
	tele "gopkg.in/telebot.v3"
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

	bot.TelegramBot.Handle("/dice", bot.feature(bot.dice, bot.DiceConfig.Enable))
	bot.TelegramBot.Handle("/qrdecode", bot.feature(bot.decodeQRCode, bot.QRConfig.Enable))
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

func (bot *Bot) getPhotoFileURL(m *tele.Message) string {
	// Get photo file ID
	var fileID string
	if m.Photo != nil {
		fileID = m.Photo.FileID
	} else if m.IsReply() && m.ReplyTo.Photo != nil {
		fileID = m.ReplyTo.Photo.FileID
	}

	if fileID == "" {
		_, err := bot.TelegramBot.Reply(m, "需要图片")
		if err != nil {
			log.Error(err)
		}
		return ""
	}

	// Get photo file URL
	file, err := bot.TelegramBot.FileByID(fileID)
	if err != nil {
		log.Error(err)
		return ""
	}
	log.Debugf("成功获取文件ID%s的URL：%s", fileID, file.FileURL)
	return file.FileURL
}

func (bot *Bot) dice(c tele.Context) error {
	var err error
	cmd := strings.Split(strings.ToLower(c.Message().Payload), " ")[0]

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
			_, err = bot.TelegramBot.Reply(c.Message(), "为了保证机器人不会炸掉，请控制投掷次数≤100次")
			if err != nil {
				return err
			}

		}
		if err == nil {
			face, err := strconv.ParseInt(s[1], 10, 64)
			if face > 10000 {
				_, err = bot.TelegramBot.Reply(c.Message(), "为了保证机器人不会炸掉，请控制骰子面数≤10000")
				if err != nil {
					return err
				}

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
				_, err = bot.TelegramBot.Reply(c.Message(), fmt.Sprintf("投掷D%d骰子%d次的结果为%d\n最终合计值为%d", face, num, results,
					sum))
				if err != nil {
					return err
				}

			}
		}
	}

	_, err = bot.TelegramBot.Reply(c.Message(), "格式不正确，正确用法例：「/dice 1d6」")
	return err
}

func (bot *Bot) decodeQRCode(c tele.Context) error {
	url := bot.getPhotoFileURL(c.Message())
	if url == "" {
		return nil
	}

	msg, err := bot.TelegramBot.Reply(c.Message(), "正在分析图片...")
	if err != nil {
		return err
	}

	resp, err := req.Get(url)
	if err != nil {
		log.Errorf("获取二维码图片时发生错误：%s\n", err.Error())
		return err
	}
	defer resp.Response().Body.Close()

	img, _, err := image.Decode(resp.Response().Body)
	if err != nil {
		log.Errorf("Decode二维码图片时发生错误：%s\n", err.Error())
		return err
	}

	bmp, err := gozxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		log.Errorf("二维码图片转换为Bitmap时发生错误：%s\n", err.Error())
		return err
	}
	qrReader := qrcode.NewQRCodeReader()
	result, _ := qrReader.Decode(bmp, nil)

	if result == nil || result.String() == "" {
		bot.TelegramBot.Edit(msg, "图片中未发现二维码，分析失败")
		return err
	}

	_, err = bot.TelegramBot.Edit(msg, fmt.Sprintf("二维码分析结果：\n%s", result.String()))
	return err
}

func (bot *Bot) featureDisabled(c tele.Context) error {
	_, err := bot.TelegramBot.Reply(c.Message(), "该功能未启动，请联系管理员")
	return err
}
