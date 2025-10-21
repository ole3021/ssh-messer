package styles

import (
	"github.com/charmbracelet/lipgloss"
)

// 主题颜色（在所有视图间共享）
var (
	// 主要主题颜色
	Primary = lipgloss.Color("#00D4AA")
	Text    = lipgloss.Color("#E5E7EB")
	Meta    = lipgloss.Color("#6B7280")
	Bg      = lipgloss.Color("#0F172A")
	Border  = lipgloss.Color("#374151")
)

// 共享样式
var (
	// 标题样式：居中对齐和粗体
	TitleStyle = lipgloss.NewStyle().
			Foreground(Primary).
			Bold(true).
			Align(lipgloss.Center)

	MetaStyle = lipgloss.NewStyle().
			Foreground(Meta).
			Align(lipgloss.Center)

	// 项目样式，带条件渲染
	ItemStyle     = lipgloss.NewStyle().Foreground(Text)
	SelectedStyle = lipgloss.NewStyle().
			Foreground(Primary).
			Bold(true)
)

// ASCII 艺术样式，居中对齐
var AsciiStyle = lipgloss.NewStyle().
	Foreground(Primary).
	Align(lipgloss.Center)

// ASCII 艺术
const AsciiArt = `
    ███╗   ███████╗███████╗██╗  ██╗    ███╗   ███╗███████╗███████╗███████╗███████╗██████╗ 
   ████║   ██╔════╝██╔════╝██║  ██║    ████╗ ████║██╔════╝██╔════╝██╔════╝██╔════╝██╔══██╗
  ██╔██║   ███████╗███████╗███████║    ██╔████╔██║█████╗  ███████╗█████╗  █████╗  ██████╔╝
 ██╔╝██║   ╚════██║╚════██║██╔══██║    ██║╚██╔╝██║██╔══╝  ╚════██║██╔══╝  ██╔══╝  ██╔══██╗
██╔╝ ██║   ███████║███████║██║  ██║    ██║ ╚═╝ ██║███████╗███████║███████╗███████╗██║  ██║
╚═╝  ╚═╝   ╚══════╝╚══════╝╚═╝  ╚═╝    ╚═╝     ╚═╝╚══════╝╚══════╝╚══════╝╚═╝  ╚═╝╚═╝  ╚═╝
`

// 状态栏相关样式
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

// 弹窗样式
var (
	// 配置弹窗特定样式
	ConfigPopupStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(Border)

	MainPopupStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Border)
)
