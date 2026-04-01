package http

import "github.com/gin-gonic/gin"

type errorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func RespondOK(c *gin.Context, status int, data any) {
	c.JSON(status, gin.H{"data": data})
}

func RespondError(c *gin.Context, status int, code, message string) {
	c.JSON(status, gin.H{"error": errorPayload{Code: code, Message: message}})
}
