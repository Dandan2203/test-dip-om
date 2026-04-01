package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"finagent/backend/internal/ai"
	"finagent/backend/internal/config"
	"finagent/backend/internal/database"
	httpserver "finagent/backend/internal/delivery/http"
	"finagent/backend/internal/mono"
	"finagent/backend/internal/repository"
	"finagent/backend/internal/usecase"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	cfg, err := config.Load()
	if err != nil {
		slog.Error("не вдалося завантажити конфігурацію", "error", err)
		os.Exit(1)
	}

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		slog.Error("не вдалося підключитися до бази даних", "error", err)
		os.Exit(1)
	}
	defer func() { _ = db.Close() }()

	userRepo := repository.NewUserRepository(db)
	userUsecase := usecase.NewUserUsecase(userRepo, cfg.JWTSecret)

	categoryRepo := repository.NewCategoryRepository(db)
	categoryUsecase := usecase.NewCategoryUsecase(categoryRepo)

	transactionRepo := repository.NewTransactionRepository(db)
	transactionUsecase := usecase.NewTransactionUsecase(transactionRepo, categoryRepo)

	statsRepo := repository.NewStatsRepository(db)
	statsUsecase := usecase.NewStatsUsecase(statsRepo)

	goalRepo := repository.NewGoalRepository(db)
	goalUsecase := usecase.NewGoalUsecase(goalRepo)

	actionRepo := repository.NewActionRepository(db)
	actionUsecase := usecase.NewActionUsecase(actionRepo, transactionRepo, goalRepo)

	chatLogRepo := repository.NewChatLogRepository(db)

	dashboardRepo := repository.NewDashboardRepository(db)
	dashboardUsecase := usecase.NewDashboardUsecase(dashboardRepo)

	monoRepo := repository.NewMonoRepository(db)
	monoClient := mono.NewClient()
	monoUsecase := usecase.NewMonoUsecase(monoRepo, monoClient, transactionRepo, categoryRepo, cfg.EncryptionKey)

	aiClient := ai.NewClient(cfg.AIServiceURL, cfg.AIInternalToken)

	server := &http.Server{
		Addr: ":" + cfg.AppPort,
		Handler: httpserver.NewRouter(
			db, userUsecase, categoryUsecase, transactionUsecase,
			statsUsecase, goalUsecase, actionUsecase, actionRepo, chatLogRepo, dashboardUsecase, monoUsecase,
			aiClient, transactionRepo, cfg.JWTSecret,
		),
	}

	go func() {
		slog.Info("запуск FinAgent backend", "port", cfg.AppPort)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("сервер завершив роботу з помилкою", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("отримано сигнал завершення, зупинка сервера...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("некоректне завершення роботи сервера", "error", err)
	}
	slog.Info("сервер зупинено")
}
