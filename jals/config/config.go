package config

import (
	"fmt"

	"github.com/caarlos0/env/v7"
)

type Config struct {
	Address         string `env:"ADDRESS" envDefault:":8080"`
	Emoji           bool   `env:"EMOJI" envDefault:"false"`
	MetricsUser     string `env:"METRICS_USER" envDefault:""`
	MetricsPassword string `env:"METRICS_PASSWORD" envDefault:""`
	RedisAddress    string `env:"REDIS_ADDRESS" envDefault:"localhost:6379"`
	RedisDB         int    `env:"REDIS_DB" envDefault:"0"`
}

func AppConfig() *Config {
	var c Config

	if err := env.Parse(&c); err != nil {
		fmt.Printf("%+v\n", err)
	}

	return &c
}
