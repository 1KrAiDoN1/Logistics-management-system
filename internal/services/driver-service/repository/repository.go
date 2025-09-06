package repository

import (
	"context"
	"fmt"
	"logistics/internal/shared/entity"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DriverRepository struct {
	pool *pgxpool.Pool
}

func NewDriverRepository(pool *pgxpool.Pool) *DriverRepository {
	return &DriverRepository{
		pool: pool,
	}
}

func (d *DriverRepository) GetAvailableDrivers(ctx context.Context) ([]*entity.Driver, error) {
	query := `SELECT id, name, phone, license_number, car, status FROM drivers WHERE status = 'available' ORDER BY id`
	rows, err := d.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query available drivers: %w", err)
	}
	defer rows.Close()

	var drivers []*entity.Driver
	for rows.Next() {
		var driver entity.Driver
		err := rows.Scan(
			&driver.ID,
			&driver.Name,
			&driver.Phone,
			&driver.LicenseNumber,
			&driver.Car,
			&driver.Status,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan driver row: %w", err)
		}
		drivers = append(drivers, &driver)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating through driver rows: %w", err)
	}

	return drivers, nil
}

func (d *DriverRepository) UpdateDriverStatus(ctx context.Context, driverID int, status string) error {
	query := `UPDATE drivers SET status = $1 WHERE id = $2`
	_, err := d.pool.Exec(ctx, query, status, driverID)
	if err != nil {
		return fmt.Errorf("failed to update driver status: %w", err)
	}
	return nil
}
