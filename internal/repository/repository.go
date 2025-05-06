package repository

import (
	"context"
	"weather-api/internal/models"
)

// CityRepository определяет методы для работы с городами
type CityRepository interface {
	GetCityByName(ctx context.Context, name string) (*models.City, error)
	GetAllCities(ctx context.Context) ([]models.City, error)
}

// WeatherRepository определяет методы для получения погоды
type WeatherRepository interface {
	WeatherToday(ctx context.Context, params models.WeatherTodayParams) (*models.WeatherResult, error)
}
