package database

import (
	"database/sql"
	"fmt"
	"secretKeeper/internal/server/migrations"

	"github.com/pressly/goose/v3"
)

func RunMigrations(db *sql.DB) error {
	goose.SetBaseFS(migrations.FS)
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("не удалось установить диалект: %w", err)
	}

	if err := goose.Up(db, "."); err != nil {
		return fmt.Errorf("не удалось выполнить миграции: %w", err)
	}

	return nil
}
