package auth_grpc_repository

import (
	"context"
	"logistics/internal/shared/entity"

	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthRepository struct {
	pool *pgxpool.Pool
}

func NewAuthRepository(pool *pgxpool.Pool) *AuthRepository {
	return &AuthRepository{
		pool: pool,
	}
}

func (a *AuthRepository) CreateUser(ctx context.Context, user *entity.User) (int64, error) {
	query := `INSERT INTO users (first_name, last_name, email, password, time_of_registration) VALUES ($1, $2, $3, $4, $5) RETURNING id`
	var userID int64
	err := a.pool.QueryRow(ctx, query, user.FirstName, user.LastName, user.Email, user.Password, user.TimeOfRegistration).Scan(&userID)
	if err != nil {
		return 0, err
	}
	return userID, nil
}

/// дальше реализация

func (a *AuthRepository) IsUserExists(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`
	var exists bool
	err := a.pool.QueryRow(ctx, query, email).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (a *AuthRepository) CheckUserVerification(ctx context.Context, email string, hashPassword string) (entity.User, error) {
	query := `SELECT id, email, first_name, last_name, password FROM users WHERE email = $1 AND password = $2`
	var user entity.User
	err := a.pool.QueryRow(ctx, query, email, hashPassword).Scan(&user.ID, &user.Email, &user.FirstName, &user.LastName, &user.Password)
	if err != nil {
		return entity.User{}, err
	}
	return user, nil

}

func (a *AuthRepository) SaveNewRefreshToken(ctx context.Context, userID int64, refreshToken string, expires_at int64) error {
	query := `INSERT INTO refresh_tokens (user_id, token, expires_at) VALUES ($1, $2, $3)`
	_, err := a.pool.Exec(ctx, query, userID, refreshToken, expires_at)
	if err != nil {
		return err
	}
	return nil
}

func (a *AuthRepository) RemoveRefreshToken(ctx context.Context, userID int64, refreshToken string) error {
	query := `DELETE FROM refresh_tokens WHERE user_id = $1 AND token = $2`
	_, err := a.pool.Exec(ctx, query, userID, refreshToken)
	if err != nil {
		return err
	}
	return nil
}

func (a *AuthRepository) GetUserIDbyRefreshToken(ctx context.Context, refreshToken string) (int64, error) {
	query := `SELECT user_id FROM refresh_tokens WHERE token = $1`
	var userID int64
	err := a.pool.QueryRow(ctx, query, refreshToken).Scan(&userID)
	if err != nil {
		return 0, err
	}
	return userID, nil
}

func (a *AuthRepository) Logout(ctx context.Context, userID int64) error {
	query := `DELETE FROM refresh_tokens WHERE user_id = $1`
	_, err := a.pool.Exec(ctx, query, userID)
	if err != nil {
		return err
	}
	return nil
}
