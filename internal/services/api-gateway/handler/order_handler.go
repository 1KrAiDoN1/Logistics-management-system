package handler

import (
	"context"
	"fmt"
	"log/slog"
	driverpb "logistics/api/protobuf/driver_service"
	orderpb "logistics/api/protobuf/order_service"
	warehousepb "logistics/api/protobuf/warehouse_service"
	"logistics/internal/services/api-gateway/middleware"
	"logistics/internal/shared/entity"
	"logistics/internal/shared/models/dto"
	"logistics/pkg/lib/logger/slogger"
	"logistics/pkg/lib/utils"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
)

type OrderHandler struct {
	logger              *slog.Logger
	orderGRPCClient     orderpb.OrderServiceClient
	driverGRPCClient    driverpb.DriverServiceClient
	warehouseGRPCClient warehousepb.WarehouseServiceClient
	stockCheckWorkers   chan struct{}
}

func NewOrderHandler(logger *slog.Logger, orderClient orderpb.OrderServiceClient, driverClient driverpb.DriverServiceClient, warehouseClient warehousepb.WarehouseServiceClient) *OrderHandler {
	return &OrderHandler{
		logger:              logger,
		orderGRPCClient:     orderClient,
		driverGRPCClient:    driverClient,
		warehouseGRPCClient: warehouseClient,
		stockCheckWorkers:   make(chan struct{}, 10),
	}
}

