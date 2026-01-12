package domain

import "io"

// SecretType тип секрета
type SecretType string

const (
	SecretTypeCredentials SecretType = "credentials" // логин/пароль
	SecretTypeFile        SecretType = "file"        // файл
)

// Secret сущность секрета
type Secret struct {
	Meta SecretMeta
	Data io.ReadCloser
}

// SecretMeta метаданные секрета
type SecretMeta struct {
	ID    string            `json:"id"`
	Type  SecretType        `json:"type"`
	Extra map[string]string `json:"extra"`
}
