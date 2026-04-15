package usecase

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"finagent/backend/internal/domain"
)

type mockCategoryRepository struct {
	mock.Mock
}

func (m *mockCategoryRepository) ListAvailable(ctx context.Context, userID int64) ([]domain.Category, error) {
	args := m.Called(ctx, userID)
	if c, ok := args.Get(0).([]domain.Category); ok {
		return c, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockCategoryRepository) GetByID(ctx context.Context, id int64) (*domain.Category, error) {
	args := m.Called(ctx, id)
	if c, ok := args.Get(0).(*domain.Category); ok {
		return c, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockCategoryRepository) Create(ctx context.Context, category *domain.Category) error {
	args := m.Called(ctx, category)
	if args.Error(0) == nil {
		category.ID = 1 // імітуємо id, згенерований базою
	}
	return args.Error(0)
}

func (m *mockCategoryRepository) Update(ctx context.Context, category *domain.Category) error {
	return m.Called(ctx, category).Error(0)
}

func (m *mockCategoryRepository) Delete(ctx context.Context, id int64) error {
	return m.Called(ctx, id).Error(0)
}

func TestCategoryUsecase_Create_Success(t *testing.T) {
	repo := new(mockCategoryRepository)
	repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Category")).Return(nil)

	uc := NewCategoryUsecase(repo)
	category, err := uc.Create(context.Background(), 5, "  Кава  ", domain.CategoryExpense)

	require.NoError(t, err)
	assert.Equal(t, "Кава", category.Name, "назву має бути обрізано від пробілів")
	assert.Equal(t, domain.CategoryExpense, category.Type)
	require.NotNil(t, category.UserID)
	assert.Equal(t, int64(5), *category.UserID)
	repo.AssertExpectations(t)
}

func TestCategoryUsecase_Update_Success(t *testing.T) {
	ownerID := int64(5)
	existing := &domain.Category{ID: 10, UserID: &ownerID, Name: "Старе", Type: domain.CategoryExpense}

	repo := new(mockCategoryRepository)
	repo.On("GetByID", mock.Anything, int64(10)).Return(existing, nil)
	repo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Category")).Return(nil)

	uc := NewCategoryUsecase(repo)
	category, err := uc.Update(context.Background(), 5, 10, "Нове")

	require.NoError(t, err)
	assert.Equal(t, "Нове", category.Name)
	repo.AssertExpectations(t)
}

func TestCategoryUsecase_Update_SystemCategoryForbidden(t *testing.T) {
	systemCategory := &domain.Category{ID: 1, UserID: nil, Name: "Транспорт", Type: domain.CategoryExpense}

	repo := new(mockCategoryRepository)
	repo.On("GetByID", mock.Anything, int64(1)).Return(systemCategory, nil)

	uc := NewCategoryUsecase(repo)
	_, err := uc.Update(context.Background(), 5, 1, "Моя назва")

	assert.ErrorIs(t, err, domain.ErrForbidden, "системну категорію змінювати не можна")
}

func TestCategoryUsecase_Delete_ForeignCategoryForbidden(t *testing.T) {
	otherUserID := int64(99)
	foreign := &domain.Category{ID: 10, UserID: &otherUserID, Name: "Чуже", Type: domain.CategoryExpense}

	repo := new(mockCategoryRepository)
	repo.On("GetByID", mock.Anything, int64(10)).Return(foreign, nil)

	uc := NewCategoryUsecase(repo)
	err := uc.Delete(context.Background(), 5, 10)

	assert.ErrorIs(t, err, domain.ErrForbidden, "чужу категорію видаляти не можна")
}

func TestCategoryUsecase_Delete_Success(t *testing.T) {
	ownerID := int64(5)
	existing := &domain.Category{ID: 10, UserID: &ownerID, Name: "Кава", Type: domain.CategoryExpense}

	repo := new(mockCategoryRepository)
	repo.On("GetByID", mock.Anything, int64(10)).Return(existing, nil)
	repo.On("Delete", mock.Anything, int64(10)).Return(nil)

	uc := NewCategoryUsecase(repo)
	err := uc.Delete(context.Background(), 5, 10)

	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestCategoryUsecase_List_Success(t *testing.T) {
	repo := new(mockCategoryRepository)
	repo.On("ListAvailable", mock.Anything, int64(5)).Return([]domain.Category{
		{ID: 1, Name: "Транспорт", Type: domain.CategoryExpense},
		{ID: 2, Name: "Зарплата", Type: domain.CategoryIncome},
	}, nil)

	uc := NewCategoryUsecase(repo)
	categories, err := uc.List(context.Background(), 5)

	require.NoError(t, err)
	assert.Len(t, categories, 2)
	repo.AssertExpectations(t)
}
