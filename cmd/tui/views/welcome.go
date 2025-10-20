package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// // WelcomeRenderer defines the interface for welcome screen rendering
type WelcomeRenderer interface {
	GetWelcomeProgress() int
	GetConfigChoices() []string
	GetConfigCursor() int
	GetWidth() int
	GetHeight() int
}

var (
	author     = "Oliver.W"
	version    = "0.1.0"
	copyright  = "2025"
	license    = "MIT"
	repository = "https://github.com/ole3021/ssh-messer"

	// ASCII art style with center alignment
	AsciiStyle = lipgloss.NewStyle().
			Foreground(Primary).
			Align(lipgloss.Center)
)

// ASCII art for the welcome screen
const AsciiArt = `
    ███╗   ███████╗███████╗██╗  ██╗    ███╗   ███╗███████╗███████╗███████╗███████╗██████╗ 
   ████║   ██╔════╝██╔════╝██║  ██║    ████╗ ████║██╔════╝██╔════╝██╔════╝██╔════╝██╔══██╗
  ██╔██║   ███████╗███████╗███████║    ██╔████╔██║█████╗  ███████╗█████╗  █████╗  ██████╔╝
 ██╔╝██║   ╚════██║╚════██║██╔══██║    ██║╚██╔╝██║██╔══╝  ╚════██║██╔══╝  ██╔══╝  ██╔══██╗
██╔╝ ██║   ███████║███████║██║  ██║    ██║ ╚═╝ ██║███████╗███████║███████╗███████╗██║  ██║
╚═╝  ╚═╝   ╚══════╝╚══════╝╚═╝  ╚═╝    ╚═╝     ╚═╝╚══════╝╚══════╝╚══════╝╚═╝  ╚═╝╚═╝  ╚═╝
`

// RenderWelcomeScreen renders the welcome screen using optimized Lip Gloss approach
func RenderWelcomeView(renderer WelcomeRenderer) string {
	// Get terminal dimensions with defaults
	width, height := renderer.GetWidth(), renderer.GetHeight()
	if width == 0 {
		width = 80
	}
	if height == 0 {
		height = 24
	}

	ascii := renderASCIIArt(renderer.GetWelcomeProgress())
	config := RenderConfigSimpleList(renderer)
	footer := RenderFooter()

	return lipgloss.Place(
		width, height,
		lipgloss.Center, lipgloss.Center,
		lipgloss.JoinVertical(
			lipgloss.Center,
			lipgloss.Place(width, height/2, lipgloss.Center, lipgloss.Center, ascii),
			lipgloss.Place(width, height/2-2, lipgloss.Center, 10, config),
			lipgloss.Place(width, 2, lipgloss.Center, lipgloss.Center, footer),
		),
		lipgloss.WithWhitespaceBackground(Bg),
	)
}

// renderASCIIArt creates animated ASCII art using Lip Gloss
func renderASCIIArt(progress int) string {
	lines := strings.Split(AsciiArt, "\n")

	// 计算每行应该显示多少字符
	var result []string
	for i, line := range lines {
		// 每行有不同的延迟，创造波浪效果
		lineDelay := i * 15 // 每行延迟15%
		lineProgress := progress - lineDelay
		if lineProgress < 0 {
			lineProgress = 0
		}

		visibleChars := (lineProgress * len(line)) / (100 - lineDelay)
		if visibleChars > len(line) {
			visibleChars = len(line)
		}

		// 逐字符显示，带闪烁效果
		var visibleLine string
		for j, char := range line {
			if j < visibleChars {
				// 添加闪烁效果
				if (progress/5)%2 == 0 {
					visibleLine += AsciiStyle.Render(string(char))
				} else {
					visibleLine += AsciiStyle.Foreground(lipgloss.Color("#FFFFFF")).Render(string(char))
				}
			} else {
				visibleLine += " "
			}
		}
		result = append(result, visibleLine)
	}

	return strings.Join(result, "\n")
}

func RenderFooter() string {
	return lipgloss.JoinVertical(
		lipgloss.Center,
		MetaStyle.Render(fmt.Sprintf("© %s %s • v%s", copyright, author, version)),
		MetaStyle.Render(fmt.Sprintf("%s • %s", license, repository)),
	)

}
