package model

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"kanji-boxes/domain"
	"kanji-boxes/storage"
	"kanji-boxes/ui/view"
)

type screen int

const (
	screenMenu screen = iota
	screenReview
	screenStats
	screenAddCard
	screenBrowse
	screenImportExport
	screenEditCard
	screenDeleteConfirm
)

type ioMode int

const (
	ioModeSelect ioMode = iota
	ioModePath
)

type reviewMode int

const (
	reviewModeRecognition reviewMode = iota
	reviewModeProduction
	reviewModeCloze
)

// レビューモード重み付け（累積パーセンテージ）
// 意味の想起を優先するため産出（英語→日本語）を中心にし、穴埋めは低頻度に抑える
// 認識: 30%, 穴埋め: 10% (40-30), 産出: 60% (100-40)
const (
	recognitionThreshold = 30 // 0-29: 認識モード (30%)
	clozeThreshold       = 40 // 30-39: 穴埋めモード (10%)
	// 40-99: 産出モード (60%)
)

type Model struct {
	choices       []string
	cursor        int
	selected      int
	quitting      bool
	screen        screen
	dataDir       string
	cards         []domain.Card
	reviewQueue   []domain.Card
	reviewModes   []reviewMode
	reviewOrder   []int
	reviewIndex   int
	reviewCardAt  time.Time
	reveal        bool
	reviewErr     error
	session       domain.SessionStats
	sessionMisses []domain.Card // 今回のセッションで不正解だったカード
	reviewWaiting int           // ボックス1の上限で待機中のカード数
	rng           *rand.Rand
	addInputs     []textinput.Model
	addFocus      int
	addMessage    string
	aiBusy        bool
	editInputs    []textinput.Model
	editFocus     int
	editMessage   string
	editID        string
	state         storage.AppState
	stateErr      error
	stateWarning  string         // 保存失敗時の警告メッセージ
	browseAll     []domain.Card  // 絞り込み前の全カード
	browseCards   []domain.Card  // 検索で絞り込んだ表示用カード
	browseWaiting map[string]int // ボックス1の枠待ちカードのIDと待ち順
	browseErr     error
	browseCursor  int
	browseQuery   textinput.Model
	browseSearch  bool // 検索入力中はキーを入力欄に渡す
	ioMode        ioMode
	ioAction      string
	ioCursor      int
	ioInput       textinput.Model
	ioMessage     string
	ioErr         error
	now           func() time.Time
}

const (
	addKanjiField = iota
	addEnglishField
	addHiraganaField
	addUsageField
	addTagsField
	addFieldCount
)

func New() Model {
	m := Model{
		choices:  []string{"Review", "Add Card", "Browse", "Import / Export", "Stats", "Quit"},
		cursor:   0,
		selected: -1,
		screen:   screenMenu,
		rng:      rand.New(rand.NewSource(time.Now().UnixNano())),
		now:      time.Now,
	}
	return m.resetBrowseSearch()
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case aiFillResultMsg:
		m.aiBusy = false
		if msg.err != nil {
			m.addMessage = "AI fill failed: " + msg.err.Error()
			return m, nil
		}
		for index, value := range msg.values {
			if index < 0 || index >= len(m.addInputs) {
				continue
			}
			m.addInputs[index].SetValue(value)
		}
		if len(msg.values) == 0 {
			m.addMessage = "No values generated."
		} else {
			m.addMessage = "AI suggestion ready. Review and edit any field before saving."
		}
		return m, nil
	case tea.KeyMsg:
		key := msg.String()
		switch key {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "q":
			if m.screen != screenAddCard && m.screen != screenImportExport && m.screen != screenEditCard && !(m.screen == screenBrowse && m.browseSearch) {
				m.quitting = true
				return m, tea.Quit
			}
		}

		switch m.screen {
		case screenMenu:
			switch key {
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(m.choices)-1 {
					m.cursor++
				}
			case "enter":
				m.selected = m.cursor
				selected := m.choices[m.selected]
				switch selected {
				case "Quit":
					m.quitting = true
					return m, tea.Quit
				case "Review":
					m.screen = screenReview
					m = m.startReview()
				case "Add Card":
					m.screen = screenAddCard
					m = m.startAddCard()
				case "Browse":
					m.screen = screenBrowse
					m = m.resetBrowseSearch()
					m = m.startBrowse()
				case "Import / Export":
					m.screen = screenImportExport
					m = m.startImportExport()
				case "Stats":
					m.screen = screenStats
					m = m.loadStats()
				}
			}
		case screenReview:
			switch key {
			case "esc":
				m.screen = screenMenu
			case " ":
				m.reveal = !m.reveal
			case "y":
				if m.reviewIndex < len(m.reviewOrder) {
					m = m.applyAnswer(true)
				}
			case "n":
				if m.reviewIndex < len(m.reviewOrder) {
					m = m.applyAnswer(false)
				}
			}
		case screenStats:
			switch key {
			case "esc":
				m.screen = screenMenu
			}
		case screenAddCard:
			switch key {
			case "esc":
				m.screen = screenMenu
			case "ctrl+f":
				updated, cmd := m.requestAIFillMissing()
				return updated, cmd
			case "ctrl+r":
				updated, cmd := m.requestAIFillFocused()
				return updated, cmd
			case "tab", "shift+tab", "up", "down":
				m = m.moveAddFocus(key)
			case "enter":
				if m.addFocus < len(m.addInputs)-1 {
					m = m.moveAddFocus("down")
				} else {
					m = m.saveAddCard()
				}
			}

			for i := range m.addInputs {
				if i == m.addFocus {
					m.addInputs[i].Focus()
				} else {
					m.addInputs[i].Blur()
				}
				m.addInputs[i], _ = m.addInputs[i].Update(msg)
			}
		case screenBrowse:
			if m.browseSearch {
				switch key {
				case "esc":
					// 入力中のescは検索を取り消す
					m = m.resetBrowseSearch()
					m = m.applyBrowseFilter()
				case "enter":
					// 絞り込みを保ったまま一覧操作に戻る
					m.browseSearch = false
					m.browseQuery.Blur()
				case "up", "down":
					m = m.moveBrowseCursor(key)
				default:
					m.browseQuery, _ = m.browseQuery.Update(msg)
					m = m.applyBrowseFilter()
				}
				return m, nil
			}
			switch key {
			case "/":
				m.browseSearch = true
				m.browseQuery.Focus()
			case "esc":
				if m.browseQuery.Value() != "" {
					m = m.resetBrowseSearch()
					m = m.applyBrowseFilter()
				} else {
					m.screen = screenMenu
				}
			case "up", "k", "down", "j":
				m = m.moveBrowseCursor(key)
			case "e":
				if len(m.browseCards) > 0 {
					m.screen = screenEditCard
					m = m.startEditCard(m.browseCards[m.browseCursor])
				}
			case "d":
				if len(m.browseCards) > 0 {
					m.screen = screenDeleteConfirm
				}
			}
		case screenImportExport:
			switch m.ioMode {
			case ioModeSelect:
				switch key {
				case "esc":
					m.screen = screenMenu
				case "up", "k":
					if m.ioCursor > 0 {
						m.ioCursor--
					}
				case "down", "j":
					if m.ioCursor < 1 {
						m.ioCursor++
					}
				case "enter":
					if m.ioCursor == 0 {
						m.ioAction = "import"
					} else {
						m.ioAction = "export"
					}
					m.ioMode = ioModePath
					m.ioInput.Focus()
				}
			case ioModePath:
				switch key {
				case "esc":
					m.ioMode = ioModeSelect
					m.ioAction = ""
					m.ioInput.Blur()
				case "enter":
					m = m.performImportExport()
				}
				m.ioInput, _ = m.ioInput.Update(msg)
			}
		case screenEditCard:
			switch key {
			case "esc":
				m.screen = screenBrowse
			case "tab", "shift+tab", "up", "down":
				m = m.moveEditFocus(key)
			case "enter":
				if m.editFocus < len(m.editInputs)-1 {
					m = m.moveEditFocus("down")
				} else {
					m = m.saveEditCard()
				}
			}

			for i := range m.editInputs {
				if i == m.editFocus {
					m.editInputs[i].Focus()
				} else {
					m.editInputs[i].Blur()
				}
				m.editInputs[i], _ = m.editInputs[i].Update(msg)
			}
		case screenDeleteConfirm:
			switch key {
			case "y":
				m = m.deleteSelectedCard()
			case "n", "esc":
				m.screen = screenBrowse
			}
		}
	}

	return m, nil
}

