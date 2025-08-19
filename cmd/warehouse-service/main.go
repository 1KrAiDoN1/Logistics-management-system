package main

import (
	"context"
	warehouseservice_config "logistics/configs/warehouse-service"
	"logistics/pkg/database/postgres"
	"logistics/pkg/lib/logger/slogger"
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

	dbpool := db.GetPool()
	_ = dbpool
	log.Info("Warehouse service configuration loaded successfully", "address", warehouseGRPCServiceConfig.Address)
}
