package storage

import (
	"bytes"
	"encoding/csv"
	"strings"
	"testing"

	"kanji-boxes/domain"
)

func TestExportCardsCSVHeaderAndTags(t *testing.T) {
	hiragana := "にち"
	cards := []domain.Card{
		{
			Kanji:    "日",
			Hiragana: &hiragana,
			English:  "day",
			Usage:    "日が昇る。",
			Tags:     []string{"jlpt5", "common"},
			Box:      3,
		},
	}

	var buffer bytes.Buffer
	if err := ExportCardsCSV(&buffer, cards); err != nil {
		t.Fatalf("export: %v", err)
	}

	reader := csv.NewReader(bytes.NewReader(buffer.Bytes()))
	header, err := reader.Read()
	if err != nil {
		t.Fatalf("read header: %v", err)
	}
	expectedHeader := []string{"kanji", "hiragana", "english", "usage", "tags", "box"}
	for i, value := range expectedHeader {
		if header[i] != value {
			t.Fatalf("header mismatch at %d: got %s, want %s", i, header[i], value)
		}
	}

	record, err := reader.Read()
	if err != nil {
		t.Fatalf("read record: %v", err)
	}
	if record[3] != "日が昇る。" {
		t.Fatalf("unexpected usage value: %s", record[3])
	}
	if record[4] != "jlpt5|common" {
		t.Fatalf("unexpected tags value: %s", record[4])
	}
	if record[5] != "3" {
		t.Fatalf("unexpected box value: %s", record[5])
	}
}

func TestImportCardsCSVDefaultsAndTags(t *testing.T) {
	today, err := domain.ParseLocalDate("2026-02-04")
	if err != nil {
		t.Fatalf("parse date: %v", err)
	}

	// boxカラムなしのv1形式（後方互換）
	data := strings.NewReader("kanji,hiragana,english,usage,tags\n日,にち,day,日が昇る。,jlpt5|common\n")
	cards, err := ImportCardsCSV(data, today)
	if err != nil {
		t.Fatalf("import: %v", err)
	}
	if len(cards) != 1 {
		t.Fatalf("expected 1 card, got %d", len(cards))
	}
	if cards[0].Box != domain.BoxMin {
		t.Fatalf("expected box %d, got %d", domain.BoxMin, cards[0].Box)
	}
	if cards[0].NextDue.String() != today.String() {
		t.Fatalf("expected nextDue %s, got %s", today.String(), cards[0].NextDue.String())
	}
	if len(cards[0].Tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(cards[0].Tags))
	}
	if cards[0].Usage != "日が昇る。" {
		t.Fatalf("unexpected usage value: %s", cards[0].Usage)
	}
}

func TestImportCardsCSVBoxRestored(t *testing.T) {
	today, err := domain.ParseLocalDate("2026-02-04")
	if err != nil {
		t.Fatalf("parse date: %v", err)
	}

	// boxカラムありの新形式でボックス番号が復元されるか確認
	data := strings.NewReader("kanji,hiragana,english,usage,tags,box\n日,にち,day,日が昇る。,jlpt5|common,4\n山,,mountain,,nature,2\n")
	cards, err := ImportCardsCSV(data, today)
	if err != nil {
		t.Fatalf("import: %v", err)
	}
	if len(cards) != 2 {
		t.Fatalf("expected 2 cards, got %d", len(cards))
	}
	if cards[0].Box != 4 {
		t.Fatalf("expected box 4, got %d", cards[0].Box)
	}
	if cards[1].Box != 2 {
		t.Fatalf("expected box 2, got %d", cards[1].Box)
	}
}
