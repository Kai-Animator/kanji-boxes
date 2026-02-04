package domain

import "sort"

func IsDue(card Card, today LocalDate) bool {
	return card.NextDue.String() <= today.String()
}

func DueCards(cards []Card, today LocalDate) []Card {
	due := make([]Card, 0, len(cards))
	for _, card := range cards {
		if IsDue(card, today) {
			due = append(due, card)
		}
	}

	sort.SliceStable(due, func(i, j int) bool {
		if due[i].NextDue.String() == due[j].NextDue.String() {
			return due[i].Kanji < due[j].Kanji
		}
		return due[i].NextDue.String() < due[j].NextDue.String()
	})

	return due
}
