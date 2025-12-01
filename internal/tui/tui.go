package tui

import (
	"context"
	"sync"

	"ssh-messer/internal/messer"
	"ssh-messer/internal/tui/commands"
	"ssh-messer/internal/tui/messages"
	"ssh-messer/internal/tui/page/ssh_messer"
	"ssh-messer/internal/tui/page/welcome"
	"ssh-messer/internal/tui/types"
	"ssh-messer/internal/tui/util"
	"ssh-messer/pkg"

	"github.com/charmbracelet/bubbles/v2/key"
	tea "github.com/charmbracelet/bubbletea/v2"
)

type appModel struct {
	keyMap KeyMap

	// Pages State
	currentPage  types.PageID
	previousPage types.PageID
	pages        map[types.PageID]util.Model
	loadedPages  map[types.PageID]bool

	// State
	appState *types.AppState
	uiState  *types.UIState

	// TODO: Tidy, Program init in subscribe, used to send messages to the program.
	program        *tea.Program
	EventsCtx      context.Context
	EventsCancel   context.CancelFunc
	EventsCancelWG *sync.WaitGroup
}

// New creates and initializes a new TUI application model.
func New() *appModel {
	appState := types.NewAppState()
	uiState := types.NewUIState()
	ctx, cancel := context.WithCancel(context.Background())

	model := &appModel{
		currentPage:  types.WelcomePageID,
		previousPage: types.WelcomePageID,

		loadedPages: make(map[types.PageID]bool),
		keyMap:      DefaultKeyMap(),

		appState: appState,
		uiState:  uiState,

		EventsCtx:      ctx,
		EventsCancel:   cancel,
		EventsCancelWG: &sync.WaitGroup{},

		pages: map[types.PageID]util.Model{
			types.WelcomePageID:   welcome.New(appState, uiState),
			types.SSHMesserPageID: ssh_messer.New(appState, uiState),
		},
	}

	return model
}

func (a *appModel) Init() tea.Cmd {
	page, ok := a.pages[a.currentPage]
	if !ok {
		return nil
	}

	var cmds []tea.Cmd
	cmd := page.Init()
	cmds = append(cmds, cmd)
	// TODO: Combine loadedPages to pages map.
	a.loadedPages[a.currentPage] = true

	return tea.Batch(cmds...)
}

func (a *appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, a.keyMap.Quit) {
			return a, tea.Quit
		}

	case tea.WindowSizeMsg:
		for _, page := range a.pages {
			// Delegate msg to sub model with window size msg.
			subCmds := util.DelegateMsgToSubModel(cmds, msg, &page)
			cmds = append(cmds, subCmds...)
		}

	case messages.ConfigLoadedMsg:
		if msg.Err != nil {
			return a, commands.ReportAppErrCmd(ErrConfigNotFound, false)
		}
		// Save configs to app state
		a.appState.SetConfigs(msg.Configs)
		// Don't return , ConfigList Component will handle the same message.

	case messages.SSHStartConnectMsg:
		a.appState.SetCurrentConfigFileName(msg.ConfigFileName)
		// Create and start messer hop connect
		config := a.appState.GetConfig(msg.ConfigFileName)
		if config == nil {
			return a, commands.ReportAppErrCmd(ErrConfigNotFound, false)
		}
		messerHop := messer.NewMesserHops(config, a.EventsCtx)
		go messerHop.ConnectHops()
		a.appState.UpSetMesserHops(msg.ConfigFileName, messerHop)
		// Subscribe to messer events
		go a.subscribeMesserEvents(msg.ConfigFileName, messerHop.SubCh)

		a.moveToPage(types.SSHMesserPageID)
		return a, nil

	// App error
	case messages.AppErrMsg:
		model, cmd := a.handleAppErrorMsg(msg)
		return model, cmd
	}

	// Delegate msg to current page.
	currentPage := a.pages[a.currentPage]
	return a, tea.Batch(util.DelegateMsgToSubModel(cmds, msg, &currentPage)...)
}

// Render view from view model.
func (a *appModel) View() string {
	return a.pages[a.currentPage].View()
}

// moveToPage handles navigation between different pages in the application.
func (a *appModel) moveToPage(pageID types.PageID) tea.Cmd {
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

// handleAppErrorMsg handles application error messages
// TODO: Handling app error message.
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

func (a *appModel) Cleanup() {

	// cancel context, this will trigger all subscribed channels to be closed
	if a.EventsCancel != nil {
		pkg.Logger.Info().Msg("[tui:Cleanup] Cancelling events context")
		a.EventsCancel()
	}

	// close all MesserHops brokers (ensure fully closed)
	for configFileName, hops := range a.appState.MesserHops {
		if hops != nil && hops.Broker != nil {
			pkg.Logger.Info().Str("config", configFileName).Msg("[tui:Cleanup] Closing MesserHops broker")
			hops.Broker.Shutdown()
		}
	}

	// wait for subscribeMesserEvents goroutines to finish
	if a.EventsCancelWG != nil {
		a.EventsCancelWG.Wait()
		pkg.Logger.Info().Msg("[tui:Cleanup] Events cancel wait group finished")
	}
}
