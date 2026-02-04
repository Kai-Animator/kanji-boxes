package domain

type Card struct {
	ID           string
	Kanji        string
	Hiragana     *string
	English      string
	Box          int
	CreatedAt    LocalDate
	LastReviewed *LocalDate
	NextDue      LocalDate
	Tags         []string
	ReviewCount  int
	CorrectCount int
	Suspended    bool
}

type SessionStats struct {
	Reviewed  int
	Correct   int
	Incorrect int
}
