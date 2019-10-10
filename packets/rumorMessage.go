package packets

// RumorMessage : Contains the text to gossip and metadata for the rumormongering protocol
type RumorMessage struct {
	Origin string
	ID     string
	Text   string
}
