package http

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"finagent/backend/internal/delivery/http/middleware"
	"finagent/backend/internal/domain"
)

type GoalHandler struct {
	uc domain.GoalUsecase
}

func NewGoalHandler(uc domain.GoalUsecase) *GoalHandler {
	return &GoalHandler{uc: uc}
}

type createGoalRequest struct {
	Title        string  `json:"title"        binding:"required,max=255"`
	TargetAmount float64 `json:"targetAmount" binding:"required,gt=0"`
	Deadline     *string `json:"deadline"`
}

type updateGoalRequest struct {
	Title        string  `json:"title"        binding:"required,max=255"`
	TargetAmount float64 `json:"targetAmount" binding:"required,gt=0"`
	Deadline     *string `json:"deadline"`
}

type contributeRequest struct {
	Amount float64 `json:"amount" binding:"required,gt=0"`
}

type goalResponse struct {
	ID            int64    `json:"id"`
	Title         string   `json:"title"`
	TargetAmount  float64  `json:"targetAmount"`
	CurrentAmount float64  `json:"currentAmount"`
	Deadline      *string  `json:"deadline"`
	CreatedAt     string   `json:"createdAt"`
}

func toGoalResponse(g *domain.Goal) goalResponse {
	var deadline *string
	if g.Deadline != nil {
		s := g.Deadline.Format(dateLayout)
		deadline = &s
	}
	return goalResponse{
		ID:            g.ID,
		Title:         g.Title,
		TargetAmount:  g.TargetAmount,
		CurrentAmount: g.CurrentAmount,
		Deadline:      deadline,
		CreatedAt:     g.CreatedAt.Format(time.RFC3339),
	}
}

func parseOptionalDate(s *string) (*time.Time, error) {
	if s == nil || *s == "" {
		return nil, nil
	}
	t, err := time.Parse(dateLayout, *s)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (h *GoalHandler) List(c *gin.Context) {
	goals, err := h.uc.List(c.Request.Context(), middleware.UserIDFromContext(c))
	if err != nil {
		HandleError(c, err)
		return
	}
	result := make([]goalResponse, 0, len(goals))
	for i := range goals {
		result = append(result, toGoalResponse(&goals[i]))
	}
	RespondOK(c, http.StatusOK, result)
}

func (h *GoalHandler) Create(c *gin.Context) {
	var req createGoalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}
	deadline, err := parseOptionalDate(req.Deadline)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "INVALID_INPUT", "deadline має бути у форматі YYYY-MM-DD")
		return
	}

	goal, err := h.uc.Create(c.Request.Context(), middleware.UserIDFromContext(c), domain.CreateGoalInput{
		Title:        req.Title,
		TargetAmount: req.TargetAmount,
		Deadline:     deadline,
	})
	if err != nil {
		HandleError(c, err)
		return
	}
	RespondOK(c, http.StatusCreated, toGoalResponse(goal))
}

func (h *GoalHandler) Update(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	var req updateGoalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}
	deadline, err := parseOptionalDate(req.Deadline)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "INVALID_INPUT", "deadline має бути у форматі YYYY-MM-DD")
		return
	}

	goal, err := h.uc.Update(c.Request.Context(), middleware.UserIDFromContext(c), id, domain.UpdateGoalInput{
		Title:        req.Title,
		TargetAmount: req.TargetAmount,
		Deadline:     deadline,
	})
	if err != nil {
		HandleError(c, err)
		return
	}
	RespondOK(c, http.StatusOK, toGoalResponse(goal))
}

func (h *GoalHandler) Contribute(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	var req contributeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	goal, err := h.uc.Contribute(c.Request.Context(), middleware.UserIDFromContext(c), id, req.Amount)
	if err != nil {
		HandleError(c, err)
		return
	}
	RespondOK(c, http.StatusOK, toGoalResponse(goal))
}

func (h *GoalHandler) Delete(c *gin.Context) {
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
