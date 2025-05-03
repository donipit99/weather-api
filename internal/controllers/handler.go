package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"weather-api/internal/dto"
	"weather-api/internal/models"
)

type WeatherUseCase interface {
	GetWeatherToday(ctx context.Context, params dto.GetWeatherTodayParams) (*dto.WeatherResult, error)
	GetWeatherByCity(ctx context.Context, cityName string) (*dto.WeatherResult, error)
	GetAllCities(ctx context.Context) ([]models.City, error)
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
	Lat string `json:"lat"`
	Lon string `json:"lon"`
}

type GetWeatherByCityRequest struct {
	City string `json:"city"`
}

type GetWeatherTodayResponse struct {
	Temperature float64 `json:"temperature"`
	WeatherCode int     `json:"weather_code"`
	WeatherDesc string  `json:"weather_description"`
}

func (controller *WeatherController) GetWeatherToday(rw http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(rw, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	ctx := r.Context()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(rw, http.StatusBadRequest, "failed to read request body")
		return
	}
	defer r.Body.Close()

	request := new(GetWeatherTodayRequest)
	if err := json.Unmarshal(body, request); err != nil {
		writeError(rw, http.StatusBadRequest, "invalid request body")
		return
	}

	lat, err := strconv.ParseFloat(request.Lat, 64)
	if err != nil {
		writeError(rw, http.StatusBadRequest, "invalid latitude")
		return
	}
	lon, err := strconv.ParseFloat(request.Lon, 64)
	if err != nil {
		writeError(rw, http.StatusBadRequest, "invalid longitude")
		return
	}
	if lat < -90 || lat > 90 {
		writeError(rw, http.StatusBadRequest, "latitude must be between -90 and 90")
		return
	}
	if lon < -180 || lon > 180 {
		writeError(rw, http.StatusBadRequest, "longitude must be between -180 and 180")
		return
	}

	result, err := controller.options.WeatherUseCase.GetWeatherToday(ctx,
		dto.GetWeatherTodayParams{
			Lat: lat,
			Lon: lon,
		})

	if err != nil {
		writeError(rw, http.StatusInternalServerError, fmt.Sprintf("failed to get weather data: %v", err))
		return
	}

	response := &GetWeatherTodayResponse{
		Temperature: result.CurrentWeather.Temperature,
		WeatherCode: result.CurrentWeather.WeatherCode,
		WeatherDesc: result.CurrentWeather.WeatherDesc,
	}

	writeResponse(rw, http.StatusOK, response)
}

func (controller *WeatherController) GetWeatherByCity(rw http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(rw, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	ctx := r.Context()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(rw, http.StatusBadRequest, "failed to read request body")
		return
	}

	defer r.Body.Close()

	var request GetWeatherByCityRequest
	if err := json.Unmarshal(body, &request); err != nil {
		writeError(rw, http.StatusBadRequest, "invalid request body")
		return
	}
	if request.City == "" {
		writeError(rw, http.StatusBadRequest, "city name is required")
		return
	}

	result, err := controller.options.WeatherUseCase.GetWeatherByCity(ctx, request.City)
	if err != nil {
		writeError(rw, http.StatusInternalServerError, fmt.Sprintf("failed to get weather data: %v", err))
		return
	}

	response := &GetWeatherTodayResponse{
		Temperature: result.CurrentWeather.Temperature,
		WeatherCode: result.CurrentWeather.WeatherCode,
		WeatherDesc: result.CurrentWeather.WeatherDesc,
	}
	writeResponse(rw, http.StatusOK, response)
}

func writeResponse(rw http.ResponseWriter, i int, response *GetWeatherTodayResponse) {
	panic("unimplemented")
}

func writeError(rw http.ResponseWriter, i int, s string) {
	panic("unimplemented")
}
