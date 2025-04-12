package controllers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"weather-api/internal/dto"

	"golang.org/x/exp/slog"
)

type WeatherUseCase interface {
	GetWeatherToday(ctx context.Context, params dto.GetWeatherTodayParams) (*dto.WeatherResult, error)
}

type WeatherController struct {
	options WeatherControllerOptions
}

type WeatherControllerOptions struct {
	WeatherUseCase WeatherUseCase
}

func NewWeatherController(options WeatherControllerOptions) *WeatherController {
	return &WeatherController{options: options}
}

type GetWeatherTodayRequest struct {
	Lat  string `json:"lat"`
	Lon  string `json:"lon"`
	Lang string `json:"lang"`
}

type GetWeatherTodayResponse struct {
	Temperature float64 `json:"temperature"`
	WeatherCode int     `json:"weather_code"`
}

func (controller *WeatherController) GetWeatherToday(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Читаем параметры из тела запроса
	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("read request body failed", "err", err)
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte(`{"error":"failed to read request body"}`))
		return
	}
	defer r.Body.Close()

	var request GetWeatherTodayRequest
	err = json.Unmarshal(body, &request)
	if err != nil {
		slog.Error("unmarshal incoming request failed", "err", err)
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte(`{"error":"invalid request body"}`))
		return
	}

	// Валидация параметров
	lat, err := strconv.ParseFloat(request.Lat, 64)
	if err != nil {
		slog.Error("invalid latitude", "lat", request.Lat, "err", err)
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte(`{"error":"invalid latitude"}`))
		return
	}

	lon, err := strconv.ParseFloat(request.Lon, 64)
	if err != nil {
		slog.Error("invalid longitude", "lon", request.Lon, "err", err)
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte(`{"error":"invalid longitude"}`))
		return
	}

	result, err := controller.options.WeatherUseCase.GetWeatherToday(ctx,
		dto.GetWeatherTodayParams{
			Lat: lat,
			Lon: lon,
		},
	)
	if err != nil {
		slog.Error("call usecase get weather today failed", "err", err)
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte(`{"error":"failed to get weather data"}`))
		return
	}

	response := &GetWeatherTodayResponse{
		Temperature: result.CurrentWeather.Temperature,
		WeatherCode: result.CurrentWeather.WeatherCode,
	}

	buf, err := json.Marshal(response)
	if err != nil {
		slog.Error("marshal result failed", "err", err)
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte(`{"error":"failed to marshal response"}`))
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	rw.Write(buf)
}
