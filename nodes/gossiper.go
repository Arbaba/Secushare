package nodes

import (
	"Peerster/packets"
	"fmt"
	"net"
	"protobuf"
)

// Gossiper : Represents the gossiper
type Gossiper struct {
	GossipAddr *net.UDPAddr
	GossipConn *net.UDPConn
	ClientAddr *net.UDPAddr
	ClientConn *net.UDPConn
	Name       string
	Peers      []string
	SimpleMode bool
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