func (m Model) View() string {
	if m.quitting {
		return "Goodbye!"
	}

	switch m.screen {
	case screenReview:
		return m.reviewView()
	case screenStats:
		return m.statsView()
	case screenAddCard:
		return m.addCardView()
	case screenBrowse:
		return m.browseView()
	case screenImportExport:
		return m.importExportView()
	case screenEditCard:
		return m.editCardView()
	case screenDeleteConfirm:
		return m.deleteConfirmView()
	default:
		return m.menuView()
	}
}

func (m Model) menuView() string {
	var builder strings.Builder

	// メニュー項目を構築
	var menuContent strings.Builder
	for i, choice := range m.choices {
		cursor := " "
		choiceStyle := view.MenuStyle
		if m.cursor == i {
			cursor = view.CursorStyle.Render(">")
			choiceStyle = view.MenuActiveStyle
		}
		menuContent.WriteString(cursor + " " + choiceStyle.Render(choice) + "\n")
	}

	// タイトルと装飾メニュー表示
	builder.WriteString(view.TitleStyle.Render("漢字ボックス") + " " + view.HintStyle.Render("Kanji Boxes") + "\n\n")
	builder.WriteString(view.MenuBoxStyle.Render(strings.TrimSuffix(menuContent.String(), "\n")))
	builder.WriteString("\n\n")
	builder.WriteString(view.HintStyle.Render("up/down: move • enter: select • q: quit") + "\n")

	return builder.String()
}

