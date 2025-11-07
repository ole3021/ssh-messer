package tui

import (
	"context"
	"sync"

	"ssh-messer/internal/pubsub"
	"ssh-messer/internal/tui/commands"
	"ssh-messer/internal/tui/components/core/status"
	"ssh-messer/internal/tui/messages"
	"ssh-messer/internal/tui/page/ssh_messer"
	"ssh-messer/internal/tui/page/welcome"
	"ssh-messer/internal/tui/types"
	"ssh-messer/internal/tui/util"

	"github.com/charmbracelet/bubbles/v2/key"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
)

// appModel represents the main application model that manages pages and UI state.
type appModel struct {
	keyMap KeyMap

	// Pages State
	currentPage  messages.PageID
	previousPage messages.PageID
	pages        map[messages.PageID]util.Model
	loadedPages  map[messages.PageID]bool

	// Status
	status status.StatusCmp

	// State
	appState *types.AppState
	uiState  *types.UIState

	// SSH status updates via pubsub
	events          chan tea.Msg
	eventsCtx       context.Context
	eventsCancel    context.CancelFunc
	serviceEventsWG *sync.WaitGroup
}

// Init initializes the application model and returns initial commands.
func (a *appModel) Init() tea.Cmd {
	item, ok := a.pages[a.currentPage]
	if !ok {
		return nil
	}

	var cmds []tea.Cmd

	// 加载配置（系统配置 + TOML 配置）
	cmds = append(cmds, commands.LoadAllConfigs())

	// 初始化当前页面
	cmd := item.Init()
	cmds = append(cmds, cmd)
	a.loadedPages[a.currentPage] = true

	// 初始化状态栏
	cmd = a.status.Init()
	cmds = append(cmds, cmd)

	return tea.Batch(cmds...)
}

// Update handles incoming messages and updates the application state.
func (a *appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.uiState.Width, a.uiState.Height = msg.Width, msg.Height
		return a, a.handleWindowResize(msg.Width, msg.Height)

	// Page change messages
	case messages.PageChangeMsg:
		return a, a.moveToPage(msg.ID)

	// Status Messages
	case util.InfoMsg, util.ClearStatusMsg:
		s, statusCmd := a.status.Update(msg)
		a.status = s.(status.StatusCmp)
		cmds = append(cmds, statusCmd)
		return a, tea.Batch(cmds...)

	// Configuration messages
	case messages.LoadConfigsMsg:
		// 处理配置加载消息
		if msg.Err != nil {
			return a, util.ReportError(msg.Err)
		}
		// 更新应用状态
		a.appState.SetConfigs(msg.Configs)
		// 消息需要传递到当前页面，所以继续执行，不 return
	case messages.ConfigSelectedMsg:
		model, cmd := a.handleConfigSelectedMsg(msg)
		return model, cmd

	// App error
	case messages.AppErrMsg:
		model, cmd := a.handleAppErrorMsg(msg)
		return model, cmd

	// SSH status updates via pubsub
	case pubsub.Event[types.EveSSHStatusUpdate]:
		model, cmd := a.handleSSHStatusUpdate(msg)
		return model, cmd

	case tea.KeyMsg:
		return a, a.handleKeyPressMsg(msg)
	}

	// Update status bar
	s, _ := a.status.Update(msg)
	a.status = s.(status.StatusCmp)

	// Delegate to current page
	item, ok := a.pages[a.currentPage]
	if !ok {
		return a, nil
	}

	updated, cmd := item.Update(msg)
	a.pages[a.currentPage] = updated
	cmds = append(cmds, cmd)

	return a, tea.Batch(cmds...)
}

// View renders the complete application interface including pages and status bar.
func (a *appModel) View() string {
	page := a.pages[a.currentPage]
	pageView := page.View()

	// 组合页面视图和状态栏
	components := []string{
		pageView,
	}
	if a.status != nil {
		components = append(components, a.status.View())
	}

	return lipgloss.JoinVertical(lipgloss.Top, components...)
}

// handleWindowResize processes window resize events and updates all components.
func (a *appModel) handleWindowResize(width, height int) tea.Cmd {
	var cmds []tea.Cmd

	// Update uiState instead of local fields
	a.uiState.Width = width
	a.uiState.Height = height

	// Update status bar
	s, cmd := a.status.Update(tea.WindowSizeMsg{Width: width, Height: height})
	if model, ok := s.(status.StatusCmp); ok {
		a.status = model
	}
	cmds = append(cmds, cmd)

	// Update all pages
	for p, page := range a.pages {
		updated, pageCmd := page.Update(tea.WindowSizeMsg{Width: width, Height: height})
		a.pages[p] = updated
		cmds = append(cmds, pageCmd)
	}

	return tea.Batch(cmds...)
}

