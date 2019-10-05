package packets

import "fmt"

// SimpleMessage : The simplest type of message a packet can contain
type SimpleMessage struct {
	OriginalName  string
	RelayPeerAddr string
	Contents      string
}

func (s SimpleMessage) String() string {
	return fmt.Sprintf("OriginalName: %s\nRelayPeerAddr: %s\nContents: %s\n", s.OriginalName, s.RelayPeerAddr, s.Contents)
}
