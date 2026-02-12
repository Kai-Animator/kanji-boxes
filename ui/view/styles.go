package view

import "github.com/charmbracelet/lipgloss"

var (
	TitleStyle      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	CursorStyle     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("208"))
	HintStyle       = lipgloss.NewStyle().Faint(true).Foreground(lipgloss.Color("244"))
	HighlightStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214"))
	SubtitleStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("75"))
	LabelStyle      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("250"))
	DividerStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
	ErrorStyle      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196"))
	SuccessStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("82"))
	MenuStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("111"))
	MenuActiveStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("220"))
	ValueStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("180"))
	InputStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("151"))
	InputFocusStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("45"))
)
