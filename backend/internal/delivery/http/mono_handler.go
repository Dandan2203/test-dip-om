package http

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"finagent/backend/internal/delivery/http/middleware"
	"finagent/backend/internal/domain"
	"finagent/backend/internal/mono"
	"finagent/backend/internal/usecase"
)

type MonoHandler struct {
	uc *usecase.MonoUsecase
}

func NewMonoHandler(uc *usecase.MonoUsecase) *MonoHandler {
	return &MonoHandler{uc: uc}
}

type monoConnectRequest struct {
	Token string `json:"token" binding:"required"`
}

type monoAccountResponse struct {
	ID           string  `json:"id"`
	Type         string  `json:"type"`
	Balance      float64 `json:"balance"`
	CurrencyCode int     `json:"currencyCode"`
	MaskedPan    string  `json:"maskedPan"`
	IBAN         string  `json:"iban"`
}

func toAccountResponses(accounts []mono.Account) []monoAccountResponse {
	out := make([]monoAccountResponse, 0, len(accounts))
	for _, a := range accounts {
		pan := ""
		if len(a.MaskedPan) > 0 {
			pan = a.MaskedPan[0]
		}
		out = append(out, monoAccountResponse{
			ID:           a.ID,
			Type:         a.Type,
			Balance:      float64(a.Balance) / 100.0,
			CurrencyCode: a.CurrencyCode,
			MaskedPan:    pan,
			IBAN:         a.IBAN,
		})
	}
	return out
}

func (h *MonoHandler) Connect(c *gin.Context) {
	var req monoConnectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	info, err := h.uc.Connect(c.Request.Context(), middleware.UserIDFromContext(c), req.Token)
	if err != nil {
		h.handleMonoError(c, err)
		return
	}
	RespondOK(c, http.StatusOK, gin.H{
		"name":     info.Name,
		"accounts": toAccountResponses(info.Accounts),
	})
}

func (h *MonoHandler) Disconnect(c *gin.Context) {
	if err := h.uc.Disconnect(c.Request.Context(), middleware.UserIDFromContext(c)); err != nil {
		HandleError(c, err)
		return
	}
	RespondOK(c, http.StatusOK, gin.H{"disconnected": true})
}

func (h *MonoHandler) Status(c *gin.Context) {
	conn, err := h.uc.Status(c.Request.Context(), middleware.UserIDFromContext(c))
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			RespondOK(c, http.StatusOK, gin.H{"connected": false})
			return
		}
		HandleError(c, err)
		return
	}
	RespondOK(c, http.StatusOK, gin.H{
		"connected":    true,
		"connectedAt":  conn.ConnectedAt,
		"lastImportAt": conn.LastImportAt,
	})
}

func (h *MonoHandler) Accounts(c *gin.Context) {
	info, err := h.uc.Accounts(c.Request.Context(), middleware.UserIDFromContext(c))
	if err != nil {
		h.handleMonoError(c, err)
		return
	}
	RespondOK(c, http.StatusOK, gin.H{"name": info.Name, "accounts": toAccountResponses(info.Accounts)})
}

func (h *MonoHandler) Currency(c *gin.Context) {
	rates, err := h.uc.Currency(c.Request.Context())
	if err != nil {
		h.handleMonoError(c, err)
		return
	}
	RespondOK(c, http.StatusOK, rates)
}

type monoImportRequest struct {
	Account      string `json:"account" binding:"required"`
	CurrencyCode int    `json:"currencyCode"`
	From         string `json:"from"`
	To           string `json:"to"`
}

func (h *MonoHandler) Import(c *gin.Context) {
	var req monoImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	to := time.Now()
	from := to.AddDate(0, 0, -31)
	if req.From != "" {
		if t, err := time.Parse(dateLayout, req.From); err == nil {
			from = t
		}
	}
	if req.To != "" {
		if t, err := time.Parse(dateLayout, req.To); err == nil {
			to = t.Add(24 * time.Hour)
		}
	}
	if req.CurrencyCode == 0 {
		req.CurrencyCode = 980
	}

	imported, err := h.uc.Import(c.Request.Context(), middleware.UserIDFromContext(c), req.Account, req.CurrencyCode, from, to)
	if err != nil {
		h.handleMonoError(c, err)
		return
	}
	RespondOK(c, http.StatusOK, gin.H{"imported": imported})
}

func (h *MonoHandler) handleMonoError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, mono.ErrInvalidToken):
		RespondError(c, http.StatusBadRequest, "MONO_INVALID_TOKEN", "недійсний токен Monobank")
	case errors.Is(err, mono.ErrRateLimited):
		RespondError(c, http.StatusTooManyRequests, "MONO_RATE_LIMITED", "перевищено ліміт Monobank, спробуйте за хвилину")
	case errors.Is(err, domain.ErrNotFound):
		RespondError(c, http.StatusBadRequest, "MONO_NOT_CONNECTED", "Monobank не підключено")
	default:
		HandleError(c, err)
	}
}
