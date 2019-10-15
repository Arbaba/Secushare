package packets


// SimpleMessage : The simplest type of message a packet can contain
type SimpleMessage struct {
	OriginalName  string
	RelayPeerAddr string
	Contents      string
}
