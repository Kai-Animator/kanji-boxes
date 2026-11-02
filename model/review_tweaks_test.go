package model

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Kai-Animator/kanji-boxes/domain"
	"github.com/Kai-Animator/kanji-boxes/storage"
)

func saveReviewCards(t *testing.T, cards []domain.Card) string {
	t.Helper()
	dataDir := tempDataDir(t)
	if err := os.Setenv(storage.EnvDataDir, dataDir); err != nil {
		t.Fatalf("set env: %v", err)
	}
	if err := storage.SaveCards(storage.CardsPath(dataDir), cards); err != nil {
		t.Fatalf("save cards: %v", err)
	}
	return dataDir
}

func TestApplyAnswerMeasuresTimeToFlipNotToAnswer(t *testing.T) {
	today := domain.Today()
	saveReviewCards(t, []domain.Card{
		{ID: "due", Kanji: "日", English: "day", Box: 2, CreatedAt: today, NextDue: today},
	})

	start := time.Date(2026, 2, 4, 12, 0, 0, 0, time.UTC)
	m := New()
	m.now = func() time.Time { return start }
	m = m.startReview()
	m.screen = screenReview
	// 2秒で裏返し、確認に時間をかけてから正解を押す
	m.now = func() time.Time { return start.Add(2 * time.Second) }
	m, _ = updateModel(m, keyMsg(" "))
	m.now = func() time.Time { return start.Add(10 * time.Second) }
	m = m.applyAnswer(true)

	if got := m.cards[0].Box; got != 3 {
		t.Fatalf("expected promotion based on flip time, got box %d", got)
	}
}

func TestFlipTimeIsTakenFromFirstFlipOnly(t *testing.T) {
	today := domain.Today()
	saveReviewCards(t, []domain.Card{
		{ID: "due", Kanji: "日", English: "day", Box: 2, CreatedAt: today, NextDue: today},
	})

	start := time.Date(2026, 2, 4, 12, 0, 0, 0, time.UTC)
	m := New()
	m.now = func() time.Time { return start }
	m = m.startReview()
	m.screen = screenReview
	m.now = func() time.Time { return start.Add(4 * time.Second) }
	m, _ = updateModel(m, keyMsg(" "))
	m, _ = updateModel(m, keyMsg(" "))
	m.now = func() time.Time { return start.Add(1 * time.Second) }
	m, _ = updateModel(m, keyMsg(" "))
	m = m.applyAnswer(true)

	if got := m.cards[0].Box; got != 2 {
		t.Fatalf("expected slow first flip to block promotion, got box %d", got)
	}
}

func TestReviewViewFreezesElapsedAfterFlip(t *testing.T) {
	card := domain.Card{Kanji: "日", English: "day", Box: 2}
	start := time.Date(2026, 2, 4, 12, 0, 0, 0, time.UTC)
	m := New()
	m.reviewCardAt = start
	m.reviewFlipAt = start.Add(1500 * time.Millisecond)
	m.reveal = true
	m.now = func() time.Time { return start.Add(30 * time.Second) }
	m.reviewQueue = []domain.Card{card}
	m.reviewModes = []reviewMode{reviewModeRecognition}

	view := m.reviewView()
	if !strings.Contains(view, "1.5s") {
		t.Fatalf("expected elapsed frozen at flip time, got %q", view)
	}
}

func TestReviewViewHintSaysNoSpeedGateInBoxOne(t *testing.T) {
	m := New()
	m.reviewCardAt = time.Now()
	m.reviewQueue = []domain.Card{{Kanji: "日", English: "day", Box: 1}}
	m.reviewModes = []reviewMode{reviewModeRecognition}

	view := m.reviewView()
	if strings.Contains(view, "<=3s advances") {
		t.Fatalf("box 1 should not show the speed hint, got %q", view)
	}
	if !strings.Contains(view, "no speed gate in box 1") {
		t.Fatalf("expected box 1 hint, got %q", view)
	}
}

