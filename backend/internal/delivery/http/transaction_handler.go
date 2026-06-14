package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"finagent/backend/internal/delivery/http/middleware"
	"finagent/backend/internal/domain"
)

const dateLayout = "2006-01-02"

type TransactionHandler struct {
	uc domain.TransactionUsecase
}

func NewTransactionHandler(uc domain.TransactionUsecase) *TransactionHandler {
	return &TransactionHandler{uc: uc}
}

type createTransactionRequest struct {
	CategoryID      *int64  `json:"categoryId"`
	Type            string  `json:"type"            binding:"required,oneof=income expense"`
	Amount          float64 `json:"amount"          binding:"required,gt=0"`
	Description     string  `json:"description"`
	TransactionDate string  `json:"transactionDate" binding:"required"`
}

type updateTransactionRequest struct {
	CategoryID      *int64  `json:"categoryId"`
	Type            string  `json:"type"            binding:"required,oneof=income expense"`
	Amount          float64 `json:"amount"          binding:"required,gt=0"`
	Description     string  `json:"description"`
	TransactionDate string  `json:"transactionDate" binding:"required"`
}

type transactionResponse struct {
	ID              int64    `json:"id"`
	CategoryID      *int64   `json:"categoryId"`
	Type            string   `json:"type"`
	Amount          float64  `json:"amount"`
	Description     string   `json:"description"`
	TransactionDate string   `json:"transactionDate"`
	IsAICategorized bool     `json:"isAiCategorized"`
	CreatedAt       string   `json:"createdAt"`
	Source          string   `json:"source"`
	CurrencyCode    int      `json:"currencyCode"`
	OriginalAmount  *float64 `json:"originalAmount"`
}

type transactionListResponse struct {
	Items []transactionResponse `json:"items"`
	Total int64                 `json:"total"`
}

func toTransactionResponse(tx *domain.Transaction) transactionResponse {
	return transactionResponse{
		ID:              tx.ID,
		CategoryID:      tx.CategoryID,
		Type:            string(tx.Type),
		Amount:          tx.Amount,
		Description:     tx.Description,
		TransactionDate: tx.TransactionDate.Format(dateLayout),
		IsAICategorized: tx.IsAICategorized,
		CreatedAt:       tx.CreatedAt.Format(time.RFC3339),
		Source:          tx.Source,
		CurrencyCode:    tx.CurrencyCode,
		OriginalAmount:  tx.OriginalAmount,
	}
}

func (h *TransactionHandler) Create(c *gin.Context) {
	var req createTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	date, err := time.Parse(dateLayout, req.TransactionDate)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "INVALID_INPUT", "transactionDate має бути у форматі YYYY-MM-DD")
		return
	}

	tx, err := h.uc.Create(c.Request.Context(), middleware.UserIDFromContext(c), domain.CreateTransactionInput{
		CategoryID:      req.CategoryID,
		Type:            domain.TransactionType(req.Type),
		Amount:          req.Amount,
		Description:     req.Description,
		TransactionDate: date,
	})
	if err != nil {
		HandleError(c, err)
		return
	}
	RespondOK(c, http.StatusCreated, toTransactionResponse(tx))
}

func (h *TransactionHandler) List(c *gin.Context) {
	filter, ok := parseTransactionFilter(c)
	if !ok {
		return
	}

	txs, total, err := h.uc.List(c.Request.Context(), middleware.UserIDFromContext(c), filter)
	if err != nil {
		HandleError(c, err)
		return
	}

	items := make([]transactionResponse, 0, len(txs))
	for i := range txs {
		items = append(items, toTransactionResponse(&txs[i]))
	}
	RespondOK(c, http.StatusOK, transactionListResponse{Items: items, Total: total})
}

func (h *TransactionHandler) Update(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}

	var req updateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	date, err := time.Parse(dateLayout, req.TransactionDate)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "INVALID_INPUT", "transactionDate має бути у форматі YYYY-MM-DD")
		return
	}

	tx, err := h.uc.Update(c.Request.Context(), middleware.UserIDFromContext(c), id, domain.UpdateTransactionInput{
		CategoryID:      req.CategoryID,
		Type:            domain.TransactionType(req.Type),
		Amount:          req.Amount,
		Description:     req.Description,
		TransactionDate: date,
	})
	if err != nil {
		HandleError(c, err)
		return
	}
	RespondOK(c, http.StatusOK, toTransactionResponse(tx))
}

func (h *TransactionHandler) Delete(c *gin.Context) {
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

func parseTransactionFilter(c *gin.Context) (domain.TransactionFilter, bool) {
	var filter domain.TransactionFilter

	if s := c.Query("from"); s != "" {
		t, err := time.Parse(dateLayout, s)
		if err != nil {
			RespondError(c, http.StatusBadRequest, "INVALID_INPUT", "from має бути у форматі YYYY-MM-DD")
			return filter, false
		}
		filter.From = &t
	}
	if s := c.Query("to"); s != "" {
		t, err := time.Parse(dateLayout, s)
		if err != nil {
			RespondError(c, http.StatusBadRequest, "INVALID_INPUT", "to має бути у форматі YYYY-MM-DD")
			return filter, false
		}
		filter.To = &t
	}
	if s := c.Query("category_id"); s != "" {
		id, err := strconv.ParseInt(s, 10, 64)
		if err != nil || id <= 0 {
			RespondError(c, http.StatusBadRequest, "INVALID_INPUT", "category_id має бути цілим числом > 0")
			return filter, false
		}
		filter.CategoryID = &id
	}
	if s := c.Query("type"); s != "" {
		if s != "income" && s != "expense" {
			RespondError(c, http.StatusBadRequest, "INVALID_INPUT", "type має бути income або expense")
			return filter, false
		}
		t := domain.TransactionType(s)
		filter.Type = &t
	}

	filter.Limit, _ = strconv.Atoi(c.DefaultQuery("limit", "0"))
	filter.Offset, _ = strconv.Atoi(c.DefaultQuery("offset", "0"))
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	return filter, true
}
