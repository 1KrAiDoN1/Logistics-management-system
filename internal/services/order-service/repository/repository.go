package repository

import (
	"context"
	"logistics/internal/shared/entity"

	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderRepository struct {
	pool *pgxpool.Pool
}

func NewOrderRepository(pool *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{
		pool: pool,
	}
}

func (o *OrderRepository) CreateOrder(ctx context.Context, order *entity.Order) (int64, error) {

}

func (o *OrderRepository) CompleteDelivery(ctx context.Context, userID, orderID int64) error {

}

func (o *OrderRepository) GetDeliveriesByUser(ctx context.Context, userID int64) ([]*entity.Order, error) {

}

func (o *OrderRepository) GetOrderDetails(ctx context.Context, orderID int64) (*entity.Order, error) {

}
