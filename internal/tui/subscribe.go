package tui

import (
	"log/slog"

	tea "github.com/charmbracelet/bubbletea/v2"
)

// Subscribe 将事件发送到 Bubbletea 程序
func (a *appModel) Subscribe(program *tea.Program) {
	defer func() {
		if r := recover(); r != nil {
			slog.Info("TUI subscription panic: attempting graceful shutdown", "error", r)
			program.Quit()
		}
	}()

	a.serviceEventsWG.Add(1)
	defer a.serviceEventsWG.Done()

	for {
		select {
		case <-a.eventsCtx.Done():
			slog.Debug("TUI message handler shutting down")
			return
		case msg, ok := <-a.events:
			if !ok {
				slog.Debug("TUI message channel closed")
				return
			}
			program.Send(msg)
		}
	}
}
