package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"finagent/backend/internal/delivery/http/middleware"
	"finagent/backend/internal/domain"
)

type CategoryHandler struct {
	uc domain.CategoryUsecase
}

func NewCategoryHandler(uc domain.CategoryUsecase) *CategoryHandler {
	return &CategoryHandler{uc: uc}
}

type createCategoryRequest struct {
	Name string `json:"name" binding:"required,max=100"`
	Type string `json:"type" binding:"required,oneof=income expense"`
}

type updateCategoryRequest struct {
	Name string `json:"name" binding:"required,max=100"`
}

type categoryResponse struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	IsSystem bool   `json:"isSystem"`
}

func toCategoryResponse(c *domain.Category) categoryResponse {
	return categoryResponse{
		ID:       c.ID,
		Name:     c.Name,
		Type:     string(c.Type),
		IsSystem: c.IsSystem(),
	}
}

func (h *CategoryHandler) List(c *gin.Context) {
	categories, err := h.uc.List(c.Request.Context(), middleware.UserIDFromContext(c))
	if err != nil {
		HandleError(c, err)
		return
	}

	result := make([]categoryResponse, 0, len(categories))
	for i := range categories {
		result = append(result, toCategoryResponse(&categories[i]))
	}
	RespondOK(c, http.StatusOK, result)
}

func (h *CategoryHandler) Create(c *gin.Context) {
	var req createCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	category, err := h.uc.Create(c.Request.Context(),
		middleware.UserIDFromContext(c), req.Name, domain.CategoryType(req.Type))
	if err != nil {
		HandleError(c, err)
		return
	}
	RespondOK(c, http.StatusCreated, toCategoryResponse(category))
}

func (h *CategoryHandler) Update(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}

	var req updateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	category, err := h.uc.Update(c.Request.Context(),
		middleware.UserIDFromContext(c), id, req.Name)
	if err != nil {
		HandleError(c, err)
		return
	}
	RespondOK(c, http.StatusOK, toCategoryResponse(category))
}

func (h *CategoryHandler) Delete(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}

	if err := h.uc.Delete(c.Request.Context(), middleware.UserIDFromContext(c), id); err != nil {
		HandleError(c, err)
		return
	}
	RespondOK(c, http.StatusOK, gin.H{"deleted": true})
}

// parseIDParam — у разі помилки сам відповідає 400 і повертає ok == false.
func parseIDParam(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		RespondError(c, http.StatusBadRequest, "INVALID_INPUT", "некоректний ідентифікатор")
		return 0, false
	}
	return id, true
}
