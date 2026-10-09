package domain

type Card struct {
	ID           string
	Kanji        string
	Hiragana     *string
	English      string
	Usage        string
	Box          int
	CreatedAt    LocalDate
	LastReviewed *LocalDate
	NextDue      LocalDate
	Tags         []string
	ReviewCount  int
	CorrectCount int
	Suspended    bool
	// 上位ボックスから落ちて再学習中のカード。ボックス1の上限の対象外
	Relearning bool
}

type SessionStats struct {
	Reviewed  int
	Correct   int
	Incorrect int
}
