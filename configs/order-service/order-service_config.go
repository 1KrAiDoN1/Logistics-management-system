package orderservice_config

import (
	"log/slog"
	"logistics/pkg/lib/logger/slogger"
	"logistics/pkg/lib/utils"
	"os"
)

func LoadOrderGRPCServiceConfig(configPath string) (utils.ServiceConfig, error) {
	serviceConfig, err := utils.LoadServiceConfig(configPath, "DB_WAREHOUSE_SERVICE_PASSWORD")
	if err != nil {
		slog.Error("Failed to load order service configuration", slogger.Err(err))
		defer os.Exit(1)
		return utils.ServiceConfig{}, err
	}
	return serviceConfig, nil
}
