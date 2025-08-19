package handler

import (
	"log/slog"
	authpb "logistics/api/protobuf/auth_service"

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
	h.logger.Info("SignUp handler called")
	// Здесь должна быть логика обработки запроса регистрации
}
func (h *AuthHandler) SignIn(c *gin.Context) {
	// Реализация логики входа пользователя
	h.logger.Info("SignIn handler called")
	// Здесь должна быть логика обработки запроса входа
}
func (h *AuthHandler) Logout(c *gin.Context) {
	// Реализация логики выхода пользователя
	h.logger.Info("Logout handler called")
	// Здесь должна быть логика обработки запроса выхода
}
