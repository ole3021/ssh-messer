package components

import tea "github.com/charmbracelet/bubbletea"

// Component 组件接口
type Component interface {
	// Update 处理消息并更新组件状态
	Update(msg tea.Msg) (Component, tea.Cmd)
	// View 渲染组件
	View() string
}
