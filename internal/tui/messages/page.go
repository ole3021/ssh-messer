package messages

type PageID string

const (
	WelcomePageID   PageID = "welcome"
	SSHMesserPageID PageID = "ssh_messer"
)

type PageChangeMsg struct {
	ID PageID
}
