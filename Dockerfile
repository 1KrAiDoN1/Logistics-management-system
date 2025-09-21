# --- Этап 1: Сборка ---
# Используем образ с Go для компиляции проекта
FROM golang:1.24.3-alpine AS builder

# Устанавливаем git, который может понадобиться для скачивания зависимостей
RUN apk add --no-cache git build-base

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

# Копируем файлы управления зависимостями
COPY go.mod go.sum ./
# Скачиваем зависимости
RUN go mod download

# Устанавливаем утилиту для миграций базы данных
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Копируем весь исходный код проекта
COPY . .

# Компилируем бинарные файлы для каждого микросервиса
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o ./bin/api-gateway         ./cmd/api-gateway/main.go
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o ./bin/auth-service        ./cmd/auth-service/main.go
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o ./bin/driver-service      ./cmd/driver-service/main.go
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o ./bin/order-service       ./cmd/order-service/main.go
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o ./bin/warehouse-service   ./cmd/warehouse-service/main.go


# --- Этап 2: Финальный образ ---
# Используем минимальный образ Alpine
FROM alpine:latest

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем скомпилированные бинарные файлы из этапа сборки
COPY --from=builder /app/bin /app/bin/
# Копируем утилиту для миграций
COPY --from=builder /go/bin/migrate /usr/local/bin/migrate
# Копируем конфигурационные файлы и миграции
COPY --from=builder /app/configs/api-gateway ./configs/api-gateway
COPY --from=builder /app/configs/auth-service ./configs/auth-service
COPY --from=builder /app/configs/driver-service ./configs/driver-service
COPY --from=builder /app/configs/order-service ./configs/order-service
COPY --from=builder /app/configs/warehouse-service ./configs/warehouse-service

COPY --from=builder /app/migrations ./migrations

COPY entrypoint.sh /usr/local/bin/entrypoint.sh

RUN chmod +x /usr/local/bin/entrypoint.sh

ENTRYPOINT ["entrypoint.sh"]
CMD ["sh", "-c", "echo 'This is a multi-service image. Please specify a binary to run, e.g., /app/bin/api-gateway' && ls -l /app/bin"]