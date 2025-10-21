package app

import (
	"ssh-messer/internal/tui/messages"
	"ssh-messer/internal/tui/models"
	"ssh-messer/internal/tui/services"
	"ssh-messer/internal/tui/views"
	sshproxy "ssh-messer/internal/tui/views/ssh-proxy"
	"ssh-messer/internal/tui/views/welcome"

	tea "github.com/charmbracelet/bubbletea"
)

// AppModel 主应用模型
type AppModel struct {
	// 状态管理
	AppState *models.AppState
	UIState  *models.UIState

	// 视图管理
	CurrentView views.View
}

// NewAppModel 创建新的应用模型
func NewAppModel() *AppModel {
	appState := models.NewAppState()
	uiState := models.NewUIState()

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

	// 特殊处理配置加载消息
	if loadMsg, ok := msg.(messages.LoadConfigs); ok {
		if loadMsg.Err == nil {
			// 更新应用状态中的配置
			m.AppState.SetConfigs(loadMsg.Configs)
		} else {
			// 如果有错误，发送错误消息给当前视图
			return m, func() tea.Msg {
				return messages.AppError{Error: &loadMsg.Err, IsFatal: false}
			}
		}
		// 继续传递给视图处理
	}

	// 委托给当前视图处理
	if m.CurrentView != nil {
		newView, cmd := m.CurrentView.Update(msg)
		m.CurrentView = newView

		// 检查视图切换消息
		if viewChangeMsg, ok := msg.(messages.ViewChangeMsg); ok {
			return m.handleViewChange(viewChangeMsg)
		}

		return m, cmd
	}

	return m, nil
}

// View 渲染当前视图
func (m *AppModel) View() string {
	if m.CurrentView != nil {
		return m.CurrentView.View()
	}
	return "No view available"
}

// handleGlobalMsg 处理全局消息
func (m *AppModel) handleGlobalMsg(msg tea.Msg) (bool, tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.handleWindowSize(msg)
	case tea.KeyMsg:
		return m.handleGlobalKeyPress(msg)
	case messages.AppError:
		return m.handleAppError(msg)
	default:
		return false, m, nil
	}
}

// handleAppError 处理应用错误
func (m *AppModel) handleAppError(appErr messages.AppError) (bool, tea.Model, tea.Cmd) {
	m.AppState.Error = appErr
	return true, m, nil
}

// handleWindowSize 处理窗口大小变化
func (m *AppModel) handleWindowSize(msg tea.WindowSizeMsg) (bool, tea.Model, tea.Cmd) {
	m.UIState.Width = msg.Width
	m.UIState.Height = msg.Height
	return true, m, nil
}

// handleGlobalKeyPress 处理全局键盘输入
func (m *AppModel) handleGlobalKeyPress(msg tea.KeyMsg) (bool, tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return true, m, tea.Quit
	default:
		return false, m, nil
	}
}

// handleViewChange 处理视图切换
func (m *AppModel) handleViewChange(msg messages.ViewChangeMsg) (tea.Model, tea.Cmd) {
	// 清理当前视图
	if m.CurrentView != nil {
		cleanupCmd := m.CurrentView.Cleanup()
		if cleanupCmd != nil {
			// 执行清理命令
			cleanupCmd()
		}
	}

	// 切换到新视图
	switch models.ViewEnum(msg.TargetView) {
	case models.SSHProxyView:
		m.CurrentView = sshproxy.NewSSHProxyView(m.UIState, m.AppState)
		return m, m.CurrentView.Init()
	case models.WelcomeView:
		m.CurrentView = welcome.NewWelcomeView(m.UIState, m.AppState)
		return m, m.CurrentView.Init()
	default:
		return m, nil
	}
}
