package config

import (
	"log"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

const ()

type (
	Config struct {
		GRPC    GRPCConfig
		Auth    AuthConfig
		Postgre PostgreConfig
		Storage StorageConfig
	}

	AuthConfig struct {
		JWT          JWTConfig
		PasswordSalt string
	}

	JWTConfig struct {
		AccessTokenTTL  time.Duration
		RefreshTokenTTL time.Duration
		SigningKey      string
	}

	GRPCConfig struct {
		Addr   string `env:"GRPC_SERVER_ADDR" envDefault:":50051"`
		Pepper string `env:"PEPPER"`
	}

	PostgreConfig struct {
		Host     string `env:"DB_HOST"`
		Port     string `env:"DB_PORT"`
		User     string `env:"DB_USER"`
		Password string `env:"DB_PASSWORD"`
		DBName   string `env:"DB_NAME"`
		SSLMode  string `env:"DB_SSL_MODE"`
	}
)

// Структура конфигурации файлового хранилища
type StorageConfig struct {
	StoragePath   string `env:"STORAGE_PATH" envDefault:"./data/uploads"`
	MaxFileSize   int64  `env:"MAX_FILE_SIZE" envDefault:"104857600"` // 100MB
	MaxSecretSize int64  `env:"MAX_SECRET_SIZE" envDefault:"65536"`   // 64KB

	Endpoint  string
	Bucket    string
	AccessKey string
	SecretKey string
	UseSSL    bool
}

func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	cfg := &Config{
		GRPC:    GRPCConfig{},
		Auth:    AuthConfig{JWT: JWTConfig{}},
		Postgre: PostgreConfig{},
		Storage: StorageConfig{},
	}

	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
