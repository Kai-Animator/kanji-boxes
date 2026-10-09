package domain

import (
	"sort"
	"strings"
)

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

// ボックス1で同時に学習する新規カードの上限。
// Leitner は区画を枚数ではなく物理的な幅で定めていたため、一日で無理なく回せる枚数として設定
const BoxOneCap = 20

// ActiveCards はボックス1の新規カードを上限まで絞り込む。
// 作成日順で固定の顔ぶれにし、昇格で空いた枠に次のカードが入る。
// 再学習中のカードとボックス2以上のカードは常に含める
func ActiveCards(cards []Card, limit int) []Card {
	newcomers := make([]Card, 0, len(cards))
	for _, card := range cards {
		if isCappedNewcomer(card) {
			newcomers = append(newcomers, card)
		}
	}
	sort.SliceStable(newcomers, func(i, j int) bool {
		if newcomers[i].CreatedAt != newcomers[j].CreatedAt {
			return newcomers[i].CreatedAt.String() < newcomers[j].CreatedAt.String()
		}
		return strings.Compare(newcomers[i].ID, newcomers[j].ID) < 0
	})
	admitted := make(map[string]bool, limit)
	for i := 0; i < len(newcomers) && i < limit; i++ {
		admitted[newcomers[i].ID] = true
	}

	active := make([]Card, 0, len(cards))
	for _, card := range cards {
		if !isCappedNewcomer(card) || admitted[card.ID] {
			active = append(active, card)
		}
	}
	return active
}

// WaitingCount は上限により待機中のボックス1カード数を返す
func WaitingCount(cards []Card, limit int) int {
	count := 0
	for _, card := range cards {
		if isCappedNewcomer(card) {
			count++
		}
	}
	if count <= limit {
		return 0
	}
	return count - limit
}

func isCappedNewcomer(card Card) bool {
	return ClampBox(card.Box) == BoxMin && !card.Relearning
}
