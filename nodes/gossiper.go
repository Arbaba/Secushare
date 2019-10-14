package nodes

import (
	"Peerster/packets"
	"fmt"
	"math/rand"
	"net"
	"protobuf"
	"sort"
	"sync"
)

// Gossiper : Represents the gossiper
type Gossiper struct {
	GossipAddr      *net.UDPAddr
	GossipConn      *net.UDPConn
	ClientAddr      *net.UDPAddr
	ClientConn      *net.UDPConn
	Name            string
	Peers           []string
	SimpleMode      bool
	StatusPacket    packets.StatusPacket
	RumorsReceived  map[string][]*packets.RumorMessage
	PendingAcks     map[string][]packets.PeerStatus
	AcksChannels    map[string]*chan packets.PeerStatus
	VectorClock     map[string]*packets.PeerStatus
	AntiEntropy     int64
	rumorsMux       sync.Mutex
	pendingAcksMux  sync.Mutex
	AcksChannelsMux sync.Mutex
	VectorClockMux  sync.Mutex
}

func (gossiper *Gossiper) AddPeer(address string) {
	containsAddr := false
	for _, paddr := range gossiper.Peers {
		if paddr == address {
			containsAddr = true
			break
		}
	}
	if !containsAddr {
		gossiper.Peers = append(gossiper.Peers, address)
	}
}

func (gossiper *Gossiper) RelayAddress() string {
	return fmt.Sprintf("%s:%d", gossiper.GossipAddr.IP, gossiper.GossipAddr.Port)
}

func (gossiper *Gossiper) SendPacket(packet packets.GossipPacket, address string) {

	encodedPacket, err := protobuf.Encode(&packet)
	//conn, err := net.Dial("udp", address)
	udpAddr, _ := net.ResolveUDPAddr("udp4", address)
	_, err = gossiper.GossipConn.WriteToUDP(encodedPacket, udpAddr)
	if err != nil {
		fmt.Println("Error : ", err)
	}
}

func (gossiper *Gossiper) SimpleBroadcast(packet packets.GossipPacket, sourceAddress string) {
	for _, peer := range gossiper.Peers {
		if sourceAddress != peer {
			gossiper.SendPacket(packet, peer)
		}
	}
}

func (gossiper *Gossiper) SendPacketRandom(packet packets.GossipPacket) string {
	idx := rand.Intn(len(gossiper.Peers))
	gossiper.SendPacket(packet, gossiper.Peers[idx])
	return gossiper.Peers[idx]
}

func (gossiper *Gossiper) SendPacketRandomExcept(packet packets.GossipPacket, exceptsAddresss string) string {
	if len(gossiper.Peers) == 0 || len(gossiper.Peers) == 1 && gossiper.Peers[0] == exceptsAddresss {
		return ""
	} else {
		for {
			idx := rand.Intn(len(gossiper.Peers))
			target := gossiper.Peers[idx]
			if target != exceptsAddresss {
				gossiper.SendPacket(packet, target)
				return target
			}
		}
	}
}

func (gossiper *Gossiper) StoreRumor(packet packets.GossipPacket) {
	if rumor := packet.Rumor; rumor != nil {
		gossiper.rumorsMux.Lock()
		defer gossiper.rumorsMux.Unlock()

		list := make([]*packets.RumorMessage, len(gossiper.RumorsReceived[rumor.Origin]))
		copy(list, gossiper.RumorsReceived[rumor.Origin])

		id := rumor.ID
		idx := sort.Search(len(list), func(i int) bool {
			return list[i].ID > id
		})
		if idx < len(list) {
			gossiper.RumorsReceived[rumor.Origin] = append(append(gossiper.RumorsReceived[rumor.Origin][:idx], rumor), list[idx:]...)
		} else {
			gossiper.RumorsReceived[rumor.Origin] = append(list, rumor)

		}
		gossiper.UpdateVectorClock(rumor)

	}

}

func (gossiper *Gossiper) UpdateVectorClock(rumor *packets.RumorMessage) {
	gossiper.VectorClockMux.Lock()
	gossiper.VectorClockMux.Unlock()
	status, found := gossiper.VectorClock[rumor.Origin]
	if found && rumor.ID >= status.NextID {
		status.NextID = rumor.ID + 1
	} else if !found {
		status = &packets.PeerStatus{Identifier: rumor.Origin, NextID: rumor.ID + 1}
		gossiper.VectorClock[rumor.Origin] = status
	}
}

