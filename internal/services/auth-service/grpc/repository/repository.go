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

func (r *AuthRepository) CheckUserVerification(ctx context.Context, email string, hashpassword string) (*entity.User, error) {
	// Реализация логики проверки учетных данных пользователя
	// Например, выполнение запроса к таблице users по email и hashPassword
	return &entity.User{
		ID:       1,
		Email:    email,
		Password: hashpassword,
	}, nil // Возвращаем пользователя, если учетные данные верны
}

func (r *AuthRepository) SaveNewRefreshToken(ctx context.Context, userID int64, refreshToken string) error {
	// Реализация логики сохранения нового refresh токена в базе данных
	// Например, вставка данных в таблицу refresh_tokens
	return nil
}

func (r *AuthRepository) RemoveRefreshToken(ctx context.Context, userID int64, refreshToken string) error {
	// Реализация логики удаления refresh токена из базы данных
	// Например, удаление записи из таблицы refresh_tokens
	return nil
}

func (r *AuthRepository) GetUserIDbyRefreshToken(ctx context.Context, refreshToken string) (int64, error) {
	// Реализация логики получения userID по refresh токену из базы данных
	// Например, выполнение запроса к таблице refresh_tokens
	return 1, nil // Возвращаем userID, если токен найден
}

func (r *AuthRepository) Logout(ctx context.Context, userID int64) error {
	// Реализация логики выхода пользователя из системы
	// Например, удаление всех refresh токенов пользователя из базы данных
	return nil
}
