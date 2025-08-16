package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	pool *pgxpool.Pool
}

func NewDatabase(ctx context.Context, databaseURL string) (*Storage, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database connection string: %w", err)
	}
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Проверка соединения
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	return &Storage{
			pool: pool},
		nil

}

func (d *Storage) GetPool() *pgxpool.Pool {
	return d.pool
}
func (d *Storage) Close() error {
	if d.pool != nil {
		d.pool.Close()
	}
	return nil
}
