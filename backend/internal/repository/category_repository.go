package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"

	"finagent/backend/internal/domain"
)

type CategoryRepository struct {
	db *sqlx.DB
}

func NewCategoryRepository(db *sqlx.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) ListAvailable(ctx context.Context, userID int64) ([]domain.Category, error) {
	const query = `
		SELECT id, user_id, name, type, created_at
		FROM categories
		WHERE user_id IS NULL OR user_id = $1
		ORDER BY type, name`

	categories := make([]domain.Category, 0)
	if err := r.db.SelectContext(ctx, &categories, query, userID); err != nil {
		return nil, fmt.Errorf("repository: список категорій: %w", err)
	}
	return categories, nil
}

func (r *CategoryRepository) GetByID(ctx context.Context, id int64) (*domain.Category, error) {
	const query = `SELECT id, user_id, name, type, created_at FROM categories WHERE id = $1`

	var category domain.Category
	if err := r.db.GetContext(ctx, &category, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("repository: пошук категорії: %w", err)
	}
	return &category, nil
}

func (r *CategoryRepository) Create(ctx context.Context, category *domain.Category) error {
	const query = `
		INSERT INTO categories (user_id, name, type)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`

	err := r.db.QueryRowxContext(ctx, query, category.UserID, category.Name, string(category.Type)).
		Scan(&category.ID, &category.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
			return domain.ErrConflict
		}
		return fmt.Errorf("repository: створення категорії: %w", err)
	}
	return nil
}

func (r *CategoryRepository) Update(ctx context.Context, category *domain.Category) error {
	const query = `UPDATE categories SET name = $1 WHERE id = $2`

	if _, err := r.db.ExecContext(ctx, query, category.Name, category.ID); err != nil {
		return fmt.Errorf("repository: оновлення категорії: %w", err)
	}
	return nil
}

func (r *CategoryRepository) Delete(ctx context.Context, id int64) error {
	const query = `DELETE FROM categories WHERE id = $1`

	if _, err := r.db.ExecContext(ctx, query, id); err != nil {
		return fmt.Errorf("repository: видалення категорії: %w", err)
	}
	return nil
}
