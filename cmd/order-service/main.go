package main

import (
	"context"
	orderservice_config "logistics/configs/order-service"
	orderservice "logistics/internal/services/order-service"
	"logistics/internal/services/order-service/grpc/app"
	"logistics/internal/services/order-service/repository"
	"logistics/pkg/cache/redis"
	"logistics/pkg/database/postgres"
	"logistics/pkg/lib/logger/slogger"
	"os"
)

func main() {
	ctx := context.Background()
	log := slogger.SetupLogger()
	orderGRPCServiceConfig, err := orderservice_config.LoadOrderGRPCServiceConfig("configs/order-service/order_service_config.yaml")
	if err != nil {
		log.Error("Failed to load order service configuration", slogger.Err(err))
		os.Exit(1)
	}
	db, err := postgres.NewDatabase(ctx, "")
	if err != nil {
		log.Error("Failed to connect to the database", slogger.Err(err))
		os.Exit(1)
	}
	dbpool := db.GetPool()

	redis, err := redis.NewRedisClient(orderGRPCServiceConfig.RedisConfig)
	if err != nil {
		log.Error("Failed to connect to Redis", slogger.Err(err))
		os.Exit(1)
	}

	orderGRPCRepository := repository.NewOrderRepository(dbpool)
	orderGRPCService := orderservice.NewOrderGRPCService(log, orderGRPCRepository, redis.Client)
	orderGRPCApp := app.NewApp(log, orderGRPCService, orderGRPCServiceConfig)
	log.Info("Auth service configuration loaded successfully", "address", orderGRPCServiceConfig.Address)

	if err := orderGRPCApp.Run(); err != nil {
		log.Error("Failed to run auth gRPC application", slogger.Err(err))
		os.Exit(1)
	}

	log.Info("Order service configuration loaded successfully", "address", orderGRPCServiceConfig.Address)
}
