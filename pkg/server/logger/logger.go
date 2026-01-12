package logger

import "go.uber.org/zap"

// Функция создания нового логгера
func NewLogger() (*zap.SugaredLogger, error) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}

	return logger.Sugar(), nil
}
