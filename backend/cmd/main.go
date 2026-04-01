// Package main — точка входу FinAgent backend
package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// main — старт сервера
func main() {
	// JSON-логер
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	gin.SetMode(gin.ReleaseMode)

	slog.Info("Starting FinAgent", "port", 8080)

	if err := setupRouter().Run(":8080"); err != nil {
		slog.Error("Server failed", "error", err)
		os.Exit(1)
	}
}

// setupRouter — роутер + middleware
func setupRouter() *gin.Engine {
	r := gin.New()
	r.Use(ginLogger(), gin.Recovery())
	r.GET("/ping", handlePing)
	return r
}

// handlePing — health check
func handlePing(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
		"service": "finagent-backend",
	})
}

// ginLogger — логування запитів
func ginLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		slog.Info("request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
		)
	}
}
