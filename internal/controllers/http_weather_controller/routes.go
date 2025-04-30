package http_weather_controler

import (
	"net/http"
	"weather-api/internal/controllers"
)

func SetupRoutes(weatherController *controllers.WeatherController) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/weather", weatherController.GetWeatherToday)
	mux.HandleFunc("/api/v1/weather/city", weatherController.GetWeatherByCity)

	return mux
}
