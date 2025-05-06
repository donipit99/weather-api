package weather_cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"weather-api/internal/adapters/redis"
	"weather-api/internal/models"
	"weather-api/internal/repository"
	"weather-api/pkg/metrics"
)

// Проверка, что тип реализует интерфейс
var _ repository.WeatherRepository = (*WeatherCache)(nil)

// WeatherCache - кэширующий прокси для погодных данных
type WeatherCache struct {
	redisClient *redis.Client
	weatherRepo repository.WeatherRepository
	metrics     *metrics.Metrics
}

// NewWeatherCache создает новый кэш для погоды
func NewWeatherCache(redisClient *redis.Client, weatherRepo repository.WeatherRepository, metrics *metrics.Metrics) *WeatherCache {
	return &WeatherCache{
		redisClient: redisClient,
		weatherRepo: weatherRepo,
		metrics:     metrics,
	}
}

// WeatherToday получает погоду на сегодня с кэшированием
func (c *WeatherCache) WeatherToday(ctx context.Context, params models.WeatherTodayParams) (*models.WeatherResult, error) {
	start := time.Now()

	// 1. Проверяем кэш
	cacheKey := fmt.Sprintf("weather:lat:%f:lon:%f", params.Lat, params.Lon)
	cachedData, err := c.redisClient.Get(ctx, cacheKey)

	// Если нашли в кэше - возвращаем
	if err == nil {
		var result models.WeatherResult
		if err := json.Unmarshal([]byte(cachedData), &result); err == nil {
			// Увеличиваем счетчик попаданий в кэш
			if c.metrics != nil {
				c.metrics.CacheHits.WithLabelValues("weather").Inc()
			}
			return &result, nil
		}
	}

	// Увеличиваем счетчик промахов кэша
	if c.metrics != nil {
		c.metrics.CacheMisses.WithLabelValues("weather").Inc()
	}

	// Если нет в кэше - идем в API через оригинальный репозиторий
	apiStart := time.Now()
	result, err := c.weatherRepo.WeatherToday(ctx, params)
	apiDuration := time.Since(apiStart).Seconds()

	// Сохраняем метрики о запросе к API
	if c.metrics != nil {
		status := "success"
		if err != nil {
			status = "error"
		}
		c.metrics.WeatherRequestsTotal.WithLabelValues(status).Inc()
		c.metrics.WeatherRequestDuration.WithLabelValues().Observe(apiDuration)
	}

	if err != nil {
		return nil, err
	}

	// Сохраняем в Redis
	resultJSON, err := json.Marshal(result)
	if err == nil {
		_ = c.redisClient.Set(ctx, cacheKey, resultJSON)
	}

	// Общее время выполнения метода
	if c.metrics != nil {
		c.metrics.HttpRequestDuration.WithLabelValues("WeatherToday", "internal").Observe(time.Since(start).Seconds())
	}

	return result, nil
}
