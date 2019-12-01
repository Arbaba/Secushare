package packets

// RumorMessage : Contains the text to gossip and metadata for the rumormongering protocol
type RumorMessage struct {
	Origin string
	ID     uint32
	Text   string
}

func (rumor * RumorMessage) GetOrigin() string {
	return rumor.Origin
}

func (rumor * RumorMessage) GetID() uint32 {
	return rumor.ID
}


func (rumor * RumorMessage) GetText() string {
	return rumor.Text
}

func (rumor *RumorMessage) IsRumor() bool  {
	return true
}
func (rumor *RumorMessage) IsTLCMsg() bool  {
	return false
}
type Rumorable interface {
	GetOrigin() string
	GetID()		uint32
	IsRumor()	bool
	IsTLCMsg()	bool 		
}




func RumorPacket(origin string, id uint32, text string) GossipPacket{
	return GossipPacket{Rumor: &RumorMessage{origin, id, text}}
}
