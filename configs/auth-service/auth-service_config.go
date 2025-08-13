package authservice_config

import (
	"log"
	"logistics/pkg/lib/logger/slogger"
	"os"

	"gopkg.in/yaml.v2"
)

type AuthGRPCServiceConfig struct {
	Address string `yaml:"auth_grpc_service_address"`
}

func LoadAuthGRPCServiceConfig(configPath string) (*AuthGRPCServiceConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Println("Error reading config file", slogger.Err(err))
		defer os.Exit(1)
		return nil, err
	}

	var address AuthGRPCServiceConfig
	err = yaml.Unmarshal(data, &address)
	if err != nil {
		log.Println("Parsing YAML failed", slogger.Err(err))
		return nil, err
	}

	return &AuthGRPCServiceConfig{
		Address: address.Address,
	}, nil
}
