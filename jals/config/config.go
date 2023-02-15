package config

import (
	"fmt"

	"github.com/caarlos0/env/v7"
)

type Config struct {
	Address         string `env:"ADDRESS" envDefault:":8080"`
	Debug           bool   `env:"DEBUG" envDefault:"false"`
	Emoji           bool   `env:"EMOJI" envDefault:"false"`
	MetricsUser     string `env:"METRICS_USER"`
	MetricsPassword string `env:"METRICS_PASSWORD"`
	RedisURI        string `env:"REDIS_URI" envDefault:""`
}

func AppConfig() *Config {
	var c Config

	if err := env.Parse(&c); err != nil {
		fmt.Printf("%+v\n", err)
	}

	return &c
}
