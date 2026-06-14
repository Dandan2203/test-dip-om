package usecase

import (
	"context"
	"strings"

	"finagent/backend/internal/domain"
)

const (
	defaultTxLimit = 50
	maxTxLimit     = 200
)

type TransactionUsecase struct {
	repo domain.TransactionRepository
}

func NewTransactionUsecase(repo domain.TransactionRepository) *TransactionUsecase {
	return &TransactionUsecase{repo: repo}
}

func (uc *TransactionUsecase) Create(ctx context.Context, userID int64, input domain.CreateTransactionInput) (*domain.Transaction, error) {
	tx := &domain.Transaction{
		UserID:          userID,
		CategoryID:      input.CategoryID,
		Type:            input.Type,
		Amount:          input.Amount,
		Description:     strings.TrimSpace(input.Description),
		TransactionDate: input.TransactionDate,
	}
	if err := uc.repo.Create(ctx, tx); err != nil {
		return nil, err
	}
	return tx, nil
}

func (uc *TransactionUsecase) List(ctx context.Context, userID int64, filter domain.TransactionFilter) ([]domain.Transaction, int64, error) {
	if filter.Limit <= 0 {
		filter.Limit = defaultTxLimit
	} else if filter.Limit > maxTxLimit {
		filter.Limit = maxTxLimit
	}
	return uc.repo.List(ctx, userID, filter)
}

func (uc *TransactionUsecase) Update(ctx context.Context, userID, transactionID int64, input domain.UpdateTransactionInput) (*domain.Transaction, error) {
	tx, err := uc.repo.GetByID(ctx, transactionID)
	if err != nil {
		return nil, err
	}
	if tx.UserID != userID {
		return nil, domain.ErrForbidden
	}

	tx.CategoryID = input.CategoryID
	tx.Type = input.Type
	tx.Amount = input.Amount
	tx.Description = strings.TrimSpace(input.Description)
	tx.TransactionDate = input.TransactionDate

	if err := uc.repo.Update(ctx, tx); err != nil {
		return nil, err
	}
	return tx, nil
}

func (uc *TransactionUsecase) Delete(ctx context.Context, userID, transactionID int64) error {
	tx, err := uc.repo.GetByID(ctx, transactionID)
	if err != nil {
		return err
	}
	if tx.UserID != userID {
		return domain.ErrForbidden
	}
	return uc.repo.Delete(ctx, transactionID)
}
