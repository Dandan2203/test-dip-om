package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	"finagent/backend/internal/ai"
	"finagent/backend/internal/delivery/http/middleware"
	"finagent/backend/internal/domain"
	"finagent/backend/internal/usecase"
)

func NewRouter(
	db *sqlx.DB,
	userUC domain.UserUsecase,
	categoryUC domain.CategoryUsecase,
	transactionUC domain.TransactionUsecase,
	statsUC domain.StatsUsecase,
	goalUC domain.GoalUsecase,
	actionUC domain.ActionUsecase,
	actionRepo domain.ActionRepository,
	chatLogRepo domain.ChatLogRepository,
	dashboardUC domain.DashboardUsecase,
	monoUC *usecase.MonoUsecase,
	aiClient *ai.Client,
	txRepo domain.TransactionRepository,
	jwtSecret string,
) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(middleware.Logger(), gin.Recovery())

	router.GET("/health", healthHandler(db))

	userHandler := NewUserHandler(userUC)
	categoryHandler := NewCategoryHandler(categoryUC)
	transactionHandler := NewTransactionHandler(transactionUC)
	statsHandler := NewStatsHandler(statsUC, goalUC)
	goalHandler := NewGoalHandler(goalUC)
	chatHandler := NewChatHandler(aiClient, txRepo, transactionUC, goalUC, categoryUC, actionRepo, chatLogRepo)
	auditHandler := NewAuditHandler(aiClient, txRepo, categoryUC)
	actionHandler := NewActionHandler(actionUC)
	dashboardHandler := NewDashboardHandler(dashboardUC)
	monoHandler := NewMonoHandler(monoUC)

	// Обмеження частоти: 20 запитів/с зі сплеском до 40 на IP.
	api := router.Group("/api", middleware.RateLimit(20, 40))
	{
		authGroup := api.Group("/auth")
		{
			authGroup.POST("/register", userHandler.Register)
			authGroup.POST("/login", userHandler.Login)
		}

		users := api.Group("/users")
		users.Use(middleware.Auth(jwtSecret))
		{
			users.GET("/me", userHandler.GetMe)
			users.PUT("/me", userHandler.UpdateMe)
		}

		categories := api.Group("/categories")
		categories.Use(middleware.Auth(jwtSecret))
		{
			categories.GET("", categoryHandler.List)
			categories.POST("", categoryHandler.Create)
			categories.PUT("/:id", categoryHandler.Update)
			categories.DELETE("/:id", categoryHandler.Delete)
		}

		transactions := api.Group("/transactions")
		transactions.Use(middleware.Auth(jwtSecret))
		{
			transactions.POST("", transactionHandler.Create)
			transactions.GET("", transactionHandler.List)
			transactions.PUT("/:id", transactionHandler.Update)
			transactions.DELETE("/:id", transactionHandler.Delete)
		}

		stats := api.Group("/stats")
		stats.Use(middleware.Auth(jwtSecret))
		{
			stats.GET("/summary", statsHandler.Summary)
			stats.GET("/by-category", statsHandler.ByCategory)
		}

		api.GET("/dashboard", middleware.Auth(jwtSecret), statsHandler.Dashboard)
		api.GET("/transactions/export", middleware.Auth(jwtSecret), statsHandler.Export)
		api.POST("/chat", middleware.Auth(jwtSecret), chatHandler.Chat)
		api.GET("/anomalies", middleware.Auth(jwtSecret), auditHandler.Anomalies)
		api.POST("/actions/undo", middleware.Auth(jwtSecret), actionHandler.Undo)
		api.POST("/actions/execute", middleware.Auth(jwtSecret), chatHandler.Execute)

		dashboards := api.Group("/dashboards")
		dashboards.Use(middleware.Auth(jwtSecret))
		{
			dashboards.GET("/:name", dashboardHandler.Get)
			dashboards.PUT("/:name", dashboardHandler.Save)
		}

		mono := api.Group("/mono")
		mono.Use(middleware.Auth(jwtSecret))
		{
			mono.POST("/connect", monoHandler.Connect)
			mono.DELETE("/connect", monoHandler.Disconnect)
			mono.GET("/status", monoHandler.Status)
			mono.GET("/accounts", monoHandler.Accounts)
			mono.GET("/currency", monoHandler.Currency)
			mono.POST("/import", monoHandler.Import)
		}

		goals := api.Group("/goals")
		goals.Use(middleware.Auth(jwtSecret))
		{
			goals.GET("", goalHandler.List)
			goals.POST("", goalHandler.Create)
			goals.PUT("/:id", goalHandler.Update)
			goals.POST("/:id/contribute", goalHandler.Contribute)
			goals.DELETE("/:id", goalHandler.Delete)
		}
	}

	return router
}

func healthHandler(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := db.Ping(); err != nil {
			RespondError(c, http.StatusServiceUnavailable,
				"DB_UNAVAILABLE", "база даних недоступна")
			return
		}
		RespondOK(c, http.StatusOK, gin.H{"status": "ok", "db": "up"})
	}
}
