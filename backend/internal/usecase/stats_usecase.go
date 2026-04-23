package usecase

import (
	"context"
	"time"

	"finagent/backend/internal/domain"
)

type StatsUsecase struct {
	repo domain.StatsRepository
}

func NewStatsUsecase(repo domain.StatsRepository) *StatsUsecase {
	return &StatsUsecase{repo: repo}
}

func (uc *StatsUsecase) Summary(ctx context.Context, userID int64, period domain.StatsPeriod) (domain.StatsSummary, error) {
	return uc.repo.Summary(ctx, userID, period)
}

func (uc *StatsUsecase) ByCategory(ctx context.Context, userID int64, txType domain.TransactionType, period domain.StatsPeriod) ([]domain.CategoryStat, error) {
	return uc.repo.ByCategory(ctx, userID, txType, period)
}

func (uc *StatsUsecase) Dashboard(ctx context.Context, userID int64, goalUC domain.GoalUsecase) (domain.DashboardData, error) {
	now := time.Now()
	period := domain.StatsPeriod{
		From: time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC),
		To:   now,
	}

	summary, err := uc.repo.Summary(ctx, userID, period)
	if err != nil {
		return domain.DashboardData{}, err
	}

	recent, err := uc.repo.RecentTransactions(ctx, userID, 10)
	if err != nil {
		return domain.DashboardData{}, err
	}

	goals, err := goalUC.List(ctx, userID)
	if err != nil {
		return domain.DashboardData{}, err
	}

	return domain.DashboardData{
		Income:             summary.Income,
		Expense:            summary.Expense,
		Balance:            summary.Balance,
		RecentTransactions: recent,
		Goals:              goals,
	}, nil
}

func (uc *StatsUsecase) Export(ctx context.Context, userID int64, filter domain.TransactionFilter) ([]domain.TransactionExportRow, error) {
	return uc.repo.ExportRows(ctx, userID, filter)
}
