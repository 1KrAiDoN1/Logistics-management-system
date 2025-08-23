package utils

import (
	"fmt"
	"log/slog"
	"logistics/pkg/cache/redis"
	"logistics/pkg/lib/logger/slogger"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

func LoadServiceConfig(configPath, dbPasswordPath string) (ServiceConfig, error) {
	v := viper.New()
	v.SetConfigType("yaml")
	v.SetConfigFile(configPath)
	if err := v.ReadInConfig(); err != nil {
		return ServiceConfig{}, err
	}

	var serviceConfig ServiceConfig
	if err := v.Unmarshal(&serviceConfig); err != nil {
		return ServiceConfig{}, err
	}
	dbConnStr, err := serviceConfig.DSN(dbPasswordPath)
	if err != nil {
		slog.Error("Error generating DSN for database connection", slogger.Err(err))
		return ServiceConfig{}, err
	}
	serviceConfig.DbConfig.DBConn = dbConnStr
	return serviceConfig, nil
}

func (d ServiceConfig) DSN(dbPasswordPath string) (string, error) {
	err := godotenv.Load(".env")
	if err != nil {
		slog.Error("Error loading .env file", slogger.Err(err))
		return "", err
	}
	return fmt.Sprintf("%s://%s:%s@%s:%d/%s",
		d.DbConfig.Driver, d.DbConfig.User, os.Getenv(dbPasswordPath), d.DbConfig.Host, d.DbConfig.Port, d.DbConfig.DBName), nil
}

type ServiceConfig struct {
	Address     string            `yaml:"address"`
	DbConfig    DBConfig          `mapstructure:"database"`
	RedisConfig redis.RedisConfig `mapstructure:"redis_config"`
}
type DBConfig struct {
	Driver string `yaml:"driver"`
	Host   string `yaml:"host"`
	Port   int    `yaml:"port"`
	User   string `yaml:"user"`
	DBName string `yaml:"dbname"`
	DBConn string
}
