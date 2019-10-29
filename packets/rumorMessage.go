package packets

// RumorMessage : Contains the text to gossip and metadata for the rumormongering protocol
type RumorMessage struct {
	Origin string
	ID     uint32
	Text   string
}

func RumorPacket(origin string, id uint32, text string) GossipPacket{
	return GossipPacket{Rumor: &RumorMessage{origin, id, text}}
}
