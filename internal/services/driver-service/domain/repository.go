package domain

import (
	"context"
	"logistics/internal/shared/entity"
)

type DriverRepositoryInterface interface {
	// FindSuitableDriver(location string) ([]*entity.Driver, error)
	GetAvailableDrivers(ctx context.Context) ([]*entity.Driver, error)
	UpdateDriverStatus(ctx context.Context, driverID int, status string) error
}
