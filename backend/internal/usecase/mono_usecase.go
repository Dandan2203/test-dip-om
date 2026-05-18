package usecase

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"finagent/backend/internal/domain"
	"finagent/backend/internal/mono"
	"finagent/backend/internal/pkg/crypto"
)

const monoImportCooldown = 60 * time.Second

type MonoUsecase struct {
	repo          domain.MonoRepository
	client        *mono.Client
	txRepo        domain.TransactionRepository
	catRepo       domain.CategoryRepository
	encryptionKey string
}

func NewMonoUsecase(
	repo domain.MonoRepository,
	client *mono.Client,
	txRepo domain.TransactionRepository,
	catRepo domain.CategoryRepository,
	encryptionKey string,
) *MonoUsecase {
	return &MonoUsecase{repo: repo, client: client, txRepo: txRepo, catRepo: catRepo, encryptionKey: encryptionKey}
}

// Збереження токена Mono.
func (uc *MonoUsecase) Connect(ctx context.Context, userID int64, token string) (*mono.ClientInfo, error) {
	token = strings.TrimSpace(token)

	info, err := uc.client.ClientInfo(ctx, token)
	if err != nil {
		return nil, err
	}

	encrypted, err := crypto.Encrypt(token, uc.encryptionKey)
	if err != nil {
		return nil, err
	}
	if err := uc.repo.Save(ctx, userID, encrypted); err != nil {
		return nil, err
	}
	return info, nil
}

func (uc *MonoUsecase) Disconnect(ctx context.Context, userID int64) error {
	return uc.repo.Delete(ctx, userID)
}

func (uc *MonoUsecase) Status(ctx context.Context, userID int64) (*domain.MonoConnection, error) {
	return uc.repo.Get(ctx, userID)
}

func (uc *MonoUsecase) Accounts(ctx context.Context, userID int64) (*mono.ClientInfo, error) {
	token, err := uc.token(ctx, userID)
	if err != nil {
		return nil, err
	}
	return uc.client.ClientInfo(ctx, token)
}

func (uc *MonoUsecase) Currency(ctx context.Context) ([]mono.CurrencyRate, error) {
	return uc.client.Currency(ctx)
}

func (uc *MonoUsecase) token(ctx context.Context, userID int64) (string, error) {
	conn, err := uc.repo.Get(ctx, userID)
	if err != nil {
		return "", err
	}
	return crypto.Decrypt(conn.TokenEncrypted, uc.encryptionKey)
}

// Імпорт виписки Mono.
func (uc *MonoUsecase) Import(ctx context.Context, userID int64, accountID string, accountCurrency int, from, to time.Time) (int, error) {
	conn, err := uc.repo.Get(ctx, userID)
	if err != nil {
		return 0, err
	}
	if conn.LastImportAt != nil && time.Since(*conn.LastImportAt) < monoImportCooldown {
		return 0, mono.ErrRateLimited
	}

	token, err := crypto.Decrypt(conn.TokenEncrypted, uc.encryptionKey)
	if err != nil {
		return 0, err
	}

	items, err := uc.client.Statement(ctx, token, accountID, from, to)
	if err != nil {
		return 0, err
	}

	catByName := uc.expenseCategoryMap(ctx, userID)
	rate := uc.uahRate(ctx, accountCurrency)

	imported := 0
	for i := range items {
		it := items[i]
		amountMajor := float64(it.Amount) / 100.0
		abs := math.Abs(amountMajor)
		uah := math.Round(abs*rate*100) / 100

		txType := domain.TransactionExpense
		var categoryID *int64
		if amountMajor >= 0 {
			txType = domain.TransactionIncome
		} else if id, ok := catByName[mono.MCCToCategory(it.MCC)]; ok {
			categoryID = &id
		}

		desc := strings.TrimSpace(it.Description)
		if it.Comment != "" {
			desc = strings.TrimSpace(desc + " — " + it.Comment)
		}

		externalID := it.ID
		original := abs
		tx := &domain.Transaction{
			UserID:          userID,
			CategoryID:      categoryID,
			Type:            txType,
			Amount:          uah,
			Description:     desc,
			TransactionDate: time.Unix(it.Time, 0),
			Source:          "monobank",
			ExternalID:      &externalID,
			CurrencyCode:    accountCurrency,
			OriginalAmount:  &original,
		}
		ok, err := uc.txRepo.CreateImported(ctx, tx)
		if err != nil {
			return imported, err
		}
		if ok {
			imported++
		}
	}

	if err := uc.repo.TouchImport(ctx, userID); err != nil {
		return imported, fmt.Errorf("оновлення часу імпорту: %w", err)
	}
	return imported, nil
}

func (uc *MonoUsecase) expenseCategoryMap(ctx context.Context, userID int64) map[string]int64 {
	cats, err := uc.catRepo.ListAvailable(ctx, userID)
	if err != nil {
		return map[string]int64{}
	}
	m := make(map[string]int64, len(cats))
	for _, c := range cats {
		if c.Type == domain.CategoryExpense {
			m[c.Name] = c.ID
		}
	}
	return m
}

func (uc *MonoUsecase) uahRate(ctx context.Context, currencyCode int) float64 {
	if currencyCode == 980 {
		return 1
	}
	rates, err := uc.client.Currency(ctx)
	if err != nil {
		return 1
	}
	for _, r := range rates {
		if r.CurrencyCodeA == currencyCode && r.CurrencyCodeB == 980 {
			switch {
			case r.RateCross > 0:
				return r.RateCross
			case r.RateSell > 0:
				return r.RateSell
			case r.RateBuy > 0:
				return r.RateBuy
			}
		}
	}
	return 1
}
