package main

import (
	"context"
	warehouseservice_config "logistics/configs/warehouse-service"
	warehouseservice "logistics/internal/services/warehouse-service"
	"logistics/internal/services/warehouse-service/grpc/app"
	"logistics/internal/services/warehouse-service/repository"
	"logistics/pkg/cache/redis"
	"logistics/pkg/database/postgres"
	"logistics/pkg/lib/logger/slogger"
	"os"
)

func main() {
	ctx := context.Background()
	log := slogger.SetupLogger()
	warehouseGRPCServiceConfig, err := warehouseservice_config.LoadWarehouseGRPCServiceConfig("configs/warehouse-service/warehouse_service_config.yaml")
	if err != nil {
		log.Error("Failed to load warehouse service configuration", slogger.Err(err))
		defer panic("Failed to load warehouse service config: " + err.Error())
	}
	db, err := postgres.NewDatabase(ctx, "")
	if err != nil {
		log.Error("Failed to connect to the database", slogger.Err(err))
		defer panic("Failed to connect to the database: " + err.Error())
	}
	redis, err := redis.NewRedisClient(warehouseGRPCServiceConfig.RedisConfig)
	if err != nil {
		log.Error("Failed to connect to Redis", slogger.Err(err))
		os.Exit(1)
	}
	dbpool := db.GetPool()
	warehouseGRPCRepository := repository.NewWarehouseRepository(dbpool)
	warehouseGRPCService := warehouseservice.NewWarehouseGRPCService(log, warehouseGRPCRepository, redis.Client)

	warehouseGRPCApp := app.NewApp(log, warehouseGRPCService, warehouseGRPCServiceConfig)
	if err := warehouseGRPCApp.Run(); err != nil {
		log.Error("Failed to run warehouse gRPC application", slogger.Err(err))
		os.Exit(1)
	}
	log.Info("Warehouse service configuration loaded successfully", "address", warehouseGRPCServiceConfig.Address)
}
