package usecase

import (
	"context"
	"encoding/json"
	"fmt"

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

// Undo AI-дії
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
		switch a.ActionType {
		case domain.ActionCreate:
			return uc.goals.Delete(ctx, a.EntityID)
		case domain.ActionContribute:
			var p struct {
				Delta float64 `json:"delta"`
			}
			if err := json.Unmarshal(a.Payload, &p); err != nil {
				return fmt.Errorf("розбір журналу дії: %w", err)
			}
			return uc.goals.Contribute(ctx, a.EntityID, -p.Delta)
		default:
			return uc.goals.Restore(ctx, a.EntityID)
		}
	}
	return nil
}