func (m Model) reviewView() string {
	var builder strings.Builder
	builder.WriteString(view.TitleStyle.Render("Review") + "\n\n")

	if m.reviewErr != nil {
		builder.WriteString(view.ErrorStyle.Render("Failed to load cards: "+m.reviewErr.Error()) + "\n")
		builder.WriteString("\n")
		builder.WriteString(view.HintStyle.Render("esc: back • q: quit") + "\n")
		return builder.String()
	}

	// 警告メッセージ表示
	if m.stateWarning != "" {
		builder.WriteString(view.ErrorStyle.Render(m.stateWarning) + "\n")
	}

	builder.WriteString(view.LabelStyle.Render("Due cards: ") + view.ValueStyle.Render(fmt.Sprintf("%d", len(m.reviewQueue))) + "\n")
	if m.reviewWaiting > 0 {
		builder.WriteString(view.HintStyle.Render(fmt.Sprintf("%d new cards waiting for a box 1 slot", m.reviewWaiting)) + "\n")
	}
	if len(m.reviewQueue) == 0 {
		builder.WriteString(view.SuccessStyle.Render("No cards due today.") + "\n")
	} else {
		// プログレスバー表示
		progressBar := view.RenderProgressBar(m.reviewIndex, len(m.reviewQueue), 20)
		builder.WriteString(view.LabelStyle.Render("Progress: ") + progressBar + " " +
			view.ValueStyle.Render(fmt.Sprintf("%d/%d", m.reviewIndex, len(m.reviewQueue))) + "\n\n")

		if m.reviewIndex >= len(m.reviewQueue) {
			builder.WriteString(view.SuccessStyle.Render("Session complete!") + "\n")
			builder.WriteString(view.LabelStyle.Render("Correct: ") + view.CorrectStatStyle.Render(fmt.Sprintf("%d", m.session.Correct)) + " | ")
			builder.WriteString(view.LabelStyle.Render("Incorrect: ") + view.IncorrectStatStyle.Render(fmt.Sprintf("%d", m.session.Incorrect)) + "\n")
			builder.WriteString(m.missedCardsView())
		} else {
			card := m.reviewQueue[m.reviewIndex]

			// レビューモードバッジ表示
			modeBadge := m.reviewModeBadge()
			boxStyle := view.BoxStyleForLevel(card.Box)
			builder.WriteString(modeBadge + " " + view.LabelStyle.Render("Box: ") + boxStyle.Render(fmt.Sprintf("%d", card.Box)) + "\n\n")
			builder.WriteString(view.LabelStyle.Render("Elapsed: ") + view.ValueStyle.Render(formatElapsed(m.currentElapsed())) + " " + view.HintStyle.Render("(<=3s advances)") + "\n\n")

			switch m.currentReviewMode() {
			case reviewModeProduction:
				if m.reveal {
					builder.WriteString(view.KanjiStyle.Render(card.Kanji) + "\n")
					if card.Hiragana != nil {
						builder.WriteString(view.HiraganaStyle.Render(*card.Hiragana) + "\n")
					}
					if card.Usage != "" {
						builder.WriteString(view.UsageStyle.Render(highlightUsage(card.Usage, card.Kanji, card.Hiragana)) + "\n")
					}
				} else {
					builder.WriteString(view.ValueStyle.Render(card.English) + "\n")
					builder.WriteString(view.HintStyle.Render("(press space to flip)") + "\n")
				}
			case reviewModeCloze:
				if m.reveal {
					builder.WriteString(view.ValueStyle.Render(clozeBack(card)) + "\n")
				} else {
					if card.Usage != "" {
						builder.WriteString(view.UsageStyle.Render(clozeFront(card.Usage, card.Kanji, card.Hiragana)) + "\n")
					} else {
						builder.WriteString(view.KanjiStyle.Render(card.Kanji) + "\n")
					}
					builder.WriteString(view.HintStyle.Render("(press space to flip)") + "\n")
				}
			default:
				if m.reveal {
					if card.Hiragana != nil {
						builder.WriteString(view.HiraganaStyle.Render(*card.Hiragana) + "\n")
					} else {
						builder.WriteString(view.KanjiStyle.Render(card.Kanji) + "\n")
					}
					builder.WriteString(view.ValueStyle.Render(card.English) + "\n")
					if card.Usage != "" {
						builder.WriteString(view.UsageStyle.Render(highlightUsage(card.Usage, card.Kanji, card.Hiragana)) + "\n")
					}
				} else {
					builder.WriteString(view.KanjiStyle.Render(card.Kanji) + "\n")
					if card.Hiragana != nil && *card.Hiragana != "" {
						builder.WriteString(view.HiraganaStyle.Render(*card.Hiragana) + "\n")
					}
					builder.WriteString(view.HintStyle.Render("(press space to flip)") + "\n")
				}
			}
		}
	}

	builder.WriteString("\n")
	builder.WriteString(view.HintStyle.Render("space: flip • y: correct • n: incorrect • esc: back • q: quit") + "\n")

	return builder.String()
}

// reviewModeBadge はレビューモードのバッジを返す
func (m Model) reviewModeBadge() string {
	switch m.currentReviewMode() {
	case reviewModeRecognition:
		return view.RecognitionModeStyle.Render("[認識]")
	case reviewModeProduction:
		return view.ProductionModeStyle.Render("[産出]")
	case reviewModeCloze:
		return view.ClozeModeStyle.Render("[穴埋め]")
	default:
		return view.RecognitionModeStyle.Render("[認識]")
	}
}

func highlightUsage(sentence string, kanji string, hiragana *string) string {
	highlighted := sentence
	if kanji != "" {
		highlighted = strings.ReplaceAll(highlighted, kanji, view.HighlightStyle.Render(kanji))
	}
	if hiragana != nil {
		value := *hiragana
		if value != "" {
			highlighted = strings.ReplaceAll(highlighted, value, view.HighlightStyle.Render(value))
		}
	}
	return highlighted
}

func statusMessageStyle(message string) lipgloss.Style {
	messageLower := strings.ToLower(message)
	if strings.Contains(messageLower, "failed") || strings.Contains(messageLower, "required") || strings.Contains(messageLower, "error") {
		return view.ErrorStyle
	}
	return view.SuccessStyle
}

func (m Model) statsView() string {
	var builder strings.Builder
	builder.WriteString(view.TitleStyle.Render("Stats") + "\n\n")
	if m.stateErr != nil {
		builder.WriteString(view.ErrorStyle.Render("Failed to load stats: "+m.stateErr.Error()) + "\n\n")
		builder.WriteString(view.HintStyle.Render("esc: back • q: quit") + "\n")
		return builder.String()
	}

	if len(m.state.DailyStats) == 0 {
		builder.WriteString(view.SuccessStyle.Render("No stats yet.") + "\n\n")
		builder.WriteString(view.HintStyle.Render("esc: back • q: quit") + "\n")
		return builder.String()
	}

	date := domain.Today().String()
	builder.WriteString(view.SubtitleStyle.Render(fmt.Sprintf("Today (%s)", date)) + "\n")
	builder.WriteString(view.LabelStyle.Render("Reviewed: ") + view.NeutralStatStyle.Render(fmt.Sprintf("%d", m.session.Reviewed)) + "\n")
	builder.WriteString(view.LabelStyle.Render("Correct:  ") + view.CorrectStatStyle.Render(fmt.Sprintf("%d", m.session.Correct)) + "\n")
	builder.WriteString(view.LabelStyle.Render("Incorrect:") + view.IncorrectStatStyle.Render(fmt.Sprintf(" %d", m.session.Incorrect)) + "\n")

	// 正答率表示（レビューがある場合）
	if m.session.Reviewed > 0 {
		accuracy := (m.session.Correct * 100) / m.session.Reviewed
		progressBar := view.RenderProgressBar(m.session.Correct, m.session.Reviewed, 15)
		builder.WriteString(view.LabelStyle.Render("Accuracy: ") + progressBar + " " +
			view.ValueStyle.Render(fmt.Sprintf("%d%%", accuracy)) + "\n")
	}
	builder.WriteString("\n")

	builder.WriteString(view.SubtitleStyle.Render("History (last 7 days)") + "\n")
	builder.WriteString(view.HintStyle.Render("Date       Rev  ✓   ✗") + "\n")
	for _, day := range recentDates(m.state.DailyStats, 7) {
		stats := m.state.DailyStats[day]
		builder.WriteString(view.ValueStyle.Render(day) + "  " +
			view.NeutralStatStyle.Render(fmt.Sprintf("%3d", stats.Reviewed)) + "  " +
			view.CorrectStatStyle.Render(fmt.Sprintf("%3d", stats.Correct)) + " " +
			view.IncorrectStatStyle.Render(fmt.Sprintf("%3d", stats.Incorrect)) + "\n")
	}
	builder.WriteString("\n")
	builder.WriteString(view.HintStyle.Render("esc: back • q: quit") + "\n")
	return builder.String()
}

