package auth_grpc_service

import (
	"context"
	"fmt"
	"log/slog"
	authpb "logistics/api/protobuf/auth_service"
	"logistics/internal/services/auth-service/domain"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type AuthGRPCService struct {
	authpb.UnimplementedAuthServiceServer
	log            *slog.Logger
	authrepository domain.AuthRepositoryInterface
	redisClient    *redis.Client
}

func NewAuthGRPCService(log *slog.Logger, repository domain.AuthRepositoryInterface, redisClient *redis.Client) *AuthGRPCService {
	return &AuthGRPCService{
		log:            log,
		authrepository: repository,
		redisClient:    redisClient,
	}
}

func RegisterAuthServiceServer(s *grpc.Server, srv *AuthGRPCService) {
	authpb.RegisterAuthServiceServer(s, srv)
}
func (s *AuthGRPCService) SignUp(ctx context.Context, req *authpb.SignUpRequest) (*authpb.SignUpResponse, error) {
	// Реализация логики регистрации пользователя
	fmt.Println("SignUp called with:", req)
	return &authpb.SignUpResponse{
		UserId:    1,
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}, nil
}

func (s *AuthGRPCService) SignIn(ctx context.Context, req *authpb.SignInRequest) (*authpb.SignInResponse, error) {
	// Реализация логики входа пользователя
	fmt.Println("SignIn called with:", req)
	return &authpb.SignInResponse{
		UserId:      1,
		Email:       req.Email,
		FirstName:   "John",
		LastName:    "Doe",
		AccessToken: "example_token",
	}, nil
}

func (s *AuthGRPCService) Logout(ctx context.Context, req *authpb.LogoutRequest) (*emptypb.Empty, error) {
	// Реализация логики выхода пользователя
	fmt.Println("Logout called with:", req)
	return &emptypb.Empty{}, nil
}

func (s *AuthGRPCService) ValidateToken(ctx context.Context, req *authpb.ValidateTokenRequest) (*authpb.ValidateTokenResponse, error) {
	// Реализация логики валидации токена
	fmt.Println("ValidateToken called with:", req)
	if req.AccessToken == "valid_token" {
		return &authpb.ValidateTokenResponse{UserId: 1}, nil
	}
	return nil, fmt.Errorf("invalid token")
}
func (s *AuthGRPCService) GetUserIDbyRefreshToken(ctx context.Context, req *authpb.GetUserIDbyRefreshTokenRequest) (*authpb.GetUserIDbyRefreshTokenResponse, error) {
	// Реализация логики получения UserID по refresh token
	fmt.Println("GetUserIDbyRefreshToken called with:", req)
	if req.RefreshToken == "valid_refresh_token" {
		return &authpb.GetUserIDbyRefreshTokenResponse{UserId: 1}, nil
	}
	return nil, fmt.Errorf("invalid refresh token")
}
func (s *AuthGRPCService) GenerateAccessToken(ctx context.Context, req *authpb.GenerateAccessTokenRequest) (*authpb.GenerateAccessTokenResponse, error) {
	//
	// Реализация логики генерации access token
	fmt.Println("GenerateAccessToken called with:", req)
	return &authpb.GenerateAccessTokenResponse{
		AccessToken: "new_access_token",
	}, nil
}

func (s *AuthGRPCService) GenerateRefreshToken(ctx context.Context, req *authpb.GenerateRefreshTokenRequest) (*authpb.GenerateRefreshTokenResponse, error) {
	return nil, nil
}

func (s *AuthGRPCService) RemoveOldRefreshToken(ctx context.Context, req *authpb.RemoveOldRefreshTokenRequest) (*emptypb.Empty, error) {
	// Реализация логики удаления старого refresh token
	fmt.Println("RemoveOldRefreshToken called with:", req)
	return &emptypb.Empty{}, nil
}

func (s *AuthGRPCService) SaveNewRefreshToken(ctx context.Context, req *authpb.SaveNewRefreshTokenRequest) (*emptypb.Empty, error) {
	// Реализация логики сохранения нового refresh token
	fmt.Println("SaveNewRefreshToken called with:", req)
	return &emptypb.Empty{}, nil
}
