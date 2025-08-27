package main

import (
	"context"
	authservice_config "logistics/configs/auth-service"
	auth_grpc_server "logistics/internal/services/auth-service/grpc"
	"logistics/internal/services/auth-service/grpc/app"
	auth_grpc_repository "logistics/internal/services/auth-service/grpc/repository"
	"logistics/pkg/cache/redis"
	"logistics/pkg/database/postgres"
	"logistics/pkg/lib/logger/slogger"
	"os"
)

func main() {
	ctx := context.Background()
	log := slogger.SetupLogger()
	authGRPCServiceConfig, err := authservice_config.LoadAuthGRPCServiceConfig("configs/auth-service/auth_service_config.yaml")
	if err != nil {
		log.Error("Failed to load auth service configuration", slogger.Err(err))
		os.Exit(1)
	}
	dbConnstr, err := authGRPCServiceConfig.DSN("DB_AUTH_SERVICE_PASSWORD")
	if err != nil {
		log.Error("Failed to get database connection string", slogger.Err(err))
		os.Exit(1)
	}
	log.Info("Connecting to the database", "connection_string", dbConnstr)
	db, err := postgres.NewDatabase(ctx, dbConnstr)
	if err != nil {
		log.Error("Failed to connect to the database", slogger.Err(err))
		os.Exit(1)
	}
	dbpool := db.GetPool()

	redis, err := redis.NewRedisClient(authGRPCServiceConfig.RedisConfig)
	if err != nil {
		log.Error("Failed to connect to Redis", slogger.Err(err))
		os.Exit(1)
	}

	authGRPCRepository := auth_grpc_repository.NewAuthRepository(dbpool)
	authGRPCService := auth_grpc_server.NewAuthGRPCService(log, authGRPCRepository, redis.Client)
	authGRPCApp := app.NewApp(log, authGRPCService, authGRPCServiceConfig)
	log.Info("Auth service configuration loaded successfully", "address", authGRPCServiceConfig.Address)

	if err := authGRPCApp.Run(); err != nil {
		log.Error("Failed to run auth gRPC application", slogger.Err(err))
		os.Exit(1)
	}
}
