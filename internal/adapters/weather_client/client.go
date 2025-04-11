package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
	"weather-api/internal/models"
)

var (
	ErrStatusWeatherAPI = fmt.Errorf("error response from weather api")
)

type Client struct {
	options ClientOptions
}

type ClientOptions struct {
	// weather api https://api.open-meteo.com
	URL string
}

func NewClient(options ClientOptions) *Client {
	return &Client{
		options: options,
	}
}

func (c *Client) WeatherToday(ctx context.Context, params models.WeatherTodayParams) (*models.WeatherResult, error) {
	url := c.options.URL + fmt.Sprintf("/v1/forecast?latitude=%f&longitude=%f&current_weather=true", params.Lat, params.Lon)

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		slog.Error("failed to create request", "err", err)
		err = fmt.Errorf("http.NewRequestWithContext(...): %w", err)
		return nil, err
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	rsp, err := client.Do(request)
	if err != nil {
		slog.Error("failed to perform request", "err", err)
		err = fmt.Errorf("http.Do(...): %w", err)
		return nil, err
	}
	defer rsp.Body.Close()

	body, err := io.ReadAll(rsp.Body)
	if err != nil {
		slog.Error("failed to read response body", "err", err)
		err = fmt.Errorf("io.ReadAll(...): %w", err)
		return nil, err
	}

	if rsp.StatusCode != http.StatusOK {
		slog.Error("weather api returned non-OK status", "status", rsp.StatusCode, "body", string(body))
		return nil, fmt.Errorf("%w: %s", ErrStatusWeatherAPI, body)
	}

	var result models.WeatherResult
	err = json.Unmarshal(body, &result)
	if err != nil {
		slog.Error("failed to unmarshal response", "err", err)
		err = fmt.Errorf("json.Unmarshal(...): %w", err)
		return nil, err
	}

	return &result, nil
}
