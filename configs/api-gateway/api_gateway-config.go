package apigateway_config

import (
	"github.com/spf13/viper"
)

type Config struct {
	HTTPServer HTTPServer `mapstructure:"http_server"`
}

type HTTPServer struct {
	Address string `yaml:"address" env-default:"0.0.0.0:9091"`
}

func LoadConfigServer(configPath string) (*Config, error) {
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
		HTTPServer: HTTPServer{
			Address: apiConfig.HTTPServer.Address,
		},
	}, nil
}