func TestApplyAnswerTracksSlowCorrectCards(t *testing.T) {
	today := domain.Today()
	saveReviewCards(t, []domain.Card{
		{ID: "slow", Kanji: "日", English: "day", Box: 2, CreatedAt: today, NextDue: today},
	})

	start := time.Date(2026, 2, 4, 12, 0, 0, 0, time.UTC)
	m := New()
	m.now = func() time.Time { return start }
	m = m.startReview()
	m.now = func() time.Time { return start.Add(5 * time.Second) }
	m = m.applyAnswer(true)

	if len(m.sessionSlow) != 1 || m.sessionSlow[0].ID != "slow" {
		t.Fatalf("expected slow card tracked, got %+v", m.sessionSlow)
	}
	if len(m.sessionMisses) != 0 {
		t.Fatal("slow correct is not a miss")
	}
	if !strings.Contains(m.lastFeedback, "slow") || !strings.Contains(m.lastFeedback, "5.0s") {
		t.Fatalf("expected slow feedback with elapsed, got %q", m.lastFeedback)
	}
}

func TestApplyAnswerFastCorrectClearsFeedback(t *testing.T) {
	today := domain.Today()
	saveReviewCards(t, []domain.Card{
		{ID: "fast", Kanji: "日", English: "day", Box: 2, CreatedAt: today, NextDue: today},
	})

	start := time.Date(2026, 2, 4, 12, 0, 0, 0, time.UTC)
	m := New()
	m.now = func() time.Time { return start }
	m = m.startReview()
	m.lastFeedback = "stale"
	m.now = func() time.Time { return start.Add(1 * time.Second) }
	m = m.applyAnswer(true)

	if m.lastFeedback != "" {
		t.Fatalf("expected feedback cleared, got %q", m.lastFeedback)
	}
	if len(m.sessionSlow) != 0 {
		t.Fatal("fast correct should not be tracked as slow")
	}
}

func TestReviewViewShowsLastFeedbackOnNextCard(t *testing.T) {
	m := New()
	m.reviewCardAt = time.Now()
	m.reviewQueue = []domain.Card{{Kanji: "日", English: "day", Box: 2}, {Kanji: "月", English: "moon", Box: 2}}
	m.reviewModes = []reviewMode{reviewModeRecognition, reviewModeRecognition}
	m.reviewIndex = 1
	m.lastFeedback = "Correct but slow (5.0s): 日 stays in box 2"

	view := m.reviewView()
	if !strings.Contains(view, "Correct but slow (5.0s)") {
		t.Fatalf("expected feedback line, got %q", view)
	}
}

func TestReviewViewCompleteListsSlowCards(t *testing.T) {
	slow := domain.Card{ID: "slow", Kanji: "日", English: "day", Box: 2}
	m := New()
	m.reviewQueue = []domain.Card{slow}
	m.reviewModes = []reviewMode{reviewModeRecognition}
	m.reviewIndex = 1
	m.sessionSlow = []domain.Card{slow}

	view := m.reviewView()
	if !strings.Contains(view, "Correct but slow (1)") {
		t.Fatalf("expected slow heading, got %q", view)
	}
	if !strings.Contains(view, "日") {
		t.Fatalf("expected slow card listed, got %q", view)
	}
}

func TestStartReviewResetsSlowAndFeedback(t *testing.T) {
	saveReviewCards(t, []domain.Card{})
	m := New()
	m.sessionSlow = []domain.Card{{ID: "stale"}}
	m.lastFeedback = "stale"
	m = m.startReview()
	if len(m.sessionSlow) != 0 || m.lastFeedback != "" {
		t.Fatal("expected slow list and feedback reset")
	}
}

func TestReviewViewShowsLeechWarning(t *testing.T) {
	m := New()
	m.reviewCardAt = time.Now()
	m.reviewQueue = []domain.Card{{Kanji: "収支", English: "balance", Box: 1, ReviewCount: 11, CorrectCount: 0}}
	m.reviewModes = []reviewMode{reviewModeRecognition}

	view := m.reviewView()
	if !strings.Contains(view, "Leech") || !strings.Contains(view, "11 reviews") || !strings.Contains(view, "0% correct") {
		t.Fatalf("expected leech warning with stats, got %q", view)
	}
	if !strings.Contains(view, "s: suspend") {
		t.Fatalf("expected suspend hint, got %q", view)
	}
}

