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
}
