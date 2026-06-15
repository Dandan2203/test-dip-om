package domain

import (
	"context"
	"encoding/json"
	"time"
)

// ChatLog — запис одного звернення до ШІ-агента: текст користувача,
// класифікований інтент та відповідь моделі (JSONB).
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
	// IncrementDailyUsage атомарно збільшує лічильник звернень користувача за поточну
	// добу і повертає НОВЕ значення — для надійного денного ліміту без гонок.
	IncrementDailyUsage(ctx context.Context, userID int64) (int, error)
}
