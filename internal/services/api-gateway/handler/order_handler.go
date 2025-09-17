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

// @Summary Создание нового заказа
// @Description Создает новый заказ после проверки наличия товаров на складе
// @Tags orders
// @Accept  json
// @Produce  json
// @Param   request body dto.CreateOrderRequest true "Данные для создания заказа"
// @Success 201 {object} dto.CreateOrderResponse
// @Failure 400 {object} object{error=string,message=string} "Некорректные данные или товара нет в наличии"
// @Failure 500 {object} object{error=string,message=string} "Ошибка сервера"
// @Security ApiKeyAuth
// @Router /orders [post]
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
		available bool
		mu        sync.RWMutex
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
		o.logger.Info("Stock check completed", slog.String("status", "success"), slog.Bool("available", result.Available))

		mu.Lock()
		available = result.Available
		mu.Unlock()

		return nil
	})

	// Ждем завершения всех параллельных операций
	if err := g.Wait(); err != nil {
		o.logger.Error("Parallel operations failed", "error", slogger.Err(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Проверяем результаты (с защитой мьютексом)
	mu.RLock()
	if !available {
		mu.RUnlock()
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Stock not available",
			"message": "Some items are out of stock",
		})
		return
	}
	mu.RUnlock()
	o.logger.Info("All parallel operations completed successfully", slog.String("status", "success"), slog.Bool("stock_available", available))

	// Создаем заказ
	orderItems := make([]*orderpb.OrderItem, len(req.Items))
	for i, item := range req.Items {
		info, err := o.orderGRPCClient.GetOrderItemInfo(ctx, &orderpb.GetOrderItemInfoRequest{
			ProductName: item.ProductName,
		})
		if err != nil {
			o.logger.Error("Failed to get item price", "error", slogger.Err(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to get item price",
				"message": err.Error(),
			})
			return
		}

		totalPrice := info.Price * float64(item.Quantity)

		orderItems[i] = &orderpb.OrderItem{
			ProductId:   info.ProductId,
			ProductName: item.ProductName,
			Quantity:    int32(item.Quantity),
			Price:       info.Price,
			TotalPrice:  totalPrice,
		}
	}
	orderReq := &orderpb.CreateOrderRequest{
		UserId:          int64(userID),
		Items:           orderItems, // Используем уже заполненный слайс
		DeliveryAddress: req.DeliveryAddress,
		Time:            time.Now().Unix(),
	}

	orderResp, err := o.orderGRPCClient.CreateOrder(ctx, orderReq)
	if err != nil {
		o.logger.Error("Failed to create order", "error", slogger.Err(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create order",
			"message": err.Error(),
		})
		return
	}
	_, err = o.warehouseGRPCClient.UpdateStock(ctx, &warehousepb.UpdateStockRequest{
		Items: utils.ConvertOrderItemToWarehouseStockItem(orderItems, orderReq.Time),
	})
	if err != nil {
		o.logger.Error("Failed to update stock after order creation", "error", slogger.Err(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update stock after order creation",
			"message": "Stock update failed",
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

// @Summary Получение списка заказов пользователя
// @Description Возвращает все заказы текущего авторизованного пользователя
// @Tags orders
// @Produce  json
// @Success 200 {object} object{orders=[]entity.Order} "Успешный ответ"
// @Failure 500 {object} object{error=string,message=string} "Ошибка сервера"
// @Security ApiKeyAuth
// @Router /orders [get]
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
		o.logger.Error("Failed to get orders", "error", slogger.Err(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get orders",
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, orders)
}

// @Summary Получение деталей заказа
// @Description Возвращает детальную информацию о конкретном заказе пользователя
// @Tags orders
// @Produce  json
// @Param   order_id path int true "ID заказа"
// @Success 200 {object} entity.Order "Детали заказа"
// @Failure 400 {object} object{error=string} "Неверный ID заказа"
// @Failure 500 {object} object{error=string,message=string} "Ошибка сервера"
// @Security ApiKeyAuth
// @Router /orders/{order_id} [get]
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
		o.logger.Error("Failed to get order details", "error", slogger.Err(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get order details",
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, order)
}

// @Summary Назначение водителя на заказ
// @Description Назначает подходящего водителя на заказ и обновляет его статус
// @Tags orders
// @Produce  json
// @Param   order_id path int true "ID заказа"
// @Success 200 {object} object{driver_id=int64,order_id=int64,success=bool,message=string} "Успешное назначение"
// @Failure 400 {object} object{error=string,message=string} "Заказ не в pending статусе"
// @Failure 500 {object} object{error=string,message=string} "Ошибка сервера"
// @Security ApiKeyAuth
// @Router /orders/{order_id}/assign [post]
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
	orderStatus, err := o.orderGRPCClient.CheckOrderStatus(ctx, &orderpb.CheckOrderStatusRequest{
		UserId:  int64(userID),
		OrderId: int64(orderID),
	})
	if err != nil {
		o.logger.Error("Failed to check order status", "error", slogger.Err(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to check order status",
			"message": err.Error(),
		})
		return
	}
	if orderStatus.Status != string(entity.StatusPending) {
		o.logger.Error("Order is not in pending status, other driver assignment is not possible", slog.String("status", orderStatus.Status))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Order is not in pending status, other driver assignment is not possible",
			"message": fmt.Sprintf("Current order status: %s", orderStatus.Status),
		})
		return
	}

	_, err = o.driverGRPCClient.FindSuitableDriver(ctx, &driverpb.FindDriverRequest{
		OrderId: int64(orderID),
	})
	if err != nil {
		o.logger.Error("Failed to find suitable driver", "error", slogger.Err(err))
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
		o.logger.Error("Failed to assign driver", "error", slogger.Err(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to assign driver",
			"message": err.Error(),
		})
		return
	}
	resp, err := o.driverGRPCClient.UpdateDriverStatus(ctx, &driverpb.UpdateDriverStatusRequest{
		DriverId: assignResp.DriverId,
		Status:   string(entity.DriverStatusBusy),
	})

	if !resp.Success {
		o.logger.Error("Failed to update driver status", slog.String("status", "error"), slogger.Err(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update driver status",
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"driver_id": assignResp.DriverId,
		"order_id":  assignResp.OrderId,
		"success":   true,
		"message":   assignResp.Message,
	})

}

// @Summary Получение списка доставок
// @Description Возвращает все доставки текущего авторизованного пользователя
// @Tags deliveries
// @Produce  json
// @Success 200 {object} object{deliveries=[]entity.Order} "Список доставок"
// @Success 200 {object} object{message=string} "Если доставок нет"
// @Failure 500 {object} object{error=string,message=string} "Ошибка сервера"
// @Security ApiKeyAuth
// @Router /deliveries [get]
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
		o.logger.Error("Failed to get deliveries", "error", slogger.Err(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get deliveries",
			"message": err.Error(),
		})
		return
	}
	if len(deliveries.Deliveries) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"message": "No deliveries found",
		})
	}
	c.JSON(http.StatusOK, deliveries)
}

// @Summary Завершение доставки
// @Description Отмечает доставку как завершенную и обновляет статус водителя
// @Tags deliveries
// @Produce  json
// @Param   order_id path int true "ID заказа"
// @Success 200 {object} object{success=bool,driver_id=int64,message=string} "Успешное завершение"
// @Failure 400 {object} object{error=string} "Неверный ID заказа"
// @Failure 500 {object} object{error=string,message=string} "Ошибка сервера"
// @Security ApiKeyAuth
// @Router /deliveries/{order_id}/complete [post]
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
		o.logger.Error("Failed to complete order", "error", slogger.Err(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to complete order",
			"message": err.Error(),
		})
		return
	}
	_, err = o.driverGRPCClient.UpdateDriverStatus(ctx, &driverpb.UpdateDriverStatusRequest{
		DriverId: completeResp.DriverId,
		Status:   string(entity.DriverStatusAvailable),
	})
	if err != nil {
		o.logger.Error("Failed to update driver status to available", "error", slogger.Err(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update driver status to available",
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success":   completeResp.Success,
		"driver_id": completeResp.DriverId,
		"message":   completeResp.Message,
	})
}
