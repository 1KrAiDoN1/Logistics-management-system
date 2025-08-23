package warehouseservice_config

import (
	"log/slog"
	"logistics/pkg/lib/logger/slogger"
	"logistics/pkg/lib/utils"
)

func LoadWarehouseGRPCServiceConfig(configPath string) (utils.ServiceConfig, error) {
	serviceConfig, err := utils.LoadServiceConfig(configPath, "DB_WAREHOUSE_SERVICE_PASSWORD")
	if err != nil {
		slog.Error("Failed to load warehouse service configuration", slogger.Err(err))
		return utils.ServiceConfig{}, err
	}
	return serviceConfig, nil
}
