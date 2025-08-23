package apigateway_config

import (
	"fmt"
	"log/slog"
	"logistics/pkg/cache/redis"
	"logistics/pkg/lib/logger/slogger"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	DB_config_path string
	HTTPServer     HTTPServer        `mapstructure:"http_server"`
	RedisConfig    redis.RedisConfig `mapstructure:"redis_config"`
}

type HTTPServer struct {
	Address string `yaml:"address" env-default:"0.0.0.0:9091"`
}

func SetConfig() (string, error) {
	err := godotenv.Load(".env")
	if err != nil {
		slog.Error("Error loading .env file", slogger.Err(err))
		return "", err
	}
	DB_config_path := fmt.Sprintf("%s://%s:%s@%s:%s/%s", os.Getenv("DB_DRIVER"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_HOST"), os.Getenv("DB_apiConfig"), os.Getenv("DB_NAME"))
	return DB_config_path, nil
}

func LoadConfigServer(configPath string) (*Config, error) {
	db_configPath, err := SetConfig()
	if err != nil {
		return nil, err
	}
	v := viper.New()
	v.SetConfigType("yaml")
	v.SetConfigFile(configPath)
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	var apiConfig *Config
	if err := v.Unmarshal(&apiConfig); err != nil {
		return nil, err
	}

	return &Config{
		DB_config_path: db_configPath,
		HTTPServer: HTTPServer{
			Address: apiConfig.HTTPServer.Address,
		},
		RedisConfig: redis.RedisConfig{
			Address:      apiConfig.RedisConfig.Address,
			Password:     apiConfig.RedisConfig.Password,
			DB:           apiConfig.RedisConfig.DB,
			PoolSize:     apiConfig.RedisConfig.PoolSize,
			MinIdleConns: apiConfig.RedisConfig.MinIdleConns,
			MaxRetries:   apiConfig.RedisConfig.MaxRetries,
			DialTimeout:  apiConfig.RedisConfig.DialTimeout,
			ReadTimeout:  apiConfig.RedisConfig.ReadTimeout,
			WriteTimeout: apiConfig.RedisConfig.WriteTimeout,
		},
	}, nil
}
