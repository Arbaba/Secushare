package nodes

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
	VectorClock *packets.StatusPacket
	Fitness		float 32
}