package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

const ()

type (
	// Общий конфиг сервера
	Config struct {
		GRPC    GRPCConfig
		Auth    AuthConfig
		Postgre PostgreConfig
		Storage StorageConfig
	}

	// Конфиг авторизации
	AuthConfig struct {
		JWT    JWTConfig
		Pepper string `env:"PEPPER"`
	}

	// Конфиг JWT
	JWTConfig struct {
		SigningKey      string        `env:"JWT_SIGNING_KEY"`
		AccessTokenTTL  time.Duration `env:"JWT_ACCESS_TTL" envDefault:"15m"`
		RefreshTokenTTL time.Duration `env:"JWT_REFRESH_TTL" envDefault:"168h"`
	}

	// Конфиг gRPC
	GRPCConfig struct {
		Addr string `env:"GRPC_SERVER_ADDR" envDefault:":50051"`
		Auth AuthConfig
	}

	// Конфиг Postgre
	PostgreConfig struct {
		Host           string        `env:"DB_HOST"`
		Port           string        `env:"DB_PORT"`
		User           string        `env:"DB_USER"`
		Password       string        `env:"DB_PASSWORD"`
		DBName         string        `env:"DB_NAME"`
		SSLMode        string        `env:"DB_SSL_MODE"`
		ContextTimeout time.Duration `env:"DB_CONTEXT_TIMEOUT"`
	}
)

// Структура конфигурации файлового хранилища
type StorageConfig struct {
	Endpoint      string `env:"MINIO_ENDPOINT"`
	Bucket        string `env:"MINIO_BUCKET"`
	AccessKey     string `env:"MINIO_ACCESS_KEY"`
	SecretKey     string `env:"MINIO_SECRET_KEY"`
	UseSSL        bool   `env:"MINIO_USE_SSL" envDefault:"false"`
	MaxFileSize   int64  `env:"MAX_FILE_SIZE" envDefault:"104857600"` // 100MB
	MaxSecretSize int64  `env:"MAX_SECRET_SIZE" envDefault:"65536"`   // 64KB
}

// Функция загрузки конфига
func LoadConfig() (*Config, error) {
	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		return nil, fmt.Errorf("CONFIG_PATH is not set")
	}

	if err := godotenv.Load(path); err != nil {
		log.Println("Файл .env не найден, используются системные переменные окружения")
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
