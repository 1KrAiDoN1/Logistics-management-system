package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(authGRPCService authpb.AuthServiceClient) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		log := logger.New("middleware", true)
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			// Извлекаем токен из заголовка "Bearer TOKEN"
			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
				log.Error("Invalid authorization header format", map[string]interface{}{
					"header": authHeader,
					"status": http.StatusUnauthorized,
				})
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
				c.Abort()
				return
			} else {
				req := dto.AccessTokenRequest{
					AccessToken: tokenParts[1],
				}

				// Валидация токена через сервис
				userID, err := authService.ValidateToken(ctx, req)
				if err != nil {
					log.Error("Invalid token", map[string]interface{}{
						"error":  err,
						"status": http.StatusUnauthorized,
					})
					c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
					c.Abort()
					return
				}
				c.Set("user_id", uint(userID.UserID))
				c.Next()
				return
			}
		} else {
			refresh_token, err := c.Cookie("refresh_token")
			if err != nil {
				log.Error("Refresh token is required", map[string]interface{}{
					"error":  err,
					"status": http.StatusUnauthorized,
				})
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token is required"})
				c.Abort()
				return
			}
			if refresh_token != "" {
				userID, err := authService.GetUserIDbyRefreshToken(ctx, refresh_token) //нужно проверку сделать что refresh_token не истек
				if err != nil {
					log.Error("Invalid refresh token", map[string]interface{}{
						"error":  err,
						"status": http.StatusUnauthorized,
					})
					c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
					c.Abort()
					return
				}
				if userID != 0 {
					err = authService.RemoveOldRefreshToken(ctx, userID)
					if err != nil {
						log.Error("Failed to remove old refresh token", map[string]interface{}{
							"error":  err,
							"status": http.StatusInternalServerError,
						})
						c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove old refresh token"})
						c.Abort()
						return
					}
					new_access_token, err := authService.GenerateAccessToken(userID)
					if err != nil {
						log.Error("Failed to generate new access token", map[string]interface{}{
							"error":  err,
							"status": http.StatusInternalServerError,
						})
						c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate new access token"})
						c.Abort()
						return
					}
					new_refresh_token, err := authService.GenerateRefreshToken()
					if err != nil {
						log.Error("Failed to generate new refresh token", map[string]interface{}{
							"error":  err,
							"status": http.StatusInternalServerError,
						})
						c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate new refresh token"})
						c.Abort()
						return
					}
					err = authService.SaveNewRefreshToken(ctx, userID, new_refresh_token)
					if err != nil {
						log.Error("Failed to save new refresh token", map[string]interface{}{
							"error":  err,
							"status": http.StatusInternalServerError,
						})
						c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save new refresh token"})
						c.Abort()
						return
					}
					c.Header("Authorization", "Bearer "+new_access_token.AccessToken)
					SetRefreshTokenCookie(c, new_refresh_token.RefreshToken)
					c.Set("user_id", uint(userID))
					c.Next()
					return

				} else {
					log.Error("Authorization is required", map[string]interface{}{
						"error":  err,
						"status": http.StatusUnauthorized,
					})
					c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization is required"})
					c.Abort()
					return
				}

			} else {
				log.Error("Authorization is required", map[string]interface{}{
					"error":  err,
					"status": http.StatusUnauthorized,
				})
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

func SetRefreshTokenCookie(c *gin.Context, refresh_token string) {
	c.SetCookie(
		"refresh_token",
		refresh_token,
		int(services.RefreshTokenTTL.Seconds()),
		"/",
		"",
		true,
		true,
	)
}
