package main

import (
	"context"
	routeservice_config "logistics/configs/route-service"
	"logistics/pkg/database/postgres"
	"logistics/pkg/lib/logger/slogger"
)

func main() {
	ctx := context.Background()
	log := slogger.SetupLogger()
	routeGRPCServiceConfig, err := routeservice_config.LoadRouteGRPCServiceConfig("configs/route-service/route_service_config.yaml")
	if err != nil {
		log.Error("Failed to load route service configuration", slogger.Err(err))
		defer panic("Failed to load route service config: " + err.Error())
	}
	db, err := postgres.NewDatabase(ctx, "")
	if err != nil {
		log.Error("Failed to connect to the database", slogger.Err(err))
		defer panic("Failed to connect to the database: " + err.Error())
	}

	dbpool := db.GetPool()
	_ = dbpool
	log.Info("Route service configuration loaded successfully", "address", routeGRPCServiceConfig.Address)
}
