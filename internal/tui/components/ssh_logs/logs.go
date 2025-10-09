package ssh_logs

import (
	"fmt"
	"strings"

	"ssh-messer/internal/pubsub"
	"ssh-messer/internal/ssh_proxy"
	"ssh-messer/internal/tui/components/core/layout"
	"ssh-messer/internal/tui/types"
	"ssh-messer/internal/tui/util"

	"github.com/charmbracelet/bubbles/v2/viewport"
	tea "github.com/charmbracelet/bubbletea/v2"
)

const (
	ellipsisLength = 3 // 省略号长度
)

// LogsCmp SSH 日志组件接口
type LogsCmp interface {
	util.Model
	layout.Sizeable
	AddLog(log string) tea.Cmd
	AddLogWithID(requestID string, log string) tea.Cmd
	UpdateLog(requestID string, log string) tea.Cmd
}

// logEntry 日志条目
type logEntry struct {
	requestID string
	text      string
}

// logsCmp SSH 日志组件实现
type logsCmp struct {
	width, height int
	appState      *types.AppState
	logs          []logEntry
	maxLogs       int
	viewport      viewport.Model
	allLines      []string // 存储所有日志行，用于检查是否在底部
	needsUpdate   bool     // 标记是否需要更新 allLines
}

// New 创建新的日志组件
func New(appState *types.AppState) LogsCmp {
	return &logsCmp{
		appState: appState,
		logs:     make([]logEntry, 0),
		maxLogs:  100, // 最多保留 100 条日志
	}
}

func (l *logsCmp) Init() tea.Cmd {
	// 初始化 viewport，如果 width 和 height 为 0，使用默认值
	width := l.width
	height := l.height
	if width <= 0 {
		width = 80
	}
	if height <= 0 {
		height = 24
	}
	l.viewport = viewport.New(viewport.WithWidth(width), viewport.WithHeight(height))
	l.viewport.MouseWheelEnabled = true
	l.viewport.MouseWheelDelta = 3
	l.viewport.SetContent("No logs yet...\nWaiting for SSH proxy activity...")
	return nil
}

func (l *logsCmp) Update(msg tea.Msg) (util.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// 处理键盘事件：方向键、Page Up/Down、Home/End
		switch msg.String() {
		case "up", "k":
			l.viewport.ScrollUp(1)
		case "down", "j":
			l.viewport.ScrollDown(1)
		case "pgup":
			l.viewport.PageUp()
		case "pgdown":
			l.viewport.PageDown()
		case "home":
			l.viewport.GotoTop()
		case "end":
			l.viewport.GotoBottom()
		default:
			// 将其他键盘事件传递给 viewport（viewport 有自己的 keymap）
			updatedViewport, cmd := l.viewport.Update(msg)
			l.viewport = updatedViewport
			return l, cmd
		}
		return l, nil

	case tea.MouseMsg:
		// 将鼠标事件传递给 viewport（viewport 会自动处理滚轮和拖动）
		updatedViewport, cmd := l.viewport.Update(msg)
		l.viewport = updatedViewport
		return l, cmd

	case pubsub.Event[ssh_proxy.ServiceProxyLogEvent]:
		// 处理 Service Proxy 日志事件
		event := msg.Payload
		logText := l.formatServiceProxyLog(event)

		// 如果是更新事件，更新已有日志；否则添加新日志
		if event.IsUpdate {
			cmd = l.UpdateLog(event.RequestID, logText)
		} else {
			cmd = l.AddLogWithID(event.RequestID, logText)
		}
		return l, cmd
	}

	// 将其他事件传递给 viewport
	updatedViewport, cmd := l.viewport.Update(msg)
	l.viewport = updatedViewport
	return l, cmd
}

func (l *logsCmp) View() string {
	if len(l.logs) == 0 {
		l.viewport.SetContent("No logs yet...\nWaiting for SSH proxy activity...")
		return l.viewport.View()
	}

	// 只在需要时更新 allLines
	if l.needsUpdate {
		// 检查是否在底部（在更新内容之前）
		wasAtBottom := l.isAtBottom()

		// 使用 updateAllLines 方法处理所有日志条目
		l.updateAllLines()
		l.needsUpdate = false

		// 合并所有行为单个字符串
		content := strings.Join(l.allLines, "\n")

		// 设置 viewport 内容
		l.viewport.SetContent(content)

		// 如果之前在底部，自动滚动到底部
		if wasAtBottom {
			l.viewport.GotoBottom()
		}
	}

	return l.viewport.View()
}

