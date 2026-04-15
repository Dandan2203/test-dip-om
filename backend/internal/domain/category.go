package domain

import (
	"context"
	"time"
)

type CategoryType string

const (
	CategoryIncome  CategoryType = "income"
	CategoryExpense CategoryType = "expense"
)

// UserID nil системна.
type Category struct {
	ID        int64        `db:"id"`
	UserID    *int64       `db:"user_id"`
	Name      string       `db:"name"`
	Type      CategoryType `db:"type"`
	CreatedAt time.Time    `db:"created_at"`
}

func (c *Category) IsSystem() bool {
	return c.UserID == nil
}

type CategoryRepository interface {
	ListAvailable(ctx context.Context, userID int64) ([]Category, error)
	GetByID(ctx context.Context, id int64) (*Category, error)
	Create(ctx context.Context, category *Category) error
	Update(ctx context.Context, category *Category) error
	Delete(ctx context.Context, id int64) error
}

type CategoryUsecase interface {
	List(ctx context.Context, userID int64) ([]Category, error)
	Create(ctx context.Context, userID int64, name string, catType CategoryType) (*Category, error)
	Update(ctx context.Context, userID, categoryID int64, name string) (*Category, error)
	Delete(ctx context.Context, userID, categoryID int64) error
}
