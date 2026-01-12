package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// AuthRepo репозиторий авторизации PostgreSQL
type AuthRepo struct {
	db *sql.DB
}

// NewAuthRepo создает новый AuthRepo
func NewAuthRepo(db *sql.DB) *AuthRepo {
	return &AuthRepo{db: db}
}

// CreateUser создает пользователя
func (r *AuthRepo) CreateUser(ctx context.Context, username, hashedPassword string) (string, error) {
	var id string

	query := `
		INSERT INTO users (username, password_hash) 
		VALUES ($1, $2)
		RETURNING id
	`

	err := r.db.QueryRowContext(ctx, query, username, hashedPassword).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return "", errors.New("пользователь уже существует")
			}
		}

		return "", fmt.Errorf("не удалось зарегистрировать пользователя: %w", err)
	}

	return id, nil
}

// LoginUser ищет пользователя для входа
func (r *AuthRepo) LoginUser(ctx context.Context, username string) (string, string, error) {
	var id string
	var hashedPassword string

	query := `
        SELECT id, password_hash
        FROM users
        WHERE username = $1
    `

	err := r.db.QueryRowContext(ctx, query, username).Scan(&id, &hashedPassword)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", "", errors.New("пользователь не найден")
		}
		return "", "", fmt.Errorf("не удалось получить пользователя по имени: %w", err)
	}

	return id, hashedPassword, nil
}

// CreateSession создает сессию обновления токена
func (r *AuthRepo) CreateSession(ctx context.Context, id, refreshTokenHash string, expiresIn time.Duration) error {
	query := `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`

	expiresAt := time.Now().Add(expiresIn)

	_, err := r.db.ExecContext(ctx, query, id, refreshTokenHash, expiresAt)
	if err != nil {
		return fmt.Errorf("не удалось создать сессию: %w", err)
	}

	return nil
}
