package postgres

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"secretKeeper/internal/server/domain"
	"time"
)

type dbFile struct {
	ID         string       `db:"id"`
	CreatedAt  time.Time    `db:"created_at"`
	PublicMeta dbPublicMeta `db:"public_meta"`
}

type dbPublicMeta struct {
	Type  string            `json:"type"`
	Extra map[string]string `json:"extra"`
}

// Scan реализует интерфейс sql.Scanner для чтения JSONB
func (m *dbPublicMeta) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("не удалось демаршалировать JSONB: %v", value)
	}
	return json.Unmarshal(bytes, m)
}

// Value реализует интерфейс driver.Valuer для записи JSONB
func (m dbPublicMeta) Value() (driver.Value, error) {
	return json.Marshal(m)
}

// ToDomain конвертирует модель БД в доменную сущность
func (dbf *dbFile) ToDomain() domain.File {
	meta := domain.FileMetaData{
		Name:        dbf.PublicMeta.Extra["name"],
		SecretType:  dbf.PublicMeta.Type,
		Mime:        dbf.PublicMeta.Extra["mime"],
		Extension:   dbf.PublicMeta.Extra["ext"],
		Description: dbf.PublicMeta.Extra["description"],
		Extra:       dbf.PublicMeta.Extra,
	}

	return domain.File{
		ID:        dbf.ID,
		CreatedAt: dbf.CreatedAt,
		Meta:      meta,
	}
}
