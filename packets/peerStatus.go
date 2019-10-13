package packets

import "fmt"

// PeerStatus : Status of the messages received a given peer
type PeerStatus struct {
	Identifier string
	NextID     uint32
}

func (peerStatus *PeerStatus) String() string {
	return fmt.Sprintf("peer %s nextID %d", peerStatus.Identifier, peerStatus.NextID)
}
