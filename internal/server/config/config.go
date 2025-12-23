package config

import (
	"log"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

// Структура конфигурации приложения объединяет все под-конфигурации
type Config struct {
	// Указатели тут Ок, env/v11 их создаст (инициализирует) сама
	GRPC    *GRPCConfig
	Postgre *PostgreConfig
	Storage *StorageConfig
	JWT     *JWTConfig
}

// Структура конфигурации gRPC сервера
type GRPCConfig struct {
	Addr   string `env:"GRPC_SERVER_ADDR" envDefault:":50051"`
	Pepper string `env:"PEPPER"`
}

// Структура конфигурации JWT токенов
type JWTConfig struct {
	SecretKey  string        `env:"JWT_SECRET_KEY" envDefault:"super-secret"`
	AccessTTL  time.Duration `env:"JWT_ACCESS_TTL" envDefault:"15m"`
	RefreshTTL time.Duration `env:"JWT_REFRESH_TTL" envDefault:"720h"` // 30 days
}

// Структура конфигурации подключения к PostgreSQL
type PostgreConfig struct {
	Host     string `env:"DB_HOST"`
	Port     string `env:"DB_PORT"`
	User     string `env:"DB_USER"`
	Password string `env:"DB_PASSWORD"`
	DBName   string `env:"DB_NAME"`
	SSLMode  string `env:"DB_SSL_MODE"`
}

// Структура конфигурации файлового хранилища
type StorageConfig struct {
	StoragePath   string `env:"STORAGE_PATH" envDefault:"./data/uploads"`
	MaxFileSize   int64  `env:"MAX_FILE_SIZE" envDefault:"104857600"` // 100MB
	MaxSecretSize int64  `env:"MAX_SECRET_SIZE" envDefault:"65536"`   // 64KB

	// MinIO Config
	MinIOEndpoint  string `env:"MINIO_ENDPOINT" envDefault:"localhost:9000"`
	MinIOAccessKey string `env:"MINIO_ACCESS_KEY"`
	MinIOSecretKey string `env:"MINIO_SECRET_KEY"`
	MinIOBucket    string `env:"MINIO_BUCKET" envDefault:"secrets"`
	MinIOUseSSL    bool   `env:"MINIO_USE_SSL" envDefault:"false"`
}

// Функция загрузки конфигурации из .env файла или переменных окружения
func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Инициализируем структуру с пустыми вложенными структурами
	cfg := &Config{
		GRPC:    &GRPCConfig{},
		Postgre: &PostgreConfig{},
		Storage: &StorageConfig{},
		JWT:     &JWTConfig{},
	}

	// Парсим ENV переменные в структуру
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
