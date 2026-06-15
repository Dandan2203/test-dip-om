package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"finagent/backend/internal/ai"
	"finagent/backend/internal/delivery/http/middleware"
	"finagent/backend/internal/domain"
)

type ChatHandler struct {
	aiClient    *ai.Client
	txRepo      domain.TransactionRepository
	txUC        domain.TransactionUsecase
	goalUC      domain.GoalUsecase
	categoryUC  domain.CategoryUsecase
	actionRepo  domain.ActionRepository
	chatLogRepo domain.ChatLogRepository
}

func NewChatHandler(
	aiClient *ai.Client,
	txRepo domain.TransactionRepository,
	txUC domain.TransactionUsecase,
	goalUC domain.GoalUsecase,
	categoryUC domain.CategoryUsecase,
	actionRepo domain.ActionRepository,
	chatLogRepo domain.ChatLogRepository,
) *ChatHandler {
	return &ChatHandler{
		aiClient:    aiClient,
		txRepo:      txRepo,
		txUC:        txUC,
		goalUC:      goalUC,
		categoryUC:  categoryUC,
		actionRepo:  actionRepo,
		chatLogRepo: chatLogRepo,
	}
}

const (
	maxChatMessageLen = 500 // максимум символів на один запит до чату
	dailyChatLimit    = 50  // максимум повідомлень до чату на добу на користувача
)

type chatRequest struct {
	Message string           `json:"message" binding:"required,max=500"`
	History []chatHistoryMsg `json:"history" binding:"max=10,dive"`
}

// chatHistoryMsg — попередня репліка діалогу (короткий контекст для LLM).
type chatHistoryMsg struct {
	Role    string `json:"role" binding:"required,oneof=user assistant"`
	Content string `json:"content" binding:"required,max=4000"`
}

// PendingAction — дія, розпізнана ШІ й підготована до виконання,
// яку клієнт показує користувачу для підтвердження (так/ні) і повертає назад.
type PendingAction struct {
	ActionType      string  `json:"actionType"`
	EntityType      string  `json:"entityType"`
	Amount          float64 `json:"amount,omitempty"`
	TransactionType string  `json:"transactionType,omitempty"`
	Description     string  `json:"description,omitempty"`
	CategoryID      *int64  `json:"categoryId,omitempty"`
	CategoryName    string  `json:"categoryName,omitempty"`
	TransactionDate string  `json:"transactionDate,omitempty"`
	TransactionID   int64   `json:"transactionId,omitempty"`
	GoalID          int64   `json:"goalId,omitempty"`
	GoalTitle       string  `json:"goalTitle,omitempty"`
	TargetAmount    float64 `json:"targetAmount,omitempty"`
	Deadline        *string `json:"deadline,omitempty"`
	ConfirmText     string  `json:"confirmText"`
	// RequiresConfirm — чи показувати «Так/Ні». true лише для створення цілі
	// та видалень; решта дій виконуються одразу з можливістю undo.
	RequiresConfirm bool `json:"requiresConfirm"`
}

