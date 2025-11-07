package ssh_logs

import (
	"strings"

	"ssh-messer/internal/tui/components/core/layout"
	"ssh-messer/internal/tui/types"
	"ssh-messer/internal/tui/util"

	tea "github.com/charmbracelet/bubbletea/v2"
)

// LogsCmp SSH 日志组件接口
type LogsCmp interface {
	util.Model
	layout.Sizeable
	AddLog(log string) tea.Cmd
}

// logsCmp SSH 日志组件实现
type logsCmp struct {
	width, height int
	appState      *types.AppState
	logs          []string
	maxLogs       int
}

// New 创建新的日志组件
func New(appState *types.AppState) LogsCmp {
	return &logsCmp{
		appState: appState,
		logs:     make([]string, 0),
		maxLogs:  100, // 最多保留 100 条日志
	}
}

func (l *logsCmp) Init() tea.Cmd {
	return nil
}

func (l *logsCmp) Update(msg tea.Msg) (util.Model, tea.Cmd) {
	return l, nil
}

func (l *logsCmp) View() string {
	if len(l.logs) == 0 {
		return "No logs yet...\nWaiting for SSH proxy activity..."
	}

	// 显示最近的日志（根据高度）
	start := 0
	if len(l.logs) > l.height {
		start = len(l.logs) - l.height
	}

	var displayLogs []string
	for i := start; i < len(l.logs); i++ {
		log := l.logs[i]
		// 截断过长的日志
		if len(log) > l.width {
			log = log[:l.width-3] + "..."
		}
		displayLogs = append(displayLogs, log)
	}

	return strings.Join(displayLogs, "\n")
}

// AddLog 添加日志（直接更新状态）
func (l *logsCmp) AddLog(log string) tea.Cmd {
	l.logs = append(l.logs, log)
	if len(l.logs) > l.maxLogs {
		l.logs = l.logs[1:] // 移除最早的日志
	}
	return nil
}

func (l *logsCmp) SetSize(width, height int) tea.Cmd {
	l.width = width
	l.height = height
	return nil
}

func (l *logsCmp) GetSize() (int, int) {
	return l.width, l.height
}
