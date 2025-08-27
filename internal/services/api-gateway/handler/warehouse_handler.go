package handler

import (
	"log/slog"
	warehousepb "logistics/api/protobuf/warehouse_service"

	"github.com/gin-gonic/gin"
)

type WarehouseHandler struct {
	logger              *slog.Logger
	warehouseGRPCClient warehousepb.WarehouseServiceClient
}

func NewWarehouseHandler(logger *slog.Logger, warehouseClient warehousepb.WarehouseServiceClient) *WarehouseHandler {
	return &WarehouseHandler{
		logger:              logger,
		warehouseGRPCClient: warehouseClient,
	}
}

func (w *WarehouseHandler) GetAvailableProducts(c *gin.Context) {
	// Implementation for getting available products from the warehouse
}
