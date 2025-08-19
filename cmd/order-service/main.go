package main

import (
	"context"
	orderservice_config "logistics/configs/order-service"
	"logistics/pkg/database/postgres"
	"logistics/pkg/lib/logger/slogger"
)

func main() {
	ctx := context.Background()
	log := slogger.SetupLogger()
	orderGRPCServiceConfig, err := orderservice_config.LoadOrderGRPCServiceConfig("configs/order-service/order_service_config.yaml")
	if err != nil {
		log.Error("Failed to load order service configuration", slogger.Err(err))
		defer panic("Failed to load order service config: " + err.Error())
	}
	db, err := postgres.NewDatabase(ctx, "")
	if err != nil {
		log.Error("Failed to connect to the database", slogger.Err(err))
		defer panic("Failed to connect to the database: " + err.Error())
	}

	dbpool := db.GetPool()
	_ = dbpool
	log.Info("Order service configuration loaded successfully", "address", orderGRPCServiceConfig.Address)
}
