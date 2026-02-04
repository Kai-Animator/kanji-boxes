package domain

import "time"

const localDateLayout = "2006-01-02"

type LocalDate string

func Today() LocalDate {
	return LocalDate(time.Now().In(time.Local).Format(localDateLayout))
}

func ParseLocalDate(value string) (LocalDate, error) {
	_, err := time.ParseInLocation(localDateLayout, value, time.Local)
	if err != nil {
		return "", err
	}
	return LocalDate(value), nil
}

func (d LocalDate) String() string {
	return string(d)
}

func (d LocalDate) AddDays(days int) (LocalDate, error) {
	parsed, err := time.ParseInLocation(localDateLayout, string(d), time.Local)
	if err != nil {
		return "", err
	}
	return LocalDate(parsed.AddDate(0, 0, days).Format(localDateLayout)), nil
}
