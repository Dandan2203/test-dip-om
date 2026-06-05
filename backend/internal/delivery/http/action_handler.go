package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"finagent/backend/internal/delivery/http/middleware"
	"finagent/backend/internal/domain"
)

type ActionHandler struct {
	uc domain.ActionUsecase
}

func NewActionHandler(uc domain.ActionUsecase) *ActionHandler {
	return &ActionHandler{uc: uc}
}

func (h *ActionHandler) Undo(c *gin.Context) {
	userID := middleware.UserIDFromContext(c)

	action, err := h.uc.Undo(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			RespondError(c, http.StatusNotFound, "NO_ACTION", "немає дій для відміни")
			return
		}
		HandleError(c, err)
		return
	}

	RespondOK(c, http.StatusOK, gin.H{
		"undone":      true,
		"action_type": action.ActionType,
		"entity_type": action.EntityType,
		"entity_id":   action.EntityID,
	})
}
