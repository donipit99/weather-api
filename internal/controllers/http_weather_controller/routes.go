package http_weather_controller

import (
	"net/http"
	"weather-api/internal/controllers"
	"weather-api/internal/middleware"
	"weather-api/pkg/metrics"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// SetupRoutes настраивает маршруты для HTTP API
func SetupRoutes(controller *controllers.WeatherController, metrics *metrics.Metrics) *mux.Router {
	router := mux.NewRouter()

	// Применяем middleware для сбора метрик ко всем маршрутам
	router.Use(middleware.MetricsMiddleware(metrics))

	// API маршруты
	api := router.PathPrefix("/api").Subrouter()

	// Маршрут для получения погоды по координатам
	api.HandleFunc("/weather", controller.GetWeather).Methods(http.MethodGet)

	// Маршрут для получения погоды по названию города
	api.HandleFunc("/weather/city/{city}", controller.GetWeatherByCity).Methods(http.MethodGet)

	// Маршрут для получения списка всех городов
	api.HandleFunc("/cities", controller.GetAllCities).Methods(http.MethodGet)

	// Маршрут для метрик Prometheus
	router.Handle("/metrics", promhttp.Handler())

	// Маршрут для проверки состояния сервиса
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods(http.MethodGet)

	return router
}