func (m Model) addCardView() string {
	var builder strings.Builder
	builder.WriteString(view.TitleStyle.Render("Add Card") + "\n\n")
	builder.WriteString(view.DividerStyle.Render(strings.Repeat("─", 24)) + "\n\n")
	labels := []string{"Kanji*", "English*", "Hiragana", "Usage sentence", "Tags (| separated)"}
	for i, input := range m.addInputs {
		inputStyle := view.InputStyle
		if i == m.addFocus {
			inputStyle = view.InputFocusStyle
		}
		builder.WriteString(view.LabelStyle.Render(labels[i]) + "\n")
		builder.WriteString(inputStyle.Render(input.View()) + "\n\n")
	}

	if m.addMessage != "" {
		builder.WriteString(statusMessageStyle(m.addMessage).Render(m.addMessage) + "\n\n")
	}
	if m.aiBusy {
		builder.WriteString(view.HintStyle.Render("AI is generating suggestions...") + "\n\n")
	}
	builder.WriteString("\n")
	builder.WriteString(view.HintStyle.Render("ctrl+f: fill missing • ctrl+r: regenerate focused • enter: save • esc: back • q: quit") + "\n")
	return builder.String()
}

func (m Model) browseView() string {
	var builder strings.Builder
	builder.WriteString(view.TitleStyle.Render("Browse") + "\n\n")

	if m.browseErr != nil {
		builder.WriteString(view.ErrorStyle.Render("Failed to load cards: "+m.browseErr.Error()) + "\n\n")
		builder.WriteString(view.HintStyle.Render("esc: back • q: quit") + "\n")
		return builder.String()
	}

	if m.browseSearch || m.browseQuery.Value() != "" {
		builder.WriteString(view.LabelStyle.Render("Search: ") + m.browseQuery.View() + "\n\n")
	}

	if len(m.browseCards) == 0 {
		if m.browseQuery.Value() != "" {
			builder.WriteString(view.HintStyle.Render("No matching cards.") + "\n\n")
			builder.WriteString(view.HintStyle.Render(m.browseHint()) + "\n")
			return builder.String()
		}
		builder.WriteString(view.SuccessStyle.Render("No cards available.") + "\n")
		builder.WriteString("\n")
		builder.WriteString(view.HintStyle.Render("esc: back • q: quit") + "\n")
		return builder.String()
	}

	// カード数表示（絞り込み中は全件数も表示）
	count := fmt.Sprintf("%d", len(m.browseCards))
	if len(m.browseCards) != len(m.browseAll) && len(m.browseAll) > 0 {
		count = fmt.Sprintf("%d / %d", len(m.browseCards), len(m.browseAll))
	}
	builder.WriteString(view.LabelStyle.Render("Cards: ") + view.ValueStyle.Render(count) + "\n\n")

	start := m.browseCursor - 5
	if start < 0 {
		start = 0
	}
	end := start + 10
	if end > len(m.browseCards) {
		end = len(m.browseCards)
	}

	for i := start; i < end; i++ {
		card := m.browseCards[i]
		cursor := " "
		rowStyle := view.MenuStyle
		if i == m.browseCursor {
			cursor = view.CursorStyle.Render(">")
			rowStyle = view.MenuActiveStyle
		}
		// ボックスレベルに応じた色でインジケーター表示
		boxIndicator := view.BoxStyleForLevel(card.Box).Render(fmt.Sprintf("[%d]", card.Box))
		if _, waiting := m.browseWaiting[card.ID]; waiting {
			// 枠待ちのカードはまだボックスに入っていないため番号を出さない
			boxIndicator = view.HintStyle.Render("[-]")
		}
		builder.WriteString(fmt.Sprintf("%s %s %s\n", cursor, boxIndicator, rowStyle.Render(fmt.Sprintf("%s - %s", card.Kanji, card.English))))
	}

	selected := m.browseCards[m.browseCursor]
	builder.WriteString("\n")
	builder.WriteString(view.SubtitleStyle.Render("Details") + "\n")
	if selected.Hiragana != nil {
		builder.WriteString(view.LabelStyle.Render("Hiragana: ") + view.HiraganaStyle.Render(*selected.Hiragana) + "\n")
	}
	if selected.Usage != "" {
		builder.WriteString(view.LabelStyle.Render("Usage: ") + view.UsageStyle.Render(selected.Usage) + "\n")
	}
	boxStyle := view.BoxStyleForLevel(selected.Box)
	if position, waiting := m.browseWaiting[selected.ID]; waiting {
		builder.WriteString(view.LabelStyle.Render("Box: ") + view.HintStyle.Render(fmt.Sprintf("waiting for a box 1 slot (#%d in line)", position)) + "\n")
		builder.WriteString(view.LabelStyle.Render("Next due: ") + view.HintStyle.Render("when a slot opens") + "\n")
	} else {
		builder.WriteString(view.LabelStyle.Render("Box: ") + boxStyle.Render(fmt.Sprintf("%d", selected.Box)) + "\n")
		builder.WriteString(view.LabelStyle.Render("Next due: ") + view.ValueStyle.Render(selected.NextDue.String()) + "\n")
	}
	if len(selected.Tags) > 0 {
		builder.WriteString(view.LabelStyle.Render("Tags: ") + view.ValueStyle.Render(strings.Join(selected.Tags, ", ")) + "\n")
	}

	builder.WriteString("\n")
	builder.WriteString(view.HintStyle.Render(m.browseHint()) + "\n")
	return builder.String()
}

func (m Model) browseHint() string {
	if m.browseSearch {
		return "type to filter • up/down: move • enter: done • esc: clear"
	}
	if m.browseQuery.Value() != "" {
		return "up/down: move • /: search • e: edit • d: delete • esc: clear search • q: quit"
	}
	return "up/down: move • /: search • e: edit • d: delete • esc: back • q: quit"
}

