package warehouseservice_config

import (
	"logistics/pkg/lib/logger/slogger"
	"os"

	"log"

	"gopkg.in/yaml.v2"
)

type WarehouseGRPCServiceConfig struct {
	Address string `yaml:"warehouse_grpc_service_address"`
}

func LoadWarehouseGRPCServiceConfig(configPath string) (*WarehouseGRPCServiceConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Println("Error reading config file", slogger.Err(err))
		defer os.Exit(1)
		return nil, err
	}

	var address WarehouseGRPCServiceConfig
	err = yaml.Unmarshal(data, &address)
	if err != nil {
		log.Println("Parsing YAML failed", slogger.Err(err))
		return nil, err
	}

	return &WarehouseGRPCServiceConfig{
		Address: address.Address,
	}, nil
}
