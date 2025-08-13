package driverservice_config

import (
	"log"
	"logistics/pkg/lib/logger/slogger"
	"os"

	"gopkg.in/yaml.v2"
)

type DriverGRPCServiceConfig struct {
	Address string `yaml:"driver_grpc_service_address"`
}

func LoadDriverGRPCServiceConfig(configPath string) (*DriverGRPCServiceConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Println("Error reading config file", slogger.Err(err))
		defer os.Exit(1)
		return nil, err
	}

	var address DriverGRPCServiceConfig
	err = yaml.Unmarshal(data, &address)
	if err != nil {
		log.Println("Parsing YAML failed", slogger.Err(err))
		return nil, err
	}

	return &DriverGRPCServiceConfig{
		Address: address.Address,
	}, nil
}
