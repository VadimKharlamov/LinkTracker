package config

import (
	"flag"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env     string        `yaml:"env" env-default:"local"`
	Scraper ScraperConfig `yaml:"scraper"`
	Clients ClientsConfig `yaml:"scraper_clients"`
}

type ScraperConfig struct {
	Address        string        `yaml:"address"`
	TransportType  string        `yaml:"message_transport"`
	AccessType     string        `yaml:"access_type" env-default:"ORM"`
	MaxConn        int32         `yaml:"max_conn" env-default:"15"`
	MinConn        int32         `yaml:"min_conn" env-default:"5"`
	BatchSize      uint64        `yaml:"batch_size" env-default:"10"`
	Token          string        `yaml:"token" env-default:""`
	StoragePath    string        `yaml:"storage_path" env-default:""`
	Timeout        time.Duration `yaml:"timeout" env-default:"10s"`
	RateLimit      int           `yaml:"rate_limit" env-default:"50"`
	LinksRateLimit int           `yaml:"links_rate_limit" env-default:"25"`
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
	Bot            Client   `yaml:"bot"`
	Kafka          Client   `yaml:"kafka"`
	Github         Client   `yaml:"github"`
	StackOverFlow  Client   `yaml:"stack_overflow"`
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

	cfg.Clients.Github.Token = os.Getenv("GITHUB_TOKEN")
	cfg.Clients.StackOverFlow.Token = os.Getenv("STACK_OVERFLOW_TOKEN")

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
