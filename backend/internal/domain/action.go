package domain

import (
	"context"
	"encoding/json"
	"time"
)

type ActionType string

const (
	ActionCreate ActionType = "create"
	ActionDelete ActionType = "delete"
)

type EntityType string

const (
	EntityTransaction EntityType = "transaction"
	EntityGoal        EntityType = "goal"
)

// ActionLog — запис дії агента для функції «Відмінити останню дію».
type ActionLog struct {
	ID         int64           `db:"id"`
	UserID     int64           `db:"user_id"`
	ActionType ActionType      `db:"action_type"`
	EntityType EntityType      `db:"entity_type"`
	EntityID   int64           `db:"entity_id"`
	Payload    json.RawMessage `db:"payload"`
	Undone     bool            `db:"undone"`
	CreatedAt  time.Time       `db:"created_at"`
}

type ActionRepository interface {
	Record(ctx context.Context, a *ActionLog) error
	LastUndoable(ctx context.Context, userID int64) (*ActionLog, error)
	MarkUndone(ctx context.Context, id int64) error
}

type ActionUsecase interface {
	Undo(ctx context.Context, userID int64) (*ActionLog, error)
}
