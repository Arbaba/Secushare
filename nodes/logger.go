package nodes

import (
	"fmt"
	"github.com/Arbaba/Peerster/packets"
	"strconv"
	"strings"
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

func (gossiper *Gossiper) LogSearchReply(reply *packets.SearchReply) {

}

func (gossiper *Gossiper) LogMatch(reply *packets.SearchReply, result *packets.SearchResult) {
	s := ""
	for _, chunknb := range result.ChunkMap {
		s += strconv.FormatUint(chunknb+1, 10) + ","
	}
	s = s[:(len(s) - 1)]
	fmt.Printf("FOUND match %s at %s metafile=%s chunks=%s\n",
		result.FileName,
		reply.Origin,
		HexToString(result.MetafileHash[:]),
		s,
	)

}

func (gossiper *Gossiper) LogUnconfirmed(tlc *packets.TLCMessage) {
	fmt.Printf("UNCONFIRMED GOSSIP origin %s ID %d file name %s size %d metahash %s\n",
		tlc.Origin,
		tlc.ID,
		tlc.TxBlock.Transaction.Name,
		tlc.TxBlock.Transaction.Size,
		HexToString(tlc.TxBlock.Transaction.MetafileHash),
	)

}

func (gossiper *Gossiper) LogTLC(tlc *packets.TLCMessage) {
	if tlc.Confirmed == -1 {
		gossiper.LogUnconfirmed(tlc)
	} else {
		gossiper.LogConfirmedTLC(tlc)
	}
}

func (gossiper *Gossiper) LogSendTLCAck(ack *packets.TLCAck, origin string) {
	fmt.Printf("SENDING ACK origin %s ID %d\n", origin, ack.ID)
}

func (gossiper *Gossiper) LogRebroadcast(id int, witnesses []string) {
	fmt.Printf("RE-BROADCAST ID %d WITNESSES %s\n", id, strings.Join(witnesses, ","))
}

func (gossiper *Gossiper) LogConfirmedTLC(tlc *packets.TLCMessage) {
	s := fmt.Sprintf("CONFIRMED GOSSIP origin %s ID %d file name %s size %d metahash %s\n",
		tlc.Origin,
		tlc.ID,
		tlc.TxBlock.Transaction.Name,
		tlc.TxBlock.Transaction.Size,
		HexToString(tlc.TxBlock.Transaction.MetafileHash),
	)
	fmt.Println(s)
	gossiper.LogsContainer.Add(s)

}

func (gossiper *Gossiper) LogAdvance(round int) {
	tlcs := gossiper.RoundState.RoundTLCMessages(round - 1)
	s := "ADVANCING TO ROUND " + fmt.Sprintf("%d", round-1) + " BASED ON CONFIRMED MESSAGES "
	for idx, tlc := range tlcs {
		s += fmt.Sprintf("origin%d %s ID%d %d,", idx+1, tlc.Origin, idx+1, tlc.Confirmed)
	}
	s = s[:len(s)-1]
	if gossiper.Hw3ex3 || gossiper.Hw3ex4 {
		fmt.Println(s)
	}
	gossiper.LogsContainer.Add(s)

}

func (gossiper *Gossiper) LogConsensus(origin string, id uint32) {
	newblock := gossiper.Blockchain.GetHead()
	names := gossiper.Blockchain.GetNames()
	h := newblock.Hash()
	s := fmt.Sprintf("CONSENSUS ON QSC round %d message origin %s ID %d %s size %d metahash %s\n",
		gossiper.RoundState.GetRound(),
		origin,
		id,
		strings.Join(names, " "),
		newblock.Transaction.Size,
		HexToString(h[:]))

	fmt.Println(s)
	gossiper.LogsContainer.Add(s)
}
