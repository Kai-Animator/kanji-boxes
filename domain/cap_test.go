package domain

import (
	"fmt"
	"testing"
	"time"
)

func TestActiveCardsCapsBoxOneNewcomersByCreationOrder(t *testing.T) {
	cards := []Card{
		{ID: "late", Box: 1, CreatedAt: mustDate(t, "2026-10-08")},
		{ID: "early", Box: 1, CreatedAt: mustDate(t, "2026-10-01")},
		{ID: "mid", Box: 1, CreatedAt: mustDate(t, "2026-10-05")},
		{ID: "relearn", Box: 1, Relearning: true, CreatedAt: mustDate(t, "2026-10-09")},
		{ID: "box3", Box: 3, CreatedAt: mustDate(t, "2026-10-09")},
	}

	active := ActiveCards(cards, 2)
	got := map[string]bool{}
	for _, card := range active {
		got[card.ID] = true
	}
	for _, id := range []string{"early", "mid", "relearn", "box3"} {
		if !got[id] {
			t.Errorf("expected %s to be active", id)
		}
	}
	if got["late"] {
		t.Error("expected newest box 1 card to wait")
	}
	if waiting := WaitingCount(cards, 2); waiting != 1 {
		t.Errorf("expected 1 waiting, got %d", waiting)
	}
	if position := WaitingQueue(cards, 2)["late"]; position != 1 {
		t.Errorf("expected late to be #1 in line, got %d", position)
	}
}

func TestActiveCardsFillsFreedSlot(t *testing.T) {
	cards := make([]Card, 0, 3)
	for i := 0; i < 3; i++ {
		cards = append(cards, Card{ID: fmt.Sprintf("c%d", i), Box: 1, CreatedAt: mustDate(t, "2026-10-08")})
	}
	if len(ActiveCards(cards, 2)) != 2 {
		t.Fatal("expected 2 active before promotion")
	}
	// c0 が昇格すると c2 が枠に入る
	cards[0].Box = 2
	active := ActiveCards(cards, 2)
	if len(active) != 3 {
		t.Fatalf("expected promoted card plus 2 newcomers, got %d", len(active))
	}
}

func TestApplyReviewTracksRelearning(t *testing.T) {
	today := mustDate(t, "2026-10-09")

	dropped, err := ApplyReview(Card{Box: 2}, false, time.Second, today)
	if err != nil {
		t.Fatal(err)
	}
	if dropped.Box != 1 || !dropped.Relearning {
		t.Fatalf("expected box 1 relearning, got box=%d relearning=%v", dropped.Box, dropped.Relearning)
	}

	// ボックス1内で不正解でも再学習フラグは維持
	stillLow, _ := ApplyReview(dropped, false, time.Second, today)
	if !stillLow.Relearning {
		t.Error("expected relearning kept while in box 1")
	}

	promoted, _ := ApplyReview(dropped, true, time.Second, today)
	if promoted.Box != 2 || promoted.Relearning {
		t.Errorf("expected flag cleared on promotion, got box=%d relearning=%v", promoted.Box, promoted.Relearning)
	}

	fresh, _ := ApplyReview(Card{Box: 1}, false, time.Second, today)
	if fresh.Relearning {
		t.Error("new box 1 miss should not be relearning")
	}
}