func TestReviewViewHidesLeechWarningForHealthyCard(t *testing.T) {
	m := New()
	m.reviewCardAt = time.Now()
	m.reviewQueue = []domain.Card{{Kanji: "日", English: "day", Box: 1, ReviewCount: 11, CorrectCount: 9}}
	m.reviewModes = []reviewMode{reviewModeRecognition}

	if view := m.reviewView(); strings.Contains(view, "Leech") {
		t.Fatalf("did not expect leech warning, got %q", view)
	}
}

func TestSuspendKeyInReviewPersistsAndSkips(t *testing.T) {
	today := domain.Today()
	dataDir := saveReviewCards(t, []domain.Card{
		{ID: "leech", Kanji: "収支", English: "balance", Box: 1, CreatedAt: today, NextDue: today, ReviewCount: 11},
	})

	m := New()
	m = m.startReview()
	m.screen = screenReview
	m, _ = updateModel(m, keyMsg("s"))

	if m.reviewIndex != 1 {
		t.Fatalf("expected review to advance past suspended card, got index %d", m.reviewIndex)
	}
	if m.session.Reviewed != 0 {
		t.Fatal("suspending is not a review")
	}
	saved, err := storage.LoadCards(storage.CardsPath(dataDir))
	if err != nil {
		t.Fatal(err)
	}
	if !saved[0].Suspended {
		t.Fatal("expected suspended flag persisted")
	}
}

func TestStartReviewExcludesSuspendedCards(t *testing.T) {
	today := domain.Today()
	saveReviewCards(t, []domain.Card{
		{ID: "susp", Kanji: "収支", English: "balance", Box: 1, CreatedAt: today, NextDue: today, Suspended: true},
		{ID: "due", Kanji: "日", English: "day", Box: 1, CreatedAt: today, NextDue: today},
	})

	m := New()
	m = m.startReview()
	if len(m.reviewQueue) != 1 || m.reviewQueue[0].ID != "due" {
		t.Fatalf("expected only unsuspended card, got %+v", m.reviewQueue)
	}
}

func TestBrowseViewMarksLeechAndSuspended(t *testing.T) {
	today := domain.Today()
	saveReviewCards(t, []domain.Card{
		{ID: "leech", Kanji: "収支", English: "balance", Box: 1, CreatedAt: today, NextDue: today, ReviewCount: 11},
		{ID: "susp", Kanji: "方策", English: "measure", Box: 1, CreatedAt: today, NextDue: today, Suspended: true},
	})

	m := New()
	m = m.startBrowse()
	view := m.browseView()
	if !strings.Contains(view, "[S]") {
		t.Fatalf("expected suspended marker, got %q", view)
	}
	if !strings.Contains(view, "leech") {
		t.Fatalf("expected leech marker, got %q", view)
	}
	if !strings.Contains(view, "0% correct") {
		t.Fatalf("expected accuracy in details, got %q", view)
	}
}

func TestBrowseSuspendKeyToggles(t *testing.T) {
	today := domain.Today()
	dataDir := saveReviewCards(t, []domain.Card{
		{ID: "a", Kanji: "収支", English: "balance", Box: 1, CreatedAt: today, NextDue: today},
	})

	m := New()
	m = m.startBrowse()
	m.screen = screenBrowse
	m, _ = updateModel(m, keyMsg("s"))

	saved, err := storage.LoadCards(storage.CardsPath(dataDir))
	if err != nil {
		t.Fatal(err)
	}
	if !saved[0].Suspended {
		t.Fatal("expected card suspended after first press")
	}
	if !m.browseCards[0].Suspended {
		t.Fatal("expected browse list refreshed")
	}

	m, _ = updateModel(m, keyMsg("s"))
	saved, _ = storage.LoadCards(storage.CardsPath(dataDir))
	if saved[0].Suspended {
		t.Fatal("expected card unsuspended after second press")
	}
}
