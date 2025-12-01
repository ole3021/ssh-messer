package tui

import (
	"fmt"
	"log/slog"
	"time"

	"ssh-messer/internal/messer"
	"ssh-messer/internal/pubsub"
	"ssh-messer/internal/tui/messages"
	"ssh-messer/pkg"

	tea "github.com/charmbracelet/bubbletea/v2"
)

// Subscribe to messer events and send them to the Bubbletea program.
func (a *appModel) Subscribe(program *tea.Program) {
	a.program = program
	pkg.Logger.Debug().Msg("[Subscribe] Start subscribing to messer events")
	defer func() {
		if r := recover(); r != nil {
			slog.Info("TUI subscription panic: attempting graceful shutdown", "error", r)
			program.Quit()
		}
	}()

	// Subscribe to events from all active MesserHops.
	for configFileName, hops := range a.appState.MesserHops {
		if hops == nil || hops.SubCh == nil {
			continue
		}

		go a.subscribeMesserEvents(configFileName, hops.SubCh)
	}
}

// subscribeMesserEvents subscribe to events from a single MesserHops.
func (a *appModel) subscribeMesserEvents(
	configFileName string,
	subCh <-chan pubsub.Event[messer.MesserEventType, any],
) {
	a.EventsCancelWG.Add(1)
	pkg.Logger.Info().Str("config", configFileName).Msg("[subscribe::subscribeMesserEvents] Subscribe to messer events")
	defer func() {
		a.EventsCancelWG.Done()
		if r := recover(); r != nil {
			slog.Error("messer event subscriber panic", "error", r, "config", configFileName)
		}
	}()

	for event := range subCh {
		var msg tea.Msg
		pkg.Logger.Debug().Str("config", configFileName).Str("event", string(event.Type)).Msg("[subscribe::subscribeMesserEvents] Messer event received")
		switch event.Type {
		case messer.EveSSHStatusUpdate:
			if payload, ok := event.Payload.(messer.EveSSHStatusUpdatePayload); ok {
				msg = messages.SSHStatusUpdateMsg{
					ConfigFileName: configFileName,
					Info:           payload.Info,
					Error:          payload.Error,
				}
			}
		case messer.EveServiceProxyLog:
			if payload, ok := event.Payload.(messer.EveSSHServiceProxyLogPayload); ok {
				// 格式化 duration 为字符串
				durationStr := formatDuration(payload.Duration)
				msg = messages.SSHServiceProxyLogMsg{
					ConfigFileName: configFileName,
					RequestID:      payload.RequestID,
					Method:         payload.Method,
					URL:            payload.URL,
					StatusCode:     payload.StatusCode,
					Duration:       durationStr,
					Error:          payload.Error,
				}
			}
		}

		if msg != nil {
			a.program.Send(msg)
		}
	}
	slog.Debug("messer event channel closed", "config", configFileName)
}

// formatDuration 格式化时间间隔为字符串
func formatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%.3fμs", float64(d.Nanoseconds())/1000.0)
	} else if d < time.Second {
		return fmt.Sprintf("%.3fms", float64(d.Nanoseconds())/1000000.0)
	} else {
		return fmt.Sprintf("%.3fs", d.Seconds())
	}
}
