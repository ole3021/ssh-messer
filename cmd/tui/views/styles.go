package views

import (
	"github.com/charmbracelet/lipgloss"
)

// Theme colors (shared across all views)
var (
	// Primary theme colors
	Primary = lipgloss.Color("#00D4AA")
	Text    = lipgloss.Color("#E5E7EB")
	Meta    = lipgloss.Color("#6B7280")
	Bg      = lipgloss.Color("#0F172A")
	Border  = lipgloss.Color("#374151")
)

// Shared styles for all views
var (
	// Title style with center alignment and bold
	TitleStyle = lipgloss.NewStyle().
			Foreground(Primary).
			Bold(true).
			Align(lipgloss.Center)

	MetaStyle = lipgloss.NewStyle().
			Foreground(Meta).
			Align(lipgloss.Center)

	// Item styles with conditional rendering
	ItemStyle     = lipgloss.NewStyle().Foreground(Text)
	SelectedStyle = lipgloss.NewStyle().
			Foreground(Primary).
			Bold(true)
)

// Status Bar 相关样式
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
