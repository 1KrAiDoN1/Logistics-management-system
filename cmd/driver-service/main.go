package main

import (
	"context"
	driverservice_config "logistics/configs/driver-service"
	driverservice "logistics/internal/services/driver-service"
	"logistics/internal/services/driver-service/grpc/app"
	"logistics/internal/services/driver-service/repository"
	"logistics/pkg/cache/redis"
	"logistics/pkg/database/postgres"
	"logistics/pkg/lib/logger/slogger"
	"os"
)

func main() {
	ctx := context.Background()
	log := slogger.SetupLogger()
	driverGRPCServiceConfig, err := driverservice_config.LoadDriverGRPCServiceConfig("configs/driver-service/driver_service_config.yaml")
	if err != nil {
		log.Error("Failed to load driver service configuration", slogger.Err(err))
		os.Exit(1)
	}
	dbConnstr, err := driverGRPCServiceConfig.DSN("DB_DRIVER_SERVICE_PASSWORD")
	if err != nil {
		log.Error("Failed to get database connection string", slogger.Err(err))
		os.Exit(1)
	}
	_ = dbConnstr
	db, err := postgres.NewDatabase(ctx, "")
	if err != nil {
		log.Error("Failed to connect to the database", slogger.Err(err))
		os.Exit(1)
	}
	defer db.Close()

	redis, err := redis.NewRedisClient(driverGRPCServiceConfig.RedisConfig)
	if err != nil {
		log.Error("Failed to connect to Redis", slogger.Err(err))
		os.Exit(1)
	}
	defer redis.Close()

	dbpool := db.GetPool()
	driverGRPCRepository := repository.NewDriverRepository(dbpool)
	driverGRPCService := driverservice.NewDriverGRPCService(log, driverGRPCRepository, redis.Client)

	driverGRPCApp := app.NewApp(log, driverGRPCService, driverGRPCServiceConfig)
	if err := driverGRPCApp.Run(); err != nil {
		log.Error("Failed to run driver gRPC application", slogger.Err(err))
		os.Exit(1)
	}
	if err := driverGRPCApp.Run(); err != nil {
		log.Error("Failed to run driver gRPC application", slogger.Err(err))
		os.Exit(1)
	}
	log.Info("Driver service configuration loaded successfully", "address", driverGRPCServiceConfig.Address)
}
