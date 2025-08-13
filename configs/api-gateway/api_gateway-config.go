package apigateway_config

import (
	"fmt"
	"log"
	"logistics/pkg/lib/logger/slogger"
	"os"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v2"
)

type Config struct {
	DB_config_path string
	HTTPServer     `yaml:"http_server"`
}

func SetConfig() (string, error) {
	log := slogger.SetupLogger()
	err := godotenv.Load(".env")
	if err != nil {
		log.Error("Error loading .env file", slogger.Err(err))
		os.Exit(1)
		return "", err
	}
	DB_config_path := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_NAME"))
	return DB_config_path, nil
}

type HTTPServer struct {
	Address string `yaml:"address" env-default:"0.0.0.0:9091"`
}

func LoadConfigServer(configPath string) (*Config, error) {
	db_configPath, err := SetConfig()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatal("Error reading config file", slogger.Err(err))
		return nil, err
	}

	var port Config
	err = yaml.Unmarshal(data, &port)
	if err != nil {
		log.Fatal("Parsing YAML failed", slogger.Err(err))
		return nil, err
	}

	return &Config{
		DB_config_path: db_configPath,
		HTTPServer: HTTPServer{
			Address: port.HTTPServer.Address,
		},
	}, nil
}
