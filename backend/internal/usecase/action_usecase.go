package usecase

import (
	"context"

	"finagent/backend/internal/domain"
)

type ActionUsecase struct {
	actions domain.ActionRepository
	txRepo  domain.TransactionRepository
	goals   domain.GoalRepository
}

func NewActionUsecase(actions domain.ActionRepository, txRepo domain.TransactionRepository, goals domain.GoalRepository) *ActionUsecase {
	return &ActionUsecase{actions: actions, txRepo: txRepo, goals: goals}
}

// Undo відкочує останню невідмінену дію користувача:
// create → м'яко видаляє сутність, delete → відновлює її.
func (uc *ActionUsecase) Undo(ctx context.Context, userID int64) (*domain.ActionLog, error) {
	last, err := uc.actions.LastUndoable(ctx, userID)
	if err != nil {
		return nil, err
	}

	if err := uc.reverse(ctx, last); err != nil {
		return nil, err
	}
	if err := uc.actions.MarkUndone(ctx, last.ID); err != nil {
		return nil, err
	}
	return last, nil
}

func (uc *ActionUsecase) reverse(ctx context.Context, a *domain.ActionLog) error {
	switch a.EntityType {
	case domain.EntityTransaction:
		if a.ActionType == domain.ActionCreate {
			return uc.txRepo.Delete(ctx, a.EntityID)
		}
		return uc.txRepo.Restore(ctx, a.EntityID)
	case domain.EntityGoal:
		if a.ActionType == domain.ActionCreate {
			return uc.goals.Delete(ctx, a.EntityID)
		}
		return uc.goals.Restore(ctx, a.EntityID)
	}
	return nil
}
