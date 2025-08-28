package domain

import (
	"context"
	"logistics/internal/shared/entity"
)

type AuthRepositoryInterface interface {
	IsUserExists(ctx context.Context, email string) (bool, error)
	CreateUser(ctx context.Context, user *entity.User) (int64, error)
	// SignUp creates a new user in the database.
	// SignUp(email, password, firstName, lastName string) (uint, error)
	// // SignIn checks user credentials and returns user ID if valid.
	// SignIn(email, password string) (uint, error)
	// // Logout removes the user's session or token.
	// Logout(userID uint) error
	// // ValidateToken checks if the provided access token is valid and returns the user ID.
	// ValidateToken(accessToken string) (uint, error)
	// // GetUserIDbyRefreshToken retrieves the user ID associated with the provided refresh token.
	// GetUserIDbyRefreshToken(refreshToken string) (uint, error)
	// // GenerateAccessToken creates a new access token for the user.
	// GenerateAccessToken(userID uint) (string, error)
	// // GenerateRefreshToken creates a new refresh token for the user.
	// GenerateRefreshToken(userID uint) (string, error)
	// // RemoveOldRefreshToken deletes the old refresh token for the user.
	// RemoveOldRefreshToken(userID uint, refreshToken string) error
	// // GetUserByID retrieves user information by user ID.
	// GetUserByID(userID uint) (string, string, string, string, error) // returns email, firstName, lastName, and error
}
