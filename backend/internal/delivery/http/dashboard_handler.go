package http

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"

	"finagent/backend/internal/delivery/http/middleware"
	"finagent/backend/internal/domain"
)

type DashboardHandler struct {
	uc domain.DashboardUsecase
}

func NewDashboardHandler(uc domain.DashboardUsecase) *DashboardHandler {
	return &DashboardHandler{uc: uc}
}

func (h *DashboardHandler) Get(c *gin.Context) {
	layout, err := h.uc.Get(c.Request.Context(), middleware.UserIDFromContext(c), c.Param("name"))
	if err != nil {
		HandleError(c, err)
		return
	}
	RespondOK(c, http.StatusOK, layout)
}

type dashboardSaveRequest struct {
	Layout json.RawMessage `json:"layout" binding:"required"`
}

func (h *DashboardHandler) Save(c *gin.Context) {
	var req dashboardSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	if err := h.uc.Save(c.Request.Context(), middleware.UserIDFromContext(c), c.Param("name"), req.Layout); err != nil {
		HandleError(c, err)
		return
	}
	RespondOK(c, http.StatusOK, gin.H{"saved": true})
}
