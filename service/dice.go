package service

type DiceConfig struct {
	Enable bool `yaml:"Enable"`
}

type DiceService struct {
	*DiceConfig
}
