package domain

import (
	"context"
	"logistics/internal/shared/entity"
)

type OrderRepositoryInterface interface {
	// Define methods for order repository
	CreateOrder(ctx context.Context, order *entity.Order) (int64, error)
	CompleteDelivery(ctx context.Context, userID, orderID int64) error
	GetDeliveriesByUser(ctx context.Context, userID int64) ([]*entity.Order, error)
	GetOrderDetails(ctx context.Context, userPD int64, orderID int64) (*entity.Order, error)
	GetOrdersByUser(ctx context.Context, userID int64) ([]*entity.Order, error)
	UpdateOrderStatus(ctx context.Context, userID, orderID int64, status string) error
}
