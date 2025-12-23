package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"secretKeeper/internal/server/config"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	maxTimeout = 5 * time.Second // timeout подключения для бд
)

var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrUserNotFound      = errors.New("user not found")
)

// Структура обертки над подключением к БД
type DB struct {
	DB *sql.DB
}

// Функция создания нового подключения к базе данных PostgreSQL
func NewDBConnection(cfg *config.PostgreConfig) (*DB, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.SSLMode)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), maxTimeout)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	return &DB{DB: db}, nil
}

// Функция закрытия подключения к базе данных
func (db *DB) Close() error {
	return db.DB.Close()
}

// Функция проверки работоспособности подключения к базе данных (Health Check)
func (db *DB) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	if err := db.DB.PingContext(ctx); err != nil {
		return fmt.Errorf("db ping failed: %w", err)
	}
	return nil
}

// Функция регистрации нового пользователя в базе данных
func (db *DB) RegisterUser(ctx context.Context, username, passwordHash string) (string, error) {
	var id string

	query := `
		INSERT INTO users (username, password_hash) 
		VALUES ($1, $2)
		RETURNING id
	`

	err := db.DB.QueryRowContext(ctx, query, username, passwordHash).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.Is(err, pgErr) {
			if pgErr.Code == "23505" {
				return "", ErrUserAlreadyExists
			}
		}

		return "", fmt.Errorf("failed to register user: %w", err)
	}

	return id, nil
}

// Функция получения ID и хеша пароля пользователя по его имени (username)
func (db *DB) GetUserByUsername(ctx context.Context, username string) (string, string, error) {
	var id string
	var hashedPassword string

	query := `
        SELECT id, password_hash
        FROM users
        WHERE username = $1
    `

	err := db.DB.QueryRowContext(ctx, query, username).Scan(&id, &hashedPassword)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", "", ErrUserNotFound
		}
		return "", "", fmt.Errorf("failed to get user by username: %w", err)
	}

	return id, hashedPassword, nil
}

// Функция обновления refresh токена для указанного пользователя
func (db *DB) AddToken(ctx context.Context, userID, refreshToken string) error {
	const query = `
        UPDATE users
        SET refresh_token = $2
        WHERE id = $1
    `

	_, err := db.DB.ExecContext(ctx, query, userID, refreshToken)
	if err != nil {
		return fmt.Errorf("failed to update refresh token: %w", err)
	}

	return nil
}

// Функция сохранения метаданных секрета в базе данных
func (db *DB) SaveSecret(ctx context.Context, userID, secretID string, secretType int, title string, meta map[string]string, storagePath string) error {
	metaJSON, err := json.Marshal(meta)
	if err != nil {
		return fmt.Errorf("failed to marshal meta: %w", err)
	}

	query := `
		INSERT INTO secrets (id, user_id, type, title, public_meta, storage_path, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
	`

	_, err = db.DB.ExecContext(ctx, query, secretID, userID, secretType, title, metaJSON, storagePath)
	if err != nil {
		return fmt.Errorf("failed to save secret: %w", err)
	}

	return nil
}

// Функция проверки существования секрета по его ID
func (db *DB) CheckSecretExists(secretID string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM secrets WHERE id = $1)`
	err := db.DB.QueryRowContext(context.Background(), query, secretID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check secret existence: %w", err)
	}
	return exists, nil
}
