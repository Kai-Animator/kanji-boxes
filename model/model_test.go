package model

import (
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

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
	if len(m.reviewModes) != len(m.reviewQueue) {
		t.Fatalf("expected review modes for each due card, got %d modes for %d cards", len(m.reviewModes), len(m.reviewQueue))
	}
}

func TestReviewViewRecognitionMode(t *testing.T) {
	hiragana := "にち"
	card := domain.Card{Kanji: "日", Hiragana: &hiragana, English: "day", Usage: "日が昇る。"}

	m := New()
	m.reviewQueue = []domain.Card{card}
	m.reviewModes = []reviewMode{reviewModeRecognition}
	m.reviewIndex = 0

	front := m.reviewView()
	if !strings.Contains(front, "日") {
		t.Fatalf("expected recognition front to show kanji, got %q", front)
	}
	if !strings.Contains(front, "にち") {
		t.Fatalf("expected recognition front to show hiragana, got %q", front)
	}
	if strings.Contains(front, "が昇る。") {
		t.Fatalf("expected recognition front to hide usage sentence, got %q", front)
	}
	if strings.Contains(front, "day") {
		t.Fatalf("expected recognition front to hide english, got %q", front)
	}

	m.reveal = true
	back := m.reviewView()
	if !strings.Contains(back, "day") || !strings.Contains(back, "にち") {
		t.Fatalf("expected recognition back to show hiragana and english, got %q", back)
	}
	if !strings.Contains(back, "が昇る。") {
		t.Fatalf("expected recognition back to show usage sentence, got %q", back)
	}
}

func TestReviewViewProductionMode(t *testing.T) {
	hiragana := "にち"
	card := domain.Card{Kanji: "日", Hiragana: &hiragana, English: "day", Usage: "日が昇る。"}

	m := New()
	m.reviewQueue = []domain.Card{card}
	m.reviewModes = []reviewMode{reviewModeProduction}
	m.reviewIndex = 0

	front := m.reviewView()
	if !strings.Contains(front, "day") {
		t.Fatalf("expected production front to show english, got %q", front)
	}
	if strings.Contains(front, "にち") {
		t.Fatalf("expected production front to hide japanese, got %q", front)
	}
	if strings.Contains(front, "が昇る。") {
		t.Fatalf("expected production front to hide usage sentence, got %q", front)
	}

	m.reveal = true
	back := m.reviewView()
	if !strings.Contains(back, "日") || !strings.Contains(back, "にち") || !strings.Contains(back, "が昇る。") {
		t.Fatalf("expected production back to show japanese details, got %q", back)
	}
}

func TestReviewViewClozeMode(t *testing.T) {
	hiragana := "だいじ"
	card := domain.Card{Kanji: "大事", Hiragana: &hiragana, English: "importance", Usage: "体が大事にしないと病気なります。"}

	m := New()
	m.reviewQueue = []domain.Card{card}
	m.reviewModes = []reviewMode{reviewModeCloze}
	m.reviewIndex = 0

	front := m.reviewView()
	if !strings.Contains(front, "体が（　）にしないと病気なります。") {
		t.Fatalf("expected cloze front to hide japanese word in usage, got %q", front)
	}
	if strings.Contains(front, "importance") {
		t.Fatalf("expected cloze front to hide english, got %q", front)
	}

	m.reveal = true
	back := m.reviewView()
	if !strings.Contains(back, "大事, importance, だいじ") {
		t.Fatalf("expected cloze back to show japanese, english, and hiragana, got %q", back)
	}
}

