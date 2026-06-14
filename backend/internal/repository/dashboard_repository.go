package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"

	"finagent/backend/internal/domain"
)

type DashboardRepository struct {
	db *sqlx.DB
}

func NewDashboardRepository(db *sqlx.DB) *DashboardRepository {
	return &DashboardRepository{db: db}
}

func (r *DashboardRepository) Get(ctx context.Context, userID int64, name string) (*domain.DashboardLayout, error) {
	const query = `
		SELECT user_id, name, layout, updated_at
		FROM dashboard_layouts WHERE user_id = $1 AND name = $2`

	var d domain.DashboardLayout
	if err := r.db.GetContext(ctx, &d, query, userID, name); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("repository: розкладка дашборда: %w", err)
	}
	return &d, nil
}

func (r *DashboardRepository) Upsert(ctx context.Context, userID int64, name string, layout json.RawMessage) error {
	const query = `
		INSERT INTO dashboard_layouts (user_id, name, layout)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, name) DO UPDATE
		SET layout = EXCLUDED.layout, updated_at = now()`

	if _, err := r.db.ExecContext(ctx, query, userID, name, []byte(layout)); err != nil {
		return fmt.Errorf("repository: збереження розкладки: %w", err)
	}
	return nil
}
