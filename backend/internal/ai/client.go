package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"finagent/backend/internal/domain"
)

type Client struct {
	baseURL       string
	internalToken string
	httpClient    *http.Client
}

func NewClient(baseURL, internalToken string) *Client {
	return &Client{
		baseURL:       baseURL,
		internalToken: internalToken,
		httpClient:    &http.Client{Timeout: 30 * time.Second},
	}
}

// --- /categorize ---

type categorizeItem struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type categorizeRequest struct {
	Description         string           `json:"description"`
	TransactionType     string           `json:"transaction_type"`
	AvailableCategories []categorizeItem `json:"available_categories"`
}

type CategorizeResult struct {
	CategoryID   *int64  `json:"category_id"`
	Confidence   float64 `json:"confidence"`
	CategoryName *string `json:"category_name"`
}

func (c *Client) Categorize(ctx context.Context, description string, txType domain.TransactionType, categories []domain.Category) (*CategorizeResult, error) {
	items := make([]categorizeItem, 0, len(categories))
	for _, cat := range categories {
		if string(cat.Type) == string(txType) {
			items = append(items, categorizeItem{ID: cat.ID, Name: cat.Name, Type: string(cat.Type)})
		}
	}

	req := categorizeRequest{
		Description:         description,
		TransactionType:     string(txType),
		AvailableCategories: items,
	}

	var result CategorizeResult
	if err := c.post(ctx, "/categorize", req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// --- /chat ---

type ChatTransaction struct {
	ID              int64   `json:"id"`
	CategoryID      *int64  `json:"category_id"`
	CategoryName    *string `json:"category_name"`
	Type            string  `json:"type"`
	Amount          float64 `json:"amount"`
	Description     string  `json:"description"`
	TransactionDate string  `json:"transaction_date"`
}

type ChatGoal struct {
	ID            int64   `json:"id"`
	Title         string  `json:"title"`
	TargetAmount  float64 `json:"target_amount"`
	CurrentAmount float64 `json:"current_amount"`
	Deadline      *string `json:"deadline"`
}

type ChatCategory struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type ChatHistoryMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Message      string            `json:"message"`
	UserID       int64             `json:"user_id"`
	History      []ChatHistoryMsg  `json:"history"`
	Transactions []ChatTransaction `json:"transactions"`
	Goals        []ChatGoal        `json:"goals"`
	Categories   []ChatCategory    `json:"categories"`
}

type ChatActionData struct {
	ActionType      string   `json:"action_type"`
	Amount          float64  `json:"amount"`
	TransactionType string   `json:"transaction_type"`
	Description     string   `json:"description"`
	CategoryName    string   `json:"category_name"`
	TransactionDate string   `json:"transaction_date"`
	TransactionID   int64    `json:"transaction_id"`
	Title           string   `json:"title"`
	TargetAmount    float64  `json:"target_amount"`
	Deadline        *string  `json:"deadline"`
	GoalID          int64    `json:"goal_id"`
}

type ChatResult struct {
	Response string          `json:"response"`
	Intent   string          `json:"intent"`
	Action   *ChatActionData `json:"action,omitempty"`
}

func (c *Client) Chat(ctx context.Context, req ChatRequest) (*ChatResult, error) {
	var result ChatResult
	if err := c.post(ctx, "/chat", req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// --- /audit ---

type AuditTransaction struct {
	ID              int64   `json:"id"`
	CategoryID      *int64  `json:"category_id"`
	CategoryName    *string `json:"category_name"`
	Type            string  `json:"type"`
	Amount          float64 `json:"amount"`
	Description     string  `json:"description"`
	TransactionDate string  `json:"transaction_date"`
}

type auditRequest struct {
	UserID       int64              `json:"user_id"`
	Transactions []AuditTransaction `json:"transactions"`
}

type AnomalyItem struct {
	TransactionID int64   `json:"transaction_id"`
	Description   string  `json:"description"`
	CategoryName  string  `json:"category_name"`
	Amount        float64 `json:"amount"`
	Mean          float64 `json:"mean"`
	Reason        string  `json:"reason"`
}

type AuditResult struct {
	Anomalies []AnomalyItem `json:"anomalies"`
}

func (c *Client) Audit(ctx context.Context, userID int64, txs []AuditTransaction) (*AuditResult, error) {
	req := auditRequest{UserID: userID, Transactions: txs}
	var result AuditResult
	if err := c.post(ctx, "/audit", req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) post(ctx context.Context, path string, body, result any) error {
	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("ai client: marshal: %w", err)
	}

	const attempts = 3
	var lastErr error
	for attempt := 0; attempt < attempts; attempt++ {
		if attempt > 0 {
			// Лінійний backoff із повагою до скасування контексту.
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Duration(attempt) * 300 * time.Millisecond):
			}
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(data))
		if err != nil {
			return fmt.Errorf("ai client: new request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Internal-Token", c.internalToken)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("ai client: do request: %w", err) // мережна помилка — повтор
			continue
		}

		if resp.StatusCode >= 500 {
			b, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			lastErr = fmt.Errorf("ai client: status %d: %s", resp.StatusCode, b) // 5xx — повтор
			continue
		}
		if resp.StatusCode >= 400 {
			b, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return fmt.Errorf("ai client: status %d: %s", resp.StatusCode, b) // 4xx — без повтору
		}

		err = json.NewDecoder(resp.Body).Decode(result)
		resp.Body.Close()
		return err
	}
	return lastErr
}
