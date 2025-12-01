package messages

type AppErrMsg struct {
	Error   error
	IsFatal bool
}