func (o *OrderHandler) CreateOrder(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	userID, err := middleware.GetUserId(c)
	if err != nil {
		o.logger.Error("getting user_id failed", slog.String("status", fmt.Sprintf("%d", http.StatusInternalServerError)), slogger.Err(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	var req dto.CreateOrderRequest
	if err := c.BindJSON(&req); err != nil {
		o.logger.Error("Failed to bind JSON", slog.String("status", fmt.Sprintf("%d", http.StatusBadRequest)), slogger.Err(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	g, gctx := errgroup.WithContext(ctx)

	var (
		stockResult *warehousepb.CheckStockResponse
		mu          sync.RWMutex
	)

	// Проверка наличия товаров на складе
	g.Go(func() error {
		o.stockCheckWorkers <- struct{}{}
		defer func() { <-o.stockCheckWorkers }()

		stockReq := &warehousepb.CheckStockRequest{
			Items: make([]*warehousepb.StockItem, len(req.Items)),
		}
		for i, item := range req.Items {
			stockReq.Items[i] = &warehousepb.StockItem{
				ProductName: item.ProductName,
				Quantity:    int32(item.Quantity),
			}
		}

		result, err := o.warehouseGRPCClient.CheckStockAvailability(gctx, stockReq)
		if err != nil {
			return fmt.Errorf("stock check failed: %w", err)
		}

		mu.Lock()
		stockResult = result
		mu.Unlock()

		return nil
	})

	// Ждем завершения всех параллельных операций
	if err := g.Wait(); err != nil {
		o.logger.Error("Parallel operations failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Проверяем результаты (с защитой мьютексом)
	mu.RLock()
	if stockResult.Available {
		mu.RUnlock()
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Stock not available",
			"message": "Some items are out of stock",
		})
		return
	}
	mu.RUnlock()

	// Создаем заказ
	orderReq := &orderpb.CreateOrderRequest{
		UserId:          int64(userID),
		Items:           convertToOrderItems(req.Items),
		DeliveryAddress: req.DeliveryAddress,
	}

	orderResp, err := o.orderGRPCClient.CreateOrder(ctx, orderReq)
	if err != nil {
		o.logger.Error("Failed to create order", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create order",
			"message": err.Error(),
		})
		return
	}

	// Возвращаем ответ
	c.JSON(http.StatusCreated, dto.CreateOrderResponse{
		Order: &entity.Order{
			ID:              orderResp.Order.Id,
			UserID:          orderReq.UserId,
			Status:          entity.OrderStatus(orderResp.Order.Status),
			Items:           utils.ConvertOrderItemToGoodsItem(orderResp.Order.Items),
			TotalAmount:     orderResp.Order.TotalAmount,
			DeliveryAddress: orderReq.DeliveryAddress,
			CreatedAt:       orderResp.Order.CreatedAt.AsTime().Unix(),
		},
		Message: "Order created successfully",
	})
}

func convertToOrderItems(createItems []dto.CreateOrderItem) []*orderpb.OrderItem {
	orderItems := make([]*orderpb.OrderItem, len(createItems))

	for i, item := range createItems {
		orderItems[i] = &orderpb.OrderItem{
			// ProductId:   item.ProductID,
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
		}
	}

	return orderItems
}

func (o *OrderHandler) GetOrders(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	userID, err := middleware.GetUserId(c)
	if err != nil {
		o.logger.Error("getting user_id failed", slog.String("status", fmt.Sprintf("%d", http.StatusInternalServerError)), slogger.Err(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	ordersReq := &orderpb.GetOrdersByUserRequest{
		UserId: int64(userID),
	}
	orders, err := o.orderGRPCClient.GetOrdersByUser(ctx, ordersReq)
	if err != nil {
		o.logger.Error("Failed to get orders", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get orders",
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, orders)
}
func (o *OrderHandler) GetOrderByID(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	userID, err := middleware.GetUserId(c)
	if err != nil {
		o.logger.Error("getting user_id failed", slog.String("status", fmt.Sprintf("%d", http.StatusInternalServerError)), slogger.Err(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	orderID, err := strconv.Atoi(c.Param("order_id"))
	if err != nil {
		o.logger.Error("Invalid order_id", slog.String("status", fmt.Sprintf("%d", http.StatusBadRequest)), slogger.Err(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid order_id",
		})
		return
	}

	orderReq := &orderpb.GetOrderDetailsRequest{
		UserId:  int64(userID),
		OrderId: int64(orderID),
	}
	order, err := o.orderGRPCClient.GetOrderDetails(ctx, orderReq)
	if err != nil {
		o.logger.Error("Failed to get order details", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get order details",
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, order)
}
func (o *OrderHandler) AssignDriver(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	userID, err := middleware.GetUserId(c)
	if err != nil {
		o.logger.Error("getting user_id failed", slog.String("status", fmt.Sprintf("%d", http.StatusInternalServerError)), slogger.Err(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	orderID, err := strconv.Atoi(c.Param("order_id"))
	if err != nil {
		o.logger.Error("Invalid order_id", slog.String("status", fmt.Sprintf("%d", http.StatusBadRequest)), slogger.Err(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid order_id",
		})
		return
	}
	_, err = o.driverGRPCClient.FindSuitableDriver(ctx, &driverpb.FindDriverRequest{
		OrderId: int64(orderID),
	})
	if err != nil {
		o.logger.Error("Failed to find suitable driver", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to find suitable driver",
			"message": err.Error(),
		})
		return
	}
	assignReq := &orderpb.AssignDriverRequest{
		UserId:  int64(userID),
		OrderId: int64(orderID),
	}
	assignResp, err := o.orderGRPCClient.AssignDriver(ctx, assignReq)
	if err != nil {
		o.logger.Error("Failed to assign driver", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to assign driver",
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, assignResp)

}

func (o *OrderHandler) GetDeliveries(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	userID, err := middleware.GetUserId(c)
	if err != nil {
		o.logger.Error("getting user_id failed", slog.String("status", fmt.Sprintf("%d", http.StatusInternalServerError)), slogger.Err(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	deliveriesReq := &orderpb.GetDeliveriesByUserRequest{
		UserId: int64(userID),
	}
	deliveries, err := o.orderGRPCClient.GetDeliveries(ctx, deliveriesReq)
	if err != nil {
		o.logger.Error("Failed to get deliveries", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get deliveries",
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, deliveries)
}

func (o *OrderHandler) CompleteOrder(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	userID, err := middleware.GetUserId(c)
	if err != nil {
		o.logger.Error("getting user_id failed", slog.String("status", fmt.Sprintf("%d", http.StatusInternalServerError)), slogger.Err(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	orderID, err := strconv.Atoi(c.Param("order_id"))
	if err != nil {
		o.logger.Error("Invalid order_id", slog.String("status", fmt.Sprintf("%d", http.StatusBadRequest)), slogger.Err(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid order_id",
		})
		return
	}
	completeReq := &orderpb.CompleteDeliveryRequest{
		UserId:  int64(userID),
		OrderId: int64(orderID),
	}
	completeResp, err := o.orderGRPCClient.CompleteDelivery(ctx, completeReq)
	if err != nil {
		o.logger.Error("Failed to complete order", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to complete order",
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, completeResp)
}
