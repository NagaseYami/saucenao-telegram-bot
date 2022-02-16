package service

type QRConfig struct {
	Enable bool `yaml:"Enable"`
}

type QRService struct {
	*QRConfig
}
