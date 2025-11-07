package messages

import "ssh-messer/internal/ssh_proxy"

// SSHStatusMsg SSH 状态更新消息（通过 pubsub）
type SSHStatusMsg struct {
	ConfigName string
	Status     ssh_proxy.SSHProxyStatus
}
