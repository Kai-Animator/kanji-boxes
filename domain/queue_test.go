package domain

import "testing"

func TestIsDue(t *testing.T) {
	today := mustDate(t, "2026-02-04")

	if !IsDue(Card{NextDue: mustDate(t, "2026-02-04")}, today) {
		t.Fatalf("expected card due when nextDue is today")
	}

	if IsDue(Card{NextDue: mustDate(t, "2026-02-05")}, today) {
		t.Fatalf("expected card not due when nextDue is after today")
	}
}

func TestDueCardsOrdersByNextDueAndKanji(t *testing.T) {
	today := mustDate(t, "2026-02-04")
	cards := []Card{
		{ID: "1", Kanji: "b", NextDue: mustDate(t, "2026-02-03")},
		{ID: "2", Kanji: "a", NextDue: mustDate(t, "2026-02-03")},
		{ID: "3", Kanji: "c", NextDue: mustDate(t, "2026-02-05")},
	}

	due := DueCards(cards, today)
	if len(due) != 2 {
		t.Fatalf("expected 2 due cards, got %d", len(due))
	}

	if due[0].Kanji != "a" || due[1].Kanji != "b" {
		t.Fatalf("unexpected order: %s, %s", due[0].Kanji, due[1].Kanji)
	}
}
