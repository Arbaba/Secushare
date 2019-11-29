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
	MetafileHash []byte
	ChunkMap []uint64
	ChunkCount uint64
}

func (result *SearchResult) IsComplete() bool{
	return len(result.ChunkMap) == int(result.ChunkCount)
}