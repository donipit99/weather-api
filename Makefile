.PHONY: create-network up-all stop-all hard-reload-kafka monitoring-up monitoring-down all-up all-down

create-network:
	docker network create weather-network || true

up-all: create-network
	docker-compose -f docker-compose.db.yml -f docker-compose.redis.yml -f docker-compose.kafka.yml -f docker-compose.app.yml up -d

stop-all:
	docker-compose -f docker-compose.app.yml -f docker-compose.kafka.yml -f docker-compose.redis.yml -f docker-compose.db.yml stop

hard-reload-kafka:
	docker-compose -f docker-compose.kafka.yml stop
	docker-compose -f docker-compose.kafka.yml rm -v -f
	docker-compose -f docker-compose.kafka.yml up -d

# Запуск мониторинга
monitoring-up:
	docker-compose -f docker-compose.monitoring.yml up -d

# Остановка мониторинга
monitoring-down:
	docker-compose -f docker-compose.monitoring.yml down

# Запуск всех контейнеров (БД, Redis, мониторинг)
all-up:
	docker-compose -f docker-compose.db.yml -f docker-compose.redis.yml -f docker-compose.monitoring.yml up -d
	
# Остановка всех контейнеров
all-down:
	docker-compose -f docker-compose.db.yml -f docker-compose.redis.yml -f docker-compose.monitoring.yml down