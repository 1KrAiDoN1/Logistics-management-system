package configs

import (
	apigateway_config "logistics/configs/api-gateway"
	authservice_config "logistics/configs/auth-service"
	driverservice_config "logistics/configs/driver-service"
	orderservice_config "logistics/configs/order-service"
	routeservice_config "logistics/configs/route-service"
	warehouseservice_config "logistics/configs/warehouse-service"
	"logistics/pkg/lib/utils"
)

type MicroservicesConfig struct {
	ApiGatewayConfig           *apigateway_config.Config
	AuthGRPCServiceConfig      utils.ServiceConfig
	OrderGRPCServiceConfig     utils.ServiceConfig
	DriverGRPCServiceConfig    utils.ServiceConfig
	WarehouseGRPCServiceConfig utils.ServiceConfig
	RouteGRPCServiceConfig     utils.ServiceConfig
}

func NewMicroservicesConfig() (*MicroservicesConfig, error) {
	apiGatewayConfig, err := apigateway_config.LoadConfigServer("configs/api-gateway/api_gateway_config.yaml")
	if err != nil {
		defer panic("Failed to load API Gateway config: " + err.Error())
		return nil, err
	}
	authGRPCServiceConfig, err := authservice_config.LoadAuthGRPCServiceConfig("configs/auth-service/auth_service_config.yaml")
	if err != nil {
		defer panic("Failed to load auth service config: " + err.Error())
		return nil, err
	}
	orderGRPCServiceConfig, err := orderservice_config.LoadOrderGRPCServiceConfig("configs/order-service/order_service_config.yaml")
	if err != nil {
		defer panic("Failed to load order service config: " + err.Error())
		return nil, err
	}
	driverGRPCServiceConfig, err := driverservice_config.LoadDriverGRPCServiceConfig("configs/driver-service/driver_service_config.yaml")
	if err != nil {
		defer panic("Failed to load driver service config: " + err.Error())
		return nil, err
	}
	warehouseGRPCServiceConfig, err := warehouseservice_config.LoadWarehouseGRPCServiceConfig("configs/warehouse-service/warehouse_service_config.yaml")
	if err != nil {
		defer panic("Failed to load warehouse service config: " + err.Error())
		return nil, err
	}
	routeGRPCServiceConfig, err := routeservice_config.LoadRouteGRPCServiceConfig("configs/route-service/route_service_config.yaml")
	if err != nil {
		defer panic("Failed to load route service config: " + err.Error())
		return nil, err
	}

	return &MicroservicesConfig{
		ApiGatewayConfig:           apiGatewayConfig,
		AuthGRPCServiceConfig:      authGRPCServiceConfig,
		OrderGRPCServiceConfig:     orderGRPCServiceConfig,
		DriverGRPCServiceConfig:    driverGRPCServiceConfig,
		WarehouseGRPCServiceConfig: warehouseGRPCServiceConfig,
		RouteGRPCServiceConfig:     routeGRPCServiceConfig,
	}, nil
}
