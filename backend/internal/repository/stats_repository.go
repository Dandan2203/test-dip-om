package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"

	"finagent/backend/internal/domain"
)

type StatsRepository struct {
	db *sqlx.DB
}

func NewStatsRepository(db *sqlx.DB) *StatsRepository {
	return &StatsRepository{db: db}
}

func (r *StatsRepository) Summary(ctx context.Context, userID int64, p domain.StatsPeriod) (domain.StatsSummary, error) {
	const query = `
		SELECT
			COALESCE(SUM(CASE WHEN type = 'income'  THEN amount ELSE 0 END), 0) AS income,
			COALESCE(SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END), 0) AS expense
		FROM transactions
		WHERE user_id = $1
		  AND deleted_at IS NULL
		  AND transaction_date >= $2
		  AND transaction_date <= $3`

	var s domain.StatsSummary
	if err := r.db.GetContext(ctx, &s, query, userID, p.From, p.To); err != nil {
		return s, fmt.Errorf("repository: зведення статистики: %w", err)
	}
	s.Balance = s.Income - s.Expense
	return s, nil
}

func (r *StatsRepository) ByCategory(ctx context.Context, userID int64, txType domain.TransactionType, p domain.StatsPeriod) ([]domain.CategoryStat, error) {
	const query = `
		SELECT category_id, SUM(amount) AS total, COUNT(*) AS count
		FROM transactions
		WHERE user_id = $1
		  AND deleted_at IS NULL
		  AND type = $2
		  AND transaction_date >= $3
		  AND transaction_date <= $4
		GROUP BY category_id
		ORDER BY total DESC`

	stats := make([]domain.CategoryStat, 0)
	if err := r.db.SelectContext(ctx, &stats, query, userID, string(txType), p.From, p.To); err != nil {
		return nil, fmt.Errorf("repository: статистика за категоріями: %w", err)
	}
	return stats, nil
}

func (r *StatsRepository) RecentTransactions(ctx context.Context, userID int64, limit int) ([]domain.Transaction, error) {
	const query = `
		SELECT id, user_id, category_id, type, amount, description,
		       transaction_date, is_ai_categorized, created_at,
		       source, external_id, currency_code, original_amount
		FROM transactions
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY transaction_date DESC, id DESC
		LIMIT $2`

	txs := make([]domain.Transaction, 0)
	if err := r.db.SelectContext(ctx, &txs, query, userID, limit); err != nil {
		return nil, fmt.Errorf("repository: останні транзакції: %w", err)
	}
	return txs, nil
}

func (r *StatsRepository) ExportRows(ctx context.Context, userID int64, f domain.TransactionFilter) ([]domain.TransactionExportRow, error) {
	cond := []string{"t.user_id = $1"}
	args := []any{userID}

	if f.From != nil {
		args = append(args, *f.From)
		cond = append(cond, fmt.Sprintf("t.transaction_date >= $%d", len(args)))
	}
	if f.To != nil {
		args = append(args, *f.To)
		cond = append(cond, fmt.Sprintf("t.transaction_date <= $%d", len(args)))
	}
	if f.Type != nil {
		args = append(args, string(*f.Type))
		cond = append(cond, fmt.Sprintf("t.type = $%d", len(args)))
	}

	query := fmt.Sprintf(`
		SELECT t.transaction_date, t.type, COALESCE(c.name, '') AS category_name, t.amount, t.description
		FROM transactions t
		LEFT JOIN categories c ON t.category_id = c.id
		WHERE %s
		ORDER BY t.transaction_date DESC, t.id DESC`, strings.Join(cond, " AND "))

	rows := make([]domain.TransactionExportRow, 0)
	if err := r.db.SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, fmt.Errorf("repository: експорт транзакцій: %w", err)
	}
	return rows, nil
}
