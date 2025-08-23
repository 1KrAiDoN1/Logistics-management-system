package handler

import (
	"context"
	"fmt"
	"log/slog"
	authpb "logistics/api/protobuf/auth_service"
	"logistics/internal/shared/models/dto"
	"logistics/pkg/lib/logger/slogger"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	logger         *slog.Logger
	authGRPCClient authpb.AuthServiceClient
}

func NewAuthHandler(logger *slog.Logger, authClient authpb.AuthServiceClient) *AuthHandler {
	return &AuthHandler{
		logger:         logger,
		authGRPCClient: authClient,
	}
}

func (h *AuthHandler) SignUp(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	var userReg dto.RegisterRequest
	if err := c.BindJSON(&userReg); err != nil {
		h.logger.Error("Failed to bind JSON", slogger.Err(err), slog.String("status", fmt.Sprintf("%d", http.StatusBadRequest)))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.authGRPCClient.SignUp(ctx, &authpb.SignUpRequest{
		Email:           userReg.Email,
		Password:        userReg.Password,
		ConfirmPassword: userReg.ConfirmPassword,
		FirstName:       userReg.FirstName,
		LastName:        userReg.LastName,
	})
	if err != nil {
		h.logger.Error("Failed to register user", slogger.Err(err), slog.String("email", userReg.Email), slog.String("status", fmt.Sprintf("%d", http.StatusBadRequest)))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.logger.Info("User registered successfully", slog.String("email", userReg.Email), slog.String("status", fmt.Sprintf("%d", http.StatusCreated)))
	c.JSON(http.StatusCreated, gin.H{"user": user})
}
func (h *AuthHandler) SignIn(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	var userAuth dto.LoginRequest
	if err := c.BindJSON(&userAuth); err != nil {
		h.logger.Error("Failed to bind JSON", slogger.Err(err), slog.String("status", fmt.Sprintf("%d", http.StatusBadRequest)))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.authGRPCClient.SignIn(ctx, &authpb.SignInRequest{
		Email:    userAuth.Email,
		Password: userAuth.Password,
	})
	if err != nil {
		h.logger.Error("Failed to authenticate user", slogger.Err(err), slog.String("email", userAuth.Email), slog.String("status", fmt.Sprintf("%d", http.StatusUnauthorized)))
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	// middleware.SetRefreshTokenCookie(c, token.RefreshToken)

	c.Header("Authorization", "Bearer "+token.AccessToken)

	h.logger.Info("User authenticated successfully", slog.String("email", userAuth.Email), slog.String("status", fmt.Sprintf("%d", http.StatusOK)))
	c.JSON(http.StatusOK, dto.AuthResponse{
		AccessToken: token.AccessToken,
		User: dto.UserInfo{
			ID:        uint(token.UserId),
			Email:     token.Email,
			FirstName: token.FirstName,
			LastName:  token.LastName,
		},
	})
}
func (h *AuthHandler) Logout(c *gin.Context) {

	// if err != nil {
	// 	h.logger.Error("Failed to get refresh token from cookie", slogger.Err(err), slog.String("status", fmt.Sprintf("%d", http.StatusUnauthorized)))
	// 	c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
	// 	return
	// }
	// удалить токены надо из редиса
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}
