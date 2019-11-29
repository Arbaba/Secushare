package packets

// GossipPacket : The only type of packets sent to other peers
type GossipPacket struct {
	Simple      	*SimpleMessage
	Rumor       	*RumorMessage
	StatusPacket	*StatusPacket
	Private     	*PrivateMessage
	DataRequest 	*DataRequest
	DataReply 		*DataReply 
	SearchRequest 	*SearchRequest
	SearchReply 	*SearchReply
	TLCMessage 		*TLCMessage
	Ack 			*TLCAck
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

type SearchRequest struct {
	Origin string
	Budget uint64
	Keywords  []string 

}

type SearchReply struct {
	Origin string
	Destination string
	HopLimit uint32
	Results []*SearchResult
}


type SearchResult struct {
	FileName string 
	MetaFileHash []byte
	ChunkMap []uint64
	ChunkCount uint64
}

func (result *SearchResult) IsComplete() bool{
	return len(result.ChunkMap) == int(result.ChunkCount)
}


type TxPublish struct {
	Name string
	Size int64
	MetafileHash []byte
}

type BlockPublish struct {
	PrevHash [32]byte
	Transaction TxPublish
}

type TLCMessage struct{
	Origin string
	ID uint32
	Confirmed int64
	TxBlock BlockPublish
	VectorClock *StatusPacket
	Fitness		float32
}

type TLCAck PrivateMessage