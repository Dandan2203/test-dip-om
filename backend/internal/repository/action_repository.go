package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"

	"finagent/backend/internal/domain"
)

type ActionRepository struct {
	db *sqlx.DB
}

func NewActionRepository(db *sqlx.DB) *ActionRepository {
	return &ActionRepository{db: db}
}

func (r *ActionRepository) Record(ctx context.Context, a *domain.ActionLog) error {
	const query = `
		INSERT INTO action_log (user_id, action_type, entity_type, entity_id, payload)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at`

	payload := a.Payload
	if len(payload) == 0 {
		payload = []byte("{}")
	}
	err := r.db.QueryRowxContext(ctx, query,
		a.UserID, string(a.ActionType), string(a.EntityType), a.EntityID, payload,
	).Scan(&a.ID, &a.CreatedAt)
	if err != nil {
		return fmt.Errorf("repository: запис дії: %w", err)
	}
	return nil
}

func (r *ActionRepository) LastUndoable(ctx context.Context, userID int64) (*domain.ActionLog, error) {
	const query = `
		SELECT id, user_id, action_type, entity_type, entity_id, payload, undone, created_at
		FROM action_log
		WHERE user_id = $1 AND undone = FALSE
		ORDER BY created_at DESC, id DESC
		LIMIT 1`

	var a domain.ActionLog
	if err := r.db.GetContext(ctx, &a, query, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("repository: остання дія: %w", err)
	}
	return &a, nil
}

func (r *ActionRepository) MarkUndone(ctx context.Context, id int64) error {
	if _, err := r.db.ExecContext(ctx, `UPDATE action_log SET undone = TRUE WHERE id = $1`, id); err != nil {
		return fmt.Errorf("repository: позначення дії відміненою: %w", err)
	}
	return nil
}
