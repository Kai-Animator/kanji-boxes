package storage

import (
	"encoding/csv"
	"io"
	"strings"

	"kanji-boxes/domain"
)

var csvHeader = []string{"kanji", "hiragana", "english", "tags"}

func ExportCardsCSV(writer io.Writer, cards []domain.Card) error {
	csvWriter := csv.NewWriter(writer)
	if err := csvWriter.Write(csvHeader); err != nil {
		return err
	}

	for _, card := range cards {
		hiragana := ""
		if card.Hiragana != nil {
			hiragana = *card.Hiragana
		}
		record := []string{
			card.Kanji,
			hiragana,
			card.English,
			strings.Join(card.Tags, "|"),
		}
		if err := csvWriter.Write(record); err != nil {
			return err
		}
	}

	csvWriter.Flush()
	return csvWriter.Error()
}

func ImportCardsCSV(reader io.Reader, today domain.LocalDate) ([]domain.Card, error) {
	csvReader := csv.NewReader(reader)
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return []domain.Card{}, nil
	}

	startIndex := 0
	if isCSVHeader(records[0]) {
		startIndex = 1
	}

	cards := make([]domain.Card, 0, len(records))
	for i := startIndex; i < len(records); i++ {
		record := records[i]
		if len(record) < 3 {
			continue
		}

		var hiragana *string
		if record[1] != "" {
			value := record[1]
			hiragana = &value
		}

		tags := []string{}
		if len(record) >= 4 && record[3] != "" {
			tags = strings.Split(record[3], "|")
		}

		cards = append(cards, domain.Card{
			Kanji:     record[0],
			Hiragana:  hiragana,
			English:   record[2],
			Tags:      tags,
			Box:       domain.BoxMin,
			CreatedAt: today,
			NextDue:   today,
		})
	}

	return cards, nil
}

func isCSVHeader(record []string) bool {
	if len(record) < len(csvHeader) {
		return false
	}
	for i, field := range csvHeader {
		if strings.TrimSpace(strings.ToLower(record[i])) != field {
			return false
		}
	}
	return true
}
