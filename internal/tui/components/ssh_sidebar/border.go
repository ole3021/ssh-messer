package ssh_sidebar

import (
	"ssh-messer/internal/tui/styles"

	"github.com/charmbracelet/lipgloss/v2"
)

// BorderStyle 边框样式类型
type BorderStyle int

const (
	// BorderStyleNeon 方案1：霓虹单色边框
	BorderStyleNeon BorderStyle = iota
	// BorderStyleDual 方案2：双色渐变边框
	BorderStyleDual
	// BorderStyleGeometric 方案3：几何科技边框
	BorderStyleGeometric
	// BorderStyleGlow 方案4：发光边框
	BorderStyleGlow
)

// ApplyBorder 应用边框样式到内容
func ApplyBorder(content string, width, height int, style BorderStyle) string {
	switch style {
	case BorderStyleNeon:
		return applyNeonBorder(content, width, height)
	case BorderStyleDual:
		return applyDualColorBorder(content, width, height)
	case BorderStyleGeometric:
		return applyGeometricBorder(content, width, height)
	case BorderStyleGlow:
		return applyGlowBorder(content, width, height)
	default:
		return applyNeonBorder(content, width, height)
	}
}

// applyNeonBorder 方案1：霓虹单色边框
// 简洁的青色霓虹边框，使用粗边框
func applyNeonBorder(content string, width, height int) string {
	style := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Border(lipgloss.ThickBorder()).
		BorderForeground(styles.NeonCyan).
		Padding(1, 1).
		Foreground(styles.Text)

	return style.Render(content)
}

// applyDualColorBorder 方案2：双色渐变边框
// 使用青色和紫色创建视觉渐变效果
func applyDualColorBorder(content string, width, height int) string {
	// 创建自定义边框，左右使用不同颜色
	customBorder := lipgloss.Border{
		Top:         "═",
		Bottom:      "═",
		Left:        "║",
		Right:       "║",
		TopLeft:     "╔",
		TopRight:    "╗",
		BottomLeft:  "╚",
		BottomRight: "╝",
	}

	// 使用青色作为主边框
	style := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Border(customBorder).
		BorderForeground(styles.NeonCyan).
		Padding(1, 1).
		Foreground(styles.Text)

	// 为了创建双色效果，我们可以在左右边框使用不同的渲染
	// 由于 lipgloss 限制，我们使用嵌套边框来模拟双色效果
	innerStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(styles.NeonPurple).
		Padding(0, 0)

	innerContent := innerStyle.Render(content)
	return style.Render(innerContent)
}

// applyGeometricBorder 方案3：几何科技边框
// 使用特殊 Unicode 字符创建科技感边框
func applyGeometricBorder(content string, width, height int) string {
	// 使用几何字符创建科技感边框
	techBorder := lipgloss.Border{
		Top:         "▀",
		Bottom:      "▄",
		Left:        "▌",
		Right:       "▐",
		TopLeft:     "▛",
		TopRight:    "▜",
		BottomLeft:  "▙",
		BottomRight: "▟",
	}

	style := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Border(techBorder).
		BorderForeground(styles.NeonCyan).
		Padding(1, 1).
		Foreground(styles.Text)

	// 添加内部装饰线
	innerStyle := lipgloss.NewStyle().
		Border(lipgloss.Border{
			Top:    "─",
			Bottom: "─",
			Left:   "│",
			Right:  "│",
		}).
		BorderForeground(styles.NeonGreen).
		Padding(0, 0)

	innerContent := innerStyle.Render(content)
	return style.Render(innerContent)
}

// applyGlowBorder 方案4：发光边框
// 多层边框模拟发光效果
func applyGlowBorder(content string, width, height int) string {
	// 外层：浅色发光边框
	outerBorder := lipgloss.Border{
		Top:         "═",
		Bottom:      "═",
		Left:        "║",
		Right:       "║",
		TopLeft:     "╔",
		TopRight:    "╗",
		BottomLeft:  "╚",
		BottomRight: "╝",
	}

	outerStyle := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Border(outerBorder).
		BorderForeground(styles.NeonCyanLight).
		Padding(0, 0)

	// 中层：主边框
	middleBorder := lipgloss.Border{
		Top:         "─",
		Bottom:      "─",
		Left:        "│",
		Right:       "│",
		TopLeft:     "┌",
		TopRight:    "┐",
		BottomLeft:  "└",
		BottomRight: "┘",
	}

	middleStyle := lipgloss.NewStyle().
		Width(width-2).
		Height(height-2).
		Border(middleBorder).
		BorderForeground(styles.NeonCyan).
		Padding(0, 0)

	// 内层：深色边框
	innerStyle := lipgloss.NewStyle().
		Width(width-4).
		Height(height-4).
		Border(lipgloss.NormalBorder()).
		BorderForeground(styles.NeonCyanDark).
		Padding(1, 1).
		Foreground(styles.Text)

	innerContent := innerStyle.Render(content)
	middleContent := middleStyle.Render(innerContent)
	return outerStyle.Render(middleContent)
}
