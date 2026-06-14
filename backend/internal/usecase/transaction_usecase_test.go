package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"finagent/backend/internal/domain"
)

type mockTransactionRepository struct {
	mock.Mock
}

func (m *mockTransactionRepository) Create(ctx context.Context, tx *domain.Transaction) error {
	args := m.Called(ctx, tx)
	if args.Error(0) == nil {
		tx.ID = 1
		tx.CreatedAt = time.Now()
	}
	return args.Error(0)
}

func (m *mockTransactionRepository) GetByID(ctx context.Context, id int64) (*domain.Transaction, error) {
	args := m.Called(ctx, id)
	if tx, ok := args.Get(0).(*domain.Transaction); ok {
		return tx, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockTransactionRepository) List(ctx context.Context, userID int64, filter domain.TransactionFilter) ([]domain.Transaction, int64, error) {
	args := m.Called(ctx, userID, filter)
	if txs, ok := args.Get(0).([]domain.Transaction); ok {
		return txs, args.Get(1).(int64), args.Error(2)
	}
	return nil, 0, args.Error(2)
}

func (m *mockTransactionRepository) Update(ctx context.Context, tx *domain.Transaction) error {
	return m.Called(ctx, tx).Error(0)
}

func (m *mockTransactionRepository) Delete(ctx context.Context, id int64) error {
	return m.Called(ctx, id).Error(0)
}

func (m *mockTransactionRepository) Restore(ctx context.Context, id int64) error {
	return m.Called(ctx, id).Error(0)
}

func (m *mockTransactionRepository) CreateImported(ctx context.Context, tx *domain.Transaction) (bool, error) {
	args := m.Called(ctx, tx)
	return args.Bool(0), args.Error(1)
}

var testDate = time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)

func TestTransactionUsecase_Create_Success(t *testing.T) {
	repo := new(mockTransactionRepository)
	repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Transaction")).Return(nil)

	uc := NewTransactionUsecase(repo)
	tx, err := uc.Create(context.Background(), 5, domain.CreateTransactionInput{
		Type:            domain.TransactionExpense,
		Amount:          150.50,
		Description:     "  Кава  ",
		TransactionDate: testDate,
	})

	require.NoError(t, err)
	assert.Equal(t, int64(5), tx.UserID)
	assert.Equal(t, 150.50, tx.Amount)
	assert.Equal(t, "Кава", tx.Description, "опис має бути обрізаний від пробілів")
	assert.Equal(t, int64(1), tx.ID)
	repo.AssertExpectations(t)
}

func TestTransactionUsecase_List_AppliesDefaultLimit(t *testing.T) {
	repo := new(mockTransactionRepository)
	repo.On("List", mock.Anything, int64(5), mock.MatchedBy(func(f domain.TransactionFilter) bool {
		return f.Limit == defaultTxLimit
	})).Return([]domain.Transaction{}, int64(0), nil)

	uc := NewTransactionUsecase(repo)
	_, _, err := uc.List(context.Background(), 5, domain.TransactionFilter{Limit: 0})

	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestTransactionUsecase_List_CapsMaxLimit(t *testing.T) {
	repo := new(mockTransactionRepository)
	repo.On("List", mock.Anything, int64(5), mock.MatchedBy(func(f domain.TransactionFilter) bool {
		return f.Limit == maxTxLimit
	})).Return([]domain.Transaction{}, int64(0), nil)

	uc := NewTransactionUsecase(repo)
	_, _, err := uc.List(context.Background(), 5, domain.TransactionFilter{Limit: 9999})

	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestTransactionUsecase_Update_Success(t *testing.T) {
	existing := &domain.Transaction{
		ID: 10, UserID: 5, Type: domain.TransactionExpense, Amount: 100, TransactionDate: testDate,
	}

	repo := new(mockTransactionRepository)
	repo.On("GetByID", mock.Anything, int64(10)).Return(existing, nil)
	repo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Transaction")).Return(nil)

	uc := NewTransactionUsecase(repo)
	updated, err := uc.Update(context.Background(), 5, 10, domain.UpdateTransactionInput{
		Type:            domain.TransactionExpense,
		Amount:          200,
		Description:     "Нова кава",
		TransactionDate: testDate,
	})

	require.NoError(t, err)
	assert.Equal(t, 200.0, updated.Amount)
	assert.Equal(t, "Нова кава", updated.Description)
	repo.AssertExpectations(t)
}

func TestTransactionUsecase_Update_ForbiddenOtherUser(t *testing.T) {
	existing := &domain.Transaction{ID: 10, UserID: 99}

	repo := new(mockTransactionRepository)
	repo.On("GetByID", mock.Anything, int64(10)).Return(existing, nil)

	uc := NewTransactionUsecase(repo)
	_, err := uc.Update(context.Background(), 5, 10, domain.UpdateTransactionInput{})

	assert.ErrorIs(t, err, domain.ErrForbidden)
	repo.AssertExpectations(t)
}

func TestTransactionUsecase_Delete_Success(t *testing.T) {
	existing := &domain.Transaction{ID: 10, UserID: 5}

	repo := new(mockTransactionRepository)
	repo.On("GetByID", mock.Anything, int64(10)).Return(existing, nil)
	repo.On("Delete", mock.Anything, int64(10)).Return(nil)

	uc := NewTransactionUsecase(repo)
	err := uc.Delete(context.Background(), 5, 10)

	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestTransactionUsecase_Delete_ForbiddenOtherUser(t *testing.T) {
	existing := &domain.Transaction{ID: 10, UserID: 99}

	repo := new(mockTransactionRepository)
	repo.On("GetByID", mock.Anything, int64(10)).Return(existing, nil)

	uc := NewTransactionUsecase(repo)
	err := uc.Delete(context.Background(), 5, 10)

	assert.ErrorIs(t, err, domain.ErrForbidden)
	repo.AssertExpectations(t)
}

func TestTransactionUsecase_Delete_NotFound(t *testing.T) {
	repo := new(mockTransactionRepository)
	repo.On("GetByID", mock.Anything, int64(999)).Return(nil, domain.ErrNotFound)

	uc := NewTransactionUsecase(repo)
	err := uc.Delete(context.Background(), 5, 999)

	assert.ErrorIs(t, err, domain.ErrNotFound)
	repo.AssertExpectations(t)
}
