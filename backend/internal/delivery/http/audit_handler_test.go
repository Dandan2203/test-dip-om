package http

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestAnomalyResponseUsesFrontendFieldNames(t *testing.T) {
	data, err := json.Marshal(anomalyResponse{TransactionID: 7, CategoryName: "Інше"})
	if err != nil {
		t.Fatal(err)
	}

	body := string(data)
	if !strings.Contains(body, `"transactionId":7`) || !strings.Contains(body, `"categoryName":"Інше"`) {
		t.Fatalf("unexpected anomaly response: %s", body)
	}
}
