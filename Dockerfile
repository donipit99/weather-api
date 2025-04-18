# Этап сборки
FROM golang:1.24.2 AS builder

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем go.mod и go.sum для загрузки зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь исходный код
COPY . .

# Проверяем наличие migrations в этапе сборки
RUN ls -l migrations

# Собираем бинарник
RUN CGO_ENABLED=0 GOOS=linux go build -o weather-api ./cmd/weather_api

# Финальный этап
FROM alpine:3.21
WORKDIR /app

# Копируем бинарник из этапа сборки
COPY --from=builder /app/weather-api .
# Копируем миграции
COPY --from=builder /app/migrations ./migrations

# Проверяем наличие migrations в финальном образе
RUN ls -l /app/migrations

# Открываем порт
EXPOSE 8080

# Команда для запуска приложения
CMD ["./weather-api"]