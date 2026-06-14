// Package domain — бізнес-сутності та контракти предметної області.
package domain

import "errors"

// Помилки предметної області; шар delivery мапить їх на HTTP-статуси.
var (
	ErrNotFound           = errors.New("ресурс не знайдено")
	ErrConflict           = errors.New("конфлікт даних")
	ErrEmailTaken         = errors.New("користувач з таким email вже зареєстрований")
	ErrInvalidCredentials = errors.New("невірний email або пароль")
	ErrUnauthorized       = errors.New("неавторизований запит")
	ErrForbidden          = errors.New("дія заборонена")
	ErrValidation         = errors.New("некоректні дані")
)
