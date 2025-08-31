package driverservice

import (
	"context"
	"log/slog"
	driverpb "logistics/api/protobuf/driver_service"
	"logistics/internal/services/driver-service/domain"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type DriverGRPCService struct {
	driverpb.UnimplementedDriverServiceServer
	driverRepo  domain.DriverRepositoryInterface
	logger      *slog.Logger
	redisClient *redis.Client
}

func NewDriverGRPCService(logger *slog.Logger, driverRepo domain.DriverRepositoryInterface, redisClient *redis.Client) *DriverGRPCService {
	return &DriverGRPCService{
		driverRepo:  driverRepo,
		logger:      logger,
		redisClient: redisClient,
	}
}

func RegisterDriverServiceServer(s *grpc.Server, srv *DriverGRPCService) {
	driverpb.RegisterDriverServiceServer(s, srv)
}

func (d *DriverGRPCService) FindSuitableDriver(context.Context, *driverpb.FindDriverRequest) (*driverpb.FindDriverResponse, error) {

}

func (d *DriverGRPCService) GetAvailableDrivers(context.Context, *emptypb.Empty) (*driverpb.GetAvailableDriversResponse, error) {

}

func (d *DriverGRPCService) UpdateDriverStatus(context.Context, *driverpb.UpdateDriverStatusRequest) (*driverpb.UpdateDriverStatusResponse, error) {

}
