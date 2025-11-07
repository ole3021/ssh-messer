package messages

import "errors"

type AppErrMsg struct {
	Error   error
	IsFatal bool
}

var (
	ErrConfigNotFound = errors.New("configuration not found")
)
