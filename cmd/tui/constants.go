package main

// ViewState represents the current view state of the application
type ViewEnum int

const (
	WelcomeView ViewEnum = iota
	MesserView
)

type SSHConnectState int

const (
	Disconnected SSHConnectState = iota
	Connecting
	Connected
	Error
)
