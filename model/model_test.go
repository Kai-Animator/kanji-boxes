package model

import (
	"os"
	"path/filepath"
	"testing"

	"kanji-boxes/domain"
	"kanji-boxes/storage"
)

func TestStartReviewBuildsDueQueue(t *testing.T) {
	dataDir := tempDataDir(t)
	if err := os.Setenv(storage.EnvDataDir, dataDir); err != nil {
		t.Fatalf("set env: %v", err)
	}

	today := domain.Today()
	tomorrow, err := today.AddDays(1)
	if err != nil {
		t.Fatalf("add days: %v", err)
	}

	cards := []domain.Card{
		{ID: "due", Kanji: "日", English: "day", Box: domain.BoxMin, CreatedAt: today, NextDue: today},
		{ID: "later", Kanji: "月", English: "moon", Box: domain.BoxMin, CreatedAt: today, NextDue: tomorrow},
	}

	if err := storage.SaveCards(storage.CardsPath(dataDir), cards); err != nil {
		t.Fatalf("save cards: %v", err)
	}

	m := New()
	m = m.startReview()

	if len(m.reviewQueue) != 1 {
		t.Fatalf("expected 1 due card, got %d", len(m.reviewQueue))
	}
	if m.reviewQueue[0].ID != "due" {
		t.Fatalf("expected due card, got %s", m.reviewQueue[0].ID)
	}
}

func TestApplyAnswerUpdatesStatsAndPersists(t *testing.T) {
	dataDir := tempDataDir(t)
	if err := os.Setenv(storage.EnvDataDir, dataDir); err != nil {
		t.Fatalf("set env: %v", err)
	}

	today := domain.Today()
	cards := []domain.Card{
		{ID: "due", Kanji: "日", English: "day", Box: domain.BoxMin, CreatedAt: today, NextDue: today},
	}

	if err := storage.SaveCards(storage.CardsPath(dataDir), cards); err != nil {
		t.Fatalf("save cards: %v", err)
	}

	m := New()
	m = m.startReview()
	m = m.applyAnswer(true)

	if m.session.Reviewed != 1 || m.session.Correct != 1 || m.session.Incorrect != 0 {
		t.Fatalf("unexpected session stats: %+v", m.session)
	}

	state, err := storage.LoadState(storage.StatePath(dataDir))
	if err != nil {
		t.Fatalf("load state: %v", err)
	}
	stats := state.DailyStats[today.String()]
	if stats.Reviewed != 1 || stats.Correct != 1 || stats.Incorrect != 0 {
		t.Fatalf("unexpected persisted stats: %+v", stats)
	}
}

func TestEditCardUpdatesFields(t *testing.T) {
	dataDir := tempDataDir(t)
	if err := os.Setenv(storage.EnvDataDir, dataDir); err != nil {
		t.Fatalf("set env: %v", err)
	}

	today := domain.Today()
	cards := []domain.Card{
		{ID: "edit", Kanji: "日", English: "day", Box: domain.BoxMin, CreatedAt: today, NextDue: today},
	}
	if err := storage.SaveCards(storage.CardsPath(dataDir), cards); err != nil {
		t.Fatalf("save cards: %v", err)
	}

	m := New()
	m = m.startBrowse()
	m = m.startEditCard(m.browseCards[0])
	m.editInputs[0].SetValue("火")
	m.editInputs[1].SetValue("fire")
	m.editInputs[2].SetValue("ひ")
	m.editInputs[3].SetValue("火が燃える。")
	m.editInputs[4].SetValue("jlpt5|basic")
	m = m.saveEditCard()

	updated, err := storage.LoadCards(storage.CardsPath(dataDir))
	if err != nil {
		t.Fatalf("load cards: %v", err)
	}
	if updated[0].Kanji != "火" || updated[0].English != "fire" {
		t.Fatalf("unexpected card: %+v", updated[0])
	}
	if updated[0].Usage != "火が燃える。" {
		t.Fatalf("unexpected usage: %s", updated[0].Usage)
	}
	if updated[0].ID != "edit" {
		t.Fatalf("expected id preserved, got %s", updated[0].ID)
	}
}

func TestDeleteSelectedCardRemovesCard(t *testing.T) {
	dataDir := tempDataDir(t)
	if err := os.Setenv(storage.EnvDataDir, dataDir); err != nil {
		t.Fatalf("set env: %v", err)
	}

	today := domain.Today()
	cards := []domain.Card{
		{ID: "keep", Kanji: "日", English: "day", Box: domain.BoxMin, CreatedAt: today, NextDue: today},
		{ID: "delete", Kanji: "月", English: "moon", Box: domain.BoxMin, CreatedAt: today, NextDue: today},
	}
	if err := storage.SaveCards(storage.CardsPath(dataDir), cards); err != nil {
		t.Fatalf("save cards: %v", err)
	}

	m := New()
	m = m.startBrowse()
	m.browseCursor = 1
	m = m.deleteSelectedCard()

	remaining, err := storage.LoadCards(storage.CardsPath(dataDir))
	if err != nil {
		t.Fatalf("load cards: %v", err)
	}
	if len(remaining) != 1 || remaining[0].ID != "keep" {
		t.Fatalf("unexpected remaining cards: %+v", remaining)
	}
}

func TestImportExportAppendsAndExports(t *testing.T) {
	dataDir := tempDataDir(t)
	if err := os.Setenv(storage.EnvDataDir, dataDir); err != nil {
		t.Fatalf("set env: %v", err)
	}

	today := domain.Today()
	existing := []domain.Card{
		{ID: "existing", Kanji: "日", English: "day", Box: domain.BoxMin, CreatedAt: today, NextDue: today},
	}
	if err := storage.SaveCards(storage.CardsPath(dataDir), existing); err != nil {
		t.Fatalf("save cards: %v", err)
	}

	importPath := filepath.Join(dataDir, "import.csv")
	importData := "kanji,hiragana,english,tags\n月,つき,moon,jlpt5\n"
	if err := os.WriteFile(importPath, []byte(importData), 0o644); err != nil {
		t.Fatalf("write csv: %v", err)
	}

	m := New()
	m.ioInput.SetValue(importPath)
	m.ioAction = "import"
	m = m.performImportExport()

	combined, err := storage.LoadCards(storage.CardsPath(dataDir))
	if err != nil {
		t.Fatalf("load cards: %v", err)
	}
	if len(combined) != 2 {
		t.Fatalf("expected 2 cards, got %d", len(combined))
	}

	exportPath := filepath.Join(dataDir, "export.csv")
	m.ioInput.SetValue(exportPath)
	m.ioAction = "export"
	m = m.performImportExport()

	if _, err := os.Stat(exportPath); err != nil {
		t.Fatalf("export file missing: %v", err)
	}
}

func TestRecentDatesOrdersAndLimits(t *testing.T) {
	stats := map[string]domain.SessionStats{
		"2026-02-01": {Reviewed: 1},
		"2026-02-03": {Reviewed: 3},
		"2026-02-02": {Reviewed: 2},
	}

	dates := recentDates(stats, 2)
	if len(dates) != 2 {
		t.Fatalf("expected 2 dates, got %d", len(dates))
	}
	if dates[0] != "2026-02-03" || dates[1] != "2026-02-02" {
		t.Fatalf("unexpected order: %+v", dates)
	}
}

func tempDataDir(t *testing.T) string {
	t.Helper()
	dataDir, err := os.MkdirTemp("", "kanji-boxes-test-")
	if err != nil {
		t.Fatalf("mkdtemp: %v", err)
	}
	t.Cleanup(func() {
		_ = os.RemoveAll(dataDir)
	})
	return dataDir
}
