package auth_grpc_repository

import "github.com/jackc/pgx/v5/pgxpool"

type AuthRepository struct {
	pool *pgxpool.Pool
}

func NewAuthRepository(pool *pgxpool.Pool) *AuthRepository {
	return &AuthRepository{
		pool: pool,
	}
}

func (r *AuthRepository) SignUp(email, password string) (int64, error) {
	// Реализация логики регистрации пользователя в базе данных
	// Например, вставка данных пользователя в таблицу users
	return 1, nil // Возвращаем ID нового пользователя
}

/// дальше реализация
