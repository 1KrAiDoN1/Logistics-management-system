package dto

import (
	"time"
)

// Запросы для аутентификации

// RegisterRequest - данные для регистрации
type RegisterRequest struct {
	Email           string `json:"email" validate:"required,email" example:"user@example.com"`
	Password        string `json:"password" validate:"required,min=8,max=100" example:"password123"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=Password"`
	FirstName       string `json:"first_name" validate:"required,min=2,max=50" example:"John"`
	LastName        string `json:"last_name" validate:"required,min=2,max=50" example:"Doe"`
}

// LoginRequest - данные для входа
type LoginRequest struct {
	// UserID   uint   `json:"id"`
	Email    string `json:"email" validate:"required,email" example:"user@example.com"`
	Password string `json:"password" validate:"required" example:"password123"`
}

// RefreshTokenRequest - запрос обновления токена
type RefreshTokenRequest struct {
	RefreshToken string    `json:"refresh_token" validate:"required"`
	ExpiresAt    time.Time `json:"expires_at"`
}

type AccessTokenRequest struct {
	AccessToken string    `json:"access_token" validate:"required"`
	ExpiresAt   time.Time `json:"expires_at"`
}

// Ответы аутентификации

// AuthResponse - ответ после успешной аутентификации
type AuthResponse struct {
	AccessToken string   `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	User        UserInfo `json:"user"`
}

type UserInfo struct {
	ID        uint   `json:"id" example:"1"`
	Email     string `json:"email" example:"user@example.com"`
	FirstName string `json:"first_name" example:"John"`
	LastName  string `json:"last_name" example:"Doe"`
}
