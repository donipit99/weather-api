package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"weather-api/internal/dto"
	"weather-api/internal/models"
	"weather-api/internal/repository"
)

type WeatherUseCaseOptions struct {
	WeatherRepository repository.WeatherRepository
	CityRepository    repository.CityRepository
}

type WeatherUseCase struct {
	options WeatherUseCaseOptions
}

func NewWeatherUseCase(options WeatherUseCaseOptions) *WeatherUseCase {
	if options.WeatherRepository == nil {
		panic("weather repository must not be nil")
	}
	if options.CityRepository == nil {
		panic("city repository must not be nil")
	}
	return &WeatherUseCase{options: options}
}

func (usecase *WeatherUseCase) GetWeatherToday(ctx context.Context, params dto.GetWeatherTodayParams) (*dto.WeatherResult, error) {
	// Запрашиваем погоду через репозиторий
	result, err := usecase.options.WeatherRepository.WeatherToday(ctx, models.WeatherTodayParams{
		Lat: params.Lat,
		Lon: params.Lon,
	})
	if err != nil {
		slog.Error("weather repository failed", "err", err)
		return nil, fmt.Errorf("weather repository failed: %w", err)
	}

	// Преобразуем результат
	weatherResult := &dto.WeatherResult{
		CurrentWeather: dto.CurrentWeather{
			Temperature: result.CurrentWeather.Temperature,
			WeatherCode: result.CurrentWeather.WeatherCode,
			WeatherDesc: models.GetWeatherDescription(result.CurrentWeather.WeatherCode),
		},
	}

	return weatherResult, nil
}

func (usecase *WeatherUseCase) GetWeatherByCity(ctx context.Context, cityName string) (*dto.WeatherResult, error) {
	// Получаем город через репозиторий (с кэшированием)
	city, err := usecase.options.CityRepository.GetCityByName(ctx, cityName)
	if err != nil {
		return nil, fmt.Errorf("city repository failed: %w", err)
	}

	// Запрашиваем погоду по координатам города (тоже с кэшированием)
	result, err := usecase.options.WeatherRepository.WeatherToday(ctx, models.WeatherTodayParams{
		Lat: city.Latitude,
		Lon: city.Longitude,
	})
	if err != nil {
		slog.Error("weather repository failed", "err", err)
		return nil, fmt.Errorf("weather repository failed: %w", err)
	}

	// Преобразуем результат
	weatherResult := &dto.WeatherResult{
		CurrentWeather: dto.CurrentWeather{
			Temperature: result.CurrentWeather.Temperature,
			WeatherCode: result.CurrentWeather.WeatherCode,
			WeatherDesc: models.GetWeatherDescription(result.CurrentWeather.WeatherCode),
		},
	}

	return weatherResult, nil
}

func (usecase *WeatherUseCase) GetAllCities(ctx context.Context) ([]models.City, error) {
	return usecase.options.CityRepository.GetAllCities(ctx)
}
