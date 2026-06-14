package http

import "testing"

func TestGoalNameMatches(t *testing.T) {
	cases := []struct {
		name, needle string
		want         bool
	}{
		{"відпустка", "відпустку", true},
		{"відпустка", "відпустки", true},
		{"відпустка", "відпустка", true},
		{"авто", "авто", true},
		{"ремонт", "ремонту", true},
		{"відпустка", "авто", false},
	}
	for _, c := range cases {
		if got := goalNameMatches(c.name, c.needle); got != c.want {
			t.Errorf("goalNameMatches(%q,%q)=%v, want %v", c.name, c.needle, got, c.want)
		}
	}
}
