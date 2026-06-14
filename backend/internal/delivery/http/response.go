// Package http — HTTP-шар доставки: роутер, middleware, обробники.
package http

import "github.com/gin-gonic/gin"

type errorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// RespondOK — успішна відповідь у форматі {"data": ...}.
func RespondOK(c *gin.Context, status int, data any) {
	c.JSON(status, gin.H{"data": data})
}

// RespondError — відповідь з помилкою у форматі {"error": {code, message}}.
func RespondError(c *gin.Context, status int, code, message string) {
	c.JSON(status, gin.H{"error": errorPayload{Code: code, Message: message}})
}
