package config

import (
	"fmt"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	Postgres   *Postgres   `envPrefix:"POSTGRES_"`
	WeatherAPI *WeatherAPI `envPrefix:"WEATHER_API_"`
	Server     *Server     `envPrefix:"SERVER_"`
}

type Postgres struct {
	Host     string `env:"HOST"`
	Port     string `env:"PORT"`
	Username string `env:"USERNAME"`
	Password string `env:"PASSWORD"`
	DB       string `env:"DB"`
}

type WeatherAPI struct {
	URL string `env:"URL"`
}

type Server struct {
	Port string `env:"PORT"`
}

func LoadConfig() (*Config, error) {
	config := new(Config)

	config.Postgres = new(Postgres)
	config.WeatherAPI = new(WeatherAPI)
	config.Server = new(Server)

	if err := env.Parse(config); err != nil {
		err = fmt.Errorf("env.Parse(...): %v", err)
		return nil, err
	}

	return config, nil
}
