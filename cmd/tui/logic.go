package main

import (
	"log"
	"os"
	"path/filepath"
	"ssh-messer/internal/loaders"
	"ssh-messer/internal/proxy"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Welcome animation message
type welcomeTickMsg struct{}

// SSH connection animation message
type sshConnectionTickMsg struct{}

// InitialModel creates the initial application model
func InitialModel() AppModel {
	return AppModel{
		Configs:     make(map[string]loaders.TomlConfig),
		SSHInfos:    make(map[string]SSHInfo),
		CurrentView: WelcomeView,
	}
}

var sshClientChan = make(chan proxy.SSHClientResultChan)
var sshProcessChan = make(chan proxy.SSHProcessChan)

// Init initializes the model and returns initial commands
func (m AppModel) Init() tea.Cmd {
	return tea.Batch(
		loadConfigsFromHomeDir(),
		tickWelcomeAnimation(),
	)
}

// Update handles messages and updates the model
func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case loadConfigsMsg:
		if msg.err != nil {
			log.Printf("加载配置文件失败: %v", msg.err)
			return m, nil
		}
		m.Configs = msg.configs
		m.ConfigViewModel.ConfigNames = msg.configNames
		return m, nil
	case sshHopConnectMsg:
		// 直接使用消息中的数据更新 model
		m.SSHInfos[msg.configName] = msg.sshInfo
		m.CurrentInfo = msg.currentInfo

		// 如果有错误，确保记录
		if msg.err != nil {
			log.Printf("SSH连接错误: %v", msg.err)
		}

		return m, nil
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil
	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	case welcomeTickMsg:
		return m.handleWelcomeTick()
	case sshConnectionTickMsg:
		return m.handleSSHConnectionTick()
	case sshClientResultMsg:
		return m.handleSSHClientResult(msg.result)
	case sshProcessResultMsg:
		return m.handleSSHProcessResult(msg.result)
	default:
		return m, nil
	}
}

type loadConfigsMsg struct {
	configs     map[string]loaders.TomlConfig
	configNames []string
	err         error
}

// loadConfigsFromHomeDir 异步加载 ~/.ssh_messer 目录下的所有 toml 文件
func loadConfigsFromHomeDir() tea.Cmd {
	return func() tea.Msg {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return loadConfigsMsg{err: err}
		}

		configDir := filepath.Join(homeDir, ".ssh_messer")
		configs := make(map[string]loaders.TomlConfig)
		configNames := []string{}

		// 读取目录下的所有 toml 文件
		files, err := filepath.Glob(filepath.Join(configDir, "*.toml"))
		if err != nil {
			return loadConfigsMsg{err: err}
		}

		for _, file := range files {
			config, err := loaders.LoadTomlProxyConfig(filepath.Base(file), configDir)
			if err != nil {
				log.Printf("加载配置文件失败 %s: %v", file, err)
				continue
			}

			// 提取配置名称（文件名去掉扩展名）
			name := strings.TrimSuffix(filepath.Base(file), ".toml")

			configs[name] = *config
			configNames = append(configNames, name)
		}

		return loadConfigsMsg{configs: configs, configNames: configNames}
	}
}

type sshHopConnectMsg struct {
	configName  string
	sshInfo     SSHInfo
	currentInfo string
	err         error
}

// func (m AppModel) connectSSHHops() tea.Cmd {
// 	return func() tea.Msg {
// 		configName := m.CurrentConfigName
// 		sshHopsConfigs := m.Configs[configName].SSHHops
// 		sshInfo := m.SSHInfos[configName]

// 		// 对sshHopsConfigs 按照order 从小到大进行排序
// 		sort.Slice(sshHopsConfigs, func(i, j int) bool {
// 			return *sshHopsConfigs[i].Order < *sshHopsConfigs[j].Order
// 		})

// 		var client *ssh.Client

