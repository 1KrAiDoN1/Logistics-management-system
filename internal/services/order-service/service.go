package orderservice

import (
	"context"
	"fmt"
	"log/slog"
	orderpb "logistics/api/protobuf/order_service"
	"logistics/internal/services/order-service/domain"
	"logistics/internal/shared/entity"
	"logistics/pkg/lib/logger/slogger"
	"logistics/pkg/lib/utils"
	"time"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type OrderGRPCService struct {
	orderpb.UnimplementedOrderServiceServer
	orderRepo   domain.OrderRepositoryInterface
	logger      *slog.Logger
	redisClient *redis.Client
}

func NewOrderGRPCService(logger *slog.Logger, orderRepo domain.OrderRepositoryInterface, redisClient *redis.Client) *OrderGRPCService {
	return &OrderGRPCService{
		orderRepo:   orderRepo,
		logger:      logger,
		redisClient: redisClient,
	}
}

func RegisterOrderServiceServer(s *grpc.Server, srv *OrderGRPCService) {
	orderpb.RegisterOrderServiceServer(s, srv)
}

func (o *OrderGRPCService) CreateOrder(ctx context.Context, req *orderpb.CreateOrderRequest) (*orderpb.CreateOrderResponse, error) {
	order := &entity.Order{
		UserID:          req.UserId,
		Status:          entity.StatusPending,
		Items:           utils.ConvertOrderItemToGoodsItem(req.Items),
		DeliveryAddress: req.DeliveryAddress,
		CreatedAt:       time.Now(),
	}
	order.TotalAmount = 0
	for _, item := range order.Items {
		order.TotalAmount += item.Price * float64(item.Quantity)
	}

	orderID, err := o.orderRepo.CreateOrder(ctx, order)
	if err != nil {
		o.logger.Error("failed to create order", slog.String("status", "error"), slogger.Err(err))
		return nil, err
	}

	return &orderpb.CreateOrderResponse{
		Order: &orderpb.Order{
			Id:          orderID,
			UserId:      order.UserID,
			Items:       req.Items,
			TotalAmount: order.TotalAmount,
			CreatedAt:   timestamppb.New(order.CreatedAt),
			Status:      string(order.Status),
		},
	}, nil

}

func (o *OrderGRPCService) AssignDriver(ctx context.Context, req *orderpb.AssignDriverRequest) (*orderpb.AssignDriverResponse, error) {

}

func (o *OrderGRPCService) CompleteDelivery(ctx context.Context, req *orderpb.CompleteDeliveryRequest) (*orderpb.CompleteDeliveryResponse, error) {
	err := o.orderRepo.CompleteDelivery(ctx, req.UserId, req.OrderId)
	if err != nil {
		o.logger.Error("failed to complete delivery", slog.String("status", "error"), slogger.Err(err))
		return nil, err
	}
	return &orderpb.CompleteDeliveryResponse{
		Success: true,
		Message: fmt.Sprintf("Delivery completed successfully, order ID: %d", req.OrderId),
	}, nil

}

func (o *OrderGRPCService) GetDeliveries(ctx context.Context, req *orderpb.GetDeliveriesByUserRequest) (*orderpb.GetDeliveriesByUserResponse, error) {
	res, err := o.orderRepo.GetDeliveriesByUser(ctx, req.UserId)
	if err != nil {
		o.logger.Error("failed to get deliveries by user", slog.String("status", "error"), slogger.Err(err))
		return nil, err
	}
	orders := make([]*orderpb.Order, 0, len(res))
	for _, order := range res {
		items := make([]*orderpb.OrderItem, 0, len(order.Items))
		for _, item := range order.Items {
			items = append(items, &orderpb.OrderItem{
				ProductId:   item.ProductID,
				ProductName: item.ProductName,
				Price:       item.Price,
				Quantity:    item.Quantity,
			})
		}
		var driverID int64
		if order.DriverID != nil {
			driverID = *order.DriverID
		}
		orders = append(orders, &orderpb.Order{
			Id:              order.ID,
			UserId:          order.UserID,
			Status:          string(order.Status),
			DeliveryAddress: order.DeliveryAddress,
			Items:           items,
			TotalAmount:     order.TotalAmount,
			DriverId:        driverID,
			CreatedAt:       timestamppb.New(order.CreatedAt),
		})
	}
	return &orderpb.GetDeliveriesByUserResponse{
		Deliveries: orders,
	}, nil
}

func (o *OrderGRPCService) GetOrderDetails(ctx context.Context, req *orderpb.GetOrderDetailsRequest) (*orderpb.GetOrderDetailsResponse, error) {

}

func (o *OrderGRPCService) GetOrdersByUser(ctx context.Context, req *orderpb.GetOrdersByUserRequest) (*orderpb.GetOrdersByUserResponse, error) {

}

func (o *OrderGRPCService) UpdateOrderStatus(ctx context.Context, req *orderpb.UpdateOrderStatusRequest) (*orderpb.UpdateOrderStatusResponse, error) {

}
