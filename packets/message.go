package packets

// Message : message sent by the client to a gossiper
type Message struct {
	Text        string
	Destination *string
	File        *string
	Request     *[]byte
	Keywords 	*[]string
	Budget		*uint64
}

type PrivateMessage struct {
	Origin      string
	ID          uint32
	Text        string
	Destination string
	HopLimit    uint32
}
