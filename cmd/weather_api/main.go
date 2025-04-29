package main

import (
	"context"
	"fmt"
	"log"
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
	telegramController "weather-api/internal/controllers/telegram" // Добавляем импорт подпакета telegram

	usecase "weather-api/internal/usecase"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil { // Загрузка .env файла
		log.Fatal("Failed to load .env file", err)
	}

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	slog.SetDefault(slog.New(handler))

	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("load config failed", "error", err)
		os.Exit(1)
	}

	// Подключение к Postgres
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.Username, cfg.Postgres.Password, cfg.Postgres.DB)
	db, err := sqlx.Open("pgx", dsn)
	if err != nil {
		slog.Error("failed to connect to postgres", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Ожидание готовности PostgreSQL
	for i := 0; i < 10; i++ {
		if err := db.Ping(); err != nil {
			slog.Warn("failed to ping postgres, retrying", "attempt", i+1, "error", err)
			time.Sleep(2 * time.Second)
			continue
		}
		slog.Info("successfully connected to postgres")
		break
	}
	if err := db.Ping(); err != nil {
		slog.Error("failed to ping postgres after retries", "error", err)
		os.Exit(1)
	}

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

	cityRepository := postgres.NewCityRepository(postgres.CityRepositoryOptions{DB: db})
	client := adapters.NewClient(adapters.ClientOptions{URL: cfg.WeatherAPI.URL})
	weatherUsecase := usecase.NewWeatherUseCase(usecase.WeatherUseCaseOptions{
		WeatherClient:  client,
		CityRepository: cityRepository,
		Cache:          redisClient,
	})
	weatherController := controllers.NewWeatherController(controllers.WeatherControllerOptions{
		WeatherUseCase: weatherUsecase,
	})

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

	http.HandleFunc("/api/v1/weather", weatherController.GetWeatherToday)
	http.HandleFunc("/api/v1/weather/city", weatherController.GetWeatherByCity)

	slog.Info("server start", "port", cfg.Server.Port)
	if err := http.ListenAndServe(":"+cfg.Server.Port, nil); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}
