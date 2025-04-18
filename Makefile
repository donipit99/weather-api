.PHONY: build up down restart hard-reload logs clean

# Собрать образы
build:
	docker compose build

# Запустить сервисы
up:
	docker compose up -d

# Остановить сервисы
down:
	docker compose down

# Перезапустить сервисы
restart: down up

# Полная пересборка и перезапуск
hard-reload:
	docker compose down --volumes --remove-orphans
	docker compose build --no-cache
	docker compose up -d --force-recreate

# Показать логи
logs:
	docker compose logs -f

# Очистить неиспользуемые образы и контейнеры
clean:
	docker system prune -f

	