func TestReviewViewShowsElapsedTime(t *testing.T) {
	hiragana := "にち"
	card := domain.Card{Kanji: "日", Hiragana: &hiragana, English: "day", Usage: "日が昇る。"}

	start := time.Date(2026, 2, 4, 12, 0, 0, 0, time.UTC)
	m := New()
	m.now = func() time.Time { return start.Add(2500 * time.Millisecond) }
	m.reviewCardAt = start
	m.reviewQueue = []domain.Card{card}
	m.reviewModes = []reviewMode{reviewModeRecognition}
	m.reviewIndex = 0

	view := m.reviewView()
	if !strings.Contains(view, "Elapsed:") {
		t.Fatalf("expected elapsed label, got %q", view)
	}
	if !strings.Contains(view, "2.5s") {
		t.Fatalf("expected formatted elapsed time, got %q", view)
	}
	if !strings.Contains(view, "<=3s advances") {
		t.Fatalf("expected elapsed threshold hint, got %q", view)
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

func TestApplyAnswerCorrectFastMovesToNextBox(t *testing.T) {
	dataDir := tempDataDir(t)
	if err := os.Setenv(storage.EnvDataDir, dataDir); err != nil {
		t.Fatalf("set env: %v", err)
	}

	today := domain.Today()
	cards := []domain.Card{
		{ID: "due", Kanji: "日", English: "day", Box: 2, CreatedAt: today, NextDue: today},
	}
	if err := storage.SaveCards(storage.CardsPath(dataDir), cards); err != nil {
		t.Fatalf("save cards: %v", err)
	}

	start := time.Date(2026, 2, 4, 12, 0, 0, 0, time.UTC)
	m := New()
	m.now = func() time.Time { return start }
	m = m.startReview()
	m.now = func() time.Time { return start.Add(2 * time.Second) }
	m = m.applyAnswer(true)

	if got := m.cards[0].Box; got != 3 {
		t.Fatalf("expected box 3, got %d", got)
	}
}

func TestApplyAnswerCorrectSlowStaysInSameBox(t *testing.T) {
	dataDir := tempDataDir(t)
	if err := os.Setenv(storage.EnvDataDir, dataDir); err != nil {
		t.Fatalf("set env: %v", err)
	}

	today := domain.Today()
	cards := []domain.Card{
		{ID: "due", Kanji: "日", English: "day", Box: 2, CreatedAt: today, NextDue: today},
	}
	if err := storage.SaveCards(storage.CardsPath(dataDir), cards); err != nil {
		t.Fatalf("save cards: %v", err)
	}

	start := time.Date(2026, 2, 4, 12, 0, 0, 0, time.UTC)
	m := New()
	m.now = func() time.Time { return start }
	m = m.startReview()
	m.now = func() time.Time { return start.Add(4 * time.Second) }
	m = m.applyAnswer(true)

	if got := m.cards[0].Box; got != 2 {
		t.Fatalf("expected box 2, got %d", got)
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
	if updated[0].Box != domain.BoxMin {
		t.Fatalf("expected box unchanged, got %d", updated[0].Box)
	}
	if updated[0].NextDue != today {
		t.Fatalf("expected next due unchanged, got %s", updated[0].NextDue)
	}
}

func TestEditCardUpdatesBoxAndResetsNextDue(t *testing.T) {
	dataDir := tempDataDir(t)
	if err := os.Setenv(storage.EnvDataDir, dataDir); err != nil {
		t.Fatalf("set env: %v", err)
	}

	today := domain.Today()
	future, err := today.AddDays(10)
	if err != nil {
		t.Fatalf("add days: %v", err)
	}

	cards := []domain.Card{
		{ID: "edit", Kanji: "日", English: "day", Box: 5, CreatedAt: today, NextDue: future},
	}
	if err := storage.SaveCards(storage.CardsPath(dataDir), cards); err != nil {
		t.Fatalf("save cards: %v", err)
	}

	m := New()
	m = m.startBrowse()
	m = m.startEditCard(m.browseCards[0])
	m.editInputs[5].SetValue("2")
	m.editInputs[6].SetValue("y")
	m = m.saveEditCard()

	updated, err := storage.LoadCards(storage.CardsPath(dataDir))
	if err != nil {
		t.Fatalf("load cards: %v", err)
	}
	if updated[0].Box != 2 {
		t.Fatalf("expected box 2, got %d", updated[0].Box)
	}
	if updated[0].NextDue != today {
		t.Fatalf("expected next due reset to today, got %s", updated[0].NextDue)
	}
}

func TestEditCardRejectsInvalidBox(t *testing.T) {
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
	m.editInputs[5].SetValue("99")
	m = m.saveEditCard()

	if !strings.Contains(m.editMessage, "Box must be between") {
		t.Fatalf("expected box validation message, got %q", m.editMessage)
	}

	updated, err := storage.LoadCards(storage.CardsPath(dataDir))
	if err != nil {
		t.Fatalf("load cards: %v", err)
	}
	if updated[0].Box != domain.BoxMin {
		t.Fatalf("expected box unchanged after validation error, got %d", updated[0].Box)
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

func TestParseTagsEmpty(t *testing.T) {
	tags := parseTags("")
	if tags != nil {
		t.Errorf("expected nil for empty input, got %v", tags)
	}
}

func TestParseTagsSingle(t *testing.T) {
	tags := parseTags("jlpt5")
	if len(tags) != 1 || tags[0] != "jlpt5" {
		t.Errorf("expected [jlpt5], got %v", tags)
	}
}

func TestParseTagsMultiple(t *testing.T) {
	tags := parseTags("jlpt5|colors|basic")
	expected := []string{"jlpt5", "colors", "basic"}
	if len(tags) != len(expected) {
		t.Fatalf("expected %d tags, got %d", len(expected), len(tags))
	}
	for i, tag := range expected {
		if tags[i] != tag {
			t.Errorf("expected tags[%d]=%s, got %s", i, tag, tags[i])
		}
	}
}

func TestParseTagsTrimWhitespace(t *testing.T) {
	tags := parseTags("  jlpt5  |  colors  ")
	if len(tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(tags))
	}
	if tags[0] != "jlpt5" || tags[1] != "colors" {
		t.Errorf("expected trimmed tags, got %v", tags)
	}
}

func TestParseTagsSkipsEmpty(t *testing.T) {
	tags := parseTags("jlpt5||colors|")
	if len(tags) != 2 {
		t.Fatalf("expected 2 tags (skipping empty), got %d: %v", len(tags), tags)
	}
}

func TestMenuViewContainsAllChoices(t *testing.T) {
	m := New()
	view := m.menuView()

	choices := []string{"Review", "Add Card", "Browse", "Import / Export", "Stats", "Quit"}
	for _, choice := range choices {
		if !strings.Contains(view, choice) {
			t.Errorf("menu view should contain %q", choice)
		}
	}
}

// updateModel はUpdate呼び出しの結果をModel型に変換するヘルパー
func updateModel(m Model, msg tea.Msg) (Model, tea.Cmd) {
	updated, cmd := m.Update(msg)
	return updated.(Model), cmd
}

func TestMenuNavigationUp(t *testing.T) {
	m := New()
	m.cursor = 2

	m, _ = updateModel(m, keyMsg("up"))
	if m.cursor != 1 {
		t.Errorf("expected cursor=1 after up, got %d", m.cursor)
	}

	m, _ = updateModel(m, keyMsg("k"))
	if m.cursor != 0 {
		t.Errorf("expected cursor=0 after k, got %d", m.cursor)
	}

	// 上限で止まる
	m, _ = updateModel(m, keyMsg("up"))
	if m.cursor != 0 {
		t.Errorf("expected cursor=0 at top, got %d", m.cursor)
	}
}

func TestMenuNavigationDown(t *testing.T) {
	m := New()
	m.cursor = 0

	m, _ = updateModel(m, keyMsg("down"))
	if m.cursor != 1 {
		t.Errorf("expected cursor=1 after down, got %d", m.cursor)
	}

	m, _ = updateModel(m, keyMsg("j"))
	if m.cursor != 2 {
		t.Errorf("expected cursor=2 after j, got %d", m.cursor)
	}
}

func TestScreenTransitionToStats(t *testing.T) {
	dataDir := tempDataDir(t)
	if err := os.Setenv(storage.EnvDataDir, dataDir); err != nil {
		t.Fatalf("set env: %v", err)
	}
	defer os.Unsetenv(storage.EnvDataDir)

	m := New()
	m.cursor = 4 // Stats
	m, _ = updateModel(m, keyMsg("enter"))

	if m.screen != screenStats {
		t.Errorf("expected stats screen, got %d", m.screen)
	}
}

func TestScreenTransitionBackFromStats(t *testing.T) {
	m := New()
	m.screen = screenStats

	m, _ = updateModel(m, keyMsg("esc"))

	if m.screen != screenMenu {
		t.Errorf("expected menu screen after esc, got %d", m.screen)
	}
}

func TestEmptyReviewQueue(t *testing.T) {
	dataDir := tempDataDir(t)
	if err := os.Setenv(storage.EnvDataDir, dataDir); err != nil {
		t.Fatalf("set env: %v", err)
	}
	defer os.Unsetenv(storage.EnvDataDir)

	// 空のカードリストを保存
	if err := storage.SaveCards(storage.CardsPath(dataDir), []domain.Card{}); err != nil {
		t.Fatalf("save cards: %v", err)
	}

	m := New()
	m = m.startReview()

	if len(m.reviewQueue) != 0 {
		t.Errorf("expected empty queue, got %d", len(m.reviewQueue))
	}

	view := m.reviewView()
	if !strings.Contains(view, "No cards due today") {
		t.Errorf("expected 'No cards due today' message, got %q", view)
	}
}

func TestReviewModeBadge(t *testing.T) {
	m := New()
	m.reviewQueue = []domain.Card{{Kanji: "日"}}
	m.reviewModes = []reviewMode{reviewModeRecognition}
	m.reviewIndex = 0

	badge := m.reviewModeBadge()
	if !strings.Contains(badge, "認識") {
		t.Errorf("expected recognition badge, got %q", badge)
	}

	m.reviewModes[0] = reviewModeProduction
	badge = m.reviewModeBadge()
	if !strings.Contains(badge, "産出") {
		t.Errorf("expected production badge, got %q", badge)
	}

	m.reviewModes[0] = reviewModeCloze
	badge = m.reviewModeBadge()
	if !strings.Contains(badge, "穴埋め") {
		t.Errorf("expected cloze badge, got %q", badge)
	}
}

func TestBrowseViewEmptyCards(t *testing.T) {
	m := New()
	m.browseCards = []domain.Card{}

	view := m.browseView()
	if !strings.Contains(view, "No cards available") {
		t.Errorf("expected 'No cards available' message, got %q", view)
	}
}

func TestBrowseNavigation(t *testing.T) {
	m := New()
	m.screen = screenBrowse
	m.browseCards = []domain.Card{
		{ID: "1", Kanji: "日", English: "day"},
		{ID: "2", Kanji: "月", English: "moon"},
		{ID: "3", Kanji: "火", English: "fire"},
	}
	m.browseCursor = 0

	m, _ = updateModel(m, keyMsg("down"))
	if m.browseCursor != 1 {
		t.Errorf("expected cursor=1 after down, got %d", m.browseCursor)
	}

	m, _ = updateModel(m, keyMsg("j"))
	if m.browseCursor != 2 {
		t.Errorf("expected cursor=2 after j, got %d", m.browseCursor)
	}

	// 下限で止まる
	m, _ = updateModel(m, keyMsg("down"))
	if m.browseCursor != 2 {
		t.Errorf("expected cursor=2 at bottom, got %d", m.browseCursor)
	}

	m, _ = updateModel(m, keyMsg("up"))
	if m.browseCursor != 1 {
		t.Errorf("expected cursor=1 after up, got %d", m.browseCursor)
	}
}

func TestAddCardValidationEmpty(t *testing.T) {
	m := New()
	m = m.startAddCard()

	// 空のまま保存を試みる
	m = m.saveAddCard()

	if !strings.Contains(m.addMessage, "required") {
		t.Errorf("expected validation error, got %q", m.addMessage)
	}
}

func TestRequestAIFillRequiresAPIKey(t *testing.T) {
	defer os.Unsetenv("OPENAI_API_KEY")
	if err := os.Unsetenv("OPENAI_API_KEY"); err != nil {
		t.Fatalf("unset env: %v", err)
	}

	m := New()
	m = m.startAddCard()
	m.addInputs[addKanjiField].SetValue("火")

	updated, cmd := m.requestAIFillMissing()
	if cmd != nil {
		t.Fatalf("expected nil cmd when key is missing")
	}
	if !strings.Contains(updated.addMessage, "OPENAI_API_KEY is required") {
		t.Fatalf("expected missing key message, got %q", updated.addMessage)
	}
}

func TestRequestAIFillUsesEnvAPIKey(t *testing.T) {
	if err := os.Setenv("OPENAI_API_KEY", "sk-test-key"); err != nil {
		t.Fatalf("set env: %v", err)
	}
	defer os.Unsetenv("OPENAI_API_KEY")

	m := New()
	m = m.startAddCard()
	m.addInputs[addKanjiField].SetValue("火")

	updated, cmd := m.requestAIFillMissing()
	if cmd == nil {
		t.Fatalf("expected command when OPENAI_API_KEY is set")
	}
	if !updated.aiBusy {
		t.Fatalf("expected aiBusy=true after requesting fill")
	}
}

func TestQuitFromMenu(t *testing.T) {
	m := New()
	m.cursor = 5 // Quit
	m, cmd := updateModel(m, keyMsg("enter"))

	if !m.quitting {
		t.Errorf("expected quitting=true")
	}
	if cmd == nil {
		t.Errorf("expected quit command")
	}
}

func TestCtrlCQuits(t *testing.T) {
	m := New()
	m, cmd := updateModel(m, keyMsg("ctrl+c"))

	if !m.quitting {
		t.Errorf("expected quitting=true on ctrl+c")
	}
	if cmd == nil {
		t.Errorf("expected quit command")
	}
}

// keyMsg はテスト用のキーメッセージを生成する
func keyMsg(key string) tea.KeyMsg {
	switch key {
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEscape}
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	case "tab":
		return tea.KeyMsg{Type: tea.KeyTab}
	case "ctrl+c":
		return tea.KeyMsg{Type: tea.KeyCtrlC}
	case " ":
		return tea.KeyMsg{Type: tea.KeySpace}
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
	}
}

func TestShuffleWithinDueGroupsKeepsDateOrder(t *testing.T) {
	today := domain.Today()
	yesterday, err := today.AddDays(-1)
	if err != nil {
		t.Fatalf("add days: %v", err)
	}

	cards := make([]domain.Card, 0, 12)
	for i := 0; i < 6; i++ {
		cards = append(cards, domain.Card{ID: string(rune('a' + i)), NextDue: yesterday})
	}
	for i := 0; i < 6; i++ {
		cards = append(cards, domain.Card{ID: string(rune('m' + i)), NextDue: today})
	}

	shuffled := shuffleWithinDueGroups(cards, rand.New(rand.NewSource(1)))

	if len(shuffled) != len(cards) {
		t.Fatalf("expected %d cards, got %d", len(cards), len(shuffled))
	}
	for i := 0; i < 6; i++ {
		if shuffled[i].NextDue != yesterday {
			t.Fatalf("expected overdue cards first, got %s at %d", shuffled[i].NextDue, i)
		}
	}
	for i := 6; i < 12; i++ {
		if shuffled[i].NextDue != today {
			t.Fatalf("expected today cards last, got %s at %d", shuffled[i].NextDue, i)
		}
	}

	seen := map[string]bool{}
	for _, card := range shuffled {
		seen[card.ID] = true
	}
	if len(seen) != len(cards) {
		t.Fatalf("expected permutation, got duplicates: %+v", shuffled)
	}

	sameOrder := true
	for i := range cards {
		if cards[i].ID != shuffled[i].ID {
			sameOrder = false
			break
		}
	}
	if sameOrder {
		t.Fatalf("expected order to change within groups, got %+v", shuffled)
	}
}

func TestStartReviewShufflesQueueWithinDueGroups(t *testing.T) {
	dataDir := tempDataDir(t)
	if err := os.Setenv(storage.EnvDataDir, dataDir); err != nil {
		t.Fatalf("set env: %v", err)
	}

	today := domain.Today()
	yesterday, err := today.AddDays(-1)
	if err != nil {
		t.Fatalf("add days: %v", err)
	}

	cards := []domain.Card{
		{ID: "t1", Kanji: "月", English: "moon", Box: domain.BoxMin, CreatedAt: today, NextDue: today},
		{ID: "y1", Kanji: "日", English: "day", Box: domain.BoxMin, CreatedAt: today, NextDue: yesterday},
		{ID: "t2", Kanji: "火", English: "fire", Box: domain.BoxMin, CreatedAt: today, NextDue: today},
	}
	if err := storage.SaveCards(storage.CardsPath(dataDir), cards); err != nil {
		t.Fatalf("save cards: %v", err)
	}

	m := New()
	m = m.startReview()

	if len(m.reviewQueue) != 3 {
		t.Fatalf("expected 3 due cards, got %d", len(m.reviewQueue))
	}
	if m.reviewQueue[0].ID != "y1" {
		t.Fatalf("expected overdue card first, got %s", m.reviewQueue[0].ID)
	}
	for i, idx := range m.reviewOrder {
		if m.cards[idx].ID != m.reviewQueue[i].ID {
			t.Fatalf("reviewOrder out of sync at %d: %s vs %s", i, m.cards[idx].ID, m.reviewQueue[i].ID)
		}
	}
}

func TestApplyAnswerTracksMissedCards(t *testing.T) {
	dataDir := tempDataDir(t)
	if err := os.Setenv(storage.EnvDataDir, dataDir); err != nil {
		t.Fatalf("set env: %v", err)
	}

	today := domain.Today()
	hiragana := "にち"
	cards := []domain.Card{
		{ID: "miss", Kanji: "日", Hiragana: &hiragana, English: "day", Usage: "日が昇る。", Box: 3, CreatedAt: today, NextDue: today},
		{ID: "hit", Kanji: "月", English: "moon", Box: domain.BoxMin, CreatedAt: today, NextDue: today},
	}
	if err := storage.SaveCards(storage.CardsPath(dataDir), cards); err != nil {
		t.Fatalf("save cards: %v", err)
	}

	m := New()
	m = m.startReview()
	// キューはシャッフルされるため、ID で判定する
	for m.reviewIndex < len(m.reviewQueue) {
		m = m.applyAnswer(m.reviewQueue[m.reviewIndex].ID != "miss")
	}

	if len(m.sessionMisses) != 1 {
		t.Fatalf("expected 1 missed card, got %d", len(m.sessionMisses))
	}
	if m.sessionMisses[0].ID != "miss" {
		t.Fatalf("expected missed card 'miss', got %s", m.sessionMisses[0].ID)
	}
	if m.sessionMisses[0].Box != 2 {
		t.Fatalf("expected missed card to reflect demoted box 2, got %d", m.sessionMisses[0].Box)
	}
}

func TestReviewViewCompleteListsMissedCards(t *testing.T) {
	hiragana := "にち"
	missed := domain.Card{ID: "miss", Kanji: "日", Hiragana: &hiragana, English: "day", Usage: "日が昇る。", Box: 2}

	m := New()
	m.reviewQueue = []domain.Card{missed}
	m.reviewModes = []reviewMode{reviewModeRecognition}
	m.reviewIndex = 1
	m.sessionMisses = []domain.Card{missed}

	view := m.reviewView()
	if !strings.Contains(view, "Session complete!") {
		t.Fatalf("expected completion message, got %q", view)
	}
	if !strings.Contains(view, "Missed cards") {
		t.Fatalf("expected missed cards heading, got %q", view)
	}
	for _, want := range []string{"日", "にち", "day", "が昇る。"} {
		if !strings.Contains(view, want) {
			t.Fatalf("expected missed card detail %q, got %q", want, view)
		}
	}
}

func TestReviewViewCompleteNoMisses(t *testing.T) {
	m := New()
	m.reviewQueue = []domain.Card{{ID: "hit", Kanji: "月", English: "moon"}}
	m.reviewModes = []reviewMode{reviewModeRecognition}
	m.reviewIndex = 1
	m.sessionMisses = nil

	view := m.reviewView()
	if strings.Contains(view, "Missed cards") {
		t.Fatalf("expected no missed cards heading, got %q", view)
	}
	if !strings.Contains(view, "No misses") {
		t.Fatalf("expected no-misses message, got %q", view)
	}
}

func TestStartReviewResetsSessionMisses(t *testing.T) {
	dataDir := tempDataDir(t)
	if err := os.Setenv(storage.EnvDataDir, dataDir); err != nil {
		t.Fatalf("set env: %v", err)
	}
	if err := storage.SaveCards(storage.CardsPath(dataDir), []domain.Card{}); err != nil {
		t.Fatalf("save cards: %v", err)
	}

	m := New()
	m.sessionMisses = []domain.Card{{ID: "stale"}}
	m = m.startReview()

	if len(m.sessionMisses) != 0 {
		t.Fatalf("expected misses reset on new session, got %d", len(m.sessionMisses))
	}
}
