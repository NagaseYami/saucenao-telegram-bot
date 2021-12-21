package dice

type Config struct {
	Enable bool `yaml:"Enable"`
}

type Service struct {
	*Config
}

