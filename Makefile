include .env
export $(shell sed 's/=.*//' .env)

DB_URL = postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@localhost:5432/${POSTGRES_DB}?sslmode=disable


proto-auth:
	protoc -I api/protobuf/ \
		api/protobuf/auth_service/auth_service.proto \
		--go_out=./api/protobuf \
		--go_opt=paths=source_relative \
		--go-grpc_out=./api/protobuf \
		--go-grpc_opt=paths=source_relative

# 2. папки, где лежат прото файлы
# 3. полный путь к прото файлу, который нужно сгенерировать
# 4. папка, куда будут сгенерированы go файлы
# 5. параметры для генерации go файлов
# 6. папка, куда будут сгенерированы go файлы
# 7. параметры для генерации grpc файлов


proto-order:
	protoc -I api/protobuf/ \
		api/protobuf/order_service/order_service.proto \
		--go_out=./api/protobuf \
		--go_opt=paths=source_relative \
		--go-grpc_out=./api/protobuf \
		--go-grpc_opt=paths=source_relative

proto-driver:
	protoc -I api/protobuf/ \
		api/protobuf/driver_service/driver_service.proto \
		--go_out=./api/protobuf \
		--go_opt=paths=source_relative \
		--go-grpc_out=./api/protobuf \
		--go-grpc_opt=paths=source_relative

proto-warehouse:
	protoc -I api/protobuf/ \
		api/protobuf/warehouse_service/warehouse_service.proto \
		--go_out=./api/protobuf \
		--go_opt=paths=source_relative \
		--go-grpc_out=./api/protobuf \
		--go-grpc_opt=paths=source_relative


create-kafka-topics:
	kafka-topics --bootstrap-server localhost:9092 --create \
  --topic drivers \
  --partitions 3 \
  --replication-factor 1

migrate-all-up:
	migrate -path ./migrations -database "${DB_URL}" up

migrate-all-down:
	migrate -path ./migrations -database "${DB_URL}" down






# Makefile для управления Docker-контейнерами проекта

# --- Переменные ---
# Используем флаг -p для указания имени проекта, чтобы избежать конфликтов
# и сделать команды короче.
COMPOSE_CMD = docker-compose -p logistics

# --- Основные команды ---
.PHONY: help build up down start stop logs ps clean

# Цель по умолчанию, если просто запустить 'make'
.DEFAULT_GOAL := help

help: ## Показывает это справочное сообщение
  @echo "Доступные команды:"
  @grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Собрать или пересобрать образы всех сервисов
  @echo "Сборка Docker-образов..."
  $(COMPOSE_CMD) build

up: ## Запустить все сервисы в фоновом режиме (detached)
  @echo "Запуск всех сервисов..."
  $(COMPOSE_CMD) up -d --build

down: ## Остановить и удалить контейнеры, сети
  @echo "Остановка и удаление контейнеров..."
  $(COMPOSE_CMD) down

start: ## Запустить ранее остановленные контейнеры
  @echo "Запуск сервисов..."
  $(COMPOSE_CMD) start

stop: ## Остановить запущенные контейнеры, не удаляя их
  @echo "Остановка сервисов..."
  $(COMPOSE_CMD) stop

logs: ## Показать и отслеживать логи всех сервисов. Пример: make logs service=api-gateway
  @echo "Отслеживание логов..."
  $(COMPOSE_CMD) logs -f $(service)

ps: ## Показать статус запущенных контейнеров
  @echo "Текущий статус контейнеров:"
  $(COMPOSE_CMD) ps

clean: ## Полная очистка: остановить, удалить контейнеры, сети и тома (удалит данные БД!)
  @echo "ВНИМАНИЕ: Удаление всех контейнеров, сетей и данных..."
  $(COMPOSE_CMD) down -v

# --- Команды для работы с миграциями ---
.PHONY: migrate-up migrate-down migrate-create

# Базовая команда для запуска утилиты миграций внутри контейнера
MIGRATE_CMD = migrate -path /app/migrations -database 'postgres://logistics_user:logistics_password@db:5432/logistics_db?sslmode=disable' -verbose

migrate-up: ## Применить все доступные 'up' миграции
  @echo "Применение миграций..."
  $(COMPOSE_CMD) run --rm migrations sh -c "$(MIGRATE_CMD) up"

migrate-down: ## Откатить последнюю примененную миграцию
  @echo "Откат последней миграции..."
  $(COMPOSE_CMD) run --rm migrations sh -c "$(MIGRATE_CMD) down 1"

migrate-create: ## Создать новые файлы миграции. Пример: make migrate-create name=add_users_table
  @echo "Создание файла миграции '$(name)'..."
  $(COMPOSE_CMD) run --rm migrations sh -c "migrate create -ext sql -dir /app/migrations -seq $(name)"

# --- Утилиты для разработки ---
.PHONY: shell

shell: ## Зайти в интерактивную оболочку сервиса. Пример: make shell service=api-gateway
  @echo "Подключение к оболочке контейнера $(service)..."
  $(COMPOSE_CMD) exec $(service) sh