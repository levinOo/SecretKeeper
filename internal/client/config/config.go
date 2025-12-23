package config

import (
	"github.com/caarlos0/env"
	"github.com/joho/godotenv"
)

// Структура конфигурации клиента
type Config struct {
	GRPC GRPCConfig
}

// Структура конфигурации gRPC подключения
type GRPCConfig struct {
	Addr string `env:"GRPC_SERVER_ADDR" envDefault:":50051"`
}

// Функция загрузки конфигурации клиента
func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, err
	}

	cfg := &Config{}

	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
