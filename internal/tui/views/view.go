package views

import (
	"ssh-messer/internal/tui/types"

	tea "github.com/charmbracelet/bubbletea"
)

// View 视图接口
type View interface {
	// Init 初始化视图，返回初始命令
	Init() tea.Cmd
	// Update 处理消息并更新视图状态
	Update(msg tea.Msg) (View, tea.Cmd)
	// View 渲染视图
	View() string
	// GetType 获取视图类型
	GetType() types.ViewEnum
	// Cleanup 清理视图资源，返回清理命令
	Cleanup() tea.Cmd
}

// BaseView 基础视图结构
type BaseView struct {
	Type types.ViewEnum
}

// GetType 获取视图类型
func (v *BaseView) GetType() types.ViewEnum {
	return v.Type
}

// Cleanup 基础视图清理（默认实现）
func (v *BaseView) Cleanup() tea.Cmd {
	return nil
}
