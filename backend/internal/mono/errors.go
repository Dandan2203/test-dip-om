package mono

import "errors"

var (
	ErrInvalidToken = errors.New("mono: недійсний токен")
	ErrRateLimited  = errors.New("mono: перевищено ліміт запитів, спробуйте за хвилину")
)