func (h *ChatHandler) Chat(c *gin.Context) {
	var req chatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "INVALID_INPUT",
			fmt.Sprintf("повідомлення не може бути порожнім або довшим за %d символів", maxChatMessageLen))
		return
	}

	ctx := c.Request.Context()
	userID := middleware.UserIDFromContext(c)

	// Денний ліміт повідомлень (захист бюджету AI). Атомарний інкремент рахує СПРОБУ
	// до виклику ШІ — без гонок між паралельними запитами й з урахуванням невдалих викликів.
	if h.chatLogRepo != nil {
		if n, err := h.chatLogRepo.IncrementDailyUsage(ctx, userID); err == nil && n > dailyChatLimit {
			RespondError(c, http.StatusTooManyRequests, "DAILY_LIMIT",
				fmt.Sprintf("Денний ліміт %d повідомлень вичерпано. Спробуйте завтра.", dailyChatLimit))
			return
		}
	}

	txs, _, err := h.txRepo.List(ctx, userID, domain.TransactionFilter{Limit: 100, Offset: 0})
	if err != nil {
		HandleError(c, err)
		return
	}

	goals, err := h.goalUC.List(ctx, userID)
	if err != nil {
		HandleError(c, err)
		return
	}

	categories, err := h.categoryUC.List(ctx, userID)
	if err != nil {
		HandleError(c, err)
		return
	}

	catMap := make(map[int64]string, len(categories))
	for _, cat := range categories {
		catMap[cat.ID] = cat.Name
	}

	aiTxs := make([]ai.ChatTransaction, 0, len(txs))
	for _, t := range txs {
		var catName *string
		if t.CategoryID != nil {
			if name, ok := catMap[*t.CategoryID]; ok {
				catName = &name
			}
		}
		aiTxs = append(aiTxs, ai.ChatTransaction{
			ID:              t.ID,
			CategoryID:      t.CategoryID,
			CategoryName:    catName,
			Type:            string(t.Type),
			Amount:          t.Amount,
			Description:     t.Description,
			TransactionDate: t.TransactionDate.Format("2006-01-02"),
		})
	}

	aiGoals := make([]ai.ChatGoal, 0, len(goals))
	for _, g := range goals {
		var dl *string
		if g.Deadline != nil {
			s := g.Deadline.Format("2006-01-02")
			dl = &s
		}
		aiGoals = append(aiGoals, ai.ChatGoal{
			ID:            g.ID,
			Title:         g.Title,
			TargetAmount:  g.TargetAmount,
			CurrentAmount: g.CurrentAmount,
			Deadline:      dl,
		})
	}

	aiCats := make([]ai.ChatCategory, 0, len(categories))
	for _, cat := range categories {
		aiCats = append(aiCats, ai.ChatCategory{
			ID:   cat.ID,
			Name: cat.Name,
			Type: string(cat.Type),
		})
	}

	aiHistory := make([]ai.ChatHistoryMsg, 0, len(req.History))
	for _, m := range req.History {
		aiHistory = append(aiHistory, ai.ChatHistoryMsg{Role: m.Role, Content: m.Content})
	}

	result, err := h.aiClient.Chat(ctx, ai.ChatRequest{
		Message:      req.Message,
		UserID:       userID,
		History:      aiHistory,
		Transactions: aiTxs,
		Goals:        aiGoals,
		Categories:   aiCats,
	})
	if err != nil {
		RespondError(c, http.StatusServiceUnavailable, "AI_UNAVAILABLE", "ШІ-сервіс недоступний")
		return
	}

	resp := gin.H{
		"response":  result.Response,
		"intent":    result.Intent,
		"timestamp": time.Now().Format(time.RFC3339),
	}

	// Дію НЕ виконуємо одразу — готуємо її на підтвердження користувачем.
	if result.Action != nil && result.Action.ActionType != "" {
		pending, clarify := h.resolveAction(ctx, userID, result.Action, categories, goals)
		if pending != nil {
			resp["pendingAction"] = pending
		} else if clarify != "" {
			resp["response"] = clarify
		}
	}

	// Лог звернення (best-effort): зберігаємо лише текст відповіді й інтент,
	// без деталей pendingAction (суми, категорії, цілі) — мінімізуємо PII у БД.
	if h.chatLogRepo != nil {
		logged, _ := json.Marshal(gin.H{"response": resp["response"], "intent": result.Intent})
		_ = h.chatLogRepo.Record(ctx, &domain.ChatLog{
			UserID:   userID,
			Message:  req.Message,
			Intent:   result.Intent,
			Response: logged,
		})
	}

	RespondOK(c, http.StatusOK, resp)
}