// 		for i, sshHopConfig := range sshHopsConfigs {
// 			sshAddress := *sshHopConfig.Host + ":" + strconv.Itoa(*sshHopConfig.Port|22)
// 			var aliasName string
// 			if sshHopConfig.Alias != nil {
// 				aliasName = *sshHopConfig.Alias
// 			} else {
// 				aliasName = sshAddress
// 			}

// 			// 更新连接状态
// 			sshInfo.SSHConnectionState = Connecting
// 			sshInfo.SSHConnectionProcess = i + 1

// 			sshClientConfig, err := proxy.TransformSSHClientConfig(sshHopConfig)
// 			if err != nil {
// 				sshInfo.SSHConnectionState = Error
// 				return sshHopConnectMsg{
// 					configName:  configName,
// 					sshInfo:     sshInfo,
// 					currentInfo: fmt.Sprintf("配置转换失败: %v", err),
// 					err:         err,
// 				}
// 			}

// 			if i == 0 {
// 				// 第一跳：直接连接
// 				client, err = ssh.Dial("tcp", sshAddress, sshClientConfig)
// 				if err != nil {
// 					sshInfo.SSHConnectionState = Error
// 					return sshHopConnectMsg{
// 						configName:  configName,
// 						sshInfo:     sshInfo,
// 						currentInfo: fmt.Sprintf("SSH连接 [%s] 失败: %v", aliasName, err),
// 						err:         err,
// 					}
// 				}
// 			} else {
// 				// 后续跳：通过隧道连接
// 				conn, err := client.Dial("tcp", sshAddress)
// 				if err != nil {
// 					client.Close()
// 					sshInfo.SSHConnectionState = Error
// 					return sshHopConnectMsg{
// 						configName:  configName,
// 						sshInfo:     sshInfo,
// 						currentInfo: fmt.Sprintf("隧道连接 [%s] 失败: %v", aliasName, err),
// 						err:         err,
// 					}
// 				}

// 				nconn, chans, reqs, err := ssh.NewClientConn(conn, sshAddress, sshClientConfig)
// 				if err != nil {
// 					conn.Close()
// 					client.Close()
// 					sshInfo.SSHConnectionState = Error
// 					return sshHopConnectMsg{
// 						configName:  configName,
// 						sshInfo:     sshInfo,
// 						currentInfo: fmt.Sprintf("SSH连接 [%s] 失败: %v", aliasName, err),
// 						err:         err,
// 					}
// 				}

// 				client = ssh.NewClient(nconn, chans, reqs)
// 			}

// 			// 更新连接进度
// 			sshInfo.SSHConnectionProcess = i + 1
// 		}

// 		// 所有跳连接完成
// 		sshInfo.SSHClient = client
// 		sshInfo.SSHConnectionState = Connected
// 		sshInfo.SSHConnectionProcess = 0
// 		sshInfo.HTTPProxyLogs = []string{}
// 		sshInfo.DockerProxyLogs = []string{}

// 		return sshHopConnectMsg{
// 			configName:  configName,
// 			sshInfo:     sshInfo,
// 			currentInfo: "SSH 连接已建立",
// 			err:         nil,
// 		}
// 	}
// }

// handleKeyPress processes keyboard input
func (m AppModel) handleKeyPress(msg tea.KeyMsg) (AppModel, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	default:
		switch m.CurrentView {
		case WelcomeView:
			return m.handleConfigSelection(msg)
		}
	}
	return m, nil
}

// *** Welcome View ***
// ********************
// Handling welcome screen animation with tick
func tickWelcomeAnimation() tea.Cmd {
	return tea.Tick(time.Millisecond*50, func(t time.Time) tea.Msg {
		return welcomeTickMsg{}
	})
}

func (m AppModel) handleWelcomeTick() (AppModel, tea.Cmd) {
	if m.CurrentView == WelcomeView {
		m.WelcomeViewModel.WelcomeAnimationProgress += 3
		if m.WelcomeViewModel.WelcomeAnimationProgress >= 100 {
			// Animation completed, stay on welcome screen
			// User can now interact with config selection
			return m, nil
		}
		return m, tickWelcomeAnimation()
	}
	return m, nil
}

