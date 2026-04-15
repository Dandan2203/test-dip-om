package usecase

import (
	"context"
	"strings"

	"finagent/backend/internal/domain"
)

type CategoryUsecase struct {
	repo domain.CategoryRepository
}

func NewCategoryUsecase(repo domain.CategoryRepository) *CategoryUsecase {
	return &CategoryUsecase{repo: repo}
}

func (uc *CategoryUsecase) List(ctx context.Context, userID int64) ([]domain.Category, error) {
	return uc.repo.ListAvailable(ctx, userID)
}

func (uc *CategoryUsecase) Create(ctx context.Context, userID int64, name string, catType domain.CategoryType) (*domain.Category, error) {
	category := &domain.Category{
		UserID: &userID,
		Name:   strings.TrimSpace(name),
		Type:   catType,
	}
	if err := uc.repo.Create(ctx, category); err != nil {
		return nil, err
	}
	return category, nil
}

func (uc *CategoryUsecase) Update(ctx context.Context, userID, categoryID int64, name string) (*domain.Category, error) {
	category, err := uc.repo.GetByID(ctx, categoryID)
	if err != nil {
		return nil, err
	}
	if err := ensureOwnedByUser(category, userID); err != nil {
		return nil, err
	}

	category.Name = strings.TrimSpace(name)
	if err := uc.repo.Update(ctx, category); err != nil {
		return nil, err
	}
	return category, nil
}

func (uc *CategoryUsecase) Delete(ctx context.Context, userID, categoryID int64) error {
	category, err := uc.repo.GetByID(ctx, categoryID)
	if err != nil {
		return err
	}
	if err := ensureOwnedByUser(category, userID); err != nil {
		return err
	}
	return uc.repo.Delete(ctx, categoryID)
}

// Перевірка власника.
func ensureOwnedByUser(category *domain.Category, userID int64) error {
	if category.UserID == nil || *category.UserID != userID {
		return domain.ErrForbidden
	}
	return nil
}
