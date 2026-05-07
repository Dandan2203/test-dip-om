package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"

	"finagent/backend/internal/domain"
)

type GoalRepository struct {
	db *sqlx.DB
}

func NewGoalRepository(db *sqlx.DB) *GoalRepository {
	return &GoalRepository{db: db}
}

func (r *GoalRepository) Create(ctx context.Context, goal *domain.Goal) error {
	const query = `
		INSERT INTO goals (user_id, title, target_amount, deadline)
		VALUES ($1, $2, $3, $4)
		RETURNING id, current_amount, created_at`

	err := r.db.QueryRowxContext(ctx, query,
		goal.UserID, goal.Title, goal.TargetAmount, goal.Deadline,
	).Scan(&goal.ID, &goal.CurrentAmount, &goal.CreatedAt)
	if err != nil {
		return fmt.Errorf("repository: створення цілі: %w", err)
	}
	return nil
}

func (r *GoalRepository) GetByID(ctx context.Context, id int64) (*domain.Goal, error) {
	const query = `
		SELECT id, user_id, title, target_amount, current_amount, deadline, created_at
		FROM goals WHERE id = $1 AND deleted_at IS NULL`

	var g domain.Goal
	if err := r.db.GetContext(ctx, &g, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("repository: пошук цілі: %w", err)
	}
	return &g, nil
}

func (r *GoalRepository) List(ctx context.Context, userID int64) ([]domain.Goal, error) {
	const query = `
		SELECT id, user_id, title, target_amount, current_amount, deadline, created_at
		FROM goals WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC`

	goals := make([]domain.Goal, 0)
	if err := r.db.SelectContext(ctx, &goals, query, userID); err != nil {
		return nil, fmt.Errorf("repository: список цілей: %w", err)
	}
	return goals, nil
}

func (r *GoalRepository) Update(ctx context.Context, goal *domain.Goal) error {
	const query = `
		UPDATE goals SET title = $1, target_amount = $2, deadline = $3
		WHERE id = $4`

	if _, err := r.db.ExecContext(ctx, query, goal.Title, goal.TargetAmount, goal.Deadline, goal.ID); err != nil {
		return fmt.Errorf("repository: оновлення цілі: %w", err)
	}
	return nil
}

func (r *GoalRepository) Contribute(ctx context.Context, id int64, amount float64) error {
	const query = `UPDATE goals SET current_amount = GREATEST(0, current_amount + $1) WHERE id = $2`

	if _, err := r.db.ExecContext(ctx, query, amount, id); err != nil {
		return fmt.Errorf("repository: внесок до цілі: %w", err)
	}
	return nil
}

func (r *GoalRepository) Delete(ctx context.Context, id int64) error {
	const query = `UPDATE goals SET deleted_at = now() WHERE id = $1 AND deleted_at IS NULL`

	if _, err := r.db.ExecContext(ctx, query, id); err != nil {
		return fmt.Errorf("repository: видалення цілі: %w", err)
	}
	return nil
}

func (r *GoalRepository) Restore(ctx context.Context, id int64) error {
	if _, err := r.db.ExecContext(ctx, `UPDATE goals SET deleted_at = NULL WHERE id = $1`, id); err != nil {
		return fmt.Errorf("repository: відновлення цілі: %w", err)
	}
	return nil
}
