// Команда застосування міграцій БД: go run ./cmd/migrate up|down
package main

import (
	"errors"
	"log/slog"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // драйвер postgres
	_ "github.com/golang-migrate/migrate/v4/source/file"       // джерело file

	"finagent/backend/internal/config"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	if len(os.Args) < 2 {
		slog.Error("не вказано команду; використання: go run ./cmd/migrate <up|down>")
		os.Exit(1)
	}
	command := os.Args[1]

	cfg, err := config.Load()
	if err != nil {
		slog.Error("не вдалося завантажити конфігурацію", "error", err)
		os.Exit(1)
	}

	migrator, err := migrate.New("file://migrations", cfg.DatabaseURL)
	if err != nil {
		slog.Error("не вдалося ініціалізувати міграції", "error", err)
		os.Exit(1)
	}

	switch command {
	case "up":
		err = migrator.Up()
	case "down":
		err = migrator.Down()
	default:
		slog.Error("невідома команда", "command", command)
		os.Exit(1)
	}

	// ErrNoChange — застосовувати нічого, це не помилка.
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		slog.Error("помилка застосування міграцій", "error", err)
		os.Exit(1)
	}

	slog.Info("міграції успішно застосовано", "command", command)
}
