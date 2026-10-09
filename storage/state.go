package storage

import "github.com/Kai-Animator/kanji-boxes/domain"

type AppState struct {
	SessionStats domain.SessionStats            `json:"sessionStats"`
	DailyStats   map[string]domain.SessionStats `json:"dailyStats"`
}
