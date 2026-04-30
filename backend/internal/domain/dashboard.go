package domain

import (
	"context"
	"encoding/json"
	"time"
)

type DashboardLayout struct {
	UserID    int64           `db:"user_id"`
	Name      string          `db:"name"`
	Layout    json.RawMessage `db:"layout"`
	UpdatedAt time.Time       `db:"updated_at"`
}

type DashboardRepository interface {
	Get(ctx context.Context, userID int64, name string) (*DashboardLayout, error)
	Upsert(ctx context.Context, userID int64, name string, layout json.RawMessage) error
}

type DashboardUsecase interface {
	Get(ctx context.Context, userID int64, name string) (json.RawMessage, error)
	Save(ctx context.Context, userID int64, name string, layout json.RawMessage) error
}
