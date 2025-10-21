package config

import (
	"ssh-messer/internal/tui/components"
	"ssh-messer/internal/tui/messages"
	"ssh-messer/internal/tui/models"

	tea "github.com/charmbracelet/bubbletea"
)

// RenderMode 渲染模式
type RenderMode int

const (
	RenderModePopup RenderMode = iota
	RenderModeFullscreen
	RenderModeInline
)

// ConfigComponent Config 组件
type ConfigComponent struct {
	State      *models.ConfigViewState
	UI         *models.UIState
	RenderMode RenderMode
}

// NewConfigComponent 创建新的 Config 组件
func NewConfigComponent(uiState *models.UIState, renderMode RenderMode) *ConfigComponent {
	return &ConfigComponent{
		State:      models.NewConfigViewState(),
		UI:         uiState,
		RenderMode: renderMode,
	}
}

// Update 处理消息并更新组件状态
func (c *ConfigComponent) Update(msg tea.Msg) (components.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return c.handleKeyPress(msg)
	case messages.LoadConfigs:
		return c.handleLoadConfigs(msg)
	default:
		return c, nil
	}
}

// View 渲染组件
func (c *ConfigComponent) View() string {
	switch c.RenderMode {
	case RenderModePopup:
		return renderPopup(c.State, c.UI)
	case RenderModeFullscreen:
		return renderFullscreen(c.State, c.UI)
	case RenderModeInline:
		return renderInline(c.State, c.UI)
	default:
		return renderPopup(c.State, c.UI)
	}
}

// handleKeyPress 处理键盘输入
func (c *ConfigComponent) handleKeyPress(msg tea.KeyMsg) (components.Component, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if c.State.Cursor > 0 {
			c.State.Cursor--
		}
	case "down", "j":
		if c.State.Cursor < len(c.State.ConfigNames)-1 {
			c.State.Cursor++
		}
	case "enter":
		if len(c.State.ConfigNames) > 0 {
			selectedConfig := c.State.ConfigNames[c.State.Cursor]
			c.State.SelectedName = selectedConfig
			// 返回选择消息
			return c, func() tea.Msg {
				return messages.ConfigSelected{ConfigName: selectedConfig}
			}
		}
	case "esc", "ctrl+c", "q":
		// 返回取消消息
		return c, func() tea.Msg {
			return messages.ConfigCancelled{}
		}
	}
	return c, nil
}

// handleLoadConfigs 处理配置加载消息
func (c *ConfigComponent) handleLoadConfigs(msg messages.LoadConfigs) (components.Component, tea.Cmd) {
	// 更新配置列表
	var configNames []string
	for name := range msg.Configs {
		configNames = append(configNames, name)
	}
	c.State.ConfigNames = configNames
	c.State.Cursor = 0

	return c, nil
}
