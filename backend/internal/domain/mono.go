package domain

import (
	"context"
	"time"
)

// MonoConnection — підключення користувача до Monobank (токен зберігається зашифрованим).
type MonoConnection struct {
	UserID         int64      `db:"user_id"`
	TokenEncrypted string     `db:"token_encrypted"`
	ConnectedAt    time.Time  `db:"connected_at"`
	LastImportAt   *time.Time `db:"last_import_at"`
}

type MonoRepository interface {
	Save(ctx context.Context, userID int64, tokenEncrypted string) error
	Get(ctx context.Context, userID int64) (*MonoConnection, error)
	Delete(ctx context.Context, userID int64) error
	TouchImport(ctx context.Context, userID int64) error
}
