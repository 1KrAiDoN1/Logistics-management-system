package orderservice

import (
	"log/slog"
	"logistics/internal/services/order-service/domain"
)

type OrderService struct {
	orderRepo domain.OrderRepositoryInterface
	logger    *slog.Logger
}

func NewOrderService(orderRepo domain.OrderRepositoryInterface, logger *slog.Logger) *OrderService {
	return &OrderService{
		orderRepo: orderRepo,
		logger:    logger,
	}
}
