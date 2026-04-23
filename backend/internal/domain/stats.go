package domain

import (
	"context"
	"time"
)

type StatsSummary struct {
	Income  float64 `db:"income"`
	Expense float64 `db:"expense"`
	Balance float64
}

type CategoryStat struct {
	CategoryID *int64  `db:"category_id" json:"categoryId"`
	Total      float64 `db:"total" json:"total"`
	Count      int     `db:"count" json:"count"`
}

type DashboardData struct {
	Income             float64
	Expense            float64
	Balance            float64
	RecentTransactions []Transaction
	Goals              []Goal
}

type StatsPeriod struct {
	From time.Time
	To   time.Time
}

type TransactionExportRow struct {
	TransactionDate time.Time `db:"transaction_date"`
	Type            string    `db:"type"`
	CategoryName    string    `db:"category_name"`
	Amount          float64   `db:"amount"`
	Description     string    `db:"description"`
}

type StatsRepository interface {
	Summary(ctx context.Context, userID int64, period StatsPeriod) (StatsSummary, error)
	ByCategory(ctx context.Context, userID int64, txType TransactionType, period StatsPeriod) ([]CategoryStat, error)
	RecentTransactions(ctx context.Context, userID int64, limit int) ([]Transaction, error)
	ExportRows(ctx context.Context, userID int64, filter TransactionFilter) ([]TransactionExportRow, error)
}

type StatsUsecase interface {
	Summary(ctx context.Context, userID int64, period StatsPeriod) (StatsSummary, error)
	ByCategory(ctx context.Context, userID int64, txType TransactionType, period StatsPeriod) ([]CategoryStat, error)
	Dashboard(ctx context.Context, userID int64, goalUC GoalUsecase) (DashboardData, error)
	Export(ctx context.Context, userID int64, filter TransactionFilter) ([]TransactionExportRow, error)
}
