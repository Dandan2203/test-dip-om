package usecase

import (
	"context"
	"encoding/json"
	"errors"

	"finagent/backend/internal/domain"
)

var validDashboards = map[string]bool{"overview": true, "goals": true}

type DashboardUsecase struct {
	repo domain.DashboardRepository
}

func NewDashboardUsecase(repo domain.DashboardRepository) *DashboardUsecase {
	return &DashboardUsecase{repo: repo}
}

// Розкладка дашборду.
func (uc *DashboardUsecase) Get(ctx context.Context, userID int64, name string) (json.RawMessage, error) {
	if !validDashboards[name] {
		return nil, domain.ErrValidation
	}
	d, err := uc.repo.Get(ctx, userID, name)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return json.RawMessage("[]"), nil
		}
		return nil, err
	}
	return d.Layout, nil
}

func (uc *DashboardUsecase) Save(ctx context.Context, userID int64, name string, layout json.RawMessage) error {
	if !validDashboards[name] {
		return domain.ErrValidation
	}
	if !json.Valid(layout) {
		return domain.ErrValidation
	}
	return uc.repo.Upsert(ctx, userID, name, layout)
}
