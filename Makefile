include .env
export $(shell sed 's/=.*//' .env)
COMPOSE_CMD = docker-compose
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


build:
  @echo "Сборка Docker-образов..."
  $(COMPOSE_CMD) build

up:
  @echo "Запуск всех сервисов..."
  $(COMPOSE_CMD) up 

down: 
  @echo "Остановка и удаление контейнеров..."
  $(COMPOSE_CMD) down

logs: 
  @echo "Отслеживание логов..."
  $(COMPOSE_CMD) logs -f $(service)

ps: 
  @echo "Текущий статус контейнеров:"
  $(COMPOSE_CMD) ps

clean:
  @echo "ВНИМАНИЕ: Удаление всех контейнеров, сетей и данных..."
  $(COMPOSE_CMD) down -v
