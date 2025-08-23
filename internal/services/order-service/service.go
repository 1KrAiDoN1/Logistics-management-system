package orderservice

import (
	"log/slog"
	orderpb "logistics/api/protobuf/order_service"
	"logistics/internal/services/order-service/domain"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
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

// CreateOrder(context.Context, *CreateOrderRequest) (*CreateOrderResponse, error)
// UpdateOrderStatus(context.Context, *UpdateOrderStatusRequest) (*UpdateOrderStatusResponse, error)
// AssignDriver(context.Context, *AssignDriverRequest) (*AssignDriverResponse, error)
// GetOrderDetails(context.Context, *GetOrderDetailsRequest) (*GetOrderDetailsResponse, error)
// GetOrdersByUser(context.Context, *GetOrdersByUserRequest) (*GetOrdersByUserResponse, error)
// CompleteDelivery(context.Context, *CompleteDeliveryRequest) (*CompleteDeliveryResponse, error)
