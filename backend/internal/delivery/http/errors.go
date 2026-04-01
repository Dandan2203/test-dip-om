package http

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"finagent/backend/internal/domain"
)

func HandleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		RespondError(c, http.StatusNotFound, "NOT_FOUND", err.Error())
	case errors.Is(err, domain.ErrEmailTaken):
		RespondError(c, http.StatusConflict, "EMAIL_TAKEN", err.Error())
	case errors.Is(err, domain.ErrConflict):
		RespondError(c, http.StatusConflict, "CONFLICT", err.Error())
	case errors.Is(err, domain.ErrInvalidCredentials):
		RespondError(c, http.StatusUnauthorized, "INVALID_CREDENTIALS", err.Error())
	case errors.Is(err, domain.ErrUnauthorized):
		RespondError(c, http.StatusUnauthorized, "UNAUTHORIZED", err.Error())
	case errors.Is(err, domain.ErrForbidden):
		RespondError(c, http.StatusForbidden, "FORBIDDEN", err.Error())
	case errors.Is(err, domain.ErrValidation):
		RespondError(c, http.StatusBadRequest, "VALIDATION", err.Error())
	default:
		// Невідому помилку не розкриваємо клієнту, лише логуємо.
		slog.Error("необроблена помилка запиту", "error", err)
		RespondError(c, http.StatusInternalServerError, "INTERNAL", "внутрішня помилка сервера")
	}
}
