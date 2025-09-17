package main

import (
	"log/slog"
	"logistics/configs"
	_ "logistics/docs"
	httpserver "logistics/internal/services/api-gateway/http-server"
	"logistics/pkg/lib/logger/slogger"
	"os"
)

// ============================================
// SWAGGER CONFIGURATION
// ============================================

// @title Logistics Management API
// @version 1.0
// @description API для управления логистическими операциями включая аутентификацию, заказы, доставки и управление складом, построенный на базе нескольких микросервисов с использованием gRPC и REST.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:9091
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Введите 'Bearer ' followed by your JWT token
func main() {
	log := slogger.SetupLogger()
	microservices_config, err := configs.NewMicroservicesConfig()
	if err != nil {
		log.Error("Failed to load microservices configuration", slogger.Err(err))
		os.Exit(1)

	}

	log.Info("Microservices Configuration",
		slog.String("api-gateway_address", microservices_config.ApiGatewayConfig.HTTPServer.Address),
		slog.Any("auth_service", microservices_config.AuthGRPCServiceConfig),
		slog.Any("order_service", microservices_config.OrderGRPCServiceConfig),
		slog.Any("driver_service", microservices_config.DriverGRPCServiceConfig),
		slog.Any("warehouse_service", microservices_config.WarehouseGRPCServiceConfig),
	)
	server := httpserver.NewServer(log, microservices_config)
	if err := server.Run(); err != nil {
		log.Error("Failed to run HTTP server", slogger.Err(err))
		os.Exit(1)
	}

}
