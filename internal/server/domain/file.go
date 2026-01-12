package domain

import "time"

// Структур загруженного файла
type File struct {
	ID        string
	CreatedAt time.Time
	Meta      FileMetaData
}

// Структура метаданных файла
type FileMetaData struct {
	Name        string
	SecretType  string
	Mime        string
	Size        int64
	Extension   string
	Description string
	Extra       map[string]string
}