// *** Config View ***
// ********************
// Handling config selection view interactions
func (m AppModel) handleConfigSelection(msg tea.KeyMsg) (AppModel, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.ConfigViewModel.Cursor > 0 {
			m.ConfigViewModel.Cursor--
		}
	case "down", "j":
		if m.ConfigViewModel.Cursor < len(m.Configs)-1 {
			m.ConfigViewModel.Cursor++
		}
	case "enter":
		m.CurrentConfigName = m.ConfigViewModel.ConfigNames[m.ConfigViewModel.Cursor]
		m.SSHInfos[m.CurrentConfigName] = SSHInfo{
			SSHConnectionState:   Connecting,
			SSHConnectionProcess: 0,
			HTTPProxyLogs:        []string{},
			DockerProxyLogs:      []string{},
		}

		// TODO: start SSH connection and & Redirect to MainView if not in MainView
		if m.CurrentView == WelcomeView {
			m.CurrentView = MesserView
		}

		go proxy.AsyncCreateSSHHopsClient(m.Configs[m.CurrentConfigName].SSHHops, sshClientChan, &sshProcessChan)
		return m, tea.Batch(
			tickSSHConnectionAnimation(),
			listenToSSHChannels(), // 开始监听 channel
		)
	}
	return m, nil
}

// *** Messer View ***
// ********************

func tickSSHConnectionAnimation() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return sshConnectionTickMsg{}
	})
}

func (m AppModel) handleSSHConnectionTick() (AppModel, tea.Cmd) {
	sshInfo := m.SSHInfos[m.CurrentConfigName]
	if sshInfo.SSHConnectionState == Connecting {
		sshInfo.SSHConnectionProcess++
		m.SSHInfos[m.CurrentConfigName] = sshInfo
		return m, tickSSHConnectionAnimation()
	}

	return m, nil
}

// 在 logic.go 中添加这些消息类型
type sshClientResultMsg struct {
	result proxy.SSHClientResultChan
}

type sshProcessResultMsg struct {
	result proxy.SSHProcessChan
}

// 添加监听 channel 的函数
func listenToSSHChannels() tea.Cmd {
	return func() tea.Msg {
		select {
		case result := <-sshClientChan:
			return sshClientResultMsg{result: result}
		case result := <-sshProcessChan:
			return sshProcessResultMsg{result: result}
		}
	}
}

// 修改现有的处理函数，添加 model 参数
func (m AppModel) handleSSHClientResult(result proxy.SSHClientResultChan) (AppModel, tea.Cmd) {
	if result.Error != nil {
		// 处理错误
		sshInfo := m.SSHInfos[m.CurrentConfigName]
		sshInfo.SSHConnectionState = Error
		m.SSHInfos[m.CurrentConfigName] = sshInfo
		return m, nil
	}

	// 更新 SSH 客户端
	sshInfo := m.SSHInfos[m.CurrentConfigName]
	sshInfo.SSHClient = result.Client
	sshInfo.SSHConnectionState = Connected
	m.SSHInfos[m.CurrentConfigName] = sshInfo

	return m, listenToSSHChannels() // 继续监听
}

func (m AppModel) handleSSHProcessResult(result proxy.SSHProcessChan) (AppModel, tea.Cmd) {
	sshInfo := m.SSHInfos[m.CurrentConfigName]

	if result.Error != nil {
		sshInfo.SSHConnectionState = Error
		m.SSHInfos[m.CurrentConfigName] = sshInfo
		m.CurrentInfo = result.Message
		return m, nil
	}

	// 更新连接进度
	sshInfo.SSHConnectionProcess = result.CompletedHopsCount
	m.SSHInfos[m.CurrentConfigName] = sshInfo
	m.CurrentInfo = result.Message

	return m, listenToSSHChannels() // 继续监听
}
