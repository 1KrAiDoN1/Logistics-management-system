package orderservice_config

import (
	"log"
	"logistics/pkg/lib/logger/slogger"
	"os"

	"gopkg.in/yaml.v2"
)

type OrderGRPCServiceConfig struct {
	Address string `yaml:"order_grpc_service_address"`
}

func LoadOrderGRPCServiceConfig(configPath string) (*OrderGRPCServiceConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Println("Error reading config file", slogger.Err(err))
		defer os.Exit(1)
		return nil, err
	}

	var address OrderGRPCServiceConfig
	err = yaml.Unmarshal(data, &address)
	if err != nil {
		log.Println("Parsing YAML failed", slogger.Err(err))
		return nil, err
	}

	return &OrderGRPCServiceConfig{
		Address: address.Address,
	}, nil
}
