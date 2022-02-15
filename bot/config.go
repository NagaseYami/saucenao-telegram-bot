package bot

import (
	"errors"
	"os"

	"github.com/NagaseYami/saucenao-telegram-bot/service"
	"gopkg.in/yaml.v2"

	log "github.com/sirupsen/logrus"
)

type Config struct {
	DebugMode        bool                    `yaml:"DebugMode"`
	TelegramBotToken string                  `yaml:"TelegramBotToken"`
	SaucenaoConfig   *service.SaucenaoConfig `yaml:"SaucenaoConfig"`
	Ascii2dConfig    *service.Ascii2dConfig  `yaml:"Ascii2dConfig"`
	DiceConfig       *service.DiceConfig     `yaml:"DiceConfig"`
}

func LoadConfig(configFilePath string) *Config {
	var err error
	config := &Config{}

	if _, err = os.Stat(configFilePath); errors.Is(err, os.ErrNotExist) {
		config = NewConfig()
		CreateConfigFile(config)
		log.Info("没有找到配置文件，已在同目录自动生成，请编辑后再次启动")
		os.Exit(0)
	}

	var bytes []byte
	bytes, err = os.ReadFile(configFilePath)
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
		DebugMode:        false,
		TelegramBotToken: "",
		SaucenaoConfig: &service.SaucenaoConfig{
			Enable:                    false,
			ApiKey:                    "",
			LowSimilarityWarningLevel: 70,
		},
		Ascii2dConfig: &service.Ascii2dConfig{
			Enable:        false,
			TempDirectory: "./temp",
		},
		DiceConfig: &service.DiceConfig{Enable: true},
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
