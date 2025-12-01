package messages

import "ssh-messer/internal/config"

type ConfigLoadedMsg struct {
	Configs map[string]*config.MesserConfig
	Err     error
}
