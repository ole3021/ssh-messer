package config

import (
	"sort"
	"ssh-messer/internal/tui/components"
	"ssh-messer/internal/tui/messages"
	"ssh-messer/internal/tui/types"

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
	// State      *models.ConfigViewState
	UIState     *types.UIState
	AppState    *types.AppState
	RenderMode  RenderMode
	ConfigNames []string
	Cursor      int
}

// NewConfigComponent 创建新的 Config 组件
func NewConfigComponent(uiState *types.UIState, appState *types.AppState, renderMode RenderMode) *ConfigComponent {
	return &ConfigComponent{
		UIState:     uiState,
		AppState:    appState,
		RenderMode:  renderMode,
		ConfigNames: getConfigNames(appState),
		Cursor:      0,
	}
}

// Update 处理消息并更新组件状态
func (c *ConfigComponent) Update(msg tea.Msg) (components.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return c.handleKeyPress(msg)
	case messages.LoadConfigs:
		c.RefreshFromAppState()
		return c, nil
	default:
		return c, nil
	}
}

// View 渲染组件
func (c *ConfigComponent) View() string {
	switch c.RenderMode {
	case RenderModePopup:
		return renderPopup(c)
	case RenderModeInline:
		return renderInline(c)
	default:
		return renderPopup(c)
	}
}

func getConfigNames(appState *types.AppState) []string {
	configs := appState.GetConfigs()
	var configNames []string
	for name := range configs {
		configNames = append(configNames, name)
	}
	sort.Strings(configNames) // 确保顺序一致
	return configNames
}

// handleKeyPress 处理键盘输入
func (c *ConfigComponent) handleKeyPress(msg tea.KeyMsg) (components.Component, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if c.Cursor > 0 {
			c.Cursor--
		}
	case "down", "j":
		if c.Cursor < len(c.ConfigNames)-1 {
			c.Cursor++
		}
	case "enter":
		if len(c.ConfigNames) > 0 {
			selectedConfigName := c.ConfigNames[c.Cursor]
			// 返回选择消息
			return c, func() tea.Msg {
				return messages.ConfigSelected{ConfigName: selectedConfigName}
			}
		}
	}
	return c, nil
}

func (c *ConfigComponent) RefreshFromAppState() {
	c.ConfigNames = getConfigNames(c.AppState)
	c.Cursor = 0
}
