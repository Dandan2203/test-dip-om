package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"finagent/backend/internal/pkg/auth"
)

const contextUserIDKey = "userID"

// Auth — middleware перевірки JWT із заголовка Authorization.
func Auth(jwtSecret string) gin.HandlerFunc {
	const prefix = "Bearer "

	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, prefix) {
			abortUnauthorized(c, "відсутній або некоректний заголовок Authorization")
			return
		}

		claims, err := auth.ParseToken(strings.TrimPrefix(header, prefix), jwtSecret)
		if err != nil {
			abortUnauthorized(c, "недійсний або прострочений токен")
			return
		}

		c.Set(contextUserIDKey, claims.UserID)
		c.Next()
	}
}

// UserIDFromContext — лише для обробників, захищених middleware Auth.
func UserIDFromContext(c *gin.Context) int64 {
	if value, ok := c.Get(contextUserIDKey); ok {
		if id, ok := value.(int64); ok {
			return id
		}
	}
	return 0
}

func abortUnauthorized(c *gin.Context, message string) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
		"error": gin.H{"code": "UNAUTHORIZED", "message": message},
	})
}
