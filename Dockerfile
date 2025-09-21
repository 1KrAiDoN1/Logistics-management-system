FROM golang:1.24.3-alpine AS builder

RUN apk add --no-cache git build-base

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

COPY . .

RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o ./bin/api-gateway         ./cmd/api-gateway/main.go
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o ./bin/auth-service        ./cmd/auth-service/main.go
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o ./bin/driver-service      ./cmd/driver-service/main.go
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o ./bin/order-service       ./cmd/order-service/main.go
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o ./bin/warehouse-service   ./cmd/warehouse-service/main.go


FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/bin /app/bin/
COPY --from=builder /go/bin/migrate /usr/local/bin/migrate
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