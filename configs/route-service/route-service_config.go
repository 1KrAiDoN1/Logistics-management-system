package routeservice_config

import (
	"log"
	"logistics/pkg/lib/logger/slogger"
	"os"

	"gopkg.in/yaml.v2"
)

type RouteGRPCServiceConfig struct {
	Address string `yaml:"route_grpc_service_address"`
}

func LoadRouteGRPCServiceConfig(configPath string) (*RouteGRPCServiceConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Println("Error reading config file", slogger.Err(err))
		defer os.Exit(1)
		return nil, err
	}

	var address RouteGRPCServiceConfig
	err = yaml.Unmarshal(data, &address)
	if err != nil {
		log.Println("Parsing YAML failed", slogger.Err(err))
		return nil, err
	}

	return &RouteGRPCServiceConfig{
		Address: address.Address,
	}, nil
}
