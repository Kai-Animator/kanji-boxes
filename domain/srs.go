package domain

import "fmt"

const (
	BoxMin = 1
	BoxMax = 5
)

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
	return ClampBox(box)
}

func MoveBox(box int, correct bool) int {
	current := ClampBox(box)
	if correct {
		return ClampBox(current + 1)
	}
	return ClampBox(current - 1)
}

func NextDueForBox(today LocalDate, box int) (LocalDate, error) {
	interval := BoxIntervalDays(box)
	if interval < BoxMin || interval > BoxMax {
		return "", fmt.Errorf("invalid box interval: %d", interval)
	}
	return today.AddDays(interval)
}

func ApplyReview(card Card, correct bool, today LocalDate) (Card, error) {
	updated := card
	updated.Box = MoveBox(card.Box, correct)
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
