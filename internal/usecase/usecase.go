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

type WeatherUseCaseOptions struct {
	WeatherClient WeatherClient
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
		err = fmt.Errorf("usecase.options.WeatherClient.WeatherToday(...): %w", err)
		slog.Error("client failed", "err", err)

		return nil, err
	}

	return &dto.WeatherResult{
		CurrentWeather: result.CurrentWeather,
	}, nil
}
