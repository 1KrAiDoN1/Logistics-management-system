package handler

import (
	"log/slog"
	authpb "logistics/api/protobuf/auth_service"
)

type Handlers struct {
	AuthHandlerInterface
}

func NewHandlers(logger *slog.Logger, authGRPCClient authpb.AuthServiceClient) *Handlers {
	return &Handlers{
		AuthHandlerInterface: NewAuthHandler(logger, authGRPCClient),
	}
}
