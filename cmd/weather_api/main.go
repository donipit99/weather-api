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
	adapters "weather-api/internal/adapters/weather_client"
	"weather-api/internal/controllers"
	routes "weather-api/internal/controllers/http_weather_controller"

	telegramController "weather-api/internal/controllers/telegram"

	usecase "weather-api/internal/usecase"
	"weather-api/pkg/logger"
	"weather-api/pkg/postgresql"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	if err := godotenv.Load(); err != nil {
		tempLogger := logger.NewLogger("info")
		tempLogger.Warn("No .env file found, relying on environment variables", "err", err)
	}

	// Загружаем конфигурацию
	cfg, err := config.LoadConfig()
	if err != nil {
		tempLogger := logger.NewLogger("info")
		tempLogger.Fatal("load config failed", "err", err)
	}

	// Создаем логгер с уровнем из конфигурации
	log := logger.NewLogger(cfg.LogLevel)
	// Устанавливаем логгер как дефолтный для slog
	slog.SetDefault(log.Logger)

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

	// Подключение к Redis
	redisAddr := net.JoinHostPort(cfg.RedisHost, cfg.RedisPort)
	redisClient := redis.NewClient(redisAddr)

	// Контекст с таймаутом для проверки подключения
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pong, err := redisClient.Ping(ctx)
	if err != nil {
		slog.Error("failed to connect to redis", "addr", redisAddr, "error", err)
		os.Exit(1)
	}
	slog.Info("redis connected", "response", pong)

	cityRepository := postgres.NewCityRepository(postgres.CityRepositoryOptions{DB: db.DB})
	client := adapters.NewClient(adapters.ClientOptions{URL: cfg.WeatherAPI.URL})
	weatherUsecase := usecase.NewWeatherUseCase(usecase.WeatherUseCaseOptions{
		WeatherClient:  client,
		CityRepository: cityRepository,
		Cache:          redisClient,
	})
	weatherController := controllers.NewWeatherController(controllers.WeatherControllerOptions{
		WeatherUseCase: weatherUsecase,
	})

	//
	router := routes.SetupRoutes(weatherController)

	// Инициализация тг бота
	bot, err := telegram.NewBot(cfg.Telegram.Token)
	if err != nil {
		slog.Error("failed to create telegram bot", "error", err)
		os.Exit(1)
	}

	// Инициализация тг контроллера
	telegramController := telegramController.NewTelegramController(bot, weatherUsecase)

	// Запуск тг контроллера
	go func() {
		ctx := context.Background()
		if err := telegramController.Start(ctx); err != nil {
			slog.Error("telegram controller failed", "error", err)
		}
	}()

	// Запуск HTTP сервера с нашим роутером
	slog.Info("server start", "port", cfg.Server.Port)
	if err := http.ListenAndServe(":"+cfg.Server.Port, router); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}
