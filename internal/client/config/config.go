package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// ClientConfig конфигурация клиента
type ClientConfig struct {
	ServerAddr string // Адрес gRPC сервера
	CertPath   string // Путь к TLS сертификату
}

func LoadConfig(v *viper.Viper) (*ClientConfig, error) {
	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		return nil, fmt.Errorf("CONFIG_PATH is not set")
	}

	_ = godotenv.Load(path)

	v.SetEnvPrefix("SK")
	v.AutomaticEnv()

	serverAddr := v.GetString("CLIENT_SERVER_ADDR")
	if serverAddr == "" {
		serverAddr = os.Getenv("SK_CLIENT_SERVER_ADDR")
	}

	certPath := v.GetString("CLIENT_CERT_PATH")
	if certPath == "" {
		certPath = os.Getenv("SK_CLIENT_CERT_PATH")
	}

	// Если сертификат задан относительным путём, делаем его абсолютным от корня проекта
	if certPath != "" && !filepath.IsAbs(certPath) && !filepath.IsAbs(certPath) {
		// Проверяем, существует ли файл как есть
		if _, err := os.Stat(certPath); err != nil {
			// Если не существует, ищем от корня проекта
			workDir, _ := os.Getwd()
			potentialPath := filepath.Join(workDir, certPath)
			if _, err := os.Stat(potentialPath); err == nil {
				certPath = potentialPath
			}
		}
	}

	return &ClientConfig{
		ServerAddr: serverAddr,
		CertPath:   certPath,
	}, nil
}
