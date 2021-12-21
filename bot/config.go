package bot

import (
	"os"
	"time"

	"github.com/NagaseYami/saucenao-telegram-bot/service/ascii2d"
	"github.com/NagaseYami/saucenao-telegram-bot/service/dice"
	"github.com/NagaseYami/saucenao-telegram-bot/service/saucenao"
	"gopkg.in/yaml.v2"

	log "github.com/sirupsen/logrus"
)

type Config struct {
	TelegramBotToken      string           `yaml:"TelegramBotToken"`
	DeleteMessageInterval time.Duration    `yaml:"DeleteMessageInterval"`
	SaucenaoConfig        *saucenao.Config `yaml:"SaucenaoConfig"`
	Ascii2dConfig         *ascii2d.Config  `yaml:"Ascii2dConfig"`
	DiceConfig            *dice.Config     `yaml:"DiceConfig"`
}

func LoadConfig(configFilePath string) *Config {
	config := &Config{}

	if configFilePath == "" {
		config = NewConfig()
		CreateConfigFile(config)
		return config
	}

	bytes, err := os.ReadFile(configFilePath)
	if err != nil {
		log.Error(err)
		return nil
	}

	err = yaml.Unmarshal(bytes, config)
	if err != nil {
		log.Error(err)
		return nil
	}

	return config
}

func NewConfig() *Config {
	return &Config{
		TelegramBotToken:      "",
		DeleteMessageInterval: 5 * time.Second,
		SaucenaoConfig: &saucenao.Config{
			Enable:     false,
			ApiKey:     "",
			Similarity: 80,
		},
		Ascii2dConfig: &ascii2d.Config{
			Enable:         true,
			TempFolderPath: "./temp",
		},
		DiceConfig: &dice.Config{Enable: true},
	}
}

func CreateConfigFile(config *Config) {
	file, err := os.Create("./config.yaml")
	if err != nil {
		log.Error(err)
		return
	}

	var bytes []byte
	bytes, err = yaml.Marshal(config)
	_, err = file.Write(bytes)
	if err != nil {
		log.Error(err)
		return
	}

	if err = file.Close(); err != nil {
		log.Error(err)
		return
	}
}
