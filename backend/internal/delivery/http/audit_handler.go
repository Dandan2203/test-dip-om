package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"finagent/backend/internal/ai"
	"finagent/backend/internal/delivery/http/middleware"
	"finagent/backend/internal/domain"
)

type AuditHandler struct {
	aiClient *ai.Client
	txRepo   domain.TransactionRepository
	catUC    domain.CategoryUsecase
}

type anomalyResponse struct {
	TransactionID int64   `json:"transactionId"`
	Description   string  `json:"description"`
	CategoryName  string  `json:"categoryName"`
	Amount        float64 `json:"amount"`
	Mean          float64 `json:"mean"`
	Reason        string  `json:"reason"`
}

func NewAuditHandler(aiClient *ai.Client, txRepo domain.TransactionRepository, catUC domain.CategoryUsecase) *AuditHandler {
	return &AuditHandler{aiClient: aiClient, txRepo: txRepo, catUC: catUC}
}

func (h *AuditHandler) Anomalies(c *gin.Context) {
	ctx := c.Request.Context()
	userID := middleware.UserIDFromContext(c)

	txs, _, err := h.txRepo.List(ctx, userID, domain.TransactionFilter{Limit: 200, Offset: 0})
	if err != nil {
		HandleError(c, err)
		return
	}

	categories, err := h.catUC.List(ctx, userID)
	if err != nil {
		HandleError(c, err)
		return
	}

	catMap := make(map[int64]string, len(categories))
	for _, cat := range categories {
		catMap[cat.ID] = cat.Name
	}

	aiTxs := make([]ai.AuditTransaction, 0, len(txs))
	for _, t := range txs {
		var catName *string
		if t.CategoryID != nil {
			if name, ok := catMap[*t.CategoryID]; ok {
				catName = &name
			}
		}
		aiTxs = append(aiTxs, ai.AuditTransaction{
			ID:              t.ID,
			CategoryID:      t.CategoryID,
			CategoryName:    catName,
			Type:            string(t.Type),
			Amount:          t.Amount,
			Description:     t.Description,
			TransactionDate: t.TransactionDate.Format("2006-01-02"),
		})
	}

	result, err := h.aiClient.Audit(ctx, userID, aiTxs)
	if err != nil {
		RespondError(c, http.StatusServiceUnavailable, "AI_UNAVAILABLE", "ШІ-сервіс недоступний")
		return
	}

	items := make([]anomalyResponse, 0, len(result.Anomalies))
	for _, item := range result.Anomalies {
		items = append(items, anomalyResponse{
			TransactionID: item.TransactionID,
			Description:   item.Description,
			CategoryName:  item.CategoryName,
			Amount:        item.Amount,
			Mean:          item.Mean,
			Reason:        item.Reason,
		})
	}

	RespondOK(c, http.StatusOK, items)
}
