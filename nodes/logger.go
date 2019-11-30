package nodes

import (
	"fmt"
	"strconv"
	//"strings"
	"github.com/Arbaba/Peerster/packets"
)

//Prints to standard output
func (gossiper *Gossiper) LogPeers() {
	gossiper.PeersMux.Lock()
	defer gossiper.PeersMux.Unlock()
	//fmt.Println("PEERS", strings.Join(gossiper.Peers[:], ","))
}

func (gossiper *Gossiper) LogStatusPacket(packet *packets.StatusPacket, address string) {
	s := fmt.Sprintf("STATUS from %s ", address)
	for i, status := range packet.Want {
		s += fmt.Sprintf("peer %s nextID %d", status.Identifier, status.NextID)
		if i != len(packet.Want)-1 {
			s += " "
		}
	}
	//fmt.Println(s)
}

func (gossiper *Gossiper) LogRumor(rumor *packets.RumorMessage, peerAddr string) {
	/*fmt.Printf("RUMOR origin %s from %s ID %d contents %s\n",
		rumor.Origin,
		peerAddr,
		rumor.ID,
		rumor.Text)*/
}

func (gossiper *Gossiper) LogSimpleMessage(packet *packets.SimpleMessage) {
	fmt.Printf("SIMPLE MESSAGE origin %s from %s contents %s\n",
		packet.OriginalName,
		packet.RelayPeerAddr,
		packet.Contents)
}

func (gossiper *Gossiper) LogMongering(target string) {
	//fmt.Printf("MONGERING with %s\n", target)
}

func (gossiper *Gossiper) LogSync(peerAddr string) {
	//fmt.Printf("IN SYNC WITH %s\n", peerAddr)
}

func (gossiper *Gossiper) LogFlip(target string) {
	//fmt.Printf("FLIPPED COIN sending rumor to %s\n", target)
}

func (gossiper *Gossiper) LogClientMsg(msg packets.Message) {
	if msg.Destination != nil {
		fmt.Printf("CLIENT MESSAGE %s dest %s\n", msg.Text, *msg.Destination)
	} else {
		fmt.Printf("CLIENT MESSAGE %s\n", msg.Text)
	}
}

func (gossiper *Gossiper) LogDSDVRumor(rumor *packets.RumorMessage, address string) {
	if rumor.Text != "" {
		fmt.Printf("DSDV %s %s\n", rumor.Origin, address)
	}
}

func (gossiper *Gossiper) LogDSDVPrivate(private *packets.PrivateMessage, address string) {
	if private.Text != "" {
		fmt.Printf("DSDV %s %s\n", private.Origin, address)
	}
}
func (gossiper *Gossiper) LogPrivateMsg(private *packets.PrivateMessage) {
	fmt.Printf("PRIVATE origin %s hop-limit %d contents %s\n", private.Origin, private.HopLimit, private.Text)
}


func (gossiper *Gossiper) LogSearchReply(reply *packets.SearchReply){

}


func (gossiper *Gossiper) LogMatch(reply *packets.SearchReply, result*packets.SearchResult){
	s :=""
	for _, chunknb := range result.ChunkMap{
		s += strconv.FormatUint(chunknb, 10) + ","
	}
	fmt.Println(result.ChunkMap)
	s = s[:(len(s)-1)]
	fmt.Printf("FOUND match %s at %s metafile=%s chunks=%s\n", 
				result.FileName, 
				reply.Origin, 
				HexToString(result.MetaFileHash[:]),
				s,
			)

	
}