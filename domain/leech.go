package domain

// リーチ判定の下限。復習回数が少ないうちは判定しない
const LeechMinReviews = 8

// この正答率未満のカードをリーチとみなす
const LeechAccuracyFloor = 40

// AccuracyPercent は正答率を整数パーセントで返す
func AccuracyPercent(card Card) int {
	if card.ReviewCount <= 0 {
		return 0
	}
	return card.CorrectCount * 100 / card.ReviewCount
}

// IsLeech は何度復習しても定着しないカードかを返す。
// 毎日出題されて正解できないカードは学習時間を食うだけなので、使用例の追加か停止を促す
func IsLeech(card Card) bool {
	if card.ReviewCount < LeechMinReviews {
		return false
	}
	return AccuracyPercent(card) < LeechAccuracyFloor
}
