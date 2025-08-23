package main

import (
	"log/slog"
	"logistics/configs"
	"logistics/pkg/cache/redis"
	"logistics/pkg/lib/logger/slogger"
	"os"
)

func main() {
	log := slogger.SetupLogger()
	microservices_config, err := configs.NewMicroservicesConfig()
	if err != nil {
		log.Error("Failed to load microservices configuration", slogger.Err(err))
		os.Exit(1)

	}
	redisClient, err := redis.NewRedisClient(microservices_config.ApiGatewayConfig.RedisConfig)
	if err != nil {
		log.Error("Failed to connect to Redis", slogger.Err(err))
		os.Exit(1)
	}
	redis_health, err := redisClient.HealthCheck()
	if err != nil {
		log.Error("Redis health check failed", slogger.Err(err))
		os.Exit(1)
	}
	defer redisClient.Close()

	log.Info("info_redis", slog.Any("redis_hits", redis_health.Hits), slog.Any("redis_misses", redis_health.Misses),
		slog.Any("redis_uptime", redis_health.IdleConns), slog.Any("redis_idle_conns", redis_health.StaleConns),
		slog.Any("redis_total_conns", redis_health.TotalConns), slog.Any("redis_max_conns", redis_health.WaitCount))

	log.Info("Microservices Configuration",
		slog.String("api-gateway_redis_address", microservices_config.ApiGatewayConfig.RedisConfig.Address),
		slog.String("api-gateway_db", microservices_config.ApiGatewayConfig.DB_config_path),
		slog.String("api-gateway_address", microservices_config.ApiGatewayConfig.HTTPServer.Address),
		slog.Any("auth_service", microservices_config.AuthGRPCServiceConfig),
		slog.Any("order_service", microservices_config.OrderGRPCServiceConfig),
		slog.Any("driver_service", microservices_config.DriverGRPCServiceConfig),
		slog.Any("warehouse_service", microservices_config.WarehouseGRPCServiceConfig),
	)

}
