package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	pb "secretKeeper/internal/proto"
	"secretKeeper/internal/server/domain"
)

// UserRepo репозиторий пользователя PostgreSQL
type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

// SaveSecret сохраняет метаданные секрета
func (r *UserRepo) SaveSecret(ctx context.Context, userID, secretID, idempotencyKey string, meta *pb.SecretMetadata) error {
	metaJSON, err := json.Marshal(meta)
	if err != nil {
		return fmt.Errorf("не удалось маршалировать метаданные: %w", err)
	}

	query := `
        INSERT INTO user_files (
            minio_object_id, 
            user_id, 
            public_meta, 
            idempotency_key
        )
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (user_id, idempotency_key) DO NOTHING
    `

	_, err = r.db.ExecContext(ctx, query,
		secretID,
		userID,
		string(metaJSON),
		idempotencyKey,
	)

	return err
}

// CheckSecretExists проверяет существование секрета по ключу идемпотентности
func (r *UserRepo) CheckSecretExists(transactionID, userID string) (bool, error) {
	var exists bool

	query := `SELECT EXISTS(SELECT 1 FROM user_files WHERE idempotency_key = $1 AND user_id = $2)`

	err := r.db.QueryRowContext(context.Background(), query, transactionID, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("не удалось проверить существование секрета: %w", err)
	}

	return exists, nil
}

// ListUserFiles возвращает список файлов пользователя
func (r *UserRepo) ListUserFiles(ctx context.Context, userID string) ([]domain.File, error) {
	query := `
		SELECT id, public_meta, created_at
		FROM user_files
		WHERE user_id = $1
	`

	rows, err := r.db.QueryContext(context.Background(), query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	domainFiles := make([]domain.File, 0)

	for rows.Next() {
		var dbf dbFile

		if err := rows.Scan(
			&dbf.ID,
			&dbf.PublicMeta,
			&dbf.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("не удалось сканировать строку: %w", err)
		}

		domainFiles = append(domainFiles, dbf.ToDomain())
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка итерации строк: %w", err)
	}

	return domainFiles, nil
}

// FileExists проверяет существование файла по ID
func (r *UserRepo) FileExists(ctx context.Context, userID, secretID string) (bool, error) {
	var exists bool

	query := `SELECT EXISTS(SELECT 1 FROM user_files WHERE id = $1 AND user_id = $2)`

	err := r.db.QueryRowContext(context.Background(), query, secretID, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("не удалось проверить существование файла: %w", err)
	}

	return exists, nil
}

// DeleteSecret удаляет секрет из БД
func (r *UserRepo) DeleteSecret(ctx context.Context, userID, secretID string) error {
	query := `
	DELETE FROM user_files
		WHERE user_id = $1 AND id = $2
	`

	_, err := r.db.ExecContext(ctx, query, userID, secretID)

	return err
}

// GetSecretMeta получает метаданные секрета
func (r *UserRepo) GetSecretMeta(ctx context.Context, userID, secretID string) (*pb.SecretMetadata, error) {
	var metaJSON string

	query := `
		SELECT PUBLIC_META
		FROM user_files
		WHERE user_id = $1 AND id = $2
	`

	err := r.db.QueryRowContext(ctx, query, userID, secretID).Scan(&metaJSON)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить метаданные секрета: %w", err)
	}

	var meta pb.SecretMetadata
	if err := json.Unmarshal([]byte(metaJSON), &meta); err != nil {
		return nil, fmt.Errorf("не удалось демаршалировать метаданные секрета: %w", err)
	}

	return &meta, nil
}
