package handler

import (
	"log/slog"
	authpb "logistics/api/protobuf/auth_service"
	orderpb "logistics/api/protobuf/order_service"
)

type Handlers struct {
	AuthHandlerInterface
	OrderHandlerInterface
}

func NewHandlers(logger *slog.Logger, authGRPCClient authpb.AuthServiceClient, orderGRPCClient orderpb.OrderServiceClient) *Handlers {
	return &Handlers{
		AuthHandlerInterface:  NewAuthHandler(logger, authGRPCClient),
		OrderHandlerInterface: NewOrderHandler(logger, orderGRPCClient),
	}
}
