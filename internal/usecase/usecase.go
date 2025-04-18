package usecase

import (
	"context"
	"encoding/json"
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

// Интерфейс для кэширования
type Cache interface {
	Set(ctx context.Context, key string, value interface{}) error
	Get(ctx context.Context, key string) (string, error)
}

type WeatherUseCaseOptions struct {
	WeatherClient  WeatherClient
	CityRepository CityRepository
	Cache          Cache // Кэш
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
	// Ключ для кэша
	cacheKey := fmt.Sprintf("weather:%f:%f", params.Lat, params.Lon)

	// Проверяем кэш, если он настроен
	if usecase.options.Cache != nil {
		if cached, err := usecase.options.Cache.Get(ctx, cacheKey); err == nil {
			var result dto.WeatherResult
			if err := json.Unmarshal([]byte(cached), &result); err == nil {
				slog.Info("cache hit", "key", cacheKey)
				return &result, nil
			}
			slog.Warn("failed to unmarshal cached data", "key", cacheKey, "err", err)
		}
	}

	// Запрашиваем API, если кэш пуст или недоступен
	result, err := usecase.options.WeatherClient.WeatherToday(ctx, models.WeatherTodayParams{
		Lat: params.Lat,
		Lon: params.Lon,
	})
	if err != nil {
		err = fmt.Errorf("weather client failed: %w", err)
		slog.Error("client failed", "err", err)
		return nil, err
	}

	weatherResult := &dto.WeatherResult{
		CurrentWeather: dto.CurrentWeather{
			Temperature: result.CurrentWeather.Temperature,
			WeatherCode: result.CurrentWeather.WeatherCode,
			WeatherDesc: models.GetWeatherDescription(result.CurrentWeather.WeatherCode),
		},
	}

	// Сохраняем в кэш, если он настроен
	if usecase.options.Cache != nil {
		weatherJSON, err := json.Marshal(weatherResult)
		if err != nil {
			slog.Warn("failed to marshal weather result for cache", "key", cacheKey, "err", err)
		} else if err := usecase.options.Cache.Set(ctx, cacheKey, weatherJSON); err != nil {
			slog.Warn("failed to set cache", "key", cacheKey, "err", err)
		} else {
			slog.Info("cache set", "key", cacheKey)
		}
	}

	return weatherResult, nil
}

func (usecase *WeatherUseCase) GetWeatherByCity(ctx context.Context, cityName string) (*dto.WeatherResult, error) {
	if usecase.options.CityRepository == nil {
		return nil, fmt.Errorf("city repository not initialized")
	}

	// Ключ для кэша
	cacheKey := fmt.Sprintf("weather:city:%s", cityName)

	// Проверяем кэш, если он настроен
	if usecase.options.Cache != nil {
		if cached, err := usecase.options.Cache.Get(ctx, cacheKey); err == nil {
			var result dto.WeatherResult
			if err := json.Unmarshal([]byte(cached), &result); err == nil {
				slog.Info("cache hit", "key", cacheKey)
				return &result, nil
			}
			slog.Warn("failed to unmarshal cached data", "key", cacheKey, "err", err)
		}
	}

	// Получаем координаты города
	city, err := usecase.options.CityRepository.GetCityByName(ctx, cityName)
	if err != nil {
		return nil, fmt.Errorf("get city failed: %w", err)
	}

	// Запрашиваем погоду
	result, err := usecase.options.WeatherClient.WeatherToday(ctx, models.WeatherTodayParams{
		Lat: city.Latitude,
		Lon: city.Longitude,
	})
	if err != nil {
		err = fmt.Errorf("weather client failed: %w", err)
		slog.Error("client failed", "err", err)
		return nil, err
	}

	weatherResult := &dto.WeatherResult{
		CurrentWeather: dto.CurrentWeather{
			Temperature: result.CurrentWeather.Temperature,
			WeatherCode: result.CurrentWeather.WeatherCode,
			WeatherDesc: models.GetWeatherDescription(result.CurrentWeather.WeatherCode),
		},
	}

	// Сохраняем в кэш, если он настроен
	if usecase.options.Cache != nil {
		weatherJSON, err := json.Marshal(weatherResult)
		if err != nil {
			slog.Warn("failed to marshal weather result for cache", "key", cacheKey, "err", err)
		} else if err := usecase.options.Cache.Set(ctx, cacheKey, weatherJSON); err != nil {
			slog.Warn("failed to set cache", "key", cacheKey, "err", err)
		} else {
			slog.Info("cache set", "key", cacheKey)
		}
	}

	return weatherResult, nil
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
