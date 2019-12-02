package nodes

import (
	"fmt"
	"math/rand"
	"net"
	"strings"
	"sync"

	"github.com/Arbaba/Peerster/packets"
	"github.com/dedis/protobuf"
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
	RumorsReceived map[string][]packets.Rumorable         //All rumors received, indexed by origin and sorted by ID
	AcksChannels   map[string]*chan *packets.StatusPacket //Channels to communicate with the right ACK callback
	VectorClock    map[string]*packets.PeerStatus         //Gossiper Status
	AntiEntropy    int64
	LastPackets    []packets.GossipPacket
	GUIPort        string //Must be non nil to active the server
	RoutingTable   map[string]string
	Rtimer         int64
	PrivateMsgs    map[string][]*packets.PrivateMessage
	HOPLIMIT       uint32
	FilesInfo      map[string]*FileMetaData //files indexed by Metahash string

	DataBuffer      map[string]*chan packets.DataReply //Used to redirect datareplies to the right goroutine (see DownloadFile & DownloadMetafile)
	Files           map[string][]byte                  //The actual files contents indexed by chunk hash
	SearchChannel   chan packets.SearchReply
	NetworkSize     int64
	StubbornTimeout int64
	RoundTable      RoundTable
	RoundState      RoundState
	AcksReceived    AcksReceived
	Matches         Matches
	PeersMux        sync.Mutex
	rumorsMux       sync.Mutex
	AcksChannelsMux sync.Mutex
	VectorClockMux  sync.Mutex
	LastPacketsMux  sync.Mutex
	RoutingTableMux sync.Mutex
	PrivateMsgsMux  sync.Mutex
	FilesInfoMux    sync.Mutex
	FilesMux        sync.Mutex
	DataBufferMux   sync.Mutex
}

func NewGossiper(address, namee, uiport string, peers []string, simpleMode bool, antiEntropy int64, guiPort string, rtimer int64, networksize int64, stubbornTimeout int64) *Gossiper {
	splitted := strings.Split(address, ":")
	ip := splitted[0]

	gossipAddr, gossipConn := UdpConnection(address)
	clientAddr, clientConn := UdpConnection(fmt.Sprintf("%s:%s", ip, uiport))
	searchChannel := make(chan packets.SearchReply, 100)

	gossiper := &Gossiper{
		GossipAddr:      gossipAddr,
		GossipConn:      gossipConn,
		ClientAddr:      clientAddr,
		ClientConn:      clientConn,
		Name:            namee,
		Peers:           peers,
		SimpleMode:      simpleMode,
		RumorsReceived:  make(map[string][]packets.Rumorable),
		AcksChannels:    make(map[string]*chan *packets.StatusPacket),
		VectorClock:     make(map[string]*packets.PeerStatus),
		AntiEntropy:     antiEntropy,
		GUIPort:         guiPort,
		RoutingTable:    make(map[string]string),
		Rtimer:          rtimer,
		PrivateMsgs:     make(map[string][]*packets.PrivateMessage),
		HOPLIMIT:        uint32(9),
		FilesInfo:       make(map[string]*FileMetaData),
		DataBuffer:      make(map[string]*chan packets.DataReply),
		Files:           make(map[string][]byte),
		SearchChannel:   searchChannel,
		NetworkSize:     networksize,
		StubbornTimeout: stubbornTimeout,
		AcksReceived:    *CreateAcksReceived(),
		RoundTable:      *CreateRoundTable(),
	}
	InitMatches(&gossiper.Matches)
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
	gossiper.PeersMux.Lock()
	defer gossiper.PeersMux.Unlock()
	if len(gossiper.Peers) > 0 {
		idx := rand.Intn(len(gossiper.Peers))
		gossiper.SendPacket(packet, gossiper.Peers[idx])
		return gossiper.Peers[idx]
	}
	return ""

}

func (gossiper *Gossiper) SendPacketRandomExcept(packet packets.GossipPacket, exceptsAddresss string) string {
	gossiper.PeersMux.Lock()
	defer gossiper.PeersMux.Unlock()
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
	if gossiper.SimpleMode && packet.Simple != nil || !gossiper.SimpleMode && (packet.Rumor != nil || packet.TLCMessage != nil) {
		gossiper.LastPacketsMux.Lock()
		defer gossiper.LastPacketsMux.Unlock()
		gossiper.LastPackets = append(gossiper.LastPackets, packet)

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
