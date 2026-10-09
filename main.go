package main

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Kai-Animator/kanji-boxes/model"
)

func main() {
	program := tea.NewProgram(model.New())
	if _, err := program.Run(); err != nil {
		log.Fatalf("failed to start program: %v", err)
	}
}
