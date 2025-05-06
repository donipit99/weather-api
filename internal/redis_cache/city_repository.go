package redis_cache

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

// CityRepositoryRedis - кэширующий прокси для репозитория городов
type CityRepositoryRedis struct {
	redisClient  *redis.Client
	postgresRepo repository.CityRepository
	metrics      *metrics.Metrics
}

// NewCityRepositoryRedis создает новый прокси-репозиторий с Redis
func NewCityRepositoryRedis(redisClient *redis.Client, postgresRepo repository.CityRepository, metrics *metrics.Metrics) *CityRepositoryRedis {
	return &CityRepositoryRedis{
		redisClient:  redisClient,
		postgresRepo: postgresRepo,
		metrics:      metrics,
	}
}

// GetCityByName получает город по имени с кэшированием
func (r *CityRepositoryRedis) GetCityByName(ctx context.Context, name string) (*models.City, error) {
	start := time.Now()

	// 1. Проверяем кэш
	cacheKey := fmt.Sprintf("city:%s", name)
	cachedData, err := r.redisClient.Get(ctx, cacheKey)

	// Если нашли в кэше и нет ошибки - возвращаем
	if err == nil {
		var city models.City
		if err := json.Unmarshal([]byte(cachedData), &city); err == nil {
			// Увеличиваем счетчик попаданий в кэш
			if r.metrics != nil {
				r.metrics.CacheHits.WithLabelValues("city").Inc()
			}
			return &city, nil
		}
	}

	// Увеличиваем счетчик промахов кэша
	if r.metrics != nil {
		r.metrics.CacheMisses.WithLabelValues("city").Inc()
	}

	// 2. Если нет в кэше - идем в PostgreSQL
	dbStart := time.Now()
	city, err := r.postgresRepo.GetCityByName(ctx, name)
	dbDuration := time.Since(dbStart).Seconds()

	// Сохраняем метрики о запросе к базе данных
	if r.metrics != nil {
		status := "success"
		if err != nil {
			status = "error"
		}
		r.metrics.DatabaseRequestsTotal.WithLabelValues("get_city", status).Inc()
		r.metrics.DatabaseRequestDuration.WithLabelValues("get_city").Observe(dbDuration)
	}

	if err != nil {
		return nil, err
	}

	// 3. Сохраняем в Redis
	cityJSON, err := json.Marshal(city)
	if err == nil {
		_ = r.redisClient.Set(ctx, cacheKey, cityJSON)
	}

	// Общее время выполнения метода
	if r.metrics != nil {
		r.metrics.HttpRequestDuration.WithLabelValues("GetCityByName", "internal").Observe(time.Since(start).Seconds())
	}

	return city, nil
}

// GetAllCities получает все города
func (r *CityRepositoryRedis) GetAllCities(ctx context.Context) ([]models.City, error) {
	start := time.Now()

	// Реализуем аналогичную логику кэширования для списка всех городов
	cacheKey := "cities:all"
	cachedData, err := r.redisClient.Get(ctx, cacheKey)

	if err == nil {
		var cities []models.City
		if err := json.Unmarshal([]byte(cachedData), &cities); err == nil {
			// Увеличиваем счетчик попаданий в кэш
			if r.metrics != nil {
				r.metrics.CacheHits.WithLabelValues("cities_all").Inc()
			}
			return cities, nil
		}
	}

	// Увеличиваем счетчик промахов кэша
	if r.metrics != nil {
		r.metrics.CacheMisses.WithLabelValues("cities_all").Inc()
	}

	// Запрос к базе данных
	dbStart := time.Now()
	cities, err := r.postgresRepo.GetAllCities(ctx)
	dbDuration := time.Since(dbStart).Seconds()

	// Сохраняем метрики
	if r.metrics != nil {
		status := "success"
		if err != nil {
			status = "error"
		}
		r.metrics.DatabaseRequestsTotal.WithLabelValues("get_all_cities", status).Inc()
		r.metrics.DatabaseRequestDuration.WithLabelValues("get_all_cities").Observe(dbDuration)
	}

	if err != nil {
		return nil, err
	}

	citiesJSON, err := json.Marshal(cities)
	if err == nil {
		_ = r.redisClient.Set(ctx, cacheKey, citiesJSON)
	}

	// Общее время выполнения метода
	if r.metrics != nil {
		r.metrics.HttpRequestDuration.WithLabelValues("GetAllCities", "internal").Observe(time.Since(start).Seconds())
	}

	return cities, nil
}
