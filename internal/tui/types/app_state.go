package types

import (
	"ssh-messer/internal/config"
	"ssh-messer/internal/messer"
)

type AppState struct {
	Configs               map[string]*config.MesserConfig
	CurrentConfigFileName string
	MesserHops            map[string]*messer.MesserHops

	Error AppError
}

func NewAppState() *AppState {
	return &AppState{
		Configs:    make(map[string]*config.MesserConfig),
		MesserHops: make(map[string]*messer.MesserHops),
	}
}

func (s *AppState) SetConfigs(configs map[string]*config.MesserConfig) {
	s.Configs = configs
}

func (s *AppState) GetConfigs() map[string]*config.MesserConfig {
	return s.Configs
}

func (s *AppState) UpSetConfig(configName string, config *config.MesserConfig) {
	s.Configs[configName] = config
}

func (s *AppState) GetConfig(configName string) *config.MesserConfig {
	return s.Configs[configName]
}

func (s *AppState) GetCurrentConfig() *config.MesserConfig {
	return s.Configs[s.CurrentConfigFileName]
}

func (s *AppState) SetCurrentConfigFileName(configName string) {
	s.CurrentConfigFileName = configName
}

func (s *AppState) UpSetMesserHops(configName string, messerHops *messer.MesserHops) {
	s.MesserHops[configName] = messerHops
}

func (s *AppState) GetMesserHops(configName string) *messer.MesserHops {
	if proxy, exists := s.MesserHops[configName]; exists {
		return proxy
	}
	return nil
}

func (s *AppState) SetAppError(error error, isFatal bool) {
	s.Error = AppError{
		Error:   error,
		IsFatal: isFatal,
	}
}

func (s *AppState) GetAppError() *AppError {
	return &s.Error
}

type AppError struct {
	Error   error
	IsFatal bool
}
