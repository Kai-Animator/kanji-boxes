package domain

import "testing"

func TestIsLeech(t *testing.T) {
	tests := []struct {
		name     string
		reviews  int
		correct  int
		expected bool
	}{
		{"too few reviews", 7, 0, false},
		{"enough reviews, zero correct", 8, 0, true},
		{"below accuracy floor", 11, 3, true},
		{"on accuracy floor", 10, 4, false},
		{"healthy", 61, 41, false},
	}
	for _, tc := range tests {
		card := Card{ReviewCount: tc.reviews, CorrectCount: tc.correct}
		if got := IsLeech(card); got != tc.expected {
			t.Errorf("%s: expected %v, got %v", tc.name, tc.expected, got)
		}
	}
}

func TestAccuracyPercent(t *testing.T) {
	if got := AccuracyPercent(Card{}); got != 0 {
		t.Fatalf("expected 0 for no reviews, got %d", got)
	}
	if got := AccuracyPercent(Card{ReviewCount: 3, CorrectCount: 1}); got != 33 {
		t.Fatalf("expected 33, got %d", got)
	}
}
