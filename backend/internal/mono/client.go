package mono

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

const baseURL = "https://api.monobank.ua"

type Client struct {
	httpClient *http.Client

	rateMu     sync.Mutex
	rateCache  []CurrencyRate
	rateExpiry time.Time
}

func NewClient() *Client {
	return &Client{httpClient: &http.Client{Timeout: 20 * time.Second}}
}

type Account struct {
	ID           string   `json:"id"`
	SendID       string   `json:"sendId"`
	Balance      int64    `json:"balance"`      // у копійках
	CreditLimit  int64    `json:"creditLimit"`  // у копійках
	Type         string   `json:"type"`         // black, white, platinum, fop
	CurrencyCode int      `json:"currencyCode"` // ISO 4217 (980 = UAH)
	MaskedPan    []string `json:"maskedPan"`
	IBAN         string   `json:"iban"`
}

type ClientInfo struct {
	Name     string    `json:"name"`
	Accounts []Account `json:"accounts"`
}

type CurrencyRate struct {
	CurrencyCodeA int     `json:"currencyCodeA"`
	CurrencyCodeB int     `json:"currencyCodeB"`
	Date          int64   `json:"date"`
	RateSell      float64 `json:"rateSell"`
	RateBuy       float64 `json:"rateBuy"`
	RateCross     float64 `json:"rateCross"`
}

type StatementItem struct {
	ID              string `json:"id"`
	Time            int64  `json:"time"`
	Description     string `json:"description"`
	MCC             int    `json:"mcc"`
	Amount          int64  `json:"amount"`          // від'ємне у копійках
	OperationAmount int64  `json:"operationAmount"` // у валюті операції
	CurrencyCode    int    `json:"currencyCode"`
	Comment         string `json:"comment"`
}

func (c *Client) ClientInfo(ctx context.Context, token string) (*ClientInfo, error) {
	var info ClientInfo
	if err := c.get(ctx, "/personal/client-info", token, &info); err != nil {
		return nil, err
	}
	return &info, nil
}

func (c *Client) Statement(ctx context.Context, token, account string, from, to time.Time) ([]StatementItem, error) {
	path := fmt.Sprintf("/personal/statement/%s/%d/%d", account, from.Unix(), to.Unix())
	var items []StatementItem
	if err := c.get(ctx, path, token, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func (c *Client) Currency(ctx context.Context) ([]CurrencyRate, error) {
	c.rateMu.Lock()
	defer c.rateMu.Unlock()

	if time.Now().Before(c.rateExpiry) && c.rateCache != nil {
		return c.rateCache, nil
	}

	var rates []CurrencyRate
	if err := c.get(ctx, "/bank/currency", "", &rates); err != nil {
		if c.rateCache != nil {
			return c.rateCache, nil
		}
		return nil, err
	}

	c.rateCache = rates
	c.rateExpiry = time.Now().Add(30 * time.Minute)
	return rates, nil
}

func (c *Client) get(ctx context.Context, path, token string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+path, nil)
	if err != nil {
		return fmt.Errorf("mono: запит: %w", err)
	}
	if token != "" {
		req.Header.Set("X-Token", token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("mono: виконання запиту: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == http.StatusTooManyRequests {
		return ErrRateLimited
	}
	if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusUnauthorized {
		return ErrInvalidToken
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("mono: статус %d: %s", resp.StatusCode, body)
	}

	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("mono: розбір відповіді: %w", err)
	}
	return nil
}
