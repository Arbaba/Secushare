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
	GossipAddr     *net.UDPAddr
	GossipConn     *net.UDPConn
	ClientAddr     *net.UDPAddr
	ClientConn     *net.UDPConn
	Name           string
	Peers          []string
	SimpleMode     bool
	StatusPacket   packets.StatusPacket
	RumorsReceived map[string][]*packets.RumorMessage
	rumorsMux      sync.Mutex
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
	conn, err := net.Dial("udp", address)
	_, err = conn.Write(encodedPacket)
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
	}
}

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
