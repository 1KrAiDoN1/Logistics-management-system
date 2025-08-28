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

func (r *AuthRepository) CreateUser(ctx context.Context, user *entity.User) (int64, error) {
	// Реализация логики регистрации пользователя в базе данных
	// Например, вставка данных пользователя в таблицу users
	return 1, nil // Возвращаем ID нового пользователя
}

/// дальше реализация

func (r *AuthRepository) IsUserExists(ctx context.Context, email string) (bool, error) {
	// Реализация логики проверки существования пользователя в базе данных
	// Например, выполнение запроса к таблице users по email
	return false, nil // Возвращаем true, если пользователь существует
}
