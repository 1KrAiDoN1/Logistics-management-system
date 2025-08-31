package handler

import (
	"context"
	"log/slog"
	warehousepb "logistics/api/protobuf/warehouse_service"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/types/known/emptypb"
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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	products, err := w.warehouseGRPCClient.GetWarehouseStock(ctx, &emptypb.Empty{})
	if err != nil {
		w.logger.Error("Failed to get available products", slog.String("status", "500"), slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	w.logger.Info("Available products retrieved successfully", slog.String("status", "200"))
	c.JSON(http.StatusOK, gin.H{
		"message":  "Available products",
		"products": products.Stocks},
	)
}
