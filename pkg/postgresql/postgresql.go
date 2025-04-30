package postgresql

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // Импортируем драйвер pq
)

// PostgresClient обертка над sqlx.DB для работы с PostgreSQL
type PostgresClient struct {
	*sqlx.DB
}

// Config параметры подключения к PostgreSQL
type Config struct {
	Host     string
	Port     string
	Username string
	Password string
	DBName   string
	SSLMode  string
}

// Option паттерн для настройки параметров подключения
type Option func(*Config)

// Задает хост
func WithHost(host string) Option {
	return func(c *Config) {
		c.Host = host
	}
}

// Задает порт
func WithPort(port string) Option {
	return func(c *Config) {
		c.Port = port
	}
}

// WithUsername задает имя пользователя
func WithUsername(username string) Option {
	return func(c *Config) {
		c.Username = username
	}
}

// WithPassword задает пароль
func WithPassword(password string) Option {
	return func(c *Config) {
		c.Password = password
	}
}

// WithDBName задает имя базы данных
func WithDBName(dbName string) Option {
	return func(c *Config) {
		c.DBName = dbName
	}
}

// WithSSLMode задает режим SSL
func WithSSLMode(sslMode string) Option {
	return func(c *Config) {
		c.SSLMode = sslMode
	}
}

// Новый клиент
func NewPostgres(options ...Option) (*PostgresClient, error) {
	// Значения по умолчанию
	cfg := &Config{
		Host:     "localhost",
		Port:     "5432",
		Username: "postgres",
		Password: "postgres",
		DBName:   "",
		SSLMode:  "disable",
	}

	// Применяем опции
	for _, opt := range options {
		opt(cfg)
	}

	// Формируем строку подключения
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.DBName, cfg.SSLMode)

	// Подключаемся к базе с использованием sqlx и драйвера pq
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	// Проверяем подключение с несколькими попытками
	for i := 0; i < 10; i++ {
		if err := db.Ping(); err != nil {
			time.Sleep(2 * time.Second)
			continue
		}
		break
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping postgres after retries: %w", err)
	}

	return &PostgresClient{DB: db}, nil
}
