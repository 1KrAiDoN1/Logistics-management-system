package main

import (
	"context"
	driverservice_config "logistics/configs/driver-service"
	"logistics/pkg/database/postgres"
	"logistics/pkg/lib/logger/slogger"
)

func main() {
	ctx := context.Background()
	log := slogger.SetupLogger()
	driverGRPCServiceConfig, err := driverservice_config.LoadDriverGRPCServiceConfig("configs/driver-service/driver_service_config.yaml")
	if err != nil {
		log.Error("Failed to load driver service configuration", slogger.Err(err))
		panic("Failed to load driver service config: " + err.Error())
	}
	db, err := postgres.NewDatabase(ctx, "")
	if err != nil {
		log.Error("Failed to connect to the database", slogger.Err(err))
		defer panic("Failed to connect to the database: " + err.Error())
	}

	dbpool := db.GetPool()
	_ = dbpool
	log.Info("Driver service configuration loaded successfully", "address", driverGRPCServiceConfig.Address)
}
