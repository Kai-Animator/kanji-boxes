package storage

import (
	"encoding/csv"
	"io"
	"strconv"
	"strings"

	"kanji-boxes/domain"
)

var csvHeader = []string{"kanji", "hiragana", "english", "usage", "tags", "box"}
var csvHeaderV1 = []string{"kanji", "hiragana", "english", "usage", "tags"}
var csvHeaderLegacy = []string{"kanji", "hiragana", "english", "tags"}

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
			card.Usage,
			strings.Join(card.Tags, "|"),
			strconv.Itoa(card.Box),
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
	// デフォルトはレガシー4カラム形式
	columns := csvColumns{kanji: 0, hiragana: 1, english: 2, usage: -1, tags: 3, box: -1}
	if isCSVHeader(records[0]) {
		startIndex = 1
		if matchCSVHeader(records[0], csvHeader) {
			columns = csvColumns{kanji: 0, hiragana: 1, english: 2, usage: 3, tags: 4, box: 5}
		} else if matchCSVHeader(records[0], csvHeaderV1) {
			columns = csvColumns{kanji: 0, hiragana: 1, english: 2, usage: 3, tags: 4, box: -1}
		}
	} else if len(records[0]) >= 6 {
		columns = csvColumns{kanji: 0, hiragana: 1, english: 2, usage: 3, tags: 4, box: 5}
	} else if len(records[0]) >= 5 {
		columns = csvColumns{kanji: 0, hiragana: 1, english: 2, usage: 3, tags: 4, box: -1}
	}

	cards := make([]domain.Card, 0, len(records))
	for i := startIndex; i < len(records); i++ {
		record := records[i]
		if len(record) < 3 {
			continue
		}

		var hiragana *string
		if value := fieldAt(record, columns.hiragana); value != "" {
			hiraganaValue := value
			hiragana = &hiraganaValue
		}

		usage := fieldAt(record, columns.usage)

		tags := []string{}
		if value := fieldAt(record, columns.tags); value != "" {
			tags = strings.Split(value, "|")
		}

		// boxカラムがあれば復元、なければデフォルト値を使用
		box := domain.BoxMin
		if boxStr := fieldAt(record, columns.box); boxStr != "" {
			if b, err := strconv.Atoi(boxStr); err == nil && b >= domain.BoxMin && b <= domain.BoxMax {
				box = b
			}
		}

		cards = append(cards, domain.Card{
			Kanji:     fieldAt(record, columns.kanji),
			Hiragana:  hiragana,
			English:   fieldAt(record, columns.english),
			Usage:     usage,
			Tags:      tags,
			Box:       box,
			CreatedAt: today,
			NextDue:   today,
		})
	}

	return cards, nil
}

type csvColumns struct {
	kanji    int
	hiragana int
	english  int
	usage    int
	tags     int
	box      int
}

func fieldAt(record []string, index int) string {
	if index < 0 || index >= len(record) {
		return ""
	}
	return record[index]
}

func isCSVHeader(record []string) bool {
	return matchCSVHeader(record, csvHeader) || matchCSVHeader(record, csvHeaderV1) || matchCSVHeader(record, csvHeaderLegacy)
}

func matchCSVHeader(record []string, header []string) bool {
	if len(record) < len(header) {
		return false
	}
	for i, field := range header {
		if strings.TrimSpace(strings.ToLower(record[i])) != field {
			return false
		}
	}
	return true
}
