package ssh_sidebar

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"ssh-messer/internal/config"
	"ssh-messer/internal/messer"
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

	if s.appState == nil || s.appState.CurrentConfigFileName == "" {
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
	hopLines = append(hopLines, "\n"+s.appState.CurrentConfigFileName+"\n"+s.appState.GetCurrentConfig().Name+"\n")

	messerConfig := s.appState.GetCurrentConfig()
	if messerConfig == nil {
		hopLines = append(hopLines, "No configuration available")
		return hopLines
	}

	// Hopes title line
	var configName string
	if messerConfig.Name != "" {
		configName = strings.TrimSuffix(messerConfig.Name, filepath.Ext(messerConfig.Name))
	} else {
		configName = s.appState.CurrentConfigFileName
	}
	prefix := fmt.Sprintf(" %s ", configName)
	prefixLen := len([]rune(prefix))
	separatorCount := s.width - prefixLen - 1
	if separatorCount < 0 {
		separatorCount = 0
	}
	hopLines = append(hopLines, prefix+strings.Repeat("━", separatorCount))

	// Hopes Status
	messerHop := s.appState.GetMesserHops(s.appState.CurrentConfigFileName)

	if messerHop == nil {
		hopLines = append(hopLines, "\n\n⚪ Uninitialized")
		return hopLines
	}

	hopConfigs := messerHop.Config.SSHHops
	sshClients := messerHop.SSHClients

	orderedHopConfigs := make([]config.SSHConfig, 0, len(hopConfigs))
	orderedHopConfigs = append(orderedHopConfigs, hopConfigs...)
	sort.Slice(orderedHopConfigs, func(i, j int) bool {
		return orderedHopConfigs[i].Order < orderedHopConfigs[j].Order
	})

	if len(orderedHopConfigs) > 0 {
		hopLines = append(hopLines, "")
		for i, hopConfig := range orderedHopConfigs {
			hopName := hopConfig.Name

			hopStatusEmoji := ""
			hopStatus := messer.StatusDisconnected
			sshClient, ok := sshClients[hopConfig.Order]
			if ok {
				hopStatus = sshClient.Status
			}

			switch hopStatus {
			case messer.StatusConnected:
				hopStatusEmoji = "🟢"
			case messer.StatusConnecting:
				hopStatusEmoji = "🟡"
			case messer.StatusChecking:
				hopStatusEmoji = "👀"
			case messer.StatusDisconnected:
				hopStatusEmoji = "⚫️"
			}

			hopLines = append(hopLines, fmt.Sprintf("%d. %s %s", i+1, hopStatusEmoji, hopName))
		}
	}

	serviceLinks := s.generateServiceLinks(messerConfig)
	if len(serviceLinks) > 0 {
		hopLines = append(hopLines, "\n")
		hopLines = append(hopLines, strings.Repeat("─", s.width))
		hopLines = append(hopLines, " 服务链接")
		hopLines = append(hopLines, "")
		hopLines = append(hopLines, serviceLinks...)
	}

	return hopLines
}

func (s *sidebarCmp) generateServiceLinks(config *config.MesserConfig) []string {
	var links []string

	if config == nil {
		return links
	}

	for _, service := range config.ReverseServices {
		if len(service.Pages) == 0 {
			continue
		}

		// TODO: Add link to text in terminal
		for _, page := range service.Pages {
			pageName := page.Name
			// pagePath := page.Path

			linkDisplay := lipgloss.NewStyle().
				Foreground(styles.NeonCyan).
				Render(fmt.Sprintf("🔗 %s", pageName))
			// linkURL := "http://" + service.Subdomain + "." + "localhost" + ":" + config.LocalHttpPort + pagePath
			// links = append(links, "")
			links = append(links, linkDisplay)
			// links = append(links, lipgloss.NewStyle().
			// 	Foreground(styles.Meta).
			// 	Render("  "+linkURL))
		}
	}

	return links
}
