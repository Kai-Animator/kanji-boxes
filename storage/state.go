package storage

import "kanji-boxes/domain"

type AppState struct {
	SessionStats domain.SessionStats            `json:"sessionStats"`
	DailyStats   map[string]domain.SessionStats `json:"dailyStats"`
}
