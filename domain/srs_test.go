package domain

import (
	"testing"
	"time"
)

func TestMoveBox(t *testing.T) {
	if got := MoveBox(5, true, 2*time.Second); got != 6 {
		t.Fatalf("expected move to box 6, got %d", got)
	}

	if got := MoveBox(5, true, 4*time.Second); got != 5 {
		t.Fatalf("expected stay at box 5 for slow correct answer, got %d", got)
	}

	if got := MoveBox(1, false, 10*time.Second); got != 1 {
		t.Fatalf("expected stay at box 1, got %d", got)
	}

	if got := MoveBox(7, true, 2*time.Second); got != 6 {
		t.Fatalf("expected clamp to box 6, got %d", got)
	}
}

func TestBoxIntervalDays(t *testing.T) {
	tests := []struct {
		box      int
		expected int
	}{
		{box: 1, expected: 1},
		{box: 2, expected: 3},
		{box: 3, expected: 7},
		{box: 4, expected: 14},
		{box: 5, expected: 30},
		{box: 6, expected: 60},
		{box: 0, expected: 1},
		{box: 99, expected: 60},
	}

	for _, tc := range tests {
		if got := BoxIntervalDays(tc.box); got != tc.expected {
			t.Fatalf("expected %d days for box %d, got %d", tc.expected, tc.box, got)
		}
	}
}

func TestNextDueForBox(t *testing.T) {
	today := mustDate(t, "2026-02-04")

	got, err := NextDueForBox(today, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.String() != "2026-02-11" {
		t.Fatalf("expected nextDue 2026-02-11, got %s", got.String())
	}

	got, err = NextDueForBox(today, 6)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.String() != "2026-04-05" {
		t.Fatalf("expected nextDue 2026-04-05, got %s", got.String())
	}
}

func TestApplyReviewCorrect(t *testing.T) {
	today := mustDate(t, "2026-02-04")
	card := Card{Box: 2, NextDue: mustDate(t, "2026-02-01")}

	updated, err := ApplyReview(card, true, 2*time.Second, today)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Box != 3 {
		t.Fatalf("expected box 3, got %d", updated.Box)
	}
	if updated.LastReviewed == nil || updated.LastReviewed.String() != "2026-02-04" {
		t.Fatalf("expected lastReviewed 2026-02-04")
	}
	if updated.NextDue.String() != "2026-02-11" {
		t.Fatalf("expected nextDue 2026-02-11, got %s", updated.NextDue.String())
	}
	if updated.ReviewCount != 1 || updated.CorrectCount != 1 {
		t.Fatalf("expected review/correct counts to increment")
	}
}

func TestApplyReviewIncorrect(t *testing.T) {
	today := mustDate(t, "2026-02-04")
	card := Card{Box: 1, NextDue: mustDate(t, "2026-02-01")}

	updated, err := ApplyReview(card, false, 2*time.Second, today)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Box != 1 {
		t.Fatalf("expected box 1, got %d", updated.Box)
	}
	if updated.ReviewCount != 1 || updated.CorrectCount != 0 {
		t.Fatalf("expected only review count to increment")
	}
}

func TestApplyReviewCorrectSlowStaysInSameBox(t *testing.T) {
	today := mustDate(t, "2026-02-04")
	card := Card{Box: 2, NextDue: mustDate(t, "2026-02-01")}

	updated, err := ApplyReview(card, true, 4*time.Second, today)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Box != 2 {
		t.Fatalf("expected box 2, got %d", updated.Box)
	}
	if updated.NextDue.String() != "2026-02-07" {
		t.Fatalf("expected nextDue 2026-02-07, got %s", updated.NextDue.String())
	}
}
