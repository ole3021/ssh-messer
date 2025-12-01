package config_list

import (
	"fmt"
	"ssh-messer/internal/config"
	"ssh-messer/internal/tui/commands"
	"ssh-messer/internal/tui/components/core/layout"
	"ssh-messer/internal/tui/messages"
	"ssh-messer/internal/tui/util"
	"strings"

	"github.com/charmbracelet/bubbles/v2/list"
	tea "github.com/charmbracelet/bubbletea/v2"
)

// use bubbletea list
// TODO: Separate bubbletea list into a separate file
type ConfigItem struct {
	filename string
	config   *config.MesserConfig
}

func (i ConfigItem) Title() string {
	if i.config.Name == "" {
		return strings.ReplaceAll(i.filename, ".toml", "")
	}

	return i.config.Name
}

func (i ConfigItem) Description() string {
	if i.config == nil {
		return ""
	}

	httpPort := "N/A"
	if i.config.LocalHttpPort != "" {
		httpPort = i.config.LocalHttpPort
	}

	dockerPort := "N/A"
	if i.config.LocalDockerPort != "" {
		dockerPort = i.config.LocalDockerPort
	}

	return fmt.Sprintf("%2d Hops🦘, %3d Services🔗, LocalPort🕸️: %4s, DockerPort🐳: %s",
		len(i.config.SSHHops),
		len(i.config.ReverseServices),
		httpPort,
		dockerPort)
}

func (c ConfigItem) FilterValue() string {
	return c.filename
}

type ConfigListCmp interface {
	util.Model
	layout.Sizeable
}

type configListCmp struct {
	width, height int
	list          list.Model
}

func NewConfigListCmp() ConfigListCmp {
	items := []list.Item{}
	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.SetShowHelp(false)
	l.SetShowTitle(true)
	l.SetShowFilter(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Title = "Local Configs"
	styles := list.DefaultStyles(true)
	l.Styles.Title = styles.Title
	l.Styles.NoItems = styles.NoItems

	return &configListCmp{list: l}
}

func (c *configListCmp) Init() tea.Cmd {
	return commands.LoadAllConfigsCmd()
}

func (c *configListCmp) Update(msg tea.Msg) (util.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case messages.ConfigLoadedMsg:
		if msg.Err != nil {
			return c, util.ReportError(msg.Err)
		}

		var configItems []list.Item
		for configName, config := range msg.Configs {
			item := ConfigItem{
				filename: configName,
				config:   config,
			}
			configItems = append(configItems, item)
		}

		c.list.SetItems(configItems)
		return c, nil
	case tea.KeyMsg:
		if msg.String() == "enter" {
			selectedItem := c.list.SelectedItem()
			if item, ok := selectedItem.(ConfigItem); ok {
				return c, util.CmdHandler(messages.SSHStartConnectMsg{
					ConfigFileName: item.filename,
				})
			}
		}
	}

	// process bubbletea list update
	var cmd tea.Cmd
	c.list, cmd = c.list.Update(msg)
	return c, cmd
}

func (c *configListCmp) View() string {
	return c.list.View()
}

func (c *configListCmp) SetSize(width, height int) tea.Cmd {
	c.list.SetWidth(width)
	c.list.SetHeight(height)
	return nil
}

func (c *configListCmp) GetSize() (int, int) {
	return c.width, c.height
}
