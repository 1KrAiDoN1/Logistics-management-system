package main

import (
	"context"
	authservice_config "logistics/configs/auth-service"
	"logistics/pkg/database/postgres"
	"logistics/pkg/lib/logger/slogger"
)

func main() {
	ctx := context.Background()
	log := slogger.SetupLogger()
	authGRPCServiceConfig, err := authservice_config.LoadAuthGRPCServiceConfig("configs/auth-service/auth_service_config.yaml")
	if err != nil {
		log.Error("Failed to load auth service configuration", slogger.Err(err))
		defer panic("Failed to load auth service config: " + err.Error())
	}
	db, err := postgres.NewDatabase(ctx, "")
	if err != nil {
		log.Error("Failed to connect to the database", slogger.Err(err))
		defer panic("Failed to connect to the database: " + err.Error())
	}

	dbpool := db.GetPool()
	_ = dbpool
	log.Info("Auth service configuration loaded successfully", "address", authGRPCServiceConfig.Address)
}
