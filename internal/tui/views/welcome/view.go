package welcome

import (
	"ssh-messer/internal/tui/components/config"
	"ssh-messer/internal/tui/models"
	"ssh-messer/internal/tui/views/styles"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// RenderWelcomeView 渲染欢迎视图
func RenderWelcomeView(state *models.WelcomeViewState, configComponent *config.ConfigComponent, ui *models.UIState) string {
	// 获取终端尺寸，设置默认值
	width, height := ui.Width, ui.Height
	if width == 0 {
		width = 80
	}
	if height == 0 {
		height = 24
	}

	ascii := renderASCIIArt(state.WelcomeAnimationProgress)
	config := configComponent.View()
	footer := renderFooter()

	return lipgloss.Place(
		width, height,
		lipgloss.Center, lipgloss.Center,
		lipgloss.JoinVertical(
			lipgloss.Center,
			lipgloss.Place(width, height/2, lipgloss.Center, lipgloss.Center, ascii),
			lipgloss.Place(width, height/2-2, lipgloss.Center, 10, config),
			lipgloss.Place(width, 2, lipgloss.Center, lipgloss.Center, footer),
		),
		lipgloss.WithWhitespaceBackground(styles.Bg),
	)
}

// renderASCIIArt 创建动画 ASCII 艺术
func renderASCIIArt(progress int) string {
	lines := strings.Split(styles.AsciiArt, "\n")

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
					visibleLine += styles.AsciiStyle.Render(string(char))
				} else {
					visibleLine += styles.AsciiStyle.Foreground(lipgloss.Color("#FFFFFF")).Render(string(char))
				}
			} else {
				visibleLine += " "
			}
		}
		result = append(result, visibleLine)
	}

	return strings.Join(result, "\n")
}

// renderFooter 渲染页脚
func renderFooter() string {
	return lipgloss.JoinVertical(
		lipgloss.Center,
		styles.MetaStyle.Render("© 2025 Oliver.W • v0.1.0"),
		styles.MetaStyle.Render("MIT • https://github.com/ole3021/ssh-messer"),
	)
}
