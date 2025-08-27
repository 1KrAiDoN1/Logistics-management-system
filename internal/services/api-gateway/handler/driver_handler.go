package handler

import (
	"log/slog"
	driverpb "logistics/api/protobuf/driver_service"
)

type DriverHandler struct {
	logger           *slog.Logger
	driverGRPCClient driverpb.DriverServiceClient
}

func NewDriverHandler(logger *slog.Logger, driverClient driverpb.DriverServiceClient) *DriverHandler {
	return &DriverHandler{
		logger:           logger,
		driverGRPCClient: driverClient,
	}
}
