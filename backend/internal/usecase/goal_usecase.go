package usecase

import (
	"context"
	"strings"

	"finagent/backend/internal/domain"
)

type GoalUsecase struct {
	repo domain.GoalRepository
}

func NewGoalUsecase(repo domain.GoalRepository) *GoalUsecase {
	return &GoalUsecase{repo: repo}
}

func (uc *GoalUsecase) Create(ctx context.Context, userID int64, input domain.CreateGoalInput) (*domain.Goal, error) {
	goal := &domain.Goal{
		UserID:       userID,
		Title:        strings.TrimSpace(input.Title),
		TargetAmount: input.TargetAmount,
		Deadline:     input.Deadline,
	}
	if err := uc.repo.Create(ctx, goal); err != nil {
		return nil, err
	}
	return goal, nil
}

func (uc *GoalUsecase) List(ctx context.Context, userID int64) ([]domain.Goal, error) {
	return uc.repo.List(ctx, userID)
}

func (uc *GoalUsecase) Update(ctx context.Context, userID, goalID int64, input domain.UpdateGoalInput) (*domain.Goal, error) {
	goal, err := uc.repo.GetByID(ctx, goalID)
	if err != nil {
		return nil, err
	}
	if goal.UserID != userID {
		return nil, domain.ErrForbidden
	}

	goal.Title = strings.TrimSpace(input.Title)
	goal.TargetAmount = input.TargetAmount
	goal.Deadline = input.Deadline

	if err := uc.repo.Update(ctx, goal); err != nil {
		return nil, err
	}
	return goal, nil
}

func (uc *GoalUsecase) Contribute(ctx context.Context, userID, goalID int64, amount float64) (*domain.Goal, error) {
	goal, err := uc.repo.GetByID(ctx, goalID)
	if err != nil {
		return nil, err
	}
	if goal.UserID != userID {
		return nil, domain.ErrForbidden
	}
	if err := uc.repo.Contribute(ctx, goalID, amount); err != nil {
		return nil, err
	}
	goal.CurrentAmount += amount
	return goal, nil
}

func (uc *GoalUsecase) Delete(ctx context.Context, userID, goalID int64) error {
	goal, err := uc.repo.GetByID(ctx, goalID)
	if err != nil {
		return err
	}
	if goal.UserID != userID {
		return domain.ErrForbidden
	}
	return uc.repo.Delete(ctx, goalID)
}
