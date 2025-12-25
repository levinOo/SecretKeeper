package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type AuthRepo struct {
	db *sql.DB
}

func NewAuthRepo(db *sql.DB) *AuthRepo {
	return &AuthRepo{db: db}
}

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
				return "", errors.New("user already exists")
			}
		}

		return "", fmt.Errorf("failed to register user: %w", err)
	}

	return id, nil
}

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
			return "", "", errors.New("user not found")
		}
		return "", "", fmt.Errorf("failed to get user by username: %w", err)
	}

	return id, hashedPassword, nil
}
