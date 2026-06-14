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

type mockGoalRepository struct {
	mock.Mock
}

func (m *mockGoalRepository) Create(ctx context.Context, goal *domain.Goal) error {
	args := m.Called(ctx, goal)
	if args.Error(0) == nil {
		goal.ID = 1
		goal.CurrentAmount = 0
		goal.CreatedAt = time.Now()
	}
	return args.Error(0)
}

func (m *mockGoalRepository) GetByID(ctx context.Context, id int64) (*domain.Goal, error) {
	args := m.Called(ctx, id)
	if g, ok := args.Get(0).(*domain.Goal); ok {
		return g, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockGoalRepository) List(ctx context.Context, userID int64) ([]domain.Goal, error) {
	args := m.Called(ctx, userID)
	if gs, ok := args.Get(0).([]domain.Goal); ok {
		return gs, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockGoalRepository) Update(ctx context.Context, goal *domain.Goal) error {
	return m.Called(ctx, goal).Error(0)
}

func (m *mockGoalRepository) Contribute(ctx context.Context, id int64, amount float64) error {
	return m.Called(ctx, id, amount).Error(0)
}

func (m *mockGoalRepository) Delete(ctx context.Context, id int64) error {
	return m.Called(ctx, id).Error(0)
}

func (m *mockGoalRepository) Restore(ctx context.Context, id int64) error {
	return m.Called(ctx, id).Error(0)
}

func TestGoalUsecase_Create_Success(t *testing.T) {
	repo := new(mockGoalRepository)
	repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Goal")).Return(nil)

	uc := NewGoalUsecase(repo)
	goal, err := uc.Create(context.Background(), 5, domain.CreateGoalInput{
		Title:        "  Новий авто  ",
		TargetAmount: 500000,
	})

	require.NoError(t, err)
	assert.Equal(t, "Новий авто", goal.Title)
	assert.Equal(t, 500000.0, goal.TargetAmount)
	assert.Equal(t, int64(5), goal.UserID)
	repo.AssertExpectations(t)
}

func TestGoalUsecase_Contribute_Success(t *testing.T) {
	existing := &domain.Goal{ID: 1, UserID: 5, TargetAmount: 1000, CurrentAmount: 100}

	repo := new(mockGoalRepository)
	repo.On("GetByID", mock.Anything, int64(1)).Return(existing, nil)
	repo.On("Contribute", mock.Anything, int64(1), 200.0).Return(nil)

	uc := NewGoalUsecase(repo)
	goal, err := uc.Contribute(context.Background(), 5, 1, 200)

	require.NoError(t, err)
	assert.Equal(t, 300.0, goal.CurrentAmount)
	repo.AssertExpectations(t)
}

func TestGoalUsecase_Delete_ForbiddenOtherUser(t *testing.T) {
	existing := &domain.Goal{ID: 1, UserID: 99}

	repo := new(mockGoalRepository)
	repo.On("GetByID", mock.Anything, int64(1)).Return(existing, nil)

	uc := NewGoalUsecase(repo)
	err := uc.Delete(context.Background(), 5, 1)

	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestGoalUsecase_Update_ForbiddenOtherUser(t *testing.T) {
	existing := &domain.Goal{ID: 1, UserID: 99}

	repo := new(mockGoalRepository)
	repo.On("GetByID", mock.Anything, int64(1)).Return(existing, nil)

	uc := NewGoalUsecase(repo)
	_, err := uc.Update(context.Background(), 5, 1, domain.UpdateGoalInput{Title: "X", TargetAmount: 1})

	assert.ErrorIs(t, err, domain.ErrForbidden)
}
