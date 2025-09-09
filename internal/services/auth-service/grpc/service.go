package auth_grpc_service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"log/slog"
	authpb "logistics/api/protobuf/auth_service"
	"logistics/internal/services/auth-service/domain"
	"logistics/internal/shared/entity"
	"logistics/pkg/lib/utils"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	AccessTokenTTL = 15 * time.Minute
)

type AuthGRPCService struct {
	authpb.UnimplementedAuthServiceServer
	log            *slog.Logger
	authrepository domain.AuthRepositoryInterface
}

func NewAuthGRPCService(log *slog.Logger, repository domain.AuthRepositoryInterface) *AuthGRPCService {
	return &AuthGRPCService{
		log:            log,
		authrepository: repository,
	}
}

func RegisterAuthServiceServer(s *grpc.Server, srv *AuthGRPCService) {
	authpb.RegisterAuthServiceServer(s, srv)
}
func (s *AuthGRPCService) SignUp(ctx context.Context, req *authpb.SignUpRequest) (*authpb.SignUpResponse, error) {
	if req.Password != req.ConfirmPassword {
		return nil, errors.New("password and confirm password do not match")
	}
	exists, err := s.authrepository.IsUserExists(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}
	if exists {
		return nil, errors.New("user with this email already exists")
	}
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}
	user := &entity.User{
		Email:              req.Email,
		Password:           hashedPassword,
		FirstName:          req.FirstName,
		LastName:           req.LastName,
		TimeOfRegistration: req.TimeOfRegistration,
	}

	userID, err := s.authrepository.CreateUser(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Реализация логики регистрации пользователя
	return &authpb.SignUpResponse{
		UserId:    userID,
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}, nil
}

func (s *AuthGRPCService) SignIn(ctx context.Context, req *authpb.SignInRequest) (*authpb.SignInResponse, error) {
	hashPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}
	user, err := s.authrepository.CheckUserVerification(ctx, req.Email, hashPassword)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	accessToken, err := s.GenerateAccessToken(ctx, &authpb.GenerateAccessTokenRequest{
		UserId: int64(user.ID),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}
	// refreshToken, err := s.GenerateRefreshToken(ctx, &authpb.GenerateRefreshTokenRequest{
	// 	UserId: int64(user.ID),
	// })
	return &authpb.SignInResponse{
		UserId:      int64(user.ID),
		Email:       user.Email,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		AccessToken: accessToken.AccessToken,
	}, nil

	//ДОБАВИТЬ КЭШИРОВАНИЕ
}

func (s *AuthGRPCService) Logout(ctx context.Context, req *authpb.LogoutRequest) (*emptypb.Empty, error) {
	err := s.authrepository.Logout(ctx, req.UserId)
	if err != nil {
		return nil, fmt.Errorf("failed to logout user: %w", err)
	}
	return &emptypb.Empty{}, nil
}

func (s *AuthGRPCService) ValidateToken(ctx context.Context, req *authpb.ValidateTokenRequest) (*authpb.ValidateTokenResponse, error) {
	err := godotenv.Load(".env")
	if err != nil {
		return &authpb.ValidateTokenResponse{}, fmt.Errorf("failed to load environment file: %w", err)
	}

	secretSignInKey := os.Getenv("SECRET_SIGNINKEY")
	if secretSignInKey == "" {
		return &authpb.ValidateTokenResponse{}, fmt.Errorf("SECRET_SIGNINKEY environment variable is not set")
	}

	token, err := jwt.ParseWithClaims(req.AccessToken, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Проверяем метод подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretSignInKey), nil
	})

	if err != nil {
		return &authpb.ValidateTokenResponse{}, fmt.Errorf("invalid token: %w", err)
	}

	// Проверяем валидность claims
	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok && token.Valid {
		if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
			return &authpb.ValidateTokenResponse{}, err
		}

		userID, err := strconv.Atoi(claims.Subject)
		if err != nil {
			return &authpb.ValidateTokenResponse{
				UserId: int64(userID),
			}, fmt.Errorf("invalid user ID in access token: %w", err)
		}
		return &authpb.ValidateTokenResponse{
			UserId: int64(userID),
		}, nil
	}

	return nil, fmt.Errorf("invalid token claims")
}

func (s *AuthGRPCService) GetUserIDbyRefreshToken(ctx context.Context, req *authpb.GetUserIDbyRefreshTokenRequest) (*authpb.GetUserIDbyRefreshTokenResponse, error) {
	userID, err := s.authrepository.GetUserIDbyRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get user ID by refresh token: %w", err)
	}
	return &authpb.GetUserIDbyRefreshTokenResponse{
		UserId: userID,
	}, nil
}

func (s *AuthGRPCService) GenerateAccessToken(ctx context.Context, req *authpb.GenerateAccessTokenRequest) (*authpb.GenerateAccessTokenResponse, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   strconv.Itoa(int(req.UserId)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(AccessTokenTTL)),
	})
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
		return &authpb.GenerateAccessTokenResponse{
			AccessToken: "",
		}, fmt.Errorf("failed to load environment file: %w", err)
	}
	secretSignInKey := os.Getenv("SECRET_SIGNINKEY")
	if secretSignInKey == "" {
		return &authpb.GenerateAccessTokenResponse{
			AccessToken: "",
		}, fmt.Errorf("SECRET_SIGNINKEY environment variable is not set")
	}
	tokenSignedString, err := token.SignedString([]byte(secretSignInKey))
	if err != nil {
		return &authpb.GenerateAccessTokenResponse{
			AccessToken: "",
		}, fmt.Errorf("failed to sign token: %w", err)
	}
	return &authpb.GenerateAccessTokenResponse{
		AccessToken: tokenSignedString,
	}, nil
}

func (s *AuthGRPCService) GenerateRefreshToken(ctx context.Context, req *authpb.GenerateRefreshTokenRequest) (*authpb.GenerateRefreshTokenResponse, error) {
	refresh_token := make([]byte, 32)
	if _, err := rand.Read(refresh_token); err != nil {
		return &authpb.GenerateRefreshTokenResponse{
			UserId:       req.UserId,
			RefreshToken: "",
		}, err
	}
	token := base64.URLEncoding.EncodeToString(refresh_token)

	return &authpb.GenerateRefreshTokenResponse{
		UserId:       req.UserId,
		RefreshToken: token,
	}, nil
}

func (s *AuthGRPCService) RemoveOldRefreshToken(ctx context.Context, req *authpb.RemoveOldRefreshTokenRequest) (*emptypb.Empty, error) {
	err := s.authrepository.RemoveRefreshToken(ctx, req.UserId, req.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to remove old refresh token: %w", err)
	}
	return &emptypb.Empty{}, nil
}

func (s *AuthGRPCService) SaveNewRefreshToken(ctx context.Context, req *authpb.SaveNewRefreshTokenRequest) (*emptypb.Empty, error) {
	err := s.authrepository.SaveNewRefreshToken(ctx, req.UserId, req.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to save new refresh token: %w", err)
	}
	return &emptypb.Empty{}, nil

}
