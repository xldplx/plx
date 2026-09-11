package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Palette
	colorSubtle  = lipgloss.Color("#626880")
	colorDim     = lipgloss.Color("#414559")
	colorBorder  = lipgloss.Color("#303446")
	colorActive  = lipgloss.Color("#ca9ee6") // Purple accent
	colorCyan    = lipgloss.Color("#85c1dc")
	colorGreen   = lipgloss.Color("#a6d189")
	colorYellow  = lipgloss.Color("#e5c890")
	colorRed     = lipgloss.Color("#e78284")
	colorWhite   = lipgloss.Color("#c6d0f5")
	colorBgDark  = lipgloss.Color("#232634")
	colorBgLight = lipgloss.Color("#292c3c")

	// Base Styles
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#ffffff")).
			Background(lipgloss.Color("#7347eb")).
			Padding(0, 1)

	badgeStyle = lipgloss.NewStyle().
			Foreground(colorWhite).
			Background(colorBgLight).
			Padding(0, 1)

	dirtyBadgeStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorYellow)

	cleanBadgeStyle = lipgloss.NewStyle().
			Foreground(colorGreen)

	branchStyle = lipgloss.NewStyle().
			Foreground(colorCyan)

	ageStyle = lipgloss.NewStyle().
			Foreground(colorSubtle)

	aheadStyle = lipgloss.NewStyle().
			Foreground(colorGreen)

	behindStyle = lipgloss.NewStyle().
			Foreground(colorRed)

	selectedRowStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#ffffff")).
				Background(lipgloss.Color("#414559"))

	normalRowStyle = lipgloss.NewStyle().
			Foreground(colorWhite)

	previewBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(0, 1)

	footerStyle = lipgloss.NewStyle().
			Foreground(colorSubtle).
			Padding(0, 1)

	keyStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorActive)
)
