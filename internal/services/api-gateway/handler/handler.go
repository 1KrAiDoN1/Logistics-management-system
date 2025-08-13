package handler

import (
	"logistics/configs"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	authGRPCClient      authpb.UserServiceClient
	orderGRPCClient     orderpb.OrderServiceClient
	driverGRPCClient    driverpb.DriverServiceClient
	warehouseGRPCClient warehousepb.WarehouseServiceClient
	routeGRPCClient     routepb.RouteServiceClient
}

func NewHandler(
	authClient authpb.UserServiceClient,
	orderClient orderpb.OrderServiceClient,
	driverClient driverpb.DriverServiceClient,
	warehouseClient warehousepb.WarehouseServiceClient,
	routeClient routepb.RouteServiceClient,
) *Handler {
	return &Handler{
		authGRPCClient:      authClient,
		orderGRPCClient:     orderClient,
		driverGRPCClient:    driverClient,
		warehouseGRPCClient: warehouseClient,
		routeGRPCClient:     routeClient,
	}
}

func (h *Handler) SignUp(c *gin.Context) {
	var req configs.MicroservicesConfig // Из shared/models/dto
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
}
