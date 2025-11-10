package ssh_sidebar

import (
	"fmt"
	"path/filepath"
	"strings"

	"ssh-messer/internal/ssh_proxy"
	"ssh-messer/internal/tui/components/core/layout"
	"ssh-messer/internal/tui/styles"
	"ssh-messer/internal/tui/types"
	"ssh-messer/internal/tui/util"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
)

const (
	SidebarTitleHeight = 3
	sidebarBroderChar  = "░"
	messerAscii        = `
░░█▄█░░█▀▀░░█▀▀░░█▀▀░░█▀▀░░█▀▄░░
░░█░█░░█▀▀░░▀▀█░░▀▀█░░█▀▀░░█▀▄░░
░░▀░▀░░▀▀▀░░▀▀▀░░▀▀▀░░▀▀▀░░▀░▀░░
`
)

type SidebarCmp interface {
	util.Model
	layout.Sizeable
}

type sidebarCmp struct {
	width, height int
	appState      *types.AppState
}

func New(appState *types.AppState) SidebarCmp {
	return &sidebarCmp{
		appState: appState,
	}
}

func (s *sidebarCmp) Init() tea.Cmd {
	return nil
}

func (s *sidebarCmp) Update(msg tea.Msg) (util.Model, tea.Cmd) {
	return s, nil
}

func (s *sidebarCmp) View() string {

	if s.appState == nil || s.appState.CurrentConfigName == "" {
		return "No configuration\nselected"
	}

	// title
	titlePart := lipgloss.NewStyle().
		Foreground(styles.Meta).
		Width(s.width).
		Render(lipgloss.JoinVertical(
			lipgloss.Top,
			strings.Repeat(sidebarBroderChar, s.width),
			lipgloss.NewStyle().
				Width(s.width).
				Render(padAsciiToWidth(strings.TrimSpace(messerAscii), s.width, sidebarBroderChar)),
			strings.Repeat(sidebarBroderChar, s.width),
		))

	return lipgloss.NewStyle().
		Width(s.width).
		Height(s.height).
		BorderForeground(styles.NeonCyan).
		Foreground(styles.Text).
		Render(lipgloss.JoinVertical(lipgloss.Top, titlePart, strings.Join(s.generateSSHHopsPart(), "\n")))
}

func (s *sidebarCmp) SetSize(width, height int) tea.Cmd {
	s.width = width
	s.height = height
	return nil
}

func (s *sidebarCmp) GetSize() (int, int) {
	return s.width, s.height
}

// padAsciiToWidth 将 ASCII 艺术的每一行填充到指定宽度
func padAsciiToWidth(ascii string, width int, fillChar string) string {
	lines := strings.Split(ascii, "\n")
	var paddedLines []string

	for _, line := range lines {
		lineLen := len([]rune(line))
		if lineLen < width {
			padding := width - lineLen
			leftPadding := padding / 2
			rightPadding := padding - leftPadding
			paddedLine := strings.Repeat(fillChar, leftPadding) + line + strings.Repeat(fillChar, rightPadding)
			paddedLines = append(paddedLines, paddedLine)
		} else {
			paddedLines = append(paddedLines, line)
		}
	}

	return strings.Join(paddedLines, "\n")
}

func (s *sidebarCmp) generateSSHHopsPart() []string {
	var hopLines []string
	hopLines = append(hopLines, "\n")

	config := s.appState.GetCurrentConfig()
	if config == nil {
		hopLines = append(hopLines, "No configuration available")
		return hopLines
	}

	// Hopes title line
	var configName string
	if config.Name != nil {
		configName = strings.TrimSuffix(*config.Name, filepath.Ext(*config.Name))
	} else {
		configName = s.appState.CurrentConfigName
	}
	prefix := fmt.Sprintf(" %s ", configName)
	prefixLen := len([]rune(prefix))
	separatorCount := s.width - prefixLen - 1
	if separatorCount < 0 {
		separatorCount = 0
	}
	hopLines = append(hopLines, prefix+strings.Repeat("━", separatorCount))

	// Hopes Status
	proxy := s.appState.GetSSHProxy(s.appState.CurrentConfigName)

	if proxy == nil {
		hopLines = append(hopLines, "\n\n⚪ Uninitialized")
		return hopLines
	}

	hopsConfigs := proxy.GetHopsConfigs()
	if len(hopsConfigs) > 0 {
		hopLines = append(hopLines, "")
		for i, hopConfig := range hopsConfigs {
			displayName := ssh_proxy.GetHopDisplayName(hopConfig)
			port := 22
			if hopConfig.Port != nil {
				port = *hopConfig.Port
			}
			if hopConfig.Host != nil {
				if port != 22 {
					hopLines = append(hopLines, fmt.Sprintf("%d. %s (%s:%d)", i+1, displayName, *hopConfig.Host, port))
				} else {
					hopLines = append(hopLines, fmt.Sprintf("%d. %s (%s)", i+1, displayName, *hopConfig.Host))
				}
			} else {
				hopLines = append(hopLines, fmt.Sprintf("%d. %s", i+1, displayName))
			}
		}
	}

	status := proxy.Status
	if status.IsConnected {
		if status.IsChecking {
			hopLines = append(hopLines, "\n\n🟢 Connected 👀")
		} else {
			hopLines = append(hopLines, "\n\n🟢 Connected")
		}
	} else if status.IsConnecting {
		hopLines = append(hopLines, "\n\n🟡 Connecting")
	} else {
		hopLines = append(hopLines, "\n\n⚪ Disconnected")
		if status.LastError != nil {
			hopLines = append(hopLines, fmt.Sprintf("\n\n🔴 %s", status.LastError.Error()))
		}
	}

	return hopLines
}
