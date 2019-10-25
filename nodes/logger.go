package nodes

import (
	"fmt"
	"strings"

	"github.com/Arbaba/Peerster/packets"
)

//Prints to standard output
func (gossiper *Gossiper) LogPeers() {
	fmt.Println("PEERS ", strings.Join(gossiper.Peers[:], ","))
}

func (gossiper *Gossiper) LogStatusPacket(packet *packets.StatusPacket, address string) {
	s := fmt.Sprintf("STATUS from %s ", address)
	for i, status := range packet.Want {
		s += fmt.Sprintf("peer %s nextID %d", status.Identifier, status.NextID)
		if i != len(packet.Want)-1 {
			s += " "
		}
	}
	fmt.Println(s)
}

func (gossiper *Gossiper) LogRumor(rumor *packets.RumorMessage, peerAddr string) {
	fmt.Printf("RUMOR origin %s from %s ID %d contents %s\n",
		rumor.Origin,
		peerAddr,
		rumor.ID,
		rumor.Text)
}

func (gossiper *Gossiper) LogSimpleMessage(packet *packets.SimpleMessage) {
	fmt.Printf("SIMPLE MESSAGE origin %s from %s contents %s\n",
		packet.OriginalName,
		packet.RelayPeerAddr,
		packet.Contents)
}

func (gossiper *Gossiper) LogMongering(target string) {
	fmt.Printf("MONGERING with %s\n", target)
}

func (gossiper *Gossiper) LogSync(peerAddr string) {
	fmt.Printf("IN SYNC WITH %s\n", peerAddr)
}

func (gossiper *Gossiper) LogFlip(target string) {
	fmt.Printf("FLIPPED COIN sending rumor to %s\n", target)
}

func (gossiper *Gossiper) LogClientMsg(msg string) {
	fmt.Printf("CLIENT MESSAGE %s\n", msg)
}
