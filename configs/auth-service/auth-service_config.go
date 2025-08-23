package authservice_config

import (
	"log/slog"
	"logistics/pkg/lib/logger/slogger"
	"logistics/pkg/lib/utils"
)

func LoadAuthGRPCServiceConfig(configPath string) (utils.ServiceConfig, error) {
	serviceConfig, err := utils.LoadServiceConfig(configPath, "DB_AUTH_SERVICE_PASSWORD")
	if err != nil {
		slog.Error("Failed to load auth service configuration", slogger.Err(err))
		return utils.ServiceConfig{}, err
	}
	return serviceConfig, nil
}