//Might be better to return a copy
func (gossiper *Gossiper) GetRumor(origin string, id uint32) *packets.RumorMessage {
	gossiper.rumorsMux.Lock()
	defer gossiper.rumorsMux.Unlock()
	list := gossiper.RumorsReceived[origin]
	idx := sort.Search(len(list), func(i int) bool {
		return list[i].ID >= id
	})
	if idx < len(list) && list[idx].ID == id {
		return list[idx]
	}
	return nil

}

func (gossiper *Gossiper) GetRumorPacket(origin string, id uint32) *packets.GossipPacket {
	rumor := gossiper.GetRumor(origin, id)
	if rumor != nil {
		return &packets.GossipPacket{Rumor: gossiper.GetRumor(origin, id)}
	}
	return nil

}

func (gossiper *Gossiper) GetNextRumorID(origin string) uint32 {
	gossiper.rumorsMux.Lock()
	defer gossiper.rumorsMux.Unlock()
	rumors := gossiper.RumorsReceived[origin]
	if len(rumors) == 0 {
		return 1
	} else {
		prevID := uint32(0)
		for _, rumor := range rumors {
			if rumor.ID != uint32(prevID+1) {
				return prevID + 1
			}
			prevID += 1
		}
		return rumors[len(rumors)-1].ID + 1
	}
}

func (gossiper *Gossiper) GetStatus() []packets.PeerStatus {
	/*
		var status []packets.PeerStatus
		for name := range gossiper.RumorsReceived {
			status = append(status, packets.PeerStatus{name, gossiper.GetNextRumorID(name)})
		}
		return status*/
	var peerStatus []packets.PeerStatus
	gossiper.VectorClockMux.Lock()
	defer gossiper.VectorClockMux.Unlock()
	for _, v := range gossiper.VectorClock {
		peerStatus = append(peerStatus, *v)
	}
	return peerStatus
}

func (gossiper *Gossiper) GetStatusPacket() *packets.StatusPacket {
	return &packets.StatusPacket{Want: gossiper.GetStatus()}
}

/*
//Returns the currentStatuses which have a greater nextID or do not appear in ackStatuses
func (gossiper *Gossiper) CompareStatus(ackStatuses, currentStatuses []packets.PeerStatus) []packets.PeerStatus {
	var statuses []packets.PeerStatus
	for _, s := range currentStatuses {
		keep := true
		newStatus := packets.PeerStatus{Identifier: s.Identifier, NextID: uint32(1)}
		for _, t := range ackStatuses {
			if s.Identifier == t.Identifier {
				if s.NextID <= t.NextID {
					keep = false
				} else {
					newStatus.NextID = t.NextID
				}
				break
			}

		}
		if keep == true {
			statuses = append(statuses, newStatus)
		}
	}

	return statuses
}

func (gossiper *Gossiper) EnqueueForAck(peerAddress, originName string, newNextID uint32) {
	gossiper.pendingAcksMux.Lock()
	defer gossiper.pendingAcksMux.Unlock()
	gossiper.PendingAcks[originName] = append(gossiper.PendingAcks[peerAddress], packets.PeerStatus{originName, newNextID})
}

func (gossiper *Gossiper) AckStatusPacket(packet *packets.StatusPacket, peerAddress string, sendLastRumors bool) []string {
	gossiper.pendingAcksMux.Lock()
	defer gossiper.pendingAcksMux.Unlock()
	var mongeringWith []string
	for _, status := range packet.Want {
		var filteredQueue []packets.PeerStatus
		//For each peer remove all pendingStatuses where the ID is smaller than the NextID
		for _, pendingStatus := range gossiper.PendingAcks[peerAddress] {
			if pendingStatus.Identifier == status.Identifier && pendingStatus.NextID > status.NextID {
				filteredQueue = append(filteredQueue, pendingStatus)
			} else if sendLastRumors && pendingStatus.Identifier == status.Identifier && pendingStatus.NextID == status.NextID {
				packet := packets.GossipPacket{Rumor: gossiper.GetRumor(status.Identifier, status.NextID)}
				address := gossiper.SendPacketRandomExcept(packet, peerAddress)
				if address != "" {
					mongeringWith = append(mongeringWith, address)
				}
			}
		}
		gossiper.PendingAcks[peerAddress] = filteredQueue
	}
	return mongeringWith
}
*/
