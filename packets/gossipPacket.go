package packets

// GossipPacket : The only type of packets sent to other peers
type GossipPacket struct {
	Simple       *SimpleMessage
	Rumor        *RumorMessage
	StatusPacket *StatusPacket
	Private      *PrivateMessage
	DataRequest  *DataRequest
	DataReply 	 *DataReply 
}


type DataRequest struct {
	Origin string
	Destination string
	HopLimit uint32
	HashValue []byte
}

type DataReply struct {
	Origin string
	Destination string
	HopLimit uint32
	HashValue []byte 
	Data []byte
}