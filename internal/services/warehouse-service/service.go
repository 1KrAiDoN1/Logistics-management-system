package warehouseservice

import (
	"context"
	"log/slog"

	warehousepb "logistics/api/protobuf/warehouse_service"
	"logistics/internal/services/warehouse-service/domain"
	"logistics/internal/shared/entity"
	"logistics/pkg/lib/logger/slogger"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type WarehouseGRPCService struct {
	warehousepb.UnimplementedWarehouseServiceServer
	warehouseRepo domain.WarehouseRepositoryInterface
	logger        *slog.Logger
	redisClient   *redis.Client
}

func NewWarehouseGRPCService(logger *slog.Logger, warehouseRepo domain.WarehouseRepositoryInterface, redisClient *redis.Client) *WarehouseGRPCService {
	return &WarehouseGRPCService{
		warehouseRepo: warehouseRepo,
		logger:        logger,
		redisClient:   redisClient,
	}
}

func RegisterWarehouseServiceServer(s *grpc.Server, srv *WarehouseGRPCService) {
	warehousepb.RegisterWarehouseServiceServer(s, srv)
}

func (s *WarehouseGRPCService) CheckStockAvailability(ctx context.Context, req *warehousepb.CheckStockRequest) (*warehousepb.CheckStockResponse, error) {
	goodsItems := convertStockItemsToOrderItems(req.Items)
	available, err := s.warehouseRepo.CheckStockAvailability(ctx, goodsItems)
	if err != nil {
		s.logger.Error("failed to check stock availability", slog.String("status", "error"), slogger.Err(err))
		return nil, err
	}
	return &warehousepb.CheckStockResponse{Available: available}, nil
}

func (s *WarehouseGRPCService) GetWarehouseStock(ctx context.Context, req *emptypb.Empty) (*warehousepb.GetWarehouseStockResponse, error) {
	stockItem, err := s.warehouseRepo.GetWarehouseStock(ctx)
	if err != nil {
		s.logger.Error("failed to get warehouse stock", slog.String("status", "error"), slogger.Err(err))
		return nil, err
	}
	stockItems := convertOrderItemsToStock(stockItem)
	return &warehousepb.GetWarehouseStockResponse{
		Stocks: stockItems,
	}, nil
}

func (s *WarehouseGRPCService) UpdateStock(ctx context.Context, req *warehousepb.UpdateStockRequest) (*warehousepb.UpdateStockResponse, error) {
	stockItems := convertStockItemsToOrderItems(req.Items)
	err := s.warehouseRepo.UpdateStock(ctx, stockItems)
	if err != nil {
		s.logger.Error("failed to update stock", slog.String("status", "error"), slogger.Err(err))
		return nil, err
	}
	return &warehousepb.UpdateStockResponse{
		Success: true,
	}, nil
}

func convertStockItemsToOrderItems(stockItems []*warehousepb.StockItem) []*entity.GoodsItem {
	goodsItems := make([]*entity.GoodsItem, len(stockItems))

	for i, stockItem := range stockItems {
		goodsItems[i] = &entity.GoodsItem{
			ProductName: stockItem.ProductName,
			Quantity:    stockItem.Quantity,
		}
	}

	return goodsItems
}

func convertOrderItemsToStock(goodsItems []*entity.GoodsItem) []*warehousepb.Stock {
	stockItems := make([]*warehousepb.Stock, len(goodsItems))

	for i, orderItem := range goodsItems {
		stockItems[i] = &warehousepb.Stock{
			ProductId:   orderItem.ProductID,
			ProductName: orderItem.ProductName,
			Quantity:    orderItem.Quantity,
		}
	}

	return stockItems
}
