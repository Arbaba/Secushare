package nodes

import (
	"github.com/Arbaba/Peerster/packets"
	"sync"
	"time"
)

type AcksReceived struct {
	sync.Mutex
	store map[uint32][]string
}

func (acks *AcksReceived) Add(tlc *packets.TLCAck) {
	acks.Lock()
	defer acks.Unlock()
	if !Contains(acks.store[tlc.ID], tlc.Origin) {
		acks.store[tlc.ID] = append(acks.store[tlc.ID], tlc.Origin)
	}
}

func (acks *AcksReceived) Witnesses(id uint32) []string {
	acks.Lock()
	defer acks.Unlock()
	return acks.store[id]
}

func CreateAcksReceived() *AcksReceived {
	var v AcksReceived
	v.store = make(map[uint32][]string)
	return &v
}

func (gossiper *Gossiper) Stubborn(tlcMessage *packets.TLCMessage) {
	if gossiper.StubbornTimeout <= 0 {
		return
	}
	ticker := time.NewTicker(time.Second * time.Duration(gossiper.StubbornTimeout))
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if len(gossiper.AcksReceived.Witnesses(tlcMessage.ID)) < int(gossiper.NetworkSize/2) {
				pkt := packets.GossipPacket{TLCMessage: tlcMessage}
				gossiper.RumorMonger(&pkt, gossiper.Name)
			}
		}
	}
}

func (gossiper *Gossiper) ACKTLC(tlc *packets.TLCMessage) {

	tlcAck := packets.TLCAck{
		Origin:      gossiper.Name,
		ID:          tlc.ID,
		Destination: tlc.Origin,
		HopLimit:    gossiper.HOPLIMIT,
	}
	pkt := packets.GossipPacket{Ack: &tlcAck}
	gossiper.SendDirect(pkt, tlc.Origin)
	gossiper.LogSendTLCAck(&tlcAck, tlc.Origin)

}

func (gossiper *Gossiper) ConfirmAndBroadcast(tlc packets.TLCMessage, witnesses []string) {
	tlc.Confirmed = int(tlc.ID)
	tlc.ID = gossiper.GetNextRumorID(tlc.Origin)
	tlc.VectorClock = gossiper.GetStatusPacket()
	pkt := &packets.GossipPacket{TLCMessage: &tlc}
	gossiper.RumorMonger(pkt, gossiper.RelayAddress())
	gossiper.LogRebroadcast(int(tlc.ID), witnesses)
}
