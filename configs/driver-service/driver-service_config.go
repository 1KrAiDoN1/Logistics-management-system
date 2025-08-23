package driverservice_config

import (
	"log/slog"
	"logistics/pkg/lib/logger/slogger"
	"logistics/pkg/lib/utils"
)

func LoadDriverGRPCServiceConfig(configPath string) (utils.ServiceConfig, error) {
	serviceConfig, err := utils.LoadServiceConfig(configPath, "DB_DRIVER_SERVICE_PASSWORD")
	if err != nil {
		slog.Error("Failed to load driver service configuration", slogger.Err(err))
		return utils.ServiceConfig{}, err
	}
	return serviceConfig, nil
}
