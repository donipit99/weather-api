package main

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"
	"weather-api/config"
	"weather-api/internal/adapters/postgres"
	"weather-api/internal/adapters/redis"
	"weather-api/internal/adapters/telegram"
	"weather-api/internal/adapters/weather_cache"
	"weather-api/internal/adapters/weather_client"
	"weather-api/internal/controllers"
	httpController "weather-api/internal/controllers/http_weather_controller"
	telegramController "weather-api/internal/controllers/telegram"
	"weather-api/internal/redis_cache"
	"weather-api/internal/usecase"
	"weather-api/pkg/logger"
	"weather-api/pkg/metrics"
	"weather-api/pkg/postgresql"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	// Загрузка .env
	if err := godotenv.Load(); err != nil {
		tempLogger := logger.NewLogger("info")
		tempLogger.Warn("No .env file found, relying on environment variables", "err", err)
	}

	// Загрузка конфига
	cfg, err := config.LoadConfig()
	if err != nil {
		tempLogger := logger.NewLogger("info")
		tempLogger.Fatal("load config failed", "err", err)
	}

	// Создаем логгер
	log := logger.NewLogger(cfg.LogLevel)
	slog.SetDefault(log.Logger)

	// Инициализация метрик Prometheus
	appMetrics := metrics.NewMetrics()

	// PostgreSQL
	db, err := postgresql.NewPostgres(
		postgresql.WithHost(cfg.Postgres.Host),
		postgresql.WithPort(cfg.Postgres.Port),
		postgresql.WithUsername(cfg.Postgres.Username),
		postgresql.WithPassword(cfg.Postgres.Password),
		postgresql.WithDBName(cfg.Postgres.DB),
		postgresql.WithSSLMode("disable"),
	)
	if err != nil {
		log.Fatal("failed to initialize postgres", "error", err)
	}
	defer db.Close()
	log.Info("successfully connected to postgres")

	// Redis
	redisAddr := net.JoinHostPort(cfg.Redis.Host, cfg.Redis.Port)
	redisClient := redis.NewClient(redisAddr, cfg.Redis.TTL)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	pong, err := redisClient.Ping(ctx)
	if err != nil {
		log.Error("failed to connect to redis", "addr", redisAddr, "error", err)
		os.Exit(1)
	}
	log.Info("redis connected", "response", pong)

	// Репозиторий городов (PostgreSQL)
	pgCityRepo := postgres.NewCityRepository(postgres.CityRepositoryOptions{DB: db.DB})

	// Кэширующий прокси для городов
	cityRepository := redis_cache.NewCityRepositoryRedis(redisClient, pgCityRepo, appMetrics)

	// Погодный клиент
	weatherClient := weather_client.NewClient(weather_client.ClientOptions{URL: cfg.WeatherAPI.URL})

	// Кэширующий прокси для погоды
	weatherRepository := weather_cache.NewWeatherCache(redisClient, weatherClient, appMetrics)

	// UseCase
	weatherUsecase := usecase.NewWeatherUseCase(usecase.WeatherUseCaseOptions{
		WeatherRepository: weatherRepository,
		CityRepository:    cityRepository,
	})

	// HTTP контроллер
	weatherController := controllers.NewWeatherController(controllers.WeatherControllerOptions{
		WeatherUseCase: weatherUsecase,
	})

	// HTTP маршруты
	router := httpController.SetupRoutes(weatherController, appMetrics)

	// Telegram бот
	bot, err := telegram.NewBot(cfg.Telegram.Token)
	if err != nil {
		log.Error("failed to create telegram bot", "error", err)
		os.Exit(1)
	}

	// Telegram контроллер
	tgController := telegramController.NewTelegramController(bot, weatherUsecase)

	// Запуск Telegram контроллера
	go func() {
		ctx := context.Background()
		if err := tgController.Start(ctx); err != nil {
			log.Error("telegram controller failed", "error", err)
		}
	}()

	// Запуск HTTP сервера
	log.Info("server start", "port", cfg.Server.Port)
	if err := http.ListenAndServe(":"+cfg.Server.Port, router); err != nil {
		log.Error("server failed", "error", err)
		os.Exit(1)
	}
}
