package model

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

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

type Model struct {
	choices      []string
	cursor       int
	selected     int
	quitting     bool
	screen       screen
	dataDir      string
	cards        []domain.Card
	reviewQueue  []domain.Card
	reviewOrder  []int
	reviewIndex  int
	reveal       bool
	reviewErr    error
	session      domain.SessionStats
	addInputs    []textinput.Model
	addFocus     int
	addMessage   string
	editInputs   []textinput.Model
	editFocus    int
	editMessage  string
	editID       string
	state        storage.AppState
	stateErr     error
	browseCards  []domain.Card
	browseErr    error
	browseCursor int
	ioMode       ioMode
	ioAction     string
	ioCursor     int
	ioInput      textinput.Model
	ioMessage    string
	ioErr        error
}

func New() Model {
	return Model{
		choices:  []string{"Review", "Add Card", "Browse", "Import / Export", "Stats", "Quit"},
		cursor:   0,
		selected: -1,
		screen:   screenMenu,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()
		switch key {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "q":
			if m.screen != screenAddCard && m.screen != screenImportExport && m.screen != screenEditCard {
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
			switch key {
			case "esc":
				m.screen = screenMenu
			case "up", "k":
				if m.browseCursor > 0 {
					m.browseCursor--
				}
			case "down", "j":
				if m.browseCursor < len(m.browseCards)-1 {
					m.browseCursor++
				}
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
	builder.WriteString(view.TitleStyle.Render("Kanji Boxes") + "\n\n")

	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = view.CursorStyle.Render(">")
		}
		builder.WriteString(cursor + " " + choice + "\n")
	}

	builder.WriteString("\n")
	builder.WriteString(view.HintStyle.Render("up/down: move • enter: select • q: quit") + "\n")

	return builder.String()
}

func (m Model) reviewView() string {
	var builder strings.Builder
	builder.WriteString(view.TitleStyle.Render("Review") + "\n\n")

	if m.reviewErr != nil {
		builder.WriteString("Failed to load cards: " + m.reviewErr.Error() + "\n")
		builder.WriteString("\n")
		builder.WriteString(view.HintStyle.Render("esc: back • q: quit") + "\n")
		return builder.String()
	}

	builder.WriteString(fmt.Sprintf("Due cards: %d\n\n", len(m.reviewQueue)))
	if len(m.reviewQueue) == 0 {
		builder.WriteString("No cards due today.\n")
	} else {
		builder.WriteString(fmt.Sprintf("Progress: %d/%d reviewed\n\n", m.reviewIndex, len(m.reviewQueue)))
		if m.reviewIndex >= len(m.reviewQueue) {
			builder.WriteString("Session complete.\n")
		} else {
			card := m.reviewQueue[m.reviewIndex]
			builder.WriteString(fmt.Sprintf("%s\n", card.Kanji))
			if m.reveal {
				if card.Hiragana != nil {
					builder.WriteString(fmt.Sprintf("%s\n", *card.Hiragana))
				}
				builder.WriteString(fmt.Sprintf("%s\n", card.English))
			} else {
				builder.WriteString("(press space to flip)\n")
			}
		}
	}

	builder.WriteString("\n")
	builder.WriteString(view.HintStyle.Render("space: flip • y: correct • n: incorrect • esc: back • q: quit") + "\n")

	return builder.String()
}

func (m Model) statsView() string {
	var builder strings.Builder
	builder.WriteString(view.TitleStyle.Render("Stats") + "\n\n")
	if m.stateErr != nil {
		builder.WriteString("Failed to load stats: " + m.stateErr.Error() + "\n\n")
		builder.WriteString(view.HintStyle.Render("esc: back • q: quit") + "\n")
		return builder.String()
	}

	if len(m.state.DailyStats) == 0 {
		builder.WriteString("No stats yet.\n\n")
		builder.WriteString(view.HintStyle.Render("esc: back • q: quit") + "\n")
		return builder.String()
	}

	date := domain.Today().String()
	builder.WriteString(fmt.Sprintf("Today (%s)\n", date))
	builder.WriteString(fmt.Sprintf("Reviewed: %d\n", m.session.Reviewed))
	builder.WriteString(fmt.Sprintf("Correct: %d\n", m.session.Correct))
	builder.WriteString(fmt.Sprintf("Incorrect: %d\n", m.session.Incorrect))
	builder.WriteString("\n")

	builder.WriteString("History (last 7 days)\n")
	for _, day := range recentDates(m.state.DailyStats, 7) {
		stats := m.state.DailyStats[day]
		builder.WriteString(fmt.Sprintf("%s  %d/%d/%d\n", day, stats.Reviewed, stats.Correct, stats.Incorrect))
	}
	builder.WriteString("\n")
	builder.WriteString(view.HintStyle.Render("esc: back • q: quit") + "\n")
	return builder.String()
}

func (m Model) addCardView() string {
	var builder strings.Builder
	builder.WriteString(view.TitleStyle.Render("Add Card") + "\n\n")
	labels := []string{"Kanji*", "English*", "Hiragana", "Tags (| separated)"}
	for i, input := range m.addInputs {
		builder.WriteString(labels[i] + "\n")
		builder.WriteString(input.View() + "\n\n")
	}

	if m.addMessage != "" {
		builder.WriteString(m.addMessage + "\n\n")
	}
	builder.WriteString("\n")
	builder.WriteString(view.HintStyle.Render("tab: next • enter: save • esc: back • q: quit") + "\n")
	return builder.String()
}

func (m Model) browseView() string {
	var builder strings.Builder
	builder.WriteString(view.TitleStyle.Render("Browse") + "\n\n")

	if m.browseErr != nil {
		builder.WriteString("Failed to load cards: " + m.browseErr.Error() + "\n\n")
		builder.WriteString(view.HintStyle.Render("esc: back • q: quit") + "\n")
		return builder.String()
	}

	if len(m.browseCards) == 0 {
		builder.WriteString("No cards available.\n")
		builder.WriteString("\n")
		builder.WriteString(view.HintStyle.Render("esc: back • q: quit") + "\n")
		return builder.String()
	}

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
		if i == m.browseCursor {
			cursor = view.CursorStyle.Render(">")
		}
		builder.WriteString(fmt.Sprintf("%s %s - %s\n", cursor, card.Kanji, card.English))
	}

	selected := m.browseCards[m.browseCursor]
	builder.WriteString("\n")
	builder.WriteString("Details\n")
	if selected.Hiragana != nil {
		builder.WriteString(fmt.Sprintf("Hiragana: %s\n", *selected.Hiragana))
	}
	builder.WriteString(fmt.Sprintf("Box: %d\n", selected.Box))
	builder.WriteString(fmt.Sprintf("Next due: %s\n", selected.NextDue.String()))
	if len(selected.Tags) > 0 {
		builder.WriteString(fmt.Sprintf("Tags: %s\n", strings.Join(selected.Tags, ", ")))
	}

	builder.WriteString("\n")
	builder.WriteString(view.HintStyle.Render("up/down: move • e: edit • d: delete • esc: back • q: quit") + "\n")
	return builder.String()
}

func (m Model) editCardView() string {
	var builder strings.Builder
	builder.WriteString(view.TitleStyle.Render("Edit Card") + "\n\n")
	labels := []string{"Kanji*", "English*", "Hiragana", "Tags (| separated)"}
	for i, input := range m.editInputs {
		builder.WriteString(labels[i] + "\n")
		builder.WriteString(input.View() + "\n\n")
	}

	if m.editMessage != "" {
		builder.WriteString(m.editMessage + "\n\n")
	}
	builder.WriteString("\n")
	builder.WriteString(view.HintStyle.Render("tab: next • enter: save • esc: back • q: quit") + "\n")
	return builder.String()
}

func (m Model) deleteConfirmView() string {
	var builder strings.Builder
	builder.WriteString(view.TitleStyle.Render("Delete Card") + "\n\n")
	if len(m.browseCards) == 0 {
		builder.WriteString("No card selected.\n\n")
		builder.WriteString(view.HintStyle.Render("esc: back • q: quit") + "\n")
		return builder.String()
	}
	card := m.browseCards[m.browseCursor]
	builder.WriteString(fmt.Sprintf("Delete %s - %s?\n\n", card.Kanji, card.English))
	builder.WriteString(view.HintStyle.Render("y: delete • n: cancel") + "\n")
	return builder.String()
}

func (m Model) importExportView() string {
	var builder strings.Builder
	builder.WriteString(view.TitleStyle.Render("Import / Export") + "\n\n")

	if m.ioErr != nil {
		builder.WriteString("Error: " + m.ioErr.Error() + "\n\n")
	}
	if m.ioMessage != "" {
		builder.WriteString(m.ioMessage + "\n\n")
	}

	switch m.ioMode {
	case ioModeSelect:
		options := []string{"Import CSV", "Export CSV"}
		for i, option := range options {
			cursor := " "
			if i == m.ioCursor {
				cursor = view.CursorStyle.Render(">")
			}
			builder.WriteString(fmt.Sprintf("%s %s\n", cursor, option))
		}
		builder.WriteString("\n")
		builder.WriteString(view.HintStyle.Render("up/down: move • enter: select • esc: back • q: quit") + "\n")
	case ioModePath:
		label := "Import from file"
		if m.ioAction == "export" {
			label = "Export to file"
		}
		builder.WriteString(label + "\n")
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
	m.reviewQueue = domain.DueCards(cards, domain.Today())
	m.reviewOrder = indicesForCards(cards, m.reviewQueue)
	m.reviewIndex = 0
	m.reveal = false
	m.session = m.state.DailyStats[domain.Today().String()]
	_ = storage.SaveState(storage.StatePath(dataDir), m.state)
	return m
}

func (m Model) startAddCard() Model {
	inputs := make([]textinput.Model, 4)
	for i := range inputs {
		input := textinput.New()
		input.CharLimit = 120
		switch i {
		case 0:
			input.Placeholder = "例: 日"
		case 1:
			input.Placeholder = "例: day"
		case 2:
			input.Placeholder = "例: にち"
		case 3:
			input.Placeholder = "例: jlpt5|common"
		}
		inputs[i] = input
	}
	inputs[0].Focus()

	m.addInputs = inputs
	m.addFocus = 0
	m.addMessage = ""
	return m
}

func (m Model) startEditCard(card domain.Card) Model {
	inputs := make([]textinput.Model, 4)
	for i := range inputs {
		input := textinput.New()
		input.CharLimit = 120
		switch i {
		case 0:
			input.Placeholder = "例: 日"
			input.SetValue(card.Kanji)
		case 1:
			input.Placeholder = "例: day"
			input.SetValue(card.English)
		case 2:
			input.Placeholder = "例: にち"
			if card.Hiragana != nil {
				input.SetValue(*card.Hiragana)
			}
		case 3:
			input.Placeholder = "例: jlpt5|common"
			input.SetValue(strings.Join(card.Tags, "|"))
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
	m.browseCards = cards
	m.browseCursor = 0
	return m
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

func (m Model) saveAddCard() Model {
	kanji := strings.TrimSpace(m.addInputs[0].Value())
	english := strings.TrimSpace(m.addInputs[1].Value())
	hiragana := strings.TrimSpace(m.addInputs[2].Value())
	tagsInput := strings.TrimSpace(m.addInputs[3].Value())

	if kanji == "" || english == "" {
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

	var hiraPtr *string
	if hiragana != "" {
		hira := hiragana
		hiraPtr = &hira
	}

	var tags []string
	if tagsInput != "" {
		for _, tag := range strings.Split(tagsInput, "|") {
			trimmed := strings.TrimSpace(tag)
			if trimmed != "" {
				tags = append(tags, trimmed)
			}
		}
	}

	today := domain.Today()
	card := domain.Card{
		ID:        fmt.Sprintf("card-%d", time.Now().UnixNano()),
		Kanji:     kanji,
		Hiragana:  hiraPtr,
		English:   english,
		Tags:      tags,
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
	kanji := strings.TrimSpace(m.editInputs[0].Value())
	english := strings.TrimSpace(m.editInputs[1].Value())
	hiragana := strings.TrimSpace(m.editInputs[2].Value())
	tagsInput := strings.TrimSpace(m.editInputs[3].Value())

	if kanji == "" || english == "" {
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

	var hiraPtr *string
	if hiragana != "" {
		hira := hiragana
		hiraPtr = &hira
	}

	var tags []string
	if tagsInput != "" {
		for _, tag := range strings.Split(tagsInput, "|") {
			trimmed := strings.TrimSpace(tag)
			if trimmed != "" {
				tags = append(tags, trimmed)
			}
		}
	}

	updated := false
	for i := range cards {
		if cards[i].ID == m.editID {
			cards[i].Kanji = kanji
			cards[i].English = english
			cards[i].Hiragana = hiraPtr
			cards[i].Tags = tags
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

func (m Model) performImportExport() Model {
	m.ioErr = nil
	m.ioMessage = ""

	path := strings.TrimSpace(m.ioInput.Value())
	if path == "" {
		m.ioErr = fmt.Errorf("path is required")
		return m
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
		for i := range imported {
			if imported[i].ID == "" {
				imported[i].ID = fmt.Sprintf("card-%d", time.Now().UnixNano()+int64(i))
			}
		}
		existing, err := storage.LoadCards(storage.CardsPath(dataDir))
		if err != nil {
			m.ioErr = err
			return m
		}
		existing = append(existing, imported...)
		if err := storage.SaveCards(storage.CardsPath(dataDir), existing); err != nil {
			m.ioErr = err
			return m
		}
		m.ioMessage = fmt.Sprintf("Imported %d cards.", len(imported))
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

func (m Model) applyAnswer(correct bool) Model {
	idx := m.reviewOrder[m.reviewIndex]
	updated, err := domain.ApplyReview(m.cards[idx], correct, domain.Today())
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
	}
	if m.state.DailyStats == nil {
		m.state.DailyStats = map[string]domain.SessionStats{}
	}
	m.state.DailyStats[domain.Today().String()] = m.session
	_ = storage.SaveState(storage.StatePath(m.dataDir), m.state)

	m.reviewIndex++
	m.reveal = false
	return m
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