func (m Model) editCardView() string {
	var builder strings.Builder
	builder.WriteString(view.TitleStyle.Render("Edit Card") + "\n\n")
	builder.WriteString(view.DividerStyle.Render(strings.Repeat("─", 24)) + "\n\n")
	labels := []string{"Kanji*", "English*", "Hiragana", "Usage sentence", "Tags (| separated)", "Box level", "Reset next due (y/N)"}
	for i, input := range m.editInputs {
		inputStyle := view.InputStyle
		if i == m.editFocus {
			inputStyle = view.InputFocusStyle
		}
		builder.WriteString(view.LabelStyle.Render(labels[i]) + "\n")
		builder.WriteString(inputStyle.Render(input.View()) + "\n\n")
	}

	if m.editMessage != "" {
		builder.WriteString(statusMessageStyle(m.editMessage).Render(m.editMessage) + "\n\n")
	}
	builder.WriteString("\n")
	builder.WriteString(view.HintStyle.Render("tab: next • enter: save • reset next due: y/yes/true/1 • esc: back • q: quit") + "\n")
	return builder.String()
}

func (m Model) deleteConfirmView() string {
	var builder strings.Builder
	builder.WriteString(view.TitleStyle.Render("Delete Card") + "\n\n")
	if len(m.browseCards) == 0 {
		builder.WriteString(view.ErrorStyle.Render("No card selected.") + "\n\n")
		builder.WriteString(view.HintStyle.Render("esc: back • q: quit") + "\n")
		return builder.String()
	}
	card := m.browseCards[m.browseCursor]
	builder.WriteString(view.ErrorStyle.Render(fmt.Sprintf("Delete %s - %s?", card.Kanji, card.English)) + "\n\n")
	builder.WriteString(view.HintStyle.Render("y: delete • n: cancel") + "\n")
	return builder.String()
}

func (m Model) importExportView() string {
	var builder strings.Builder
	builder.WriteString(view.TitleStyle.Render("Import / Export") + "\n\n")

	if m.ioErr != nil {
		builder.WriteString(view.ErrorStyle.Render("Error: "+m.ioErr.Error()) + "\n\n")
	}
	if m.ioMessage != "" {
		builder.WriteString(view.SuccessStyle.Render(m.ioMessage) + "\n\n")
	}

	switch m.ioMode {
	case ioModeSelect:
		options := []string{"Import CSV", "Export CSV"}
		for i, option := range options {
			cursor := " "
			optionStyle := view.MenuStyle
			if i == m.ioCursor {
				cursor = view.CursorStyle.Render(">")
				optionStyle = view.MenuActiveStyle
			}
			builder.WriteString(fmt.Sprintf("%s %s\n", cursor, optionStyle.Render(option)))
		}
		builder.WriteString("\n")
		builder.WriteString(view.HintStyle.Render("up/down: move • enter: select • esc: back • q: quit") + "\n")
	case ioModePath:
		label := "Import from file"
		if m.ioAction == "export" {
			label = "Export to file"
		}
		builder.WriteString(view.SubtitleStyle.Render(label) + "\n")
		builder.WriteString(m.ioInput.View() + "\n\n")
		builder.WriteString(view.HintStyle.Render("enter: confirm • esc: back • q: quit") + "\n")
	}

	return builder.String()
}

func (m Model) startReview() Model {
	m.reviewErr = nil

	dataDir, err := storage.ResolveDataDir("")
	if err != nil {
		m.reviewErr = err
		return m
	}
	state, err := storage.LoadState(storage.StatePath(dataDir))
	if err != nil {
		m.reviewErr = err
		return m
	}

	cards, err := storage.LoadCards(storage.CardsPath(dataDir))
	if err != nil {
		m.reviewErr = err
		return m
	}

	m.dataDir = dataDir
	m.state = normalizeState(state, domain.Today())
	m.cards = cards
	// 同じ期限日のカード内でシャッフルし、並び順の手がかりによる暗記を防ぐ
	// 期限切れの古いカードを優先する順序は維持する
	// ボックス1の新規カードは上限内のみ出題し、残りは枠が空くまで待機させる
	m.reviewQueue = shuffleWithinDueGroups(domain.DueCards(domain.ActiveCards(cards, domain.BoxOneCap), domain.Today()), m.rng)
	m.reviewWaiting = domain.WaitingCount(cards, domain.BoxOneCap)
	m.reviewModes = make([]reviewMode, len(m.reviewQueue))
	for i := range m.reviewModes {
		roll := m.rng.Intn(100)
		if roll < recognitionThreshold {
			m.reviewModes[i] = reviewModeRecognition
			continue
		}
		if roll < clozeThreshold {
			m.reviewModes[i] = reviewModeCloze
			continue
		}
		m.reviewModes[i] = reviewModeProduction
	}
	m.reviewOrder = indicesForCards(cards, m.reviewQueue)
	m.reviewIndex = 0
	m.reviewCardAt = m.now()
	m.reveal = false
	m.stateWarning = ""
	m.sessionMisses = nil
	m.session = m.state.DailyStats[domain.Today().String()]
	if err := storage.SaveState(storage.StatePath(dataDir), m.state); err != nil {
		m.stateWarning = "Warning: Failed to save state"
	}
	return m
}

func (m Model) startAddCard() Model {
	inputs := make([]textinput.Model, addFieldCount)
	for i := range inputs {
		input := textinput.New()
		input.CharLimit = 120
		switch i {
		case addKanjiField:
			input.Placeholder = "例: 赤"
		case addEnglishField:
			input.Placeholder = "例: red"
		case addHiraganaField:
			input.Placeholder = "例: あか"
		case addUsageField:
			input.Placeholder = "例: 赤いシャツをあげます。"
		case addTagsField:
			input.Placeholder = "例: colors|jlpt5"
		}
		inputs[i] = input
	}
	inputs[0].Focus()

	m.addInputs = inputs
	m.addFocus = 0
	m.addMessage = ""
	m.aiBusy = false
	return m
}

