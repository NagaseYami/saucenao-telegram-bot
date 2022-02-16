package main

import (
	"github.com/NagaseYami/telegram-bot/bot"
	"github.com/NagaseYami/telegram-bot/tool"
	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
)

func main() {
	configFileFlag := flag.StringP("config", "c", "config.yaml", "Config file path.")
	flag.Parse()

	config := bot.LoadConfig(*configFileFlag)
	log.Debug("读取配置文件成功")

	if config.DebugMode {
		log.SetLevel(log.DebugLevel)
		log.Debug("已开启Debug模式")
	}

	tool.Browser.Init()
	log.Debug("初始化Headless Browser成功")
	defer tool.Browser.UnInit()

	bot := bot.NewBot(config)
	bot.Init()
	log.Info("幾重にも辛酸を舐め、七難八苦を超え、艱難辛苦の果て、満願成就に至る。")
	bot.Start()
}
