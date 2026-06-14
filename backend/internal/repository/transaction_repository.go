package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"

	"finagent/backend/internal/domain"
)

type TransactionRepository struct {
	db *sqlx.DB
}

func NewTransactionRepository(db *sqlx.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (r *TransactionRepository) Create(ctx context.Context, tx *domain.Transaction) error {
	const query = `
		INSERT INTO transactions (user_id, category_id, type, amount, description, transaction_date, is_ai_categorized)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at`

	err := r.db.QueryRowxContext(ctx, query,
		tx.UserID, tx.CategoryID, string(tx.Type),
		tx.Amount, tx.Description, tx.TransactionDate, tx.IsAICategorized,
	).Scan(&tx.ID, &tx.CreatedAt)
	if err != nil {
		return fmt.Errorf("repository: створення транзакції: %w", err)
	}
	return nil
}

// CreateImported додає транзакцію з зовнішнього джерела (Monobank) з дедуплікацією
// за (user_id, external_id). Повертає false, якщо запис уже існує.
func (r *TransactionRepository) CreateImported(ctx context.Context, tx *domain.Transaction) (bool, error) {
	const query = `
		INSERT INTO transactions
			(user_id, category_id, type, amount, description, transaction_date,
			 is_ai_categorized, source, external_id, currency_code, original_amount)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (user_id, external_id) WHERE external_id IS NOT NULL DO NOTHING
		RETURNING id, created_at`

	err := r.db.QueryRowxContext(ctx, query,
		tx.UserID, tx.CategoryID, string(tx.Type), tx.Amount, tx.Description, tx.TransactionDate,
		tx.IsAICategorized, tx.Source, tx.ExternalID, tx.CurrencyCode, tx.OriginalAmount,
	).Scan(&tx.ID, &tx.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil // дублікат — пропускаємо
	}
	if err != nil {
		return false, fmt.Errorf("repository: імпорт транзакції: %w", err)
	}
	return true, nil
}

func (r *TransactionRepository) GetByID(ctx context.Context, id int64) (*domain.Transaction, error) {
	const query = `
		SELECT id, user_id, category_id, type, amount, description,
		       transaction_date, is_ai_categorized, created_at,
		       source, external_id, currency_code, original_amount
		FROM transactions WHERE id = $1 AND deleted_at IS NULL`

	var tx domain.Transaction
	if err := r.db.GetContext(ctx, &tx, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("repository: пошук транзакції: %w", err)
	}
	return &tx, nil
}

func (r *TransactionRepository) Restore(ctx context.Context, id int64) error {
	if _, err := r.db.ExecContext(ctx, `UPDATE transactions SET deleted_at = NULL WHERE id = $1`, id); err != nil {
		return fmt.Errorf("repository: відновлення транзакції: %w", err)
	}
	return nil
}

func (r *TransactionRepository) List(ctx context.Context, userID int64, f domain.TransactionFilter) ([]domain.Transaction, int64, error) {
	cond := []string{"user_id = $1", "deleted_at IS NULL"}
	args := []any{userID}

	if f.From != nil {
		args = append(args, *f.From)
		cond = append(cond, fmt.Sprintf("transaction_date >= $%d", len(args)))
	}
	if f.To != nil {
		args = append(args, *f.To)
		cond = append(cond, fmt.Sprintf("transaction_date <= $%d", len(args)))
	}
	if f.CategoryID != nil {
		args = append(args, *f.CategoryID)
		cond = append(cond, fmt.Sprintf("category_id = $%d", len(args)))
	}
	if f.Type != nil {
		args = append(args, string(*f.Type))
		cond = append(cond, fmt.Sprintf("type = $%d", len(args)))
	}

	where := "WHERE " + strings.Join(cond, " AND ")

	var total int64
	if err := r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM transactions "+where, args...); err != nil {
		return nil, 0, fmt.Errorf("repository: підрахунок транзакцій: %w", err)
	}

	args = append(args, f.Limit, f.Offset)
	dataQuery := fmt.Sprintf(`
		SELECT id, user_id, category_id, type, amount, description,
		       transaction_date, is_ai_categorized, created_at,
		       source, external_id, currency_code, original_amount
		FROM transactions %s
		ORDER BY transaction_date DESC, id DESC
		LIMIT $%d OFFSET $%d`, where, len(args)-1, len(args))

	txs := make([]domain.Transaction, 0)
	if err := r.db.SelectContext(ctx, &txs, dataQuery, args...); err != nil {
		return nil, 0, fmt.Errorf("repository: список транзакцій: %w", err)
	}
	return txs, total, nil
}

func (r *TransactionRepository) Update(ctx context.Context, tx *domain.Transaction) error {
	const query = `
		UPDATE transactions
		SET category_id = $1, type = $2, amount = $3, description = $4, transaction_date = $5
		WHERE id = $6`

	if _, err := r.db.ExecContext(ctx, query,
		tx.CategoryID, string(tx.Type), tx.Amount, tx.Description, tx.TransactionDate, tx.ID,
	); err != nil {
		return fmt.Errorf("repository: оновлення транзакції: %w", err)
	}
	return nil
}

func (r *TransactionRepository) Delete(ctx context.Context, id int64) error {
	const query = `UPDATE transactions SET deleted_at = now() WHERE id = $1 AND deleted_at IS NULL`

	if _, err := r.db.ExecContext(ctx, query, id); err != nil {
		return fmt.Errorf("repository: видалення транзакції: %w", err)
	}
	return nil
}
