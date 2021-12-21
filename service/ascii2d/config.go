package ascii2d

type Config struct {
	Enable bool
	TempFolderPath string
}

func NewConfig() *Config{
	return &Config{
		Enable: false,
		TempFolderPath: "temp",
	}
}
