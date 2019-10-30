package packets

// GossipPacket : The only type of packets sent to other peers
type GossipPacket struct {
	Simple       *SimpleMessage
	Rumor        *RumorMessage
	StatusPacket *StatusPacket
	Private 	 *PrivateMessage
}