func (m Model) startEditCard(card domain.Card) Model {
	inputs := make([]textinput.Model, 7)
	for i := range inputs {
		input := textinput.New()
		input.CharLimit = 120
		switch i {
		case 0:
			input.Placeholder = "例: 赤"
			input.SetValue(card.Kanji)
		case 1:
			input.Placeholder = "例: red"
			input.SetValue(card.English)
		case 2:
			input.Placeholder = "例: あか"
			if card.Hiragana != nil {
				input.SetValue(*card.Hiragana)
			}
		case 3:
			input.Placeholder = "例: 赤いシャツをあげます。"
			input.SetValue(card.Usage)
		case 4:
			input.Placeholder = "例: colors|jlpt5"
			input.SetValue(strings.Join(card.Tags, "|"))
		case 5:
			input.Placeholder = fmt.Sprintf("%d-%d", domain.BoxMin, domain.BoxMax)
			input.SetValue(fmt.Sprintf("%d", card.Box))
		case 6:
			input.Placeholder = "y/N"
		}
		inputs[i] = input
	}
	inputs[0].Focus()

	m.editInputs = inputs
	m.editFocus = 0
	m.editMessage = ""
	m.editID = card.ID
	return m
}

func (m Model) startBrowse() Model {
	m.browseErr = nil

	dataDir, err := storage.ResolveDataDir("")
	if err != nil {
		m.browseErr = err
		return m
	}

	cards, err := storage.LoadCards(storage.CardsPath(dataDir))
	if err != nil {
		m.browseErr = err
		return m
	}

	m.dataDir = dataDir
	m.browseAll = cards
	m.browseWaiting = domain.WaitingQueue(cards, domain.BoxOneCap)
	m = m.applyBrowseFilter()
	return m
}

// 端で折り返す（先頭で上→末尾、末尾で下→先頭）
func (m Model) moveBrowseCursor(key string) Model {
	count := len(m.browseCards)
	if count == 0 {
		return m
	}
	step := 1
	if key == "up" || key == "k" {
		step = -1
	}
	m.browseCursor = (m.browseCursor + step + count) % count
	return m
}

func (m Model) resetBrowseSearch() Model {
	input := textinput.New()
	input.Placeholder = "kanji, hiragana or english"
	input.CharLimit = 50
	m.browseQuery = input
	m.browseSearch = false
	return m
}

// 検索語で一覧を絞り込み、カーソルを先頭に戻す
func (m Model) applyBrowseFilter() Model {
	m.browseCards = filterCards(m.browseAll, m.browseQuery.Value())
	m.browseCursor = 0
	return m
}

// 漢字・読み・英語に部分一致するカードを返す（空の検索語は全件、英語は大文字小文字を区別しない）
func filterCards(cards []domain.Card, query string) []domain.Card {
	query = strings.TrimSpace(query)
	if query == "" {
		return cards
	}
	lowerQuery := strings.ToLower(query)
	filtered := make([]domain.Card, 0, len(cards))
	for _, card := range cards {
		if strings.Contains(card.Kanji, query) ||
			(card.Hiragana != nil && strings.Contains(*card.Hiragana, query)) ||
			strings.Contains(strings.ToLower(card.English), lowerQuery) {
			filtered = append(filtered, card)
		}
	}
	return filtered
}

func (m Model) startImportExport() Model {
	m.ioErr = nil
	m.ioMessage = ""
	m.ioMode = ioModeSelect
	m.ioCursor = 0
	m.ioAction = ""

	placeholder := filepath.Join("/path", "to", "cards.csv")
	if dataDir, err := storage.ResolveDataDir(""); err == nil {
		placeholder = filepath.Join(dataDir, "cards.csv")
	}

	input := textinput.New()
	input.Placeholder = placeholder
	input.CharLimit = 200
	input.Blur()
	m.ioInput = input
	return m
}

func (m Model) moveAddFocus(key string) Model {
	count := len(m.addInputs)
	if count == 0 {
		return m
	}

	switch key {
	case "up", "shift+tab":
		m.addFocus--
	case "down", "tab":
		m.addFocus++
	}

	if m.addFocus < 0 {
		m.addFocus = count - 1
	}
	if m.addFocus >= count {
		m.addFocus = 0
	}
	return m
}

func (m Model) moveEditFocus(key string) Model {
	count := len(m.editInputs)
	if count == 0 {
		return m
	}

	switch key {
	case "up", "shift+tab":
		m.editFocus--
	case "down", "tab":
		m.editFocus++
	}

	if m.editFocus < 0 {
		m.editFocus = count - 1
	}
	if m.editFocus >= count {
		m.editFocus = 0
	}
	return m
}

// cardInput はカード入力フォームから抽出されたデータ
type cardInput struct {
	kanji    string
	english  string
	hiragana *string
	usage    string
	tags     []string
}

// parseCardInputs はテキスト入力からカードデータを抽出する
func parseCardInputs(inputs []textinput.Model) cardInput {
	kanji := strings.TrimSpace(inputs[addKanjiField].Value())
	english := strings.TrimSpace(inputs[addEnglishField].Value())
	hiragana := strings.TrimSpace(inputs[addHiraganaField].Value())
	usage := strings.TrimSpace(inputs[addUsageField].Value())
	tagsInput := strings.TrimSpace(inputs[addTagsField].Value())

	var hiraPtr *string
	if hiragana != "" {
		hiraPtr = &hiragana
	}

	return cardInput{
		kanji:    kanji,
		english:  english,
		hiragana: hiraPtr,
		usage:    usage,
		tags:     parseTags(tagsInput),
	}
}

// parseTags はパイプ区切りのタグ文字列をスライスに変換する
func parseTags(input string) []string {
	if input == "" {
		return nil
	}
	var tags []string
	for _, tag := range strings.Split(input, "|") {
		trimmed := strings.TrimSpace(tag)
		if trimmed != "" {
			tags = append(tags, trimmed)
		}
	}
	return tags
}

