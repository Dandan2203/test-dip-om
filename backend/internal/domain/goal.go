package domain

import (
	"context"
	"time"
)

type Goal struct {
	ID            int64      `db:"id"`
	UserID        int64      `db:"user_id"`
	Title         string     `db:"title"`
	TargetAmount  float64    `db:"target_amount"`
	CurrentAmount float64    `db:"current_amount"`
	Deadline      *time.Time `db:"deadline"`
	CreatedAt     time.Time  `db:"created_at"`
}

type CreateGoalInput struct {
	Title        string
	TargetAmount float64
	Deadline     *time.Time
}

type UpdateGoalInput struct {
	Title        string
	TargetAmount float64
	Deadline     *time.Time
}

type GoalRepository interface {
	Create(ctx context.Context, goal *Goal) error
	GetByID(ctx context.Context, id int64) (*Goal, error)
	List(ctx context.Context, userID int64) ([]Goal, error)
	Update(ctx context.Context, goal *Goal) error
	Contribute(ctx context.Context, id int64, amount float64) error
	Delete(ctx context.Context, id int64) error
	Restore(ctx context.Context, id int64) error
}

type GoalUsecase interface {
	Create(ctx context.Context, userID int64, input CreateGoalInput) (*Goal, error)
	List(ctx context.Context, userID int64) ([]Goal, error)
	Update(ctx context.Context, userID, goalID int64, input UpdateGoalInput) (*Goal, error)
	Contribute(ctx context.Context, userID, goalID int64, amount float64) (*Goal, error)
	Delete(ctx context.Context, userID, goalID int64) error
}
