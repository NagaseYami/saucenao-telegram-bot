package bot

import (
	"errors"
	"os"

	"github.com/NagaseYami/telegram-bot/service"
	"gopkg.in/yaml.v2"

	log "github.com/sirupsen/logrus"
)

type Config struct {
	DebugMode        bool                  `yaml:"DebugMode"`
	TelegramBotToken string                `yaml:"TelegramBotToken"`
	DiceConfig       *service.DiceConfig   `yaml:"DiceConfig"`
	QRConfig         *service.QRConfig     `yaml:"QRConfig"`
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
		},
		DiceConfig: &service.DiceConfig{Enable: true},
		QRConfig:   &service.QRConfig{Enable: true},
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
