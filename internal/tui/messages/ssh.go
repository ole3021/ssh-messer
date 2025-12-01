package messages

type SSHStartConnectMsg struct {
	ConfigFileName string
}

type SSHStatusUpdateMsg struct {
	ConfigFileName string
	Info           string
	Error          error
}

type SSHServiceProxyLogMsg struct {
	ConfigFileName string
	RequestID      string
	Method         string
	URL            string
	StatusCode     int
	Duration       string
	Error          error
}
