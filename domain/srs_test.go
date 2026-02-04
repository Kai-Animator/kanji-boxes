package domain

import "testing"

func TestMoveBox(t *testing.T) {
	if got := MoveBox(4, true); got != 5 {
		t.Fatalf("expected move to box 5, got %d", got)
	}

	if got := MoveBox(1, false); got != 1 {
		t.Fatalf("expected stay at box 1, got %d", got)
	}

	if got := MoveBox(6, true); got != 5 {
		t.Fatalf("expected clamp to box 5, got %d", got)
	}
}

func TestNextDueForBox(t *testing.T) {
	today := mustDate(t, "2026-02-04")

	got, err := NextDueForBox(today, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.String() != "2026-02-07" {
		t.Fatalf("expected nextDue 2026-02-07, got %s", got.String())
	}
}

func TestApplyReviewCorrect(t *testing.T) {
	today := mustDate(t, "2026-02-04")
	card := Card{Box: 2, NextDue: mustDate(t, "2026-02-01")}

	updated, err := ApplyReview(card, true, today)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Box != 3 {
		t.Fatalf("expected box 3, got %d", updated.Box)
	}
	if updated.LastReviewed == nil || updated.LastReviewed.String() != "2026-02-04" {
		t.Fatalf("expected lastReviewed 2026-02-04")
	}
	if updated.NextDue.String() != "2026-02-07" {
		t.Fatalf("expected nextDue 2026-02-07, got %s", updated.NextDue.String())
	}
	if updated.ReviewCount != 1 || updated.CorrectCount != 1 {
		t.Fatalf("expected review/correct counts to increment")
	}
}

func TestApplyReviewIncorrect(t *testing.T) {
	today := mustDate(t, "2026-02-04")
	card := Card{Box: 1, NextDue: mustDate(t, "2026-02-01")}

	updated, err := ApplyReview(card, false, today)
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
