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
