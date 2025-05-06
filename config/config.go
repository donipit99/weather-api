package config

import (
	"fmt"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	Postgres   *Postgres   `envPrefix:"POSTGRES_"`
	WeatherAPI *WeatherAPI `envPrefix:"WEATHER_API_"`
	Server     *Server     `envPrefix:"SERVER_"`
	Telegram   *Telegram   `envPrefix:"TELEGRAM_"`
	Redis      *Redis      `envPrefix:"REDIS_"`
	LogLevel   string      `env:"LOG_LEVEL"` // уровень логирования
}

type Postgres struct {
	Host     string `env:"HOST"`
	Port     string `env:"PORT"`
	Username string `env:"USERNAME"`
	Password string `env:"PASSWORD"`
	DB       string `env:"DB"`
}

type Redis struct {
	Host string `env:"HOST"`
	Port string `env:"PORT"`
	TTL  int    `env:"TTL"`
}

type WeatherAPI struct {
	URL string `env:"URL"`
}

type Server struct {
	Port string `env:"PORT"`
}

type Telegram struct {
	Token string `env:"TOKEN"`
}

func LoadConfig() (*Config, error) {
	config := new(Config)
	config.Postgres = new(Postgres)
	config.WeatherAPI = new(WeatherAPI)
	config.Server = new(Server)
	config.Telegram = new(Telegram)

	if err := env.Parse(config); err != nil {
		return nil, fmt.Errorf("env.Parse: %v", err)
	}

	if config.WeatherAPI.URL == "" {
		return nil, fmt.Errorf("WEATHER_API_URL is required")
	}
	if config.Server.Port == "" {
		return nil, fmt.Errorf("SERVER_PORT is required")
	}
	if config.Postgres.Host == "" {
		return nil, fmt.Errorf("POSTGRES_HOST is required")
	}
	if config.Postgres.DB == "" {
		return nil, fmt.Errorf("POSTGRES_DB is required")
	}
	if config.Telegram.Token == "" {
		return nil, fmt.Errorf("TELEGRAM_TOKEN is required")
	}

	return config, nil
}
