package handler

import (
	"context"
	"fmt"
	"log/slog"
	authpb "logistics/api/protobuf/auth_service"
	"logistics/internal/services/api-gateway/middleware"
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

// @Summary Регистрация пользователя
// @Description Создает нового пользователя в системе
// @Tags auth
// @Accept  json
// @Produce  json
// @Param   request body dto.RegisterRequest true "Данные для регистрации"
// @Success 201 {object} object{user_id=int64,email=string,first_name=string,last_name=string} "Успешная регистрация"
// @Failure 400 {object} object{error=string} "Некорректные данные"
// @Failure 500 {object} object{error=string} "Ошибка сервера"
// @Router /auth/sign-up [post]
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
		Email:              userReg.Email,
		Password:           userReg.Password,
		ConfirmPassword:    userReg.ConfirmPassword,
		FirstName:          userReg.FirstName,
		LastName:           userReg.LastName,
		TimeOfRegistration: time.Now().Unix(),
	})
	if err != nil {
		h.logger.Error("Failed to register user", slogger.Err(err), slog.String("email", userReg.Email), slog.String("status", fmt.Sprintf("%d", http.StatusInternalServerError)))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.logger.Info("User registered successfully", slog.String("email", userReg.Email), slog.String("status", fmt.Sprintf("%d", http.StatusCreated)))
	c.JSON(http.StatusCreated, gin.H{"user": user.UserId, "email": user.Email, "first_name": user.FirstName, "last_name": user.LastName})
}

// @Summary Аутентификация пользователя
// @Description Выполняет вход пользователя и возвращает токены
// @Tags auth
// @Accept  json
// @Produce  json
// @Param   request body dto.LoginRequest true "Данные для входа"
// @Success 200 {object} dto.AuthResponse
// @Failure 400 {object} object{error=string} "Некорректные данные"
// @Failure 401 {object} object{error=string} "Неверные учетные данные"
// @Failure 500 {object} object{error=string} "Ошибка сервера"
// @Router /auth/sign-in [post]
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
	_, err = h.authGRPCClient.SaveNewRefreshToken(ctx, &authpb.SaveNewRefreshTokenRequest{
		UserId:       token.UserId,
		RefreshToken: token.RefreshToken,
		ExpiresAt:    time.Now().Add(middleware.RefreshTokenTTL).Unix(),
	})
	if err != nil {
		h.logger.Error("Failed to save refresh token", slogger.Err(err), slog.String("email", userAuth.Email), slog.String("status", fmt.Sprintf("%d", http.StatusInternalServerError)))
	}
	middleware.SetRefreshTokenCookie(c, token.RefreshToken)

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

// @Summary Выход из системы
// @Description Выполняет выход пользователя и удаляет refresh token
// @Tags auth
// @Produce  json
// @Success 200 {object} object{message=string}
// @Failure 500 {object} object{error=string} "Ошибка сервера"
// @Security ApiKeyAuth
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	userID, err := middleware.GetUserId(c)
	if err != nil {
		h.logger.Error("getting user_id failed", slog.String("status", fmt.Sprintf("%d", http.StatusInternalServerError)), slogger.Err(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	_, err = h.authGRPCClient.Logout(ctx, &authpb.LogoutRequest{
		UserId: int64(userID),
	})
	if err != nil {
		h.logger.Error("logout user failed", slog.String("status", fmt.Sprintf("%d", http.StatusInternalServerError)), slogger.Err(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}
