package packets

// StatusPacket : Summarizes the set of messages the sending peer has seen so far (Vector clock)
type StatusPacket struct {
	Want []PeerStatus
}
