package domain

import (
	"fmt"
	"time"
)

const (
	BoxMin = 1
	BoxMax = 6
)

var boxIntervalsByLevel = map[int]int{
	1: 1,
	2: 3,
	3: 7,
	4: 14,
	5: 30,
	6: 60,
}

const FastAnswerThreshold = 3 * time.Second

func ClampBox(box int) int {
	if box < BoxMin {
		return BoxMin
	}
	if box > BoxMax {
		return BoxMax
	}
	return box
}

func BoxIntervalDays(box int) int {
	if interval, ok := boxIntervalsByLevel[ClampBox(box)]; ok {
		return interval
	}
	return boxIntervalsByLevel[BoxMin]
}

func MoveBox(box int, correct bool, answerDuration time.Duration) int {
	current := ClampBox(box)
	if correct {
		if answerDuration > FastAnswerThreshold {
			return current
		}
		return ClampBox(current + 1)
	}
	return ClampBox(current - 1)
}

func NextDueForBox(today LocalDate, box int) (LocalDate, error) {
	interval := BoxIntervalDays(box)
	if interval <= 0 {
		return "", fmt.Errorf("invalid box interval: %d", interval)
	}
	return today.AddDays(interval)
}

func ApplyReview(card Card, correct bool, answerDuration time.Duration, today LocalDate) (Card, error) {
	updated := card
	updated.Box = MoveBox(card.Box, correct, answerDuration)
	if updated.Box > BoxMin {
		updated.Relearning = false
	} else if ClampBox(card.Box) > BoxMin {
		updated.Relearning = true
	}
	updated.LastReviewed = &today
	updated.ReviewCount++
	if correct {
		updated.CorrectCount++
	}

	nextDue, err := NextDueForBox(today, updated.Box)
	if err != nil {
		return Card{}, err
	}
	updated.NextDue = nextDue
	return updated, nil
}