// Execute виконує дію, ПІДТВЕРДЖЕНУ користувачем (кнопка «Так»).
func (h *ChatHandler) Execute(c *gin.Context) {
	var p PendingAction
	if err := c.ShouldBindJSON(&p); err != nil {
		RespondError(c, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	ctx := c.Request.Context()
	userID := middleware.UserIDFromContext(c)

	executed, err := h.runAction(ctx, userID, &p)
	if err != nil {
		HandleError(c, err)
		return
	}

	resp := gin.H{"executed": true}
	if executed != nil {
		resp["action"] = gin.H{
			"actionType": string(executed.ActionType),
			"entityType": string(executed.EntityType),
			"entityId":   executed.EntityID,
		}
	}
	RespondOK(c, http.StatusOK, resp)
}

// History повертає збережену історію чату користувача, розгорнуту в окремі
// повідомлення (user/assistant) для відновлення панелі після перезавантаження.
func (h *ChatHandler) History(c *gin.Context) {
	userID := middleware.UserIDFromContext(c)

	logs, err := h.chatLogRepo.History(c.Request.Context(), userID, 50)
	if err != nil {
		HandleError(c, err)
		return
	}

	type histMsg struct {
		Role      string `json:"role"`
		Content   string `json:"content"`
		Intent    string `json:"intent,omitempty"`
		Timestamp string `json:"timestamp"`
	}

	messages := make([]histMsg, 0, len(logs)*2)
	for _, l := range logs {
		ts := l.CreatedAt.Format(time.RFC3339)
		messages = append(messages, histMsg{Role: "user", Content: l.Message, Timestamp: ts})

		var payload struct {
			Response string `json:"response"`
		}
		_ = json.Unmarshal(l.Response, &payload)
		if payload.Response != "" {
			messages = append(messages, histMsg{
				Role: "assistant", Content: payload.Response, Intent: l.Intent, Timestamp: ts,
			})
		}
	}

	RespondOK(c, http.StatusOK, gin.H{"messages": messages})
}

// resolveAction зіставляє розпізнану ШІ дію з реальними сутностями користувача
// й формує текст підтвердження. Повертає (дію, "") або (nil, "уточнення").
func (h *ChatHandler) resolveAction(
	ctx context.Context,
	userID int64,
	action *ai.ChatActionData,
	categories []domain.Category,
	goals []domain.Goal,
) (*PendingAction, string) {
	switch action.ActionType {
	case "create_transaction":
		p := &PendingAction{
			ActionType:      "create_transaction",
			EntityType:      "transaction",
			Amount:          action.Amount,
			TransactionType: action.TransactionType,
			Description:     action.Description,
			TransactionDate: action.TransactionDate,
		}
		if action.CategoryName != "" {
			needle := strings.ToLower(action.CategoryName)
			for _, cat := range categories {
				if strings.ToLower(cat.Name) == needle {
					id := cat.ID
					p.CategoryID = &id
					p.CategoryName = cat.Name
					break
				}
			}
		}
		verb := "витрату"
		if action.TransactionType == "income" {
			verb = "дохід"
		}
		desc := ""
		if action.Description != "" {
			desc = fmt.Sprintf(" («%s»)", action.Description)
		}
		p.ConfirmText = fmt.Sprintf("Додати %s %s ₴%s?", verb, money(action.Amount), desc)
		return p, ""

	case "contribute_goal":
		g := findGoal(goals, action.GoalID, action.Title)
		if g == nil {
			return nil, fmt.Sprintf("Не знайшов ціль «%s». Уточніть назву або спершу створіть ціль.", action.Title)
		}
		p := &PendingAction{
			ActionType: "contribute_goal",
			EntityType: "goal",
			Amount:     action.Amount,
			GoalID:     g.ID,
			GoalTitle:  g.Title,
		}
		p.ConfirmText = fmt.Sprintf("Додати %s ₴ до цілі «%s»?", money(action.Amount), g.Title)
		return p, ""

	case "withdraw_goal":
		g := findGoal(goals, action.GoalID, action.Title)
		if g == nil {
			return nil, fmt.Sprintf("Не знайшов ціль «%s». Уточніть назву.", action.Title)
		}
		p := &PendingAction{
			ActionType: "withdraw_goal",
			EntityType: "goal",
			Amount:     action.Amount,
			GoalID:     g.ID,
			GoalTitle:  g.Title,
		}
		p.ConfirmText = fmt.Sprintf("Зняти %s ₴ з цілі «%s»?", money(action.Amount), g.Title)
		return p, ""

	case "create_goal":
		p := &PendingAction{
			ActionType:      "create_goal",
			EntityType:      "goal",
			GoalTitle:       action.Title,
			TargetAmount:    action.TargetAmount,
			Deadline:        action.Deadline,
			RequiresConfirm: true,
		}
		p.ConfirmText = fmt.Sprintf("Створити ціль «%s» на %s ₴?", action.Title, money(action.TargetAmount))
		return p, ""

	case "delete_transaction":
		if action.TransactionID == 0 {
			return nil, "Уточніть, яку саме транзакцію видалити."
		}
		tx, err := h.txRepo.GetByID(ctx, action.TransactionID)
		if err != nil || tx.UserID != userID {
			return nil, "Не знайшов таку транзакцію."
		}
		p := &PendingAction{
			ActionType:      "delete_transaction",
			EntityType:      "transaction",
			TransactionID:   action.TransactionID,
			Description:     tx.Description,
			Amount:          tx.Amount,
			RequiresConfirm: true,
		}
		p.ConfirmText = fmt.Sprintf("Видалити транзакцію на %s ₴?", money(tx.Amount))
		return p, ""

	case "delete_goal":
		g := findGoal(goals, action.GoalID, action.Title)
		if g == nil {
			return nil, "Не знайшов таку ціль."
		}
		p := &PendingAction{
			ActionType:      "delete_goal",
			EntityType:      "goal",
			GoalID:          g.ID,
			GoalTitle:       g.Title,
			RequiresConfirm: true,
		}
		p.ConfirmText = fmt.Sprintf("Видалити ціль «%s»?", g.Title)
		return p, ""
	}
	return nil, ""
}

// runAction виконує підтверджену дію та (для create/delete) пише її в action_log.
func (h *ChatHandler) runAction(ctx context.Context, userID int64, p *PendingAction) (*domain.ActionLog, error) {
	switch p.ActionType {
	case "create_transaction":
		date := time.Now()
		if p.TransactionDate != "" {
			if d, err := time.Parse("2006-01-02", p.TransactionDate); err == nil {
				date = d
			}
		}
		tx, err := h.txUC.Create(ctx, userID, domain.CreateTransactionInput{
			CategoryID:      p.CategoryID,
			Type:            domain.TransactionType(p.TransactionType),
			Amount:          p.Amount,
			Description:     p.Description,
			TransactionDate: date,
		})
		if err != nil {
			return nil, err
		}
		log := &domain.ActionLog{
			UserID:     userID,
			ActionType: domain.ActionCreate,
			EntityType: domain.EntityTransaction,
			EntityID:   tx.ID,
			Payload:    mustJSON(tx),
		}
		_ = h.actionRepo.Record(ctx, log)
		return log, nil

	case "contribute_goal":
		if _, err := h.goalUC.Contribute(ctx, userID, p.GoalID, p.Amount); err != nil {
			return nil, err
		}
		log := &domain.ActionLog{
			UserID:     userID,
			ActionType: domain.ActionContribute,
			EntityType: domain.EntityGoal,
			EntityID:   p.GoalID,
			Payload:    mustJSON(map[string]float64{"delta": p.Amount}),
		}
		_ = h.actionRepo.Record(ctx, log)
		return log, nil

	case "withdraw_goal":
		// Зняття = від'ємний внесок; репозиторій не дає піти нижче 0.
		if _, err := h.goalUC.Contribute(ctx, userID, p.GoalID, -p.Amount); err != nil {
			return nil, err
		}
		log := &domain.ActionLog{
			UserID:     userID,
			ActionType: domain.ActionContribute,
			EntityType: domain.EntityGoal,
			EntityID:   p.GoalID,
			Payload:    mustJSON(map[string]float64{"delta": -p.Amount}),
		}
		_ = h.actionRepo.Record(ctx, log)
		return log, nil

	case "create_goal":
		var dl *time.Time
		if p.Deadline != nil && *p.Deadline != "" {
			if d, err := time.Parse("2006-01-02", *p.Deadline); err == nil {
				dl = &d
			}
		}
		g, err := h.goalUC.Create(ctx, userID, domain.CreateGoalInput{
			Title:        p.GoalTitle,
			TargetAmount: p.TargetAmount,
			Deadline:     dl,
		})
		if err != nil {
			return nil, err
		}
		log := &domain.ActionLog{
			UserID:     userID,
			ActionType: domain.ActionCreate,
			EntityType: domain.EntityGoal,
			EntityID:   g.ID,
			Payload:    mustJSON(g),
		}
		_ = h.actionRepo.Record(ctx, log)
		return log, nil

	case "delete_transaction":
		tx, err := h.txRepo.GetByID(ctx, p.TransactionID)
		if err != nil || tx.UserID != userID {
			return nil, domain.ErrNotFound
		}
		payload := mustJSON(tx)
		if err := h.txUC.Delete(ctx, userID, p.TransactionID); err != nil {
			return nil, err
		}
		log := &domain.ActionLog{
			UserID:     userID,
			ActionType: domain.ActionDelete,
			EntityType: domain.EntityTransaction,
			EntityID:   p.TransactionID,
			Payload:    payload,
		}
		_ = h.actionRepo.Record(ctx, log)
		return log, nil

	case "delete_goal":
		goals, _ := h.goalUC.List(ctx, userID)
		var found *domain.Goal
		for i := range goals {
			if goals[i].ID == p.GoalID {
				found = &goals[i]
				break
			}
		}
		if found == nil {
			return nil, domain.ErrNotFound
		}
		payload := mustJSON(found)
		if err := h.goalUC.Delete(ctx, userID, p.GoalID); err != nil {
			return nil, err
		}
		log := &domain.ActionLog{
			UserID:     userID,
			ActionType: domain.ActionDelete,
			EntityType: domain.EntityGoal,
			EntityID:   p.GoalID,
			Payload:    payload,
		}
		_ = h.actionRepo.Record(ctx, log)
		return log, nil
	}
	return nil, nil
}

// findGoal шукає ціль за ID, інакше за назвою (без регістру, частковий збіг).
func findGoal(goals []domain.Goal, id int64, title string) *domain.Goal {
	if id != 0 {
		for i := range goals {
			if goals[i].ID == id {
				return &goals[i]
			}
		}
	}
	needle := strings.ToLower(strings.TrimSpace(title))
	if needle == "" {
		return nil
	}
	for i := range goals {
		if goalNameMatches(strings.ToLower(goals[i].Title), needle) {
			return &goals[i]
		}
	}
	return nil
}

// goalNameMatches зіставляє назви цілей з урахуванням українських відмінків:
// порівнює за основою без останньої літери («відпустку» ≈ «Відпустка»).
func goalNameMatches(name, needle string) bool {
	if name == needle || strings.Contains(name, needle) || strings.Contains(needle, name) {
		return true
	}
	stem := func(s string) string {
		r := []rune(s)
		if len(r) > 4 {
			return string(r[:len(r)-1])
		}
		return s
	}
	a, b := stem(name), stem(needle)
	return strings.HasPrefix(a, b) || strings.HasPrefix(b, a)
}

func money(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}

func mustJSON(v any) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}
