package styles

import (
	"github.com/charmbracelet/lipgloss/v2"
)

var (
	Primary = lipgloss.Color("#00D4AA")
	Text    = lipgloss.Color("#E5E7EB")
	Meta    = lipgloss.Color("#6B7280")
	Bg      = lipgloss.Color("#0F172A")
	Border  = lipgloss.Color("#374151")
)

// Cyberpunk 配色常量
var (
	// 霓虹青色（主色）
	NeonCyan = lipgloss.Color("#00D4AA")
	// 霓虹青色（浅色，用于发光效果）
	NeonCyanLight = lipgloss.Color("#00FFCC")
	// 霓虹青色（深色，用于内层边框）
	NeonCyanDark = lipgloss.Color("#0A8A6A")
	// 霓虹紫色（用于双色渐变）
	NeonPurple = lipgloss.Color("#A855F7")
	// 霓虹绿色（用于几何边框）
	NeonGreen = lipgloss.Color("#00FF88")
)

var (
	TitleStyle = lipgloss.NewStyle().
			Foreground(Primary).
			Bold(true).
			Align(lipgloss.Center)

	MetaStyle = lipgloss.NewStyle().
			Foreground(Meta).
			Align(lipgloss.Center)

	ItemStyle     = lipgloss.NewStyle().Foreground(Text)
	SelectedStyle = lipgloss.NewStyle().
			Foreground(Primary).
			Bold(true)
)

var AsciiStyle = lipgloss.NewStyle().
	Foreground(Primary).
	Align(lipgloss.Center)

var (
	StatusBarBg = lipgloss.Color("#1F2937") // 状态栏背景色

	StatusBarStyle = lipgloss.NewStyle().
			Background(StatusBarBg).
			Foreground(Text).
			Width(100) // 会在运行时动态设置

	// SSH 状态区域样式（根据状态变化）
	SSHStatusConnectedStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#10B981")). // 绿色
				Foreground(lipgloss.Color("#FFFFFF")).
				Padding(0, 1).
				Bold(true)

	SSHStatusDisconnectedStyle = lipgloss.NewStyle().
					Background(lipgloss.Color("#374151")). // 深灰色
					Foreground(lipgloss.Color("#FFFFFF")).
					Padding(0, 1).
					Bold(true)

	SSHStatusConnectingStyle = lipgloss.NewStyle().
					Background(lipgloss.Color("#F59E0B")). // 黄色
					Foreground(lipgloss.Color("#FFFFFF")).
					Padding(0, 1).
					Bold(true)

	SSHStatusErrorStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#EF4444")). // 红色
				Foreground(lipgloss.Color("#FFFFFF")).
				Padding(0, 1).
				Bold(true)

	// 配置名称样式
	ConfigNameStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#374151")). // 深灰色
			Foreground(Primary).
			Bold(true)
)

var (
	// 配置弹窗特定样式
	ConfigPopupStyle = lipgloss.NewStyle().
		// Border(lipgloss.RoundedBorder()).
		BorderForeground(Border)

	MainPopupStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Border)
)
