package domain

import (
	"context"
	"time"
)

type TransactionType string

const (
	TransactionIncome  TransactionType = "income"
	TransactionExpense TransactionType = "expense"
)

type Transaction struct {
	ID              int64           `db:"id"`
	UserID          int64           `db:"user_id"`
	CategoryID      *int64          `db:"category_id"`
	Type            TransactionType `db:"type"`
	Amount          float64         `db:"amount"`
	Description     string          `db:"description"`
	TransactionDate time.Time       `db:"transaction_date"`
	IsAICategorized bool            `db:"is_ai_categorized"`
	CreatedAt       time.Time       `db:"created_at"`
	Source          string          `db:"source"`
	ExternalID      *string         `db:"external_id"`
	CurrencyCode    int             `db:"currency_code"`
	OriginalAmount  *float64        `db:"original_amount"`
}

type TransactionFilter struct {
	From       *time.Time
	To         *time.Time
	CategoryID *int64
	Type       *TransactionType
	Limit      int
	Offset     int
}

type CreateTransactionInput struct {
	CategoryID      *int64
	Type            TransactionType
	Amount          float64
	Description     string
	TransactionDate time.Time
}

type UpdateTransactionInput struct {
	CategoryID      *int64
	Type            TransactionType
	Amount          float64
	Description     string
	TransactionDate time.Time
}

type TransactionRepository interface {
	Create(ctx context.Context, tx *Transaction) error
	CreateImported(ctx context.Context, tx *Transaction) (bool, error)
	GetByID(ctx context.Context, id int64) (*Transaction, error)
	List(ctx context.Context, userID int64, filter TransactionFilter) ([]Transaction, int64, error)
	Update(ctx context.Context, tx *Transaction) error
	Delete(ctx context.Context, id int64) error
	Restore(ctx context.Context, id int64) error
}

type TransactionUsecase interface {
	Create(ctx context.Context, userID int64, input CreateTransactionInput) (*Transaction, error)
	List(ctx context.Context, userID int64, filter TransactionFilter) ([]Transaction, int64, error)
	Update(ctx context.Context, userID, transactionID int64, input UpdateTransactionInput) (*Transaction, error)
	Delete(ctx context.Context, userID, transactionID int64) error
}
