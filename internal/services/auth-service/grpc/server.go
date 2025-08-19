package auth_grpc_server

import (
	"context"
	"fmt"
	"log/slog"
	authpb "logistics/api/protobuf/auth_service"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
)

type AuthGRPCServer struct {
	authpb.UnimplementedAuthServiceServer
	log    *slog.Logger
	dbpool *pgxpool.Pool
}

func NewAuthGRPCServer(log *slog.Logger, dbpool *pgxpool.Pool) *AuthGRPCServer {
	return &AuthGRPCServer{
		log:    log,
		dbpool: dbpool,
	}
}

func RegisterAuthServiceServer(s *grpc.Server, srv *AuthGRPCServer) {
	authpb.RegisterAuthServiceServer(s, srv)
}
func (s *AuthGRPCServer) SignUp(ctx context.Context, req *authpb.SignUpRequest) (*authpb.SignUpResponse, error) {
	// Реализация логики регистрации пользователя
	fmt.Println("SignUp called with:", req)
	return &authpb.SignUpResponse{
		UserId:    1,
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}, nil
}
