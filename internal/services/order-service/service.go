package orderservice

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	orderpb "logistics/api/protobuf/order_service"
	kfk "logistics/internal/kafka"
	"logistics/internal/services/order-service/domain"
	"logistics/internal/shared/entity"
	"logistics/pkg/lib/logger/slogger"
	"logistics/pkg/lib/utils"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type OrderGRPCService struct {
	orderpb.UnimplementedOrderServiceServer
	orderRepo     domain.OrderRepositoryInterface
	logger        *slog.Logger
	redisClient   *redis.Client
	kafkaConsumer *kfk.KafkaConsumer
}

func NewOrderGRPCService(logger *slog.Logger, orderRepo domain.OrderRepositoryInterface, kafkaConsumer *kfk.KafkaConsumer, redisClient *redis.Client) *OrderGRPCService {
	return &OrderGRPCService{
		orderRepo:     orderRepo,
		logger:        logger,
		redisClient:   redisClient,
		kafkaConsumer: kafkaConsumer,
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

	orderJSON, err := json.Marshal(*order)
	if err != nil {
		o.logger.Error("failed to marshal order", slogger.Err(err))
	} else {
		// Сохраняем
		err = o.redisClient.Set(ctx, fmt.Sprintf("user:%d_order:%d", req.UserId, orderID), orderJSON, 15*time.Minute).Err()
		if err != nil {
			o.logger.Error("failed to cache order in redis", slogger.Err(err))
		}
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
	out := make(chan *kafka.Message, 1)
	errCh := make(chan error, 1)

	go func() {
		kafkaMessage, err := o.kafkaConsumer.ReadMessage(ctx)
		if err != nil {
			errCh <- err
			return
		}
		out <- kafkaMessage
	}()

	select {

	case kafkamessage := <-out:

		var message entity.Msg
		err := json.Unmarshal(kafkamessage.Value, &message)
		if err != nil {
			o.logger.Error("❌ Failed to unmarshal message", slog.String("error", err.Error()))
			return &orderpb.AssignDriverResponse{}, err
		}

		stats, err := o.UpdateOrderStatus(ctx, &orderpb.UpdateOrderStatusRequest{
			UserId:   req.UserId,
			OrderId:  req.OrderId,
			DriverId: message.Driver.ID,
			Status:   string(entity.StatusInProgress),
		})
		if err != nil {
			o.logger.Error("❌ Failed to update order status", slog.String("status", "error"), slog.String("error", err.Error()))
			return &orderpb.AssignDriverResponse{}, err
		}
		if !stats.Success {
			o.logger.Error("❌ Not success, failed to update order status")
			return &orderpb.AssignDriverResponse{}, err
		}

		if err := o.kafkaConsumer.CommitMessage(ctx, kafkamessage); err != nil {
			o.logger.Error("Failed to commit message", slog.String("error", err.Error()))
			return &orderpb.AssignDriverResponse{}, err
		}
		return &orderpb.AssignDriverResponse{
			DriverId: message.Driver.ID,
			OrderId:  message.OrderId,
			Success:  true,
			Message:  stats.Message,
		}, nil

	case err := <-errCh:
		o.logger.Error("❌ Failed to consume message", slog.String("error", err.Error()))
		return nil, err

	case <-ctx.Done():
		o.logger.Error("❌ Context cancelled", slog.String("error", ctx.Err().Error()))
		return nil, ctx.Err()
	}

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
	orderJSON, err := o.redisClient.Get(ctx, fmt.Sprintf("user:%d_order:%d", req.UserId, req.OrderId)).Result()
	if err != nil && err != redis.Nil {
		o.logger.Error("failed to get order from redis", slogger.Err(err))
	} else if err == nil {
		var order entity.Order
		err = json.Unmarshal([]byte(orderJSON), &order)
		if err != nil {
			o.logger.Error("failed to unmarshal order from redis", slogger.Err(err))
		} else {
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
			return &orderpb.GetOrderDetailsResponse{
				Order: &orderpb.Order{
					Id:              order.ID,
					UserId:          order.UserID,
					Status:          string(order.Status),
					DeliveryAddress: order.DeliveryAddress,
					Items:           items,
					TotalAmount:     order.TotalAmount,
					DriverId:        driverID,
					CreatedAt:       timestamppb.New(order.CreatedAt),
				},
			}, nil
		}
	}

	order, err := o.orderRepo.GetOrderDetails(ctx, req.UserId, req.OrderId)
	if err != nil {
		o.logger.Error("failed to get order details", slog.String("status", "error"), slogger.Err(err))
		return nil, err
	}
	if order == nil {
		return &orderpb.GetOrderDetailsResponse{}, nil
	}
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
	return &orderpb.GetOrderDetailsResponse{
		Order: &orderpb.Order{
			Id:              order.ID,
			UserId:          order.UserID,
			Status:          string(order.Status),
			DeliveryAddress: order.DeliveryAddress,
			Items:           items,
			TotalAmount:     order.TotalAmount,
			DriverId:        driverID,
			CreatedAt:       timestamppb.New(order.CreatedAt),
		},
	}, nil
}

func (o *OrderGRPCService) GetOrdersByUser(ctx context.Context, req *orderpb.GetOrdersByUserRequest) (*orderpb.GetOrdersByUserResponse, error) {
	res, err := o.orderRepo.GetOrdersByUser(ctx, req.UserId)
	if err != nil {
		o.logger.Error("failed to get orders by user", slog.String("status", "error"), slogger.Err(err))
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
	return &orderpb.GetOrdersByUserResponse{
		Orders: orders,
	}, nil
}

func (o *OrderGRPCService) UpdateOrderStatus(ctx context.Context, req *orderpb.UpdateOrderStatusRequest) (*orderpb.UpdateOrderStatusResponse, error) {
	err := o.orderRepo.UpdateOrderStatus(ctx, req.UserId, req.OrderId, req.DriverId, req.Status)
	if err != nil {
		o.logger.Error("failed to update order status", slog.String("status", "error"), slogger.Err(err))
		return &orderpb.UpdateOrderStatusResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to update order status to %s for order ID: %d", req.Status, req.OrderId),
		}, nil
	}
	orderJSON, err := o.redisClient.Get(ctx, fmt.Sprintf("user:%d_order:%d", req.UserId, req.OrderId)).Result()
	if err != nil && err != redis.Nil {
		o.logger.Error("failed to get order from redis", slogger.Err(err))
	} else if err == nil {
		var order entity.Order
		err = json.Unmarshal([]byte(orderJSON), &order)
		if err != nil {
			o.logger.Error("failed to unmarshal order from redis", slogger.Err(err))
		} else {
			order.Status = entity.OrderStatus(req.Status)
			if req.DriverId != 0 {
				order.DriverID = &req.DriverId
			}
			updatedOrderJSON, err := json.Marshal(order)
			if err != nil {
				o.logger.Error("failed to marshal updated order", slogger.Err(err))
			} else {
				err = o.redisClient.Set(ctx, fmt.Sprintf("user:%d_order:%d", req.UserId, req.OrderId), updatedOrderJSON, 15*time.Minute).Err()
				if err != nil {
					o.logger.Error("failed to update cached order in redis", slogger.Err(err))
				}
			}
		}
	}
	return &orderpb.UpdateOrderStatusResponse{
		Success: true,
		Message: fmt.Sprintf("Order status updated successfully to %s for order ID: %d", req.Status, req.OrderId),
	}, nil
}
