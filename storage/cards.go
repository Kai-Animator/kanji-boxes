package storage

import "github.com/Kai-Animator/kanji-boxes/domain"

func LoadCards(path string) ([]domain.Card, error) {
	var cards []domain.Card
	if err := ReadJSON(path, &cards); err != nil {
		return nil, err
	}
	if cards == nil {
		cards = []domain.Card{}
	}
	return cards, nil
}

func SaveCards(path string, cards []domain.Card) error {
	return WriteJSONAtomic(path, cards)
}

func LoadState(path string) (AppState, error) {
	state := AppState{}
	if err := ReadJSON(path, &state); err != nil {
		return AppState{}, err
	}
	return state, nil
}

func SaveState(path string, state AppState) error {
	return WriteJSONAtomic(path, state)
}
