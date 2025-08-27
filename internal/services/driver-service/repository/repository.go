package repository

import "github.com/jackc/pgx/v5/pgxpool"

type DriverRepository struct {
	pool *pgxpool.Pool
}

func NewDriverRepository(pool *pgxpool.Pool) *DriverRepository {
	return &DriverRepository{
		pool: pool,
	}
}
