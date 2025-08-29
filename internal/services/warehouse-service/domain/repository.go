package domain

import (
	"context"
	"logistics/internal/shared/entity"
)

type WarehouseRepositoryInterface interface {
	CheckStockAvailability(ctx context.Context, orders []*entity.GoodsItem) (bool, error)
	GetWarehouseStock(ctx context.Context) ([]*entity.GoodsItem, error)
	UpdateStock(ctx context.Context, items []*entity.GoodsItem) error
}
