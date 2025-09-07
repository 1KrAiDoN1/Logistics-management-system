package middleware

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	authpb "logistics/api/protobuf/auth_service"
	"logistics/internal/shared/models/dto"
	"logistics/pkg/lib/logger/slogger"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	RefreshTokenTTL = 24 * time.Hour
)

func AuthMiddleware(authGRPCService authpb.AuthServiceClient) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			// Извлекаем токен из заголовка "Bearer TOKEN"
			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
				slog.Error("Invalid authorization header format")
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
				c.Abort()
				return
			} else {
				req := dto.AccessTokenRequest{
					AccessToken: tokenParts[1],
				}

				// Валидация токена через сервис
				userID, err := authGRPCService.ValidateToken(ctx, &authpb.ValidateTokenRequest{
					AccessToken: req.AccessToken,
				})
				if err != nil {
					slog.Error("Invalid token", slogger.Err(err), slog.String("status", fmt.Sprintf("%d", http.StatusUnauthorized)))
					c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
					c.Abort()
					return
				}
				c.Set("user_id", uint(userID.UserId))
				c.Next()
				return
			}
		} else {
			refresh_token, err := c.Cookie("refresh_token")
			if err != nil {
				slog.Error("Refresh token is required", slogger.Err(err), slog.String("status", fmt.Sprintf("%d", http.StatusUnauthorized)))
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token is required"})
				c.Abort()
				return
			}
			if refresh_token != "" {
				userID, err := authGRPCService.GetUserIDbyRefreshToken(ctx, &authpb.GetUserIDbyRefreshTokenRequest{
					RefreshToken: refresh_token,
				})
				if err != nil {
					slog.Error("Invalid refresh token", slogger.Err(err), slog.String("status", fmt.Sprintf("%d", http.StatusUnauthorized)))
					c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
					c.Abort()
					return
				}
				if userID.UserId != 0 {
					_, err = authGRPCService.RemoveOldRefreshToken(ctx, &authpb.RemoveOldRefreshTokenRequest{
						UserId:       userID.UserId,
						RefreshToken: refresh_token,
					}) // удалить старый refresh token
					if err != nil {
						slog.Error("Failed to remove old refresh token", slogger.Err(err), slog.String("status", fmt.Sprintf("%d", http.StatusInternalServerError)))
						c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove old refresh token"})
						c.Abort()
						return
					}
					new_access_token, err := authGRPCService.GenerateAccessToken(ctx, &authpb.GenerateAccessTokenRequest{
						UserId: userID.UserId,
					})
					if err != nil {
						slog.Error("Failed to generate new access token", slogger.Err(err), slog.String("status", fmt.Sprintf("%d", http.StatusInternalServerError)))
						c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate new access token"})
						c.Abort()
						return
					}
					new_refresh_token, err := authGRPCService.GenerateRefreshToken(ctx, &authpb.GenerateRefreshTokenRequest{
						UserId: userID.UserId,
					})
					if err != nil {
						slog.Error("Failed to generate new refresh token", slogger.Err(err), slog.String("status", fmt.Sprintf("%d", http.StatusInternalServerError)))
						c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate new refresh token"})
						c.Abort()
						return
					}

					_, err = authGRPCService.SaveNewRefreshToken(ctx, &authpb.SaveNewRefreshTokenRequest{
						UserId:       userID.UserId,
						RefreshToken: new_refresh_token.RefreshToken,
					})
					if err != nil {
						slog.Error("Failed to save new refresh token", slogger.Err(err), slog.String("status", fmt.Sprintf("%d", http.StatusInternalServerError)))
						c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save new refresh token"})
						c.Abort()
						return
					}
					c.Header("Authorization", "Bearer "+new_access_token.AccessToken)
					SetRefreshTokenCookie(c, new_refresh_token.RefreshToken)
					c.Set("user_id", uint(userID.UserId))
					c.Next()
					return

				} else {
					slog.Error("Authorization is required", slog.String("status", fmt.Sprintf("%d", http.StatusUnauthorized)))
					c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization is required"})
					c.Abort()
					return
				}

			} else {
				slog.Error("Authorization is required", slog.String("status", fmt.Sprintf("%d", http.StatusUnauthorized)))
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization is required"})
				c.Abort()
				return

			}
		}

	})
}

func GetUserId(c *gin.Context) (uint, error) {
	userID, ok := c.Get("user_id")
	if !ok {
		return 0, errors.New("user_id not found in context")
	}

	// Проверяем тип и конвертируем
	switch v := userID.(type) {
	case uint:
		return v, nil
	case int:
		return uint(v), nil
	default:
		return 0, fmt.Errorf("invalid user_id type: %T", userID)
	}
}

func SetRefreshTokenCookie(c *gin.Context, refreshToken string) {
	c.SetCookie("refresh_token", refreshToken, int(RefreshTokenTTL.Seconds()), "/", "", true, true)
}
