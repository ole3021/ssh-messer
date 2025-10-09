package layout

import tea "github.com/charmbracelet/bubbletea/v2"

// Sizeable 可调整大小的接口
type Sizeable interface {
	SetSize(width, height int) tea.Cmd
	GetSize() (width, height int)
}

// Focusable 可聚焦的接口
type Focusable interface {
	Focus() tea.Cmd
	Blur() tea.Cmd
	IsFocused() bool
}
