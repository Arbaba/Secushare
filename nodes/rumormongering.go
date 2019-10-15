package nodes

import (
	"github.com/Arbaba/Peerster/packets"
	"math/rand"
	"time"
	"strings"
	"fmt"
)

/*
RumorMongering and status handling
*/


func (gossiper *Gossiper) AckID(identifier string, nextID uint32, senderAddress string) string {
	return (senderAddress + ";" + identifier + ";" + string(nextID))
}

func (gossiper *Gossiper) RumorMonger(rumorpkt *packets.GossipPacket, exceptIP string) string {
	rumor := rumorpkt.Rumor
	gossiper.UpdateVectorClock(rumor)
	target := gossiper.SendPacketRandomExcept(*rumorpkt, exceptIP)
	if target != "" {
		ackChannel := make(chan packets.PeerStatus)
		ackID := gossiper.AckID(rumor.Origin, rumor.ID+uint32(1), target)
		gossiper.AcksChannels[ackID] = &ackChannel
		go gossiper.WaitForAck(ackID, target, rumor.ID)
		gossiper.LogMongering(target)
	}
	return target

}

func (gossiper *Gossiper) WaitForAck(ackID string, ackSenderIP string, rumorID uint32) {

	timeout := make(chan bool)
	go func() {
		time.Sleep(10 * time.Second)
		timeout <- true
	}()
	select {
	case <-timeout:
		
		identifier := strings.Split(ackID, ";")[1]
		fmt.Println(identifier, rumorID)
		pkt:=gossiper.GetRumorPacket(identifier, rumorID)
		gossiper.RumorMonger(pkt, ackSenderIP)
		
	case status, open := <-*gossiper.AcksChannels[ackID]:
		if open {
			gossiper.AcksChannelsMux.Lock()
			defer gossiper.AcksChannelsMux.Unlock()
			gossiper.AckStatus(status, ackSenderIP, rumorID)
			delete(gossiper.AcksChannels, ackID)
		}
	}
}

//Compare statuses and either send a rumor or a status packet
//returns true if the statuses are not equal
func (gossiper *Gossiper) CompareStatusStrict(status packets.PeerStatus, ackSenderIP string) bool {
	gossiper.VectorClockMux.Lock()
	defer gossiper.VectorClockMux.Unlock()
	currentStatus, found := gossiper.VectorClock[status.Identifier]
	if found && currentStatus.NextID > status.NextID {
		packet := gossiper.GetRumorPacket(status.Identifier, status.NextID)
		if packet != nil {
			gossiper.SendPacket(*packet, ackSenderIP)
		}
		return true
	} else if found && currentStatus.NextID < status.NextID {
		gossiper.SendPacket(packets.GossipPacket{StatusPacket: gossiper.GetStatusPacket()}, ackSenderIP)
		return true
	} else {
		return false
	}
}
//Acknowledges a status packet
func (gossiper *Gossiper) AckStatus(status packets.PeerStatus, ackSenderIP string, rumorID uint32) {
	equal := !gossiper.CompareStatusStrict(status, ackSenderIP)
	if equal {
		gossiper.LogSync(ackSenderIP)
		if rand.Int()%2 == 0 {
			packet := gossiper.GetRumorPacket(status.Identifier, rumorID)
			if packet != nil {
				target := gossiper.RumorMonger(packet, ackSenderIP)
				if target != "" {
					gossiper.LogFlip(target)

				}
			}
		}
	}
}

func (gossiper *Gossiper) SendRumorRandom(origin string, id uint32, exceptAddress string) {
	packet := gossiper.GetRumorPacket(origin, id)
	if packet != nil {
		gossiper.SendPacketRandomExcept(*packet, exceptAddress)
	}
}
