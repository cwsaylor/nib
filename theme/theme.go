package theme

import "github.com/charmbracelet/lipgloss"

var (
	AppStyle = lipgloss.NewStyle().Padding(1, 2)

	TitleBar = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#303446")).
			Background(lipgloss.Color("#8AABF4")).
			Padding(0, 1).
			MarginBottom(1)

	StatusBar = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A49FA5")).
			Background(lipgloss.Color("#353533")).
			Padding(0, 1)

	PreviewBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#8AABF4")).
			Padding(0, 1)

	SelectedItem = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#303446")).
			Background(lipgloss.Color("#8AABF4")).
			Padding(0, 1)

	NormalItem = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#DDDDDD")).
			Padding(0, 1)

	SubtleText = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6C6C6C"))

	ErrorText = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true)

	ConfirmDialog = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#FF5555")).
			Padding(1, 3).
			Width(50).
			Align(lipgloss.Center)

	HelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))
)
