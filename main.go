package main

import (
	"github.com/NagaseYami/saucenao-telegram-bot/bot"
	flag "github.com/spf13/pflag"
)

func main() {
	configFileFlag := flag.String("config", "", "Config file path.")
	flag.Parse()

	config := bot.LoadConfig(*configFileFlag)
	bot.Instance = bot.NewBot(config)
}