// handleKeyPressMsg processes keyboard input and routes to appropriate handlers.
func (a *appModel) handleKeyPressMsg(msg tea.KeyMsg) tea.Cmd {
	// Check this first as the user should be able to quit no matter what.
	if key.Matches(msg, a.keyMap.Quit) {
		return tea.Quit
	}

	// Delegate to current page
	item, ok := a.pages[a.currentPage]
	if !ok {
		return nil
	}

	updated, cmd := item.Update(msg)
	a.pages[a.currentPage] = updated
	return cmd
}

// moveToPage handles navigation between different pages in the application.
func (a *appModel) moveToPage(pageID messages.PageID) tea.Cmd {
	var cmds []tea.Cmd

	// Lazy load page if not loaded
	if _, ok := a.loadedPages[pageID]; !ok {
		if page, exists := a.pages[pageID]; exists {
			cmd := page.Init()
			cmds = append(cmds, cmd)
			a.loadedPages[pageID] = true
		}
	}

	a.previousPage = a.currentPage
	a.currentPage = pageID

	// Update page size
	if a.uiState.Width > 0 && a.uiState.Height > 0 {
		item, ok := a.pages[a.currentPage]
		if ok {
			if sizable, ok := item.(interface{ SetSize(int, int) tea.Cmd }); ok {
				cmd := sizable.SetSize(a.uiState.Width, a.uiState.Height)
				cmds = append(cmds, cmd)
			}
		}
	}

	return tea.Batch(cmds...)
}

// handleLoadConfigsMsg handles configuration loading messages
func (a *appModel) handleLoadConfigsMsg(msg messages.LoadConfigsMsg) (tea.Model, tea.Cmd) {
	if msg.Err != nil {
		return a, util.ReportError(msg.Err)
	}

	// Update app state with loaded configs
	a.appState.SetConfigs(msg.Configs)
	return a, nil
}

// handleConfigSelectedMsg handles configuration selection messages
func (a *appModel) handleConfigSelectedMsg(msg messages.ConfigSelectedMsg) (tea.Model, tea.Cmd) {
	a.appState.CurrentConfigName = msg.ConfigName

	var cmds []tea.Cmd
	// 切换到 SSH Messer 页面
	cmds = append(cmds, util.CmdHandler(messages.PageChangeMsg{ID: messages.SSHMesserPageID}))
	// 初始化 SSH 代理
	cmds = append(cmds, commands.InitSSHProxy(a.appState, msg.ConfigName))

	return a, tea.Batch(cmds...)
}

// handleAppErrorMsg handles application error messages
func (a *appModel) handleAppErrorMsg(appErr messages.AppErrMsg) (tea.Model, tea.Cmd) {
	a.appState.Error = types.AppError{
		Error:   appErr.Error,
		IsFatal: appErr.IsFatal,
	}
	if appErr.IsFatal {
		return a, tea.Quit
	}
	return a, util.ReportError(appErr.Error)
}

// handleSSHStatusUpdate handles SSH status updates from pubsub
func (a *appModel) handleSSHStatusUpdate(event pubsub.Event[types.EveSSHStatusUpdate]) (tea.Model, tea.Cmd) {
	update := event.Payload
	a.appState.GetSSHProxy(update.ConfigName).UpdateSSHProxyStatus(update.Status)

	// Forward to current page
	item, ok := a.pages[a.currentPage]
	if !ok {
		return a, nil
	}

	updated, cmd := item.Update(event)
	a.pages[a.currentPage] = updated
	return a, cmd
}

// New creates and initializes a new TUI application model.
func New() *appModel {
	appState := types.NewAppState()
	uiState := types.NewUIState()

	welcomePage := welcome.New(appState, uiState)

	ctx, cancel := context.WithCancel(context.Background())

	model := &appModel{
		currentPage:     messages.WelcomePageID,
		appState:        appState,
		uiState:         uiState,
		status:          status.NewStatusCmp(),
		loadedPages:     make(map[messages.PageID]bool),
		keyMap:          DefaultKeyMap(),
		events:          make(chan tea.Msg, 100),
		eventsCtx:       ctx,
		eventsCancel:    cancel,
		serviceEventsWG: &sync.WaitGroup{},

		pages: map[messages.PageID]util.Model{
			messages.WelcomePageID:   welcomePage,
			messages.SSHMesserPageID: ssh_messer.New(appState, uiState),
		},
	}

	// 设置 SSH 状态更新订阅
	model.setupSSHStatusSubscriber()

	return model
}
