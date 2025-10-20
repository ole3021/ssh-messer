package main

import "ssh-messer/cmd/tui/views"

// View renders the current view based on the model state
func (m AppModel) View() string {
	switch m.CurrentView {
	case WelcomeView:
		return views.RenderWelcomeView(m)
	case MesserView:
		return views.RenderMesserView(m)
	default:
		return views.RenderWelcomeView(m)
	}
}

// GetWelcomeProgress implements the WelcomeRenderer interface
func (m AppModel) GetWelcomeProgress() int {
	return m.WelcomeViewModel.WelcomeAnimationProgress
}

// GetConfigChoices implements the ConfigSelectionRenderer interface
func (m AppModel) GetConfigChoices() []string {
	return m.ConfigViewModel.ConfigNames
}

// GetConfigCursor implements the ConfigSelectionRenderer interface
func (m AppModel) GetConfigCursor() int {
	return m.ConfigViewModel.Cursor
}

// GetWidth implements the WelcomeRenderer interface
func (m AppModel) GetWidth() int {
	return m.Width
}

// GetHeight implements the WelcomeRenderer interface
func (m AppModel) GetHeight() int {
	return m.Height
}

// GetConnectionProcess implements the MesserRenderer interface
func (m AppModel) GetConnectionProcess() int {
	return m.SSHInfos[m.CurrentConfigName].SSHConnectionProcess
}

// GetCurrentConfigName implements the MesserRenderer interface
func (m AppModel) GetCurrentConfigName() string {
	return m.CurrentConfigName
}

// GetSSHConnectionState implements the MesserRenderer interface
func (m AppModel) GetSSHConnectionState() string {
	switch m.SSHInfos[m.CurrentConfigName].SSHConnectionState {
	case Disconnected:
		return "Disconnected"
	case Connecting:
		return "Connecting"
	case Connected:
		return "Connected"
	case Error:
		return "Error"
	default:
		return "Unknown"
	}
}

// GetHTTPProxyLogs implements the MesserRenderer interface
func (m AppModel) GetHTTPProxyLogs() []string {
	return m.SSHInfos[m.CurrentConfigName].HTTPProxyLogs
}

// GetDockerProxyLogs implements the MesserRenderer interface
func (m AppModel) GetDockerProxyLogs() []string {
	return m.SSHInfos[m.CurrentConfigName].DockerProxyLogs
}

// GetCurrentInfo implements the MesserRenderer interface
func (m AppModel) GetCurrentInfo() string {
	return m.CurrentInfo
}
