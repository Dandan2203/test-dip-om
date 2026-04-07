package usecase

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"finagent/backend/internal/domain"
	"finagent/backend/internal/pkg/auth"
)

type mockUserRepository struct {
	mock.Mock
}

func (m *mockUserRepository) Create(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	if args.Error(0) == nil {
		user.ID = 1 // імітвція id
	}
	return args.Error(0)
}

func (m *mockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if u, ok := args.Get(0).(*domain.User); ok {
		return u, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockUserRepository) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	args := m.Called(ctx, id)
	if u, ok := args.Get(0).(*domain.User); ok {
		return u, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockUserRepository) Update(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func TestUserUsecase_Register_Success(t *testing.T) {
	repo := new(mockUserRepository)
	repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil)

	uc := NewUserUsecase(repo, "test-secret")
	user, err := uc.Register(context.Background(), "  Test@Example.COM ", "password123", "  Петро  ")

	require.NoError(t, err)
	assert.Equal(t, "test@example.com", user.Email, "email має бути нормалізовано")
	assert.Equal(t, "Петро", user.Name, "ім'я має бути обрізано від пробілів")
	assert.NotEmpty(t, user.PasswordHash)
	assert.NotEqual(t, "password123", user.PasswordHash, "пароль має бути захешовано")
	repo.AssertExpectations(t)
}

func TestUserUsecase_Register_EmailConflict(t *testing.T) {
	repo := new(mockUserRepository)
	repo.On("Create", mock.Anything, mock.Anything).Return(domain.ErrConflict)

	uc := NewUserUsecase(repo, "test-secret")
	_, err := uc.Register(context.Background(), "taken@example.com", "password123", "Тарас")

	assert.ErrorIs(t, err, domain.ErrEmailTaken)
	repo.AssertExpectations(t)
}

func TestUserUsecase_Login_Success(t *testing.T) {
	hash, _ := auth.HashPassword("password123")
	existing := &domain.User{ID: 7, Email: "user@example.com", PasswordHash: hash}

	repo := new(mockUserRepository)
	repo.On("GetByEmail", mock.Anything, "user@example.com").Return(existing, nil)

	uc := NewUserUsecase(repo, "test-secret")
	token, err := uc.Login(context.Background(), "user@example.com", "password123")

	require.NoError(t, err)
	assert.NotEmpty(t, token)
	repo.AssertExpectations(t)
}

func TestUserUsecase_Login_WrongPassword(t *testing.T) {
	hash, _ := auth.HashPassword("password123")
	existing := &domain.User{ID: 7, Email: "user@example.com", PasswordHash: hash}

	repo := new(mockUserRepository)
	repo.On("GetByEmail", mock.Anything, "user@example.com").Return(existing, nil)

	uc := NewUserUsecase(repo, "test-secret")
	_, err := uc.Login(context.Background(), "user@example.com", "wrong-password")

	assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
}

func TestUserUsecase_Login_UserNotFound(t *testing.T) {
	repo := new(mockUserRepository)
	repo.On("GetByEmail", mock.Anything, "ghost@example.com").Return(nil, domain.ErrNotFound)

	uc := NewUserUsecase(repo, "test-secret")
	_, err := uc.Login(context.Background(), "ghost@example.com", "password123")

	assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
}

func TestUserUsecase_UpdateProfile_Success(t *testing.T) {
	existing := &domain.User{ID: 7, Email: "user@example.com"}

	repo := new(mockUserRepository)
	repo.On("GetByID", mock.Anything, int64(7)).Return(existing, nil)
	repo.On("Update", mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil)

	uc := NewUserUsecase(repo, "test-secret")
	user, err := uc.UpdateProfile(context.Background(), 7, "  Петро  ")

	require.NoError(t, err)
	assert.Equal(t, "Петро", user.Name, "ім'я має бути обрізано від пробілів")
	repo.AssertExpectations(t)
}
