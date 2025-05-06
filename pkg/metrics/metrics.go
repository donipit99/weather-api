package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics содержит все метрики приложения
type Metrics struct {
	HttpRequestsTotal       *prometheus.CounterVec
	HttpRequestDuration     *prometheus.HistogramVec
	WeatherRequestsTotal    *prometheus.CounterVec
	WeatherRequestDuration  *prometheus.HistogramVec
	CacheHits               *prometheus.CounterVec
	CacheMisses             *prometheus.CounterVec
	DatabaseRequestsTotal   *prometheus.CounterVec
	DatabaseRequestDuration *prometheus.HistogramVec
}

// NewMetrics создает и регистрирует метрики Prometheus
func NewMetrics() *Metrics {
	m := &Metrics{
		// Метрики HTTP запросов
		HttpRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "weather_api_http_requests_total",
				Help: "Общее количество HTTP запросов",
			},
			[]string{"method", "endpoint", "status"},
		),
		HttpRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "weather_api_http_request_duration_seconds",
				Help:    "Длительность HTTP запросов в секундах",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "endpoint"},
		),

		// Метрики запросов погоды
		WeatherRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "weather_api_weather_requests_total",
				Help: "Общее количество запросов к API погоды",
			},
			[]string{"status"},
		),
		WeatherRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "weather_api_weather_request_duration_seconds",
				Help:    "Длительность запросов к API погоды в секундах",
				Buckets: prometheus.DefBuckets,
			},
			[]string{},
		),

		// Метрики кэша
		CacheHits: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "weather_api_cache_hits_total",
				Help: "Количество успешных обращений к кэшу",
			},
			[]string{"cache_type"},
		),
		CacheMisses: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "weather_api_cache_misses_total",
				Help: "Количество промахов кэша",
			},
			[]string{"cache_type"},
		),

		// Метрики базы данных
		DatabaseRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "weather_api_database_requests_total",
				Help: "Общее количество запросов к базе данных",
			},
			[]string{"operation", "status"},
		),
		DatabaseRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "weather_api_database_request_duration_seconds",
				Help:    "Длительность запросов к базе данных в секундах",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"operation"},
		),
	}

	return m
}
