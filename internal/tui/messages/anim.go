package messages

import "ssh-messer/internal/tui/types"

type StepMsg struct {
	AnimType types.AnimationType
	ID       int // Optional ID for char animation instances
}