func (m Model) saveAddCard() Model {
	if m.aiBusy {
		m.addMessage = "Please wait for AI generation to finish."
		return m
	}

	input := parseCardInputs(m.addInputs)

	if input.kanji == "" || input.english == "" {
		m.addMessage = "Kanji and English are required."
		return m
	}

	dataDir, err := storage.ResolveDataDir("")
	if err != nil {
		m.addMessage = "Failed to resolve data dir: " + err.Error()
		return m
	}

	cards, err := storage.LoadCards(storage.CardsPath(dataDir))
	if err != nil {
		m.addMessage = "Failed to load cards: " + err.Error()
		return m
	}

	today := domain.Today()
	card := domain.Card{
		ID:        fmt.Sprintf("card-%d", time.Now().UnixNano()),
		Kanji:     input.kanji,
		Hiragana:  input.hiragana,
		English:   input.english,
		Usage:     input.usage,
		Tags:      input.tags,
		Box:       domain.BoxMin,
		CreatedAt: today,
		NextDue:   today,
	}

	cards = append(cards, card)
	if err := storage.SaveCards(storage.CardsPath(dataDir), cards); err != nil {
		m.addMessage = "Failed to save card: " + err.Error()
		return m
	}

	m = m.startAddCard()
	m.addMessage = "Card saved."
	return m
}

func (m Model) saveEditCard() Model {
	input := parseCardInputs(m.editInputs[:5])

	if input.kanji == "" || input.english == "" {
		m.editMessage = "Kanji and English are required."
		return m
	}

	dataDir, err := storage.ResolveDataDir("")
	if err != nil {
		m.editMessage = "Failed to resolve data dir: " + err.Error()
		return m
	}

	cards, err := storage.LoadCards(storage.CardsPath(dataDir))
	if err != nil {
		m.editMessage = "Failed to load cards: " + err.Error()
		return m
	}

	boxInput := strings.TrimSpace(m.editInputs[5].Value())
	box, err := strconv.Atoi(boxInput)
	if err != nil || box < domain.BoxMin || box > domain.BoxMax {
		m.editMessage = fmt.Sprintf("Box must be between %d and %d.", domain.BoxMin, domain.BoxMax)
		return m
	}

	resetNextDue := shouldResetNextDue(m.editInputs[6].Value())

	updated := false
	for i := range cards {
		if cards[i].ID == m.editID {
			cards[i].Kanji = input.kanji
			cards[i].English = input.english
			cards[i].Hiragana = input.hiragana
			cards[i].Usage = input.usage
			cards[i].Tags = input.tags
			cards[i].Box = box
			if resetNextDue {
				cards[i].NextDue = domain.Today()
			}
			updated = true
			break
		}
	}
	if !updated {
		m.editMessage = "Card not found."
		return m
	}

	if err := storage.SaveCards(storage.CardsPath(dataDir), cards); err != nil {
		m.editMessage = "Failed to save card: " + err.Error()
		return m
	}

	m = m.startBrowse()
	m.screen = screenBrowse
	return m
}

func shouldResetNextDue(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "y", "yes", "true", "1":
		return true
	default:
		return false
	}
}

func (m Model) performImportExport() Model {
	m.ioErr = nil
	m.ioMessage = ""

	path := strings.TrimSpace(m.ioInput.Value())
	if path == "" {
		if m.ioAction == "export" {
			path = "exported.csv"
		} else {
			m.ioErr = fmt.Errorf("path is required")
			return m
		}
	}

	dataDir, err := storage.ResolveDataDir("")
	if err != nil {
		m.ioErr = err
		return m
	}

	if m.ioAction == "export" {
		cards, err := storage.LoadCards(storage.CardsPath(dataDir))
		if err != nil {
			m.ioErr = err
			return m
		}
		file, err := os.Create(path)
		if err != nil {
			m.ioErr = err
			return m
		}
		defer file.Close()
		if err := storage.ExportCardsCSV(file, cards); err != nil {
			m.ioErr = err
			return m
		}
		m.ioMessage = fmt.Sprintf("Exported %d cards.", len(cards))
	} else {
		file, err := os.Open(path)
		if err != nil {
			m.ioErr = err
			return m
		}
		defer file.Close()
		imported, err := storage.ImportCardsCSV(file, domain.Today())
		if err != nil {
			m.ioErr = err
			return m
		}
		existing, err := storage.LoadCards(storage.CardsPath(dataDir))
		if err != nil {
			m.ioErr = err
			return m
		}
		// 既存カードのkanjiインデックスを構築
		kanjiIndex := make(map[string]int, len(existing))
		for i, c := range existing {
			kanjiIndex[c.Kanji] = i
		}
		added := 0
		updated := 0
		for i, card := range imported {
			if idx, found := kanjiIndex[card.Kanji]; found {
				// 既存カードのboxを更新
				existing[idx].Box = card.Box
				updated++
			} else {
				if card.ID == "" {
					card.ID = fmt.Sprintf("card-%d", time.Now().UnixNano()+int64(i))
				}
				existing = append(existing, card)
				added++
			}
		}
		if err := storage.SaveCards(storage.CardsPath(dataDir), existing); err != nil {
			m.ioErr = err
			return m
		}
		m.ioMessage = fmt.Sprintf("Imported %d new, updated %d existing.", added, updated)
	}

	m.ioMode = ioModeSelect
	m.ioInput.SetValue("")
	m.ioInput.Blur()
	m.ioAction = ""
	return m
}

func (m Model) deleteSelectedCard() Model {
	dataDir, err := storage.ResolveDataDir("")
	if err != nil {
		m.browseErr = err
		m.screen = screenBrowse
		return m
	}

	if len(m.browseCards) == 0 {
		m.screen = screenBrowse
		return m
	}

	deleteID := m.browseCards[m.browseCursor].ID
	cards, err := storage.LoadCards(storage.CardsPath(dataDir))
	if err != nil {
		m.browseErr = err
		m.screen = screenBrowse
		return m
	}

	filtered := make([]domain.Card, 0, len(cards))
	for _, card := range cards {
		if card.ID != deleteID {
			filtered = append(filtered, card)
		}
	}

	if err := storage.SaveCards(storage.CardsPath(dataDir), filtered); err != nil {
		m.browseErr = err
		m.screen = screenBrowse
		return m
	}

	m = m.startBrowse()
	m.screen = screenBrowse
	return m
}

func indicesForCards(all []domain.Card, subset []domain.Card) []int {
	indexByID := make(map[string]int, len(all))
	for i, card := range all {
		indexByID[card.ID] = i
	}

	indices := make([]int, 0, len(subset))
	for _, card := range subset {
		if idx, ok := indexByID[card.ID]; ok {
			indices = append(indices, idx)
		}
	}
	return indices
}

