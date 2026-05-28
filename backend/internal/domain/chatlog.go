package domain

import (
	"context"
	"encoding/json"
	"time"
)

type ChatLog struct {
	ID        int64           `db:"id"`
	UserID    int64           `db:"user_id"`
	Message   string          `db:"message"`
	Intent    string          `db:"intent"`
	Response  json.RawMessage `db:"response"`
	CreatedAt time.Time       `db:"created_at"`
}

type ChatLogRepository interface {
	Record(ctx context.Context, l *ChatLog) error
	History(ctx context.Context, userID int64, limit int) ([]ChatLog, error)
	IncrementDailyUsage(ctx context.Context, userID int64) (int, error)
}
