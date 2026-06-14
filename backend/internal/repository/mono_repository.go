package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"

	"finagent/backend/internal/domain"
)

type MonoRepository struct {
	db *sqlx.DB
}

func NewMonoRepository(db *sqlx.DB) *MonoRepository {
	return &MonoRepository{db: db}
}

func (r *MonoRepository) Save(ctx context.Context, userID int64, tokenEncrypted string) error {
	const query = `
		INSERT INTO mono_connections (user_id, token_encrypted)
		VALUES ($1, $2)
		ON CONFLICT (user_id) DO UPDATE
		SET token_encrypted = EXCLUDED.token_encrypted, connected_at = now()`

	if _, err := r.db.ExecContext(ctx, query, userID, tokenEncrypted); err != nil {
		return fmt.Errorf("repository: збереження підключення Mono: %w", err)
	}
	return nil
}

func (r *MonoRepository) Get(ctx context.Context, userID int64) (*domain.MonoConnection, error) {
	const query = `
		SELECT user_id, token_encrypted, connected_at, last_import_at
		FROM mono_connections WHERE user_id = $1`

	var conn domain.MonoConnection
	if err := r.db.GetContext(ctx, &conn, query, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("repository: підключення Mono: %w", err)
	}
	return &conn, nil
}

func (r *MonoRepository) Delete(ctx context.Context, userID int64) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM mono_connections WHERE user_id = $1`, userID); err != nil {
		return fmt.Errorf("repository: відключення Mono: %w", err)
	}
	return nil
}

func (r *MonoRepository) TouchImport(ctx context.Context, userID int64) error {
	if _, err := r.db.ExecContext(ctx, `UPDATE mono_connections SET last_import_at = now() WHERE user_id = $1`, userID); err != nil {
		return fmt.Errorf("repository: оновлення часу імпорту: %w", err)
	}
	return nil
}
