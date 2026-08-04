package model_test

import (
	"testing"

	"github.com/programmerpark/subscription-aggregator/internal/model"
)

func TestOverlapMonths(t *testing.T) {
	from := model.NewYearMonth(2025, 7)
	to := model.NewYearMonth(2025, 9)

	t.Run("open-ended covers full period", func(t *testing.T) {
		start := model.NewYearMonth(2025, 1)
		got := model.OverlapMonths(start, nil, from, to)
		if got != 3 {
			t.Fatalf("got %d, want 3", got)
		}
	})

	t.Run("partial overlap", func(t *testing.T) {
		start := model.NewYearMonth(2025, 8)
		end := model.NewYearMonth(2025, 12)
		got := model.OverlapMonths(start, &end, from, to)
		if got != 2 {
			t.Fatalf("got %d, want 2", got)
		}
	})

	t.Run("no overlap", func(t *testing.T) {
		start := model.NewYearMonth(2024, 1)
		end := model.NewYearMonth(2024, 12)
		got := model.OverlapMonths(start, &end, from, to)
		if got != 0 {
			t.Fatalf("got %d, want 0", got)
		}
	})
}

func TestParseYearMonth(t *testing.T) {
	ym, err := model.ParseYearMonth("07-2025")
	if err != nil {
		t.Fatal(err)
	}
	if ym.String() != "07-2025" {
		t.Fatalf("got %s", ym.String())
	}
}
