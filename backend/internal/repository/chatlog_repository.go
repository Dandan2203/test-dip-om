package repository

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"finagent/backend/internal/domain"
)

type ChatLogRepository struct {
	db *sqlx.DB
}

func NewChatLogRepository(db *sqlx.DB) *ChatLogRepository {
	return &ChatLogRepository{db: db}
}

func (r *ChatLogRepository) Record(ctx context.Context, l *domain.ChatLog) error {
	const query = `
		INSERT INTO ai_chat_logs (user_id, message, intent, response)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at`

	response := l.Response
	if len(response) == 0 {
		response = []byte("{}")
	}
	err := r.db.QueryRowxContext(ctx, query,
		l.UserID, l.Message, l.Intent, response,
	).Scan(&l.ID, &l.CreatedAt)
	if err != nil {
		return fmt.Errorf("repository: запис логу чату: %w", err)
	}
	return nil
}

// History повертає останні limit звернень користувача у хронологічному порядку.
func (r *ChatLogRepository) History(ctx context.Context, userID int64, limit int) ([]domain.ChatLog, error) {
	const query = `
		SELECT id, user_id, message, intent, response, created_at
		FROM (
			SELECT id, user_id, message, intent, response, created_at
			FROM ai_chat_logs
			WHERE user_id = $1
			ORDER BY id DESC
			LIMIT $2
		) recent
		ORDER BY id ASC`

	var logs []domain.ChatLog
	if err := r.db.SelectContext(ctx, &logs, query, userID, limit); err != nil {
		return nil, fmt.Errorf("repository: історія чату: %w", err)
	}
	return logs, nil
}

// IncrementDailyUsage атомарно інкрементує денний лічильник звернень і повертає
// нове значення. UPSERT за (user_id, day) виключає гонку між паралельними запитами.
func (r *ChatLogRepository) IncrementDailyUsage(ctx context.Context, userID int64) (int, error) {
	const query = `
		INSERT INTO chat_usage (user_id, day, count) VALUES ($1, current_date, 1)
		ON CONFLICT (user_id, day) DO UPDATE SET count = chat_usage.count + 1
		RETURNING count`
	var n int
	if err := r.db.GetContext(ctx, &n, query, userID); err != nil {
		return 0, fmt.Errorf("repository: облік звернень за день: %w", err)
	}
	return n, nil
}
