package types

// UIState UI 状态
type UIState struct {
	Width  int
	Height int
}

// NewUIState 创建新的 UI 状态
func NewUIState() *UIState {
	return &UIState{Width: 0, Height: 0}
}
