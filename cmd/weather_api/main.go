package main

import (
	"net/http"
	"os"
	"weather-api/config"
	adapters "weather-api/internal/adapters/weather_client"
	"weather-api/internal/controllers"
	usecase "weather-api/internal/usecase"

	"golang.org/x/exp/slog"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("msg", "load config failed", "err", err)
		os.Exit(1)
	}

	client := adapters.NewClient(
		adapters.ClientOptions{
			URL: "https://api.open-meteo.com",
		},
	)

	weatherUsecase := usecase.NewWeatherUseCase(
		usecase.WeatherUseCaseOptions{
			WeatherClient: client,
		},
	)

	weatherController := controllers.NewWeatherController(
		controllers.WeatherControllerOptions{
			WeatherUseCase: weatherUsecase,
		},
	)

	http.HandleFunc("/api/v1/weather", weatherController.GetWeatherToday)

	if err := http.ListenAndServe(":"+cfg.Server.Port, nil); err != nil {
		slog.Error("server failed", "err", err)
		os.Exit(1)
	}
}
