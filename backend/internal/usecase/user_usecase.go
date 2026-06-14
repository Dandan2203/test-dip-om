// Package usecase — бізнес-логіка (сценарії використання) застосунку.
package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"finagent/backend/internal/domain"
	"finagent/backend/internal/pkg/auth"
)

type UserUsecase struct {
	repo      domain.UserRepository
	jwtSecret string
}

func NewUserUsecase(repo domain.UserRepository, jwtSecret string) *UserUsecase {
	return &UserUsecase{repo: repo, jwtSecret: jwtSecret}
}

func (uc *UserUsecase) Register(ctx context.Context, email, password, name string) (*domain.User, error) {
	email = normalizeEmail(email)

	hash, err := auth.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("usecase: хешування пароля: %w", err)
	}

	user := &domain.User{Email: email, PasswordHash: hash, Name: strings.TrimSpace(name)}
	if err := uc.repo.Create(ctx, user); err != nil {
		// Єдине унікальне поле при реєстрації — email, тож конфлікт = email зайнятий.
		if errors.Is(err, domain.ErrConflict) {
			return nil, domain.ErrEmailTaken
		}
		return nil, err
	}
	return user, nil
}

func (uc *UserUsecase) Login(ctx context.Context, email, password string) (string, error) {
	email = normalizeEmail(email)

	user, err := uc.repo.GetByEmail(ctx, email)
	if err != nil {
		// Не розкриваємо, що саме невірне — email чи пароль.
		if errors.Is(err, domain.ErrNotFound) {
			return "", domain.ErrInvalidCredentials
		}
		return "", err
	}

	if !auth.CheckPassword(password, user.PasswordHash) {
		return "", domain.ErrInvalidCredentials
	}

	token, err := auth.GenerateToken(user.ID, uc.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("usecase: генерація токена: %w", err)
	}
	return token, nil
}

func (uc *UserUsecase) GetProfile(ctx context.Context, userID int64) (*domain.User, error) {
	return uc.repo.GetByID(ctx, userID)
}

func (uc *UserUsecase) UpdateProfile(ctx context.Context, userID int64, name string) (*domain.User, error) {
	user, err := uc.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	user.Name = strings.TrimSpace(name)
	if err := uc.repo.Update(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
