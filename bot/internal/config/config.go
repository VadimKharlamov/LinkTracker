package config

import (
	"flag"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env     string        `yaml:"env" env-default:"local"`
	Bot     BotConfig     `yaml:"bot"`
	Clients ClientsConfig `yaml:"bot_clients"`
}

type BotConfig struct {
	Address     string        `yaml:"address"`
	Token       string        `env:"TOKEN"`
	StoragePath string        `yaml:"storage_path"`
	MaxIdle     int           `yaml:"max_idle" env-default:"5"`
	MaxActive   int           `yaml:"max_active" env-default:"10"`
	Timeout     time.Duration `yaml:"timeout" env-default:"10s"`
	RateLimit   int           `yaml:"rate_limit" env-default:"50"`
}

type Client struct {
	Address  string        `yaml:"address"`
	Timeout  time.Duration `yaml:"timeout"`
	Retry    uint          `yaml:"retry"`
	Backoff  time.Duration `yaml:"backoff"`
	Token    string        `env:"TOKEN"`
	Topic    string        `yaml:"base_topic"`
	DLQTopic string        `yaml:"dlq_topic"`
}

type CBConfig struct {
	MaxRequests       uint32        `yaml:"max_requests"`
	Timeout           time.Duration `yaml:"timeout"`
	SlidingWindowSize uint32        `yaml:"sliding_window_size"`
	FailureCount      uint32        `yaml:"failure_count"`
}

type ClientsConfig struct {
	Kafka          Client   `yaml:"kafka"`
	Scrapper       Client   `yaml:"scraper"`
	CircuitBreaker CBConfig `yaml:"circuit_breaker"`
}

func MustLoad() *Config {
	configPath := fetchConfigPath()
	if configPath == "" {
		panic("config path is empty")
	}

	return MustLoadPath(configPath)
}

func MustLoadPath(configPath string) *Config {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic("config file does not exist: " + configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic("cannot read config: " + err.Error())
	}

	cfg.Bot.Token = os.Getenv("BOT_TOKEN")

	return &cfg
}

func fetchConfigPath() string {
	var res string

	flag.StringVar(&res, "config", "", "path to config file")
	flag.Parse()

	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}

	return res
}
