package view

import "testing"

func TestStylesRenderText(t *testing.T) {
	input := "Kanji"
	if got := TitleStyle.Render(input); got == "" {
		t.Fatalf("TitleStyle render empty")
	}
	if got := CursorStyle.Render(input); got == "" {
		t.Fatalf("CursorStyle render empty")
	}
	if got := HintStyle.Render(input); got == "" {
		t.Fatalf("HintStyle render empty")
	}
}
