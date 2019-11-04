package nodes

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/Arbaba/Peerster/packets"
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
		ackChannel := make(chan *packets.StatusPacket)
		ackID := gossiper.AckID(rumor.Origin, rumor.ID+uint32(1), target)
		gossiper.AcksChannelsMux.Lock()
		gossiper.AcksChannels[ackID] = &ackChannel
		gossiper.AcksChannelsMux.Unlock()
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
	identifier := strings.Split(ackID, ";")[1]
	gossiper.AcksChannelsMux.Lock()
	channel := *gossiper.AcksChannels[ackID]
	gossiper.AcksChannelsMux.Unlock()
	select {

	case <-timeout:

		fmt.Println(identifier, rumorID)
		pkt := gossiper.GetRumorPacket(identifier, rumorID)
		gossiper.RumorMonger(pkt, ackSenderIP)

	case status, open := <-channel:
		if open {
			/*gossiper.AcksChannelsMux.Lock()
			defer gossiper.AcksChannelsMux.Unlock()*/
			gossiper.AckStatus(status, identifier, ackSenderIP, rumorID)
			gossiper.AcksChannelsMux.Lock()
			delete(gossiper.AcksChannels, ackID)
			gossiper.AcksChannelsMux.Unlock()

		}
	}
}

//Compare statuses and either send a rumor or a status packet
//returns true if the statuses are not equal
func (gossiper *Gossiper) CompareStatusStrict(status packets.PeerStatus, ackSenderIP string) bool {
	gossiper.VectorClockMux.Lock()
	currentStatus, found := gossiper.VectorClock[status.Identifier]
	gossiper.VectorClockMux.Unlock()

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

//Acknowledges a status packet that we were waiting for
func (gossiper *Gossiper) AckStatus(statuspkt *packets.StatusPacket, identifier string, ackSenderIP string, rumorID uint32) {
	code, origin, id := gossiper.CompareStatus2(statuspkt)

	switch code {
	case 0:
		packet := gossiper.GetRumorPacket(origin, id)
		if packet != nil {
			gossiper.SendPacket(*packet, ackSenderIP)
		}
	case 1:
		gossiper.SendPacket(packets.GossipPacket{StatusPacket: gossiper.GetStatusPacket()}, ackSenderIP)
	case 2:
		gossiper.LogSync(ackSenderIP)
		if rand.Int()%2 == 0 {
			gossiper.LogFlip(ackSenderIP)
			packet := gossiper.GetRumorPacket(identifier, rumorID)
			if packet != nil {
				target := gossiper.RumorMonger(packet, ackSenderIP)
				if target != "" {
					gossiper.LogFlip(target)

				}
			}
		}

	}
}

func (gossiper *Gossiper) AckRandomStatusPkt(statuspkt *packets.StatusPacket, peerAddr string) {
	if len(gossiper.RumorsReceived) == 0{
		useless_rumor_ID:= uint32(1)
		gossiper.AckStatus(statuspkt, "useless identifier",peerAddr, useless_rumor_ID)
		return 
	}
	randomPeeridx := rand.Intn(len(gossiper.RumorsReceived))
	counter := 0
	for k, v := range gossiper.RumorsReceived {
		if counter == randomPeeridx {
			randomRumoridx := rand.Intn(len(v))
			for idx, rumor := range gossiper.RumorsReceived[k] {
				if idx == randomRumoridx {
					gossiper.AckStatus(statuspkt, rumor.Origin, peerAddr, rumor.ID)
				}
			}
		}
		counter += 1
	}

}



func (gossiper *Gossiper) SendRumorRandom(origin string, id uint32, exceptAddress string) {
	packet := gossiper.GetRumorPacket(origin, id)
	if packet != nil {
		gossiper.SendPacketRandomExcept(*packet, exceptAddress)
	}
}

/*
CompareStatus compares the gossiper's status packet with the one given.
The return code can take 3 values (and code 0 takes precedence over code 1):
- code=0: the gossiper has messages unknown to the peer.
In this case, it also returns the origin and message id of the first message found.
- code=1: the peer has messages unknown to the gossiper.
The other return values are not set.
- code=2: they are synchronized.
The other return values are not set.
*/
func (gossiper *Gossiper) CompareStatus2(peerStat *packets.StatusPacket) (code int, origin string, id uint32) {
	stat := gossiper.GetStatusPacket()
	code = 2
	for _, s1 := range stat.Want {
		for _, s2 := range peerStat.Want {
			if s1.Identifier == s2.Identifier {
				if s1.NextID > s2.NextID {
					code = 0
					origin = s2.Identifier
					id = s2.NextID
					return
				} else if s1.NextID < s2.NextID {
					code = 1
				}
			}
		}
	}
	// We loop in the other order, in order to discover if the other peer
	// has peers unknown to our gossiper.
	// This may also indicate that we lack messages.
	for _, s1 := range peerStat.Want {
		known := false
		for _, s2 := range stat.Want {
			if s1.Identifier == s2.Identifier {
				known = true
			}
		}
		if !known {
			gossiper.UpdateVectorClock(&packets.RumorMessage{Origin: s1.Identifier, ID: uint32(0), Text:"" })
			if s1.NextID != 1 && code != 0 {
				code = 1
			}
		}
	}
	return
}
