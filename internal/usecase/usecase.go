package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"weather-api/internal/dto"
	"weather-api/internal/models"
)

type WeatherClient interface {
	WeatherToday(ctx context.Context, params models.WeatherTodayParams) (*models.WeatherResult, error)
}

type CityRepository interface {
	GetCityByName(ctx context.Context, name string) (*models.City, error)
	GetAllCities(ctx context.Context) ([]models.City, error)
}

type WeatherUseCaseOptions struct {
	WeatherClient  WeatherClient
	CityRepository CityRepository
}

type WeatherUseCase struct {
	options WeatherUseCaseOptions
}

func NewWeatherUseCase(options WeatherUseCaseOptions) *WeatherUseCase {
	if options.WeatherClient == nil {
		panic("weather client must not be nil")
	}
	return &WeatherUseCase{options: options}
}

func (usecase *WeatherUseCase) GetWeatherToday(ctx context.Context, params dto.GetWeatherTodayParams) (*dto.WeatherResult, error) {
	result, err := usecase.options.WeatherClient.WeatherToday(ctx, models.WeatherTodayParams{
		Lat: params.Lat,
		Lon: params.Lon,
	})
	if err != nil {
		err = fmt.Errorf("weather client failed: %w", err)
		slog.Error("client failed", "err", err)
		return nil, err
	}
	return &dto.WeatherResult{
		CurrentWeather: dto.CurrentWeather{
			Temperature: result.CurrentWeather.Temperature,
			WeatherCode: result.CurrentWeather.WeatherCode,
			WeatherDesc: models.GetWeatherDescription(result.CurrentWeather.WeatherCode),
		},
	}, nil
}

func (usecase *WeatherUseCase) GetWeatherByCity(ctx context.Context, cityName string) (*dto.WeatherResult, error) {
	if usecase.options.CityRepository == nil {
		return nil, fmt.Errorf("city repository not initialized")
	}
	city, err := usecase.options.CityRepository.GetCityByName(ctx, cityName)
	if err != nil {
		return nil, fmt.Errorf("get city failed: %w", err)
	}
	result, err := usecase.options.WeatherClient.WeatherToday(ctx, models.WeatherTodayParams{
		Lat: city.Latitude,
		Lon: city.Longitude,
	})
	if err != nil {
		err = fmt.Errorf("weather client failed: %w", err)
		slog.Error("client failed", "err", err)
		return nil, err
	}
	return &dto.WeatherResult{
		CurrentWeather: dto.CurrentWeather{
			Temperature: result.CurrentWeather.Temperature,
			WeatherCode: result.CurrentWeather.WeatherCode,
			WeatherDesc: models.GetWeatherDescription(result.CurrentWeather.WeatherCode),
		},
	}, nil
}

func (usecase *WeatherUseCase) GetAllCities(ctx context.Context) ([]models.City, error) {
	if usecase.options.CityRepository == nil {
		return nil, fmt.Errorf("city repository not initialized")
	}
	cities, err := usecase.options.CityRepository.GetAllCities(ctx)
	if err != nil {
		return nil, fmt.Errorf("get all cities failed: %w", err)
	}
	return cities, nil
}
