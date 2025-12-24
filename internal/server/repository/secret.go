package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	pb "secretKeeper/internal/server/transport/grpc/proto"
)

type SecretRepo struct {
	db *sql.DB
}

func NewSecretRepo(db *sql.DB) *SecretRepo {
	return &SecretRepo{db: db}
}

func (s *SecretRepo) SaveSecret(ctx context.Context, userID, secretID string, secretType int32, title string, meta *pb.SecretMetadata, storagePath string) error {
	metaJSON, err := json.Marshal(meta)
	if err != nil {
		return fmt.Errorf("failed to marshal meta: %w", err)
	}

	query := `
		INSERT INTO secrets (id, user_id, type, title, public_meta, storage_path, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
	`

	_, err = s.db.ExecContext(ctx, query, secretID, userID, secretType, title, metaJSON, storagePath)
	if err != nil {
		return fmt.Errorf("failed to save secret: %w", err)
	}

	return nil
}

func (s *SecretRepo) CheckSecretExists(transactionID string) (bool, error) {
	var exists bool

	query := `SELECT EXISTS(SELECT 1 FROM secrets WHERE id = $1)`

	err := s.db.QueryRowContext(context.Background(), query, transactionID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check secret existence: %w", err)
	}

	return exists, nil
}

func (s *SecretRepo) GetSecret(ctx context.Context, userID, secretID string) (string, error) {
	return "", nil
}

func (s *SecretRepo) DeleteSecret(ctx context.Context, userID, secretID string) error {
	return nil
}
