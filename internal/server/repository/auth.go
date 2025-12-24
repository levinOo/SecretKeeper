package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrUserNotFound      = errors.New("user not found")
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (u *UserRepository) RegisterUser(ctx context.Context, username, passwordHash string) (string, error) {
	var id string

	query := `
		INSERT INTO users (username, password_hash) 
		VALUES ($1, $2)
		RETURNING id
	`

	err := u.db.QueryRowContext(ctx, query, username, passwordHash).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return "", ErrUserAlreadyExists
			}
		}

		return "", fmt.Errorf("failed to register user: %w", err)
	}

	return id, nil
}

// Функция получения ID и хеша пароля пользователя по его имени (username)
func (u *UserRepository) LoginUser(ctx context.Context, username string) (string, string, error) {
	var id string
	var hashedPassword string

	query := `
        SELECT id, password_hash
        FROM users
        WHERE username = $1
    `

	err := u.db.QueryRowContext(ctx, query, username).Scan(&id, &hashedPassword)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", "", ErrUserNotFound
		}
		return "", "", fmt.Errorf("failed to get user by username: %w", err)
	}

	return id, hashedPassword, nil
}

// Функция добавления refresh токена пользователю
func (u *UserRepository) AddToken(ctx context.Context, userID, refreshToken string) error {
	query := `
		UPDATE users 
		SET refresh_token = $1 
		WHERE id = $2
	`

	_, err := u.db.ExecContext(ctx, query, refreshToken, userID)
	if err != nil {
		return fmt.Errorf("failed to add token: %w", err)
	}

	return nil
}
