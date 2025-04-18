package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"
	"weather-api/config"
	"weather-api/internal/adapters/postgres"
	"weather-api/internal/adapters/telegram"
	adapters "weather-api/internal/adapters/weather_client"
	"weather-api/internal/controllers"
	usecase "weather-api/internal/usecase"

	"github.com/golang-migrate/migrate/v4"
	migratepostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

func main() {
	// Загрузка .env файла
	if err := godotenv.Load(); err != nil {
		slog.Warn("No .env file found, relying on environment variables", "err", err)
	}
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	slog.SetDefault(slog.New(handler))

	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("load config failed", "err", err)
		os.Exit(1)
	}

	// Подключение к Postgres
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.Username, cfg.Postgres.Password, cfg.Postgres.DB)
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		slog.Error("failed to connect to postgres", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	// Ожидание готовности PostgreSQL
	for i := 0; i < 10; i++ {
		if err := db.Ping(); err != nil {
			slog.Warn("failed to ping postgres, retrying", "attempt", i+1, "err", err)
			time.Sleep(2 * time.Second)
			continue
		}
		slog.Info("successfully connected to postgres")
		break
	}
	if err := db.Ping(); err != nil {
		slog.Error("failed to ping postgres after retries", "err", err)
		os.Exit(1)
	}

	// Применение миграций
	driver, err := migratepostgres.WithInstance(db, &migratepostgres.Config{})
	if err != nil {
		slog.Error("failed to create migration driver", "err", err)
		os.Exit(1)
	}
	m, err := migrate.NewWithDatabaseInstance("file:///app/migrations", "postgres", driver)
	if err != nil {
		slog.Error("failed to initialize migrations", "err", err)
		os.Exit(1)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		slog.Error("failed to apply migrations", "err", err)
		os.Exit(1)
	}

	cityRepo := postgres.NewCityRepository(db)
	client := adapters.NewClient(adapters.ClientOptions{URL: cfg.WeatherAPI.URL})
	weatherUsecase := usecase.NewWeatherUseCase(usecase.WeatherUseCaseOptions{
		WeatherClient:  client,
		CityRepository: cityRepo,
	})
	weatherController := controllers.NewWeatherController(controllers.WeatherControllerOptions{
		WeatherUseCase: weatherUsecase,
	})

	// Телеграм бот
	bot, err := telegram.NewBot(cfg.Telegram.Token, weatherUsecase)
	if err != nil {
		slog.Error("failed to create telegram bot", "err", err)
		os.Exit(1)
	}
	go func() {
		ctx := context.Background()
		if err := bot.Start(ctx); err != nil {
			slog.Error("telegram bot failed", "err", err)
		}
	}()

	http.HandleFunc("/api/v1/weather", weatherController.GetWeatherToday)
	http.HandleFunc("/api/v1/weather/city", weatherController.GetWeatherByCity)

	slog.Info("server start", "port", cfg.Server.Port)
	if err := http.ListenAndServe(":"+cfg.Server.Port, nil); err != nil {
		slog.Error("server failed", "err", err)
		os.Exit(1)
	}
}
