package controllers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"weather-api/internal/dto"
	"weather-api/internal/models"
	"weather-api/internal/usecase"

	"github.com/gorilla/mux"
)

type WeatherUseCase interface {
	GetWeatherToday(ctx context.Context, params dto.GetWeatherTodayParams) (*dto.WeatherResult, error)
	GetWeatherByCity(ctx context.Context, cityName string) (*dto.WeatherResult, error)
	GetAllCities(ctx context.Context) ([]models.City, error)
}

// WeatherController обрабатывает HTTP запросы к погодному API
type WeatherController struct {
	weatherUseCase *usecase.WeatherUseCase
}

// WeatherControllerOptions параметры для создания контроллера
type WeatherControllerOptions struct {
	WeatherUseCase *usecase.WeatherUseCase
}

// NewWeatherController создает новый контроллер погоды
func NewWeatherController(options WeatherControllerOptions) *WeatherController {
	return &WeatherController{
		weatherUseCase: options.WeatherUseCase,
	}
}

// GetWeather получает погоду по координатам
func (c *WeatherController) GetWeather(w http.ResponseWriter, r *http.Request) {
	// Получаем параметры запроса
	query := r.URL.Query()

	// Парсим координаты
	latStr := query.Get("lat")
	lonStr := query.Get("lon")

	if latStr == "" || lonStr == "" {
		http.Error(w, "Missing lat or lon parameters", http.StatusBadRequest)
		return
	}

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		http.Error(w, "Invalid lat parameter", http.StatusBadRequest)
		return
	}

	lon, err := strconv.ParseFloat(lonStr, 64)
	if err != nil {
		http.Error(w, "Invalid lon parameter", http.StatusBadRequest)
		return
	}

	// Формируем запрос
	params := dto.GetWeatherTodayParams{
		Lat: lat,
		Lon: lon,
	}

	// Получаем погоду через usecase
	result, err := c.weatherUseCase.GetWeatherToday(r.Context(), params)
	if err != nil {
		slog.Error("Failed to get weather", "error", err)
		http.Error(w, "Error getting weather: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Отправляем результат
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// GetWeatherByCity получает погоду по названию города
func (c *WeatherController) GetWeatherByCity(w http.ResponseWriter, r *http.Request) {
	// Получаем название города из URL
	vars := mux.Vars(r)
	cityName := vars["city"]

	if cityName == "" {
		http.Error(w, "City name is required", http.StatusBadRequest)
		return
	}

	// Получаем погоду через usecase
	result, err := c.weatherUseCase.GetWeatherByCity(r.Context(), cityName)
	if err != nil {
		slog.Error("Failed to get weather for city", "city", cityName, "error", err)
		http.Error(w, "Error getting weather: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Отправляем результат
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// GetAllCities получает список всех городов
func (c *WeatherController) GetAllCities(w http.ResponseWriter, r *http.Request) {
	// Получаем список городов через usecase
	cities, err := c.weatherUseCase.GetAllCities(r.Context())
	if err != nil {
		slog.Error("Failed to get cities", "error", err)
		http.Error(w, "Error getting cities: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Отправляем результат
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cities)
}
