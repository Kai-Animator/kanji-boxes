package view

import "github.com/charmbracelet/lipgloss"

var (
	// 基本スタイル
	TitleStyle      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	CursorStyle     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("208"))
	HintStyle       = lipgloss.NewStyle().Faint(true).Foreground(lipgloss.Color("244"))
	HighlightStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214"))
	SubtitleStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("75"))
	LabelStyle      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("250"))
	DividerStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
	ErrorStyle      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196"))
	SuccessStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("82"))
	MenuStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("111"))
	MenuActiveStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("220"))
	ValueStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("180"))
	InputStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("151"))
	InputFocusStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("45"))

	// ボックスレベル色（SRS進捗を示すグラデーション）
	BoxStyle1 = lipgloss.NewStyle().Foreground(lipgloss.Color("196")) // 赤 - ボックス1（新規/苦手）
	BoxStyle2 = lipgloss.NewStyle().Foreground(lipgloss.Color("208")) // オレンジ - ボックス2
	BoxStyle3 = lipgloss.NewStyle().Foreground(lipgloss.Color("220")) // 黄色 - ボックス3
	BoxStyle4 = lipgloss.NewStyle().Foreground(lipgloss.Color("118")) // 黄緑 - ボックス4
	BoxStyle5 = lipgloss.NewStyle().Foreground(lipgloss.Color("46"))  // 緑 - ボックス5（習得済み）
	BoxStyle6 = lipgloss.NewStyle().Foreground(lipgloss.Color("51"))  // 水色 - ボックス6（長期記憶）

	// レビューモード表示
	RecognitionModeStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("135")) // 紫
	ProductionModeStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("33"))  // 青
	ClozeModeStyle       = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("172")) // オレンジ

	// 日本語テキスト用
	KanjiStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("231")) // 明るい白
	HiraganaStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("183"))            // 薄紫
	UsageStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))            // 薄灰色

	// プログレスバー
	ProgressFillStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("76"))  // 緑
	ProgressEmptyStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("236")) // 暗灰色

	// カードフレーム
	CardBorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(1, 2)

	// 統計表示
	CorrectStatStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("82"))  // 緑
	IncorrectStatStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196")) // 赤
	NeutralStatStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("75"))  // 青

	// メニュー枠
	MenuBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1)
)

// BoxStyleForLevel はボックスレベルに応じたスタイルを返す
func BoxStyleForLevel(box int) lipgloss.Style {
	switch box {
	case 1:
		return BoxStyle1
	case 2:
		return BoxStyle2
	case 3:
		return BoxStyle3
	case 4:
		return BoxStyle4
	case 5:
		return BoxStyle5
	case 6:
		return BoxStyle6
	default:
		return BoxStyle1
	}
}

// RenderProgressBar はプログレスバーを描画する
func RenderProgressBar(current, total, width int) string {
	if total == 0 {
		return ""
	}
	filled := (current * width) / total
	if filled > width {
		filled = width
	}
	empty := width - filled

	bar := ProgressFillStyle.Render(repeatChar('█', filled)) +
		ProgressEmptyStyle.Render(repeatChar('░', empty))
	return bar
}

func repeatChar(char rune, count int) string {
	if count <= 0 {
		return ""
	}
	result := make([]rune, count)
	for i := range result {
		result[i] = char
	}
	return string(result)
}
