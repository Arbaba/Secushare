package nodes

import (
	"fmt"
	"math/rand"
	"net"
	"sort"
	"strings"
	"sync"

	"github.com/Arbaba/Peerster/packets"
	"github.com/dedis/protobuf"
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
	RumorsReceived  map[string][]*packets.RumorMessage     //All rumors received, indexed by origin and sorted by ID
	AcksChannels    map[string]*chan *packets.StatusPacket //Channels to communicate with the right ACK callback
	VectorClock     map[string]*packets.PeerStatus         //Gossiper Status
	AntiEntropy     int64
	LastPackets     []packets.GossipPacket
	GUIPort         string //Must be non nil to active the server
	RoutingTable    map[string]string
	Rtimer          int64
	PrivateMsgs     map[string][]*packets.PrivateMessage
	HOPLIMIT		uint32

	rumorsMux       sync.Mutex
	AcksChannelsMux sync.Mutex
	VectorClockMux  sync.Mutex
	LastPacketsMux  sync.Mutex
	RoutingTableMux sync.Mutex
	PrivateMsgsMux  sync.Mutex
}

func NewGossiper(address, namee, uiport string, peers []string, simpleMode bool, antiEntropy int64, guiPort string, rtimer int64) *Gossiper {
	splitted := strings.Split(address, ":")
	ip := splitted[0]

	gossipAddr, gossipConn := UdpConnection(address)
	clientAddr, clientConn := UdpConnection(fmt.Sprintf("%s:%s", ip, uiport))

	gossiper := &Gossiper{
		GossipAddr:     gossipAddr,
		GossipConn:     gossipConn,
		ClientAddr:     clientAddr,
		ClientConn:     clientConn,
		Name:           namee,
		Peers:          peers,
		SimpleMode:     simpleMode,
		RumorsReceived: make(map[string][]*packets.RumorMessage),
		AcksChannels:   make(map[string]*chan *packets.StatusPacket),
		VectorClock:    make(map[string]*packets.PeerStatus),
		AntiEntropy:    antiEntropy,
		GUIPort:        guiPort,
		RoutingTable:   make(map[string]string),
		Rtimer:         rtimer,
		PrivateMsgs:    make(map[string][]*packets.PrivateMessage),
		HOPLIMIT:		uint32(10),
	}
	return gossiper
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

//Returns the gossiper address
func (gossiper *Gossiper) RelayAddress() string {
	return fmt.Sprintf("%s:%d", gossiper.GossipAddr.IP, gossiper.GossipAddr.Port)
}

func (gossiper *Gossiper) SendPacket(packet packets.GossipPacket, address string) {

	encodedPacket, err := protobuf.Encode(&packet)
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

//Sends the packet to a random peer and returns the peer address
func (gossiper *Gossiper) SendPacketRandom(packet packets.GossipPacket) string {
	if len(gossiper.Peers) > 0 {
		idx := rand.Intn(len(gossiper.Peers))
		gossiper.SendPacket(packet, gossiper.Peers[idx])
		return gossiper.Peers[idx]
	}
	return ""

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

//Store all Messages packets received in order. Used to supply the GUI. Discard statuses
func (gossiper *Gossiper) StoreLastPacket(packet packets.GossipPacket) {
	if gossiper.SimpleMode && packet.Simple != nil || !gossiper.SimpleMode && packet.Rumor != nil {
		gossiper.LastPacketsMux.Lock()
		defer gossiper.LastPacketsMux.Unlock()
		gossiper.LastPackets = append(gossiper.LastPackets, packet)

	}
}

//Returns the last packets as rumors (even for simpleMessages) and clears the list. Used to supply the GUI
func (gossiper *Gossiper) GetLastRumorsSince(idx int) []packets.RumorMessage {
	gossiper.LastPacketsMux.Lock()
	defer gossiper.LastPacketsMux.Unlock()
	var copy []packets.RumorMessage = nil
	if idx < len(gossiper.LastPackets) && len(gossiper.LastPackets) > 0 || idx == 0 {

		for _, packet := range gossiper.LastPackets[idx:] {
			if packet.Simple != nil {

				s := packet.Simple
				copy = append(copy, packets.RumorMessage{s.OriginalName, 0, s.Contents})
			} else if packet.Rumor != nil {
				copy = append(copy, *packet.Rumor)
			}

		}
	}
	return copy
}

//Stores the rumor in a map of list of ordered rumors. Each key of the map is a node orign
func (gossiper *Gossiper) StoreRumor(packet packets.GossipPacket) {
	if rumor := packet.Rumor; rumor != nil {
		gossiper.rumorsMux.Lock()

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
		gossiper.rumorsMux.Unlock()

		gossiper.UpdateVectorClock(rumor)

	}

}

//Update the vector clock. Add Status or update nextID
func (gossiper *Gossiper) UpdateVectorClock(rumor *packets.RumorMessage) {
	gossiper.VectorClockMux.Lock()
	defer gossiper.VectorClockMux.Unlock()
	status, found := gossiper.VectorClock[rumor.Origin]
	if found && rumor.ID >= status.NextID {
		status.NextID = rumor.ID + 1
	} else if !found {
		status = &packets.PeerStatus{Identifier: rumor.Origin, NextID: rumor.ID + 1}
		gossiper.VectorClock[rumor.Origin] = status
	}
}

//Returns the rumor
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

//Returns a GossipPacket with the rumor
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

//Returns the gossiper status
func (gossiper *Gossiper) GetStatus() []packets.PeerStatus {
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

