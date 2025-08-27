package warehouseservice

import (
	"log/slog"

	warehousepb "logistics/api/protobuf/warehouse_service"
	"logistics/internal/services/warehouse-service/domain"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
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
