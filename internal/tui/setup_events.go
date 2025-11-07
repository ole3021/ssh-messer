package tui

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"ssh-messer/internal/pubsub"
	"ssh-messer/internal/ssh_proxy"

	tea "github.com/charmbracelet/bubbletea/v2"
)

func setupSubscriber[T any](
	ctx context.Context,
	wg *sync.WaitGroup,
	outputCh chan<- tea.Msg,
	name string,
	subscriber func(context.Context) <-chan pubsub.Event[T],
) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		subCh := subscriber(ctx)
		for {
			select {
			case event, ok := <-subCh:
				if !ok {
					slog.Debug("subscription channel closed", "name", name)
					return
				}
				var msg tea.Msg = event
				select {
				case outputCh <- msg:
				case <-time.After(2 * time.Second):
					slog.Warn("message dropped due to slow consumer", "name", name)
				case <-ctx.Done():
					slog.Debug("subscription cancelled", "name", name)
					return
				}
			case <-ctx.Done():
				slog.Debug("subscription cancelled", "name", name)
				return
			}
		}
	}()
}

// setupSSHStatusSubscriber 设置 SSH 状态更新订阅
func (a *appModel) setupSSHStatusSubscriber() {
	broker := ssh_proxy.GetStatusBroker()
	setupSubscriber(
		a.eventsCtx,
		a.serviceEventsWG,
		a.events,
		"ssh-status",
		broker.Subscribe,
	)
}
