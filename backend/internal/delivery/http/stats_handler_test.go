package http

import (
	"bytes"
	"testing"
	"time"

	"finagent/backend/internal/domain"
)

func TestWriteTransactionCSV(t *testing.T) {
	rows := []domain.TransactionExportRow{
		{
			TransactionDate: time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC),
			Type:            "expense",
			CategoryName:    "Їжа",
			Amount:          1250.5,
			Description:     "Продукти",
		},
		{
			TransactionDate: time.Date(2026, 6, 14, 0, 0, 0, 0, time.UTC),
			Type:            "income",
			Amount:          5000,
			Description:     "Переказ",
		},
	}

	var out bytes.Buffer
	if err := writeTransactionCSV(&out, rows); err != nil {
		t.Fatalf("writeTransactionCSV() error = %v", err)
	}

	want := "Дата;Опис;Категорія;Тип;Сума, грн\r\n" +
		"15.06.2026;Продукти;Їжа;Витрата;1250,50\r\n" +
		"14.06.2026;Переказ;Без категорії;Дохід;5000,00\r\n"
	if out.String() != want {
		t.Fatalf("unexpected CSV:\n%s", out.String())
	}
}
