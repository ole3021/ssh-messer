package app

import (
	"ssh-messer/internal/tui/messages"
	"ssh-messer/internal/tui/services"
	"ssh-messer/internal/tui/types"
	"ssh-messer/internal/tui/views"
	"ssh-messer/internal/tui/views/welcome"

	tea "github.com/charmbracelet/bubbletea"
)

type AppModel struct {
	// 状态管理
	AppState *types.AppState
	UIState  *types.UIState

	// 视图管理
	CurrentView views.View
}

// NewAppModel 创建新的应用模型
func NewAppModel() *AppModel {
	appState := types.NewAppState()
	uiState := types.NewUIState()

	return &AppModel{
		AppState:    appState,
		UIState:     uiState,
		CurrentView: welcome.NewWelcomeView(uiState, appState),
	}
}

// Init 初始化应用模型
func (m *AppModel) Init() tea.Cmd {
	return tea.Batch(
		services.LoadConfigsFromHomeDir(),
		m.CurrentView.Init(),
	)
}

// Update 处理消息并更新模型
func (m *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// 先处理全局消息
	if handled, model, cmd := m.handleGlobalMsg(msg); handled {
		return model, cmd
	}

	// 根据状态变化决定视图切换
	if viewChangeMsg, ok := msg.(messages.ViewChangeMsg); ok {
		return m.handleViewChangeMsg(viewChangeMsg)
	}

	// 处理全局配置加载， 在视图渲染前处理
	if loadMsg, ok := msg.(messages.LoadConfigs); ok {
		m.handleLoadConfigsMsg(loadMsg)
	}

	if configSelectedMsg, ok := msg.(messages.ConfigSelected); ok {
		return m.handleConfigSelectedMsg(configSelectedMsg)
	}

	// 委托给当前视图处理视图更新
	var cmd tea.Cmd
	if m.CurrentView != nil {
		newView, newCmd := m.CurrentView.Update(msg)
		m.CurrentView = newView
		cmd = newCmd
	}

	return m, cmd
}

// View 渲染当前视图
func (m *AppModel) View() string {
	if m.CurrentView != nil {
		return m.CurrentView.View()
	}
	return "No view available"
}
