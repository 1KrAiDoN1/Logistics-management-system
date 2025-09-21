#!/bin/sh

DB_HOST=${DB_HOST:-db}
REDIS_HOST=${REDIS_HOST:-redis}
KAFKA_HOST=${KAFKA_HOST:-kafka}
AUTH_SERVICE_HOST=${AUTH_SERVICE_HOST:-auth-service}
DRIVER_SERVICE_HOST=${DRIVER_SERVICE_HOST:-driver-service}
ORDER_SERVICE_HOST=${ORDER_SERVICE_HOST:-order-service}
WAREHOUSE_SERVICE_HOST=${WAREHOUSE_SERVICE_HOST:-warehouse-service}

echo "Заменяем хосты в конфигурационных файлах..."

find /app/configs -type f -name '*.yaml' -exec sed -i "s/host: localhost/host: ${DB_HOST}/g" {} +
find /app/configs -type f -name '*.yaml' -exec sed -i "s/address: \"127.0.0.1:6379\"/address: \"${REDIS_HOST}:6379\"/g" {} +
find /app/configs -type f -name '*.yaml' -exec sed -i "s/- \"localhost:9092\"/- \"${KAFKA_HOST}:9092\"/g" {} +

find $CONFIG_PATH -type f -name '*.yaml' -exec sed -i "s/address: \"localhost:40001\"/address: \"${AUTH_SERVICE_HOST}:40001\"/g" {} +
find $CONFIG_PATH -type f -name '*.yaml' -exec sed -i "s/address: \"localhost:40002\"/address: \"${DRIVER_SERVICE_HOST}:40002\"/g" {} +
find $CONFIG_PATH -type f -name '*.yaml' -exec sed -i "s/address: \"localhost:40003\"/address: \"${ORDER_SERVICE_HOST}:40003\"/g" {} +
find $CONFIG_PATH -type f -name '*.yaml' -exec sed -i "s/address: \"localhost:40005\"/address: \"${WAREHOUSE_SERVICE_HOST}:40005\"/g" {} +

echo "Замена завершена. Запускаем основное приложение..."

exec "$@"