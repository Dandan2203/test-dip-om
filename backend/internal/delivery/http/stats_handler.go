package http

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"finagent/backend/internal/delivery/http/middleware"
	"finagent/backend/internal/domain"
)

type StatsHandler struct {
	uc     domain.StatsUsecase
	goalUC domain.GoalUsecase
}

func NewStatsHandler(uc domain.StatsUsecase, goalUC domain.GoalUsecase) *StatsHandler {
	return &StatsHandler{uc: uc, goalUC: goalUC}
}

func (h *StatsHandler) Summary(c *gin.Context) {
	period, ok := parsePeriod(c)
	if !ok {
		return
	}
	summary, err := h.uc.Summary(c.Request.Context(), middleware.UserIDFromContext(c), period)
	if err != nil {
		HandleError(c, err)
		return
	}
	RespondOK(c, http.StatusOK, gin.H{
		"income":  summary.Income,
		"expense": summary.Expense,
		"balance": summary.Balance,
	})
}

func (h *StatsHandler) ByCategory(c *gin.Context) {
	period, ok := parsePeriod(c)
	if !ok {
		return
	}
	typeParam := c.DefaultQuery("type", "expense")
	if typeParam != "income" && typeParam != "expense" {
		RespondError(c, http.StatusBadRequest, "INVALID_INPUT", "type має бути income або expense")
		return
	}

	stats, err := h.uc.ByCategory(c.Request.Context(), middleware.UserIDFromContext(c),
		domain.TransactionType(typeParam), period)
	if err != nil {
		HandleError(c, err)
		return
	}
	RespondOK(c, http.StatusOK, stats)
}

type dashboardTransactionResponse struct {
	ID              int64   `json:"id"`
	CategoryID      *int64  `json:"categoryId"`
	Type            string  `json:"type"`
	Amount          float64 `json:"amount"`
	Description     string  `json:"description"`
	TransactionDate string  `json:"transactionDate"`
}

func (h *StatsHandler) Dashboard(c *gin.Context) {
	data, err := h.uc.Dashboard(c.Request.Context(), middleware.UserIDFromContext(c), h.goalUC)
	if err != nil {
		HandleError(c, err)
		return
	}

	recentResp := make([]dashboardTransactionResponse, 0, len(data.RecentTransactions))
	for _, t := range data.RecentTransactions {
		recentResp = append(recentResp, dashboardTransactionResponse{
			ID:              t.ID,
			CategoryID:      t.CategoryID,
			Type:            string(t.Type),
			Amount:          t.Amount,
			Description:     t.Description,
			TransactionDate: t.TransactionDate.Format(dateLayout),
		})
	}

	goalsResp := make([]goalResponse, 0, len(data.Goals))
	for i := range data.Goals {
		goalsResp = append(goalsResp, toGoalResponse(&data.Goals[i]))
	}

	RespondOK(c, http.StatusOK, gin.H{
		"income":             data.Income,
		"expense":            data.Expense,
		"balance":            data.Balance,
		"recentTransactions": recentResp,
		"goals":              goalsResp,
	})
}

func (h *StatsHandler) Export(c *gin.Context) {
	filter, ok := parseTransactionFilter(c)
	if !ok {
		return
	}

	rows, err := h.uc.Export(c.Request.Context(), middleware.UserIDFromContext(c), filter)
	if err != nil {
		HandleError(c, err)
		return
	}

	filename := fmt.Sprintf("finagent_export_%s.csv", time.Now().Format("2006-01-02"))
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Header("Content-Type", "text/csv; charset=utf-8")

	// BOM для коректного відкриття в Excel
	c.Writer.Write([]byte("\xEF\xBB\xBF"))

	w := csv.NewWriter(c.Writer)
	_ = w.Write([]string{"Дата", "Тип", "Категорія", "Сума", "Опис"})

	typeLabel := map[string]string{"income": "Дохід", "expense": "Витрата"}
	for _, r := range rows {
		_ = w.Write([]string{
			r.TransactionDate.Format(dateLayout),
			typeLabel[r.Type],
			r.CategoryName,
			fmt.Sprintf("%.2f", r.Amount),
			r.Description,
		})
	}
	w.Flush()
}

func parsePeriod(c *gin.Context) (domain.StatsPeriod, bool) {
	now := time.Now().UTC()
	defaultFrom := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	var period domain.StatsPeriod

	if s := c.Query("from"); s != "" {
		t, err := time.Parse(dateLayout, s)
		if err != nil {
			RespondError(c, http.StatusBadRequest, "INVALID_INPUT", "from має бути у форматі YYYY-MM-DD")
			return period, false
		}
		period.From = t
	} else {
		period.From = defaultFrom
	}

	if s := c.Query("to"); s != "" {
		t, err := time.Parse(dateLayout, s)
		if err != nil {
			RespondError(c, http.StatusBadRequest, "INVALID_INPUT", "to має бути у форматі YYYY-MM-DD")
			return period, false
		}
		period.To = t
	} else {
		period.To = now
	}

	return period, true
}
