package handler

import (
	"log/slog"
	authpb "logistics/api/protobuf/auth_service"
	driverpb "logistics/api/protobuf/driver_service"
	orderpb "logistics/api/protobuf/order_service"
	warehousepb "logistics/api/protobuf/warehouse_service"
)

type Handlers struct {
	AuthHandlerInterface
	OrderHandlerInterface
	WarehouseHandlerInterface
}

func NewHandlers(logger *slog.Logger, authGRPCClient authpb.AuthServiceClient, orderGRPCClient orderpb.OrderServiceClient, driverGRPCClient driverpb.DriverServiceClient, warehouseGRPCClient warehousepb.WarehouseServiceClient) *Handlers {
	return &Handlers{
		AuthHandlerInterface:      NewAuthHandler(logger, authGRPCClient),
		OrderHandlerInterface:     NewOrderHandler(logger, orderGRPCClient, driverGRPCClient, warehouseGRPCClient),
		WarehouseHandlerInterface: NewWarehouseHandler(logger, warehouseGRPCClient),
	}
}