// shuffleWithinDueGroups は期限日ごとのグループ内でカードをシャッフルする
// 入力は期限日昇順にソート済みであることを前提とし、グループ間の順序は保つ
func shuffleWithinDueGroups(cards []domain.Card, rng *rand.Rand) []domain.Card {
	result := make([]domain.Card, len(cards))
	copy(result, cards)

	start := 0
	for start < len(result) {
		end := start + 1
		for end < len(result) && result[end].NextDue.String() == result[start].NextDue.String() {
			end++
		}
		group := result[start:end]
		rng.Shuffle(len(group), func(i, j int) {
			group[i], group[j] = group[j], group[i]
		})
		start = end
	}
	return result
}

// missedCardsView はセッション完了画面に不正解カードの一覧を描画する
func (m Model) missedCardsView() string {
	var builder strings.Builder
	builder.WriteString("\n")

	if len(m.sessionMisses) == 0 {
		builder.WriteString(view.SuccessStyle.Render("No misses. Nice.") + "\n")
		return builder.String()
	}

	builder.WriteString(view.SubtitleStyle.Render(fmt.Sprintf("Missed cards (%d)", len(m.sessionMisses))) + "\n")
	builder.WriteString(view.DividerStyle.Render(strings.Repeat("─", 24)) + "\n")
	for _, card := range m.sessionMisses {
		boxStyle := view.BoxStyleForLevel(card.Box)
		line := view.KanjiStyle.Render(card.Kanji)
		if card.Hiragana != nil && *card.Hiragana != "" {
			line += " " + view.HiraganaStyle.Render(*card.Hiragana)
		}
		line += "  " + view.ValueStyle.Render(card.English)
		line += "  " + view.LabelStyle.Render("Box ") + boxStyle.Render(fmt.Sprintf("%d", card.Box))
		builder.WriteString(line + "\n")
		if card.Usage != "" {
			builder.WriteString("  " + view.UsageStyle.Render(highlightUsage(card.Usage, card.Kanji, card.Hiragana)) + "\n")
		}
	}
	return builder.String()
}

func (m Model) applyAnswer(correct bool) Model {
	idx := m.reviewOrder[m.reviewIndex]
	answeredAt := m.now()
	answerDuration := time.Duration(0)
	if !m.reviewCardAt.IsZero() {
		answerDuration = answeredAt.Sub(m.reviewCardAt)
		if answerDuration < 0 {
			answerDuration = 0
		}
	}
	updated, err := domain.ApplyReview(m.cards[idx], correct, answerDuration, domain.Today())
	if err != nil {
		m.reviewErr = err
		return m
	}

	m.cards[idx] = updated
	m.reviewQueue[m.reviewIndex] = updated
	if err := storage.SaveCards(storage.CardsPath(m.dataDir), m.cards); err != nil {
		m.reviewErr = err
		return m
	}

	m.session.Reviewed++
	if correct {
		m.session.Correct++
	} else {
		m.session.Incorrect++
		m.sessionMisses = append(m.sessionMisses, updated)
	}
	if m.state.DailyStats == nil {
		m.state.DailyStats = map[string]domain.SessionStats{}
	}
	m.state.DailyStats[domain.Today().String()] = m.session
	if err := storage.SaveState(storage.StatePath(m.dataDir), m.state); err != nil {
		m.stateWarning = "Warning: Failed to save state"
	}

	m.reviewIndex++
	m.reviewCardAt = answeredAt
	m.reveal = false
	return m
}

func (m Model) currentReviewMode() reviewMode {
	if m.reviewIndex < 0 || m.reviewIndex >= len(m.reviewModes) {
		return reviewModeRecognition
	}
	return m.reviewModes[m.reviewIndex]
}

func (m Model) currentElapsed() time.Duration {
	if m.reviewCardAt.IsZero() {
		return 0
	}
	now := time.Now
	if m.now != nil {
		now = m.now
	}
	elapsed := now().Sub(m.reviewCardAt)
	if elapsed < 0 {
		return 0
	}
	return elapsed
}

func formatElapsed(elapsed time.Duration) string {
	return fmt.Sprintf("%.1fs", elapsed.Seconds())
}

func clozeFront(sentence string, kanji string, hiragana *string) string {
	hidden := sentence
	if kanji != "" && strings.Contains(hidden, kanji) {
		return strings.Replace(hidden, kanji, "（　）", 1)
	}
	if hiragana != nil && *hiragana != "" && strings.Contains(hidden, *hiragana) {
		return strings.Replace(hidden, *hiragana, "（　）", 1)
	}
	return hidden
}

func clozeBack(card domain.Card) string {
	if card.Hiragana != nil && *card.Hiragana != "" {
		return fmt.Sprintf("%s, %s, %s", card.Kanji, card.English, *card.Hiragana)
	}
	return fmt.Sprintf("%s, %s", card.Kanji, card.English)
}

func (m Model) loadStats() Model {
	m.stateErr = nil

	dataDir, err := storage.ResolveDataDir("")
	if err != nil {
		m.stateErr = err
		return m
	}
	state, err := storage.LoadState(storage.StatePath(dataDir))
	if err != nil {
		m.stateErr = err
		return m
	}

	m.dataDir = dataDir
	m.state = normalizeState(state, domain.Today())
	m.session = m.state.DailyStats[domain.Today().String()]
	return m
}

func normalizeState(state storage.AppState, today domain.LocalDate) storage.AppState {
	if state.DailyStats == nil {
		state.DailyStats = map[string]domain.SessionStats{}
	}
	if _, ok := state.DailyStats[today.String()]; !ok {
		state.DailyStats[today.String()] = state.SessionStats
	}
	return state
}

func recentDates(stats map[string]domain.SessionStats, limit int) []string {
	if limit <= 0 {
		return []string{}
	}

	keys := make([]string, 0, len(stats))
	for key := range stats {
		keys = append(keys, key)
	}

	sort.Slice(keys, func(i, j int) bool {
		left, leftErr := time.Parse("2006-01-02", keys[i])
		right, rightErr := time.Parse("2006-01-02", keys[j])
		if leftErr != nil || rightErr != nil {
			return keys[i] > keys[j]
		}
		return left.After(right)
	})

	if len(keys) > limit {
		keys = keys[:limit]
	}
	return keys
}
