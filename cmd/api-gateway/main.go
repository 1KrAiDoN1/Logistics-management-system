package main

import (
	"log/slog"
	"logistics/configs"
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
	log.Info("Microservices Configuration",
		slog.String("api-gateway_db", microservices_config.ApiGatewayConfig.DB_config_path),
		slog.String("api-gateway_address", microservices_config.ApiGatewayConfig.HTTPServer.Address),
		slog.String("auth_service", microservices_config.AuthGRPCServiceConfig),
		slog.String("order_service", microservices_config.OrderGRPCServiceConfig),
		slog.String("driver_service", microservices_config.DriverGRPCServiceConfig),
		slog.String("warehouse_service", microservices_config.WarehouseGRPCServiceConfig),
		slog.String("route_service", microservices_config.RouteGRPCServiceConfig),
	)

}