// AddLog 添加日志（直接更新状态）
func (l *logsCmp) AddLog(log string) tea.Cmd {
	return l.AddLogWithID("", log)
}

// AddLogWithID 添加带 RequestID 的日志
func (l *logsCmp) AddLogWithID(requestID string, log string) tea.Cmd {
	l.logs = append(l.logs, logEntry{
		requestID: requestID,
		text:      log,
	})
	if len(l.logs) > l.maxLogs {
		l.logs = l.logs[1:] // 移除最早的日志
	}

	// 标记需要更新
	l.needsUpdate = true

	return nil
}

// UpdateLog 根据 RequestID 更新日志
func (l *logsCmp) UpdateLog(requestID string, log string) tea.Cmd {
	if requestID == "" {
		// 如果没有 RequestID，直接添加新日志
		return l.AddLog(log)
	}

	// 从后往前查找匹配的日志条目（最新的匹配项）
	found := false
	for i := len(l.logs) - 1; i >= 0; i-- {
		if l.logs[i].requestID == requestID {
			// 找到匹配的日志，更新它
			l.logs[i].text = log
			found = true
			break
		}
	}

	if !found {
		// 如果没有找到匹配的日志，添加新日志
		return l.AddLogWithID(requestID, log)
	}

	// 标记需要更新
	l.needsUpdate = true

	return nil
}

// updateAllLines 更新 allLines 字段，处理所有日志条目
func (l *logsCmp) updateAllLines() {
	l.allLines = make([]string, 0)
	for i := 0; i < len(l.logs); i++ {
		log := l.logs[i].text
		// 按换行符分割日志条目
		lines := strings.Split(log, "\n")
		for _, line := range lines {
			// 截断过长的行以适应窗口宽度
			line = util.TruncateString(line, l.width)
			l.allLines = append(l.allLines, line)
		}
	}
}

func (l *logsCmp) SetSize(width, height int) tea.Cmd {
	l.width = width
	l.height = height
	l.viewport.SetWidth(width)
	l.viewport.SetHeight(height)
	// 尺寸变化时需要重新计算
	l.needsUpdate = true
	return nil
}

// isAtBottom 检查当前滚动位置是否在最底端
func (l *logsCmp) isAtBottom() bool {
	if len(l.allLines) == 0 {
		return true
	}

	// 使用 viewport 的 AtBottom() 方法检查是否在底部
	// 但需要先设置内容才能正确判断
	if l.viewport.GetContent() == "" {
		return true
	}

	return l.viewport.AtBottom()
}

func (l *logsCmp) GetSize() (int, int) {
	return l.width, l.height
}

// formatServiceProxyLog 格式化 service proxy 日志，限制长度不超过日志窗口宽度
// 如果有错误消息，错误消息显示在下一行
func (l *logsCmp) formatServiceProxyLog(event ssh_proxy.ServiceProxyLogEvent) string {
	logsWidth := l.width
	if logsWidth <= 0 {
		logsWidth = 80 // 默认宽度
	}

	// 基础格式：[ServiceAlias] Method URL -> StatusCode (ResponseSize bytes)
	// 估算基础文本长度（不包括 URL）
	baseText := fmt.Sprintf("[%s] %s ", event.ServiceAlias, event.Method)
	baseTextLen := len([]rune(baseText))

	// 根据状态码格式化后缀
	var suffixText string
	if event.StatusCode == 0 {
		// StatusCode 为 0 表示请求中
		suffixText = " -> pending..."
	} else {
		// 格式化用时信息，单位是秒（s），保留3位小数
		durationSec := event.Duration.Seconds()
		suffixText = fmt.Sprintf(" <> %d (%d bytes) %.3fs", event.StatusCode, event.ResponseSize, durationSec)
	}
	suffixTextLen := len([]rune(suffixText))

	// 计算 URL 可用的最大长度
	maxURLLen := logsWidth - baseTextLen - suffixTextLen - ellipsisLength // 预留省略号长度

	// 如果 URL 太长，截断它
	url := util.TruncateString(event.URL, maxURLLen)

	logText := baseText + url + suffixText

	// 最后检查：如果整个文本仍然超过宽度，再次截断
	logText = util.TruncateString(logText, logsWidth)

	// 如果有错误消息，将其添加到下一行
	if event.ErrorMessage != "" {
		// 截断错误消息以适应窗口宽度
		errorMsg := util.TruncateString(event.ErrorMessage, logsWidth)
		logText += "\n" + errorMsg
	}

	return logText
}
