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
	"net/http"
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

	// Используем errgroup для параллельных операций
	g, gctx := errgroup.WithContext(ctx)

	// Результаты параллельных операций
	var (
		stockResult *warehousepb.CheckStockResponse
		mu          sync.RWMutex // защищаем результаты
	)

	// Горутина 1: Проверка наличия товаров на складе
	g.Go(func() error {
		o.stockCheckWorkers <- struct{}{} // семафор для ограничения горутин
		defer func() { <-o.stockCheckWorkers }()

		stockReq := &warehousepb.CheckStockRequest{
			Items: make([]*warehousepb.StockItem, len(req.Items)),
		}
		for i, item := range req.Items {
			stockReq.Items[i] = &warehousepb.StockItem{
				ProductId: item.ProductID,
				Quantity:  int32(item.Quantity),
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
	if !stockResult.Available {
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
			TotalAmount:     orderResp.Order.TotalAmount,
			DeliveryAddress: orderReq.DeliveryAddress,
			CreatedAt:       orderResp.Order.CreatedAt.AsTime(),
		},
		Message: "Order created successfully",
	})
}

func convertToOrderItems(createItems []dto.CreateOrderItem) []*orderpb.OrderItem {
	orderItems := make([]*orderpb.OrderItem, len(createItems))

	for i, item := range createItems {
		orderItems[i] = &orderpb.OrderItem{
			ProductId:   item.ProductID,
			ProductName: item.ProductName,
			Price:       item.Price,
		}
	}

	return orderItems
}

func (o *OrderHandler) GetOrders(c *gin.Context) {
	// Implementation for getting orders
}
func (o *OrderHandler) GetOrderByID(c *gin.Context) {
	// Implementation for getting an order by ID
}
func (o *OrderHandler) AssignDriver(c *gin.Context) {
	// Implementation for assigning a driver to an order
}

func (o *OrderHandler) GetDeliveries(c *gin.Context) {
	// Implementation for getting deliveries
}
