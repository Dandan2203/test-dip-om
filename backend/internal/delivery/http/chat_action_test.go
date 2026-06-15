package http

import (
	"context"
	"testing"

	"finagent/backend/internal/ai"
	"finagent/backend/internal/domain"
)

func TestResolvedChangingActionsRequireConfirmation(t *testing.T) {
	h := &ChatHandler{}
	goals := []domain.Goal{{ID: 7, UserID: 1, Title: "Відпустка"}}

	cases := []struct {
		name   string
		action ai.ChatActionData
	}{
		{
			name: "create transaction",
			action: ai.ChatActionData{
				ActionType:      "create_transaction",
				Amount:          100,
				TransactionType: "expense",
			},
		},
		{
			name: "create goal",
			action: ai.ChatActionData{
				ActionType:   "create_goal",
				Title:        "Ноутбук",
				TargetAmount: 50000,
			},
		},
		{
			name: "contribute goal",
			action: ai.ChatActionData{
				ActionType: "contribute_goal",
				Title:      "Відпустка",
				Amount:     500,
			},
		},
		{
			name: "withdraw goal",
			action: ai.ChatActionData{
				ActionType: "withdraw_goal",
				Title:      "Відпустка",
				Amount:     200,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pending, clarify := h.resolveAction(context.Background(), 1, &tc.action, nil, goals)
			if clarify != "" {
				t.Fatalf("unexpected clarification: %s", clarify)
			}
			if pending == nil || !pending.RequiresConfirm {
				t.Fatal("changing AI action must require confirmation")
			}
		})
	}
}
