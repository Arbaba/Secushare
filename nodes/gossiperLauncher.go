package nodes

import (
	"github.com/Arbaba/Peerster/packets"

	"fmt"
	"net"
	"github.com/dedis/protobuf"
	"strings"
)

func (gossiper *Gossiper) LaunchGossiperCLI(){
	go gossiper.AntiEntropyLoop()
	go listenClient(gossiper)
	listenGossip(gossiper)
}

func (gossiper *Gossiper) LaunchGossiperGUI(){
	go gossiper.AntiEntropyLoop()
	go listenClient(gossiper)
	go listenGossip(gossiper)
	RunServer(gossiper)

}

func UdpConnection(address string) (*net.UDPAddr, *net.UDPConn) {
	//TODO: Check for errors
	udpAddr, _ := net.ResolveUDPAddr("udp4", address)
	udpConn, _ := net.ListenUDP("udp4", udpAddr)
	return udpAddr, udpConn
}
func listenClient(gossiper *Gossiper) {
	conn := gossiper.ClientConn
	defer conn.Close()
	for {
		message := make([]byte, 1000)
		rlen, _, err := conn.ReadFromUDP(message[:])
		if err != nil {
			panic(err)
		}
		go handleClient(gossiper, message, rlen)

	}
}

func handleClient(gossiper *Gossiper, message []byte, rlen int) {
	var msg packets.Message
	protobuf.Decode(message[:rlen], &msg)

	if gossiper.SimpleMode {
		packet := packets.GossipPacket{
			Simple: &packets.SimpleMessage{
				OriginalName:  gossiper.Name,
				RelayPeerAddr: gossiper.RelayAddress(),
				Contents:      msg.Text,
			}}
		sourceAddress := packet.Simple.RelayPeerAddr
		packet.Simple.RelayPeerAddr = gossiper.RelayAddress()
		packet.Simple.OriginalName = gossiper.Name
		gossiper.StoreLastPacket(packet)
		gossiper.SimpleBroadcast(packet, sourceAddress)
		fmt.Printf("CLIENT MESSAGE %s\n", packet.Simple.Contents)
		fmt.Println("PEERS ", strings.Join(gossiper.Peers[:], ","))
	} else {

		packet := packets.GossipPacket{
			Rumor: &packets.RumorMessage{
				Origin: gossiper.Name,
				ID:     gossiper.GetNextRumorID(gossiper.Name),
				Text:   msg.Text},
		}
		fmt.Printf("CLIENT MESSAGE %s\n", msg.Text)
		gossiper.StoreLastPacket(packet)

		gossiper.StoreRumor(packet)
		//Créer fonction rumorMonger
		//Dans cette fonction on crée une goroutine et alloue un channel + wait for 10 sec
		go gossiper.RumorMonger(&packet, gossiper.RelayAddress())
	}
}



func handleGossip(gossiper *Gossiper, message []byte, rlen int, raddr *net.UDPAddr) {
	var packet packets.GossipPacket
	protobuf.Decode(message[:rlen], &packet)
	peerAddr := fmt.Sprintf("%s:%d", raddr.IP, raddr.Port)
	gossiper.StoreLastPacket(packet)
	if packet.Simple != nil {
		gossiper.AddPeer(packet.Simple.RelayPeerAddr)
		gossiper.LogPeers()
		gossiper.LogSimpleMessage(packet.Simple)

		sourceAddress := packet.Simple.RelayPeerAddr
		packet.Simple.RelayPeerAddr = gossiper.RelayAddress()
		gossiper.SimpleBroadcast(packet, sourceAddress)
		gossiper.LogPeers()
	} else if rumor := packet.Rumor; packet.Rumor != nil {
		gossiper.LogRumor(rumor, peerAddr)
		gossiper.AddPeer(peerAddr)
		gossiper.LogPeers()

		rumor := gossiper.GetRumor(rumor.Origin, rumor.ID)
		if rumor != nil {
			return
		}
		gossiper.StoreRumor(packet)
		gossiper.SendPacket(packets.GossipPacket{StatusPacket: gossiper.GetStatusPacket()}, peerAddr)

		target := gossiper.SendPacketRandomExcept(packet, peerAddr)
		if target != "" {
			gossiper.LogMongering(target)
		}

	} else if packet.StatusPacket != nil {
		gossiper.LogStatusPacket(packet.StatusPacket, peerAddr)
		for _, status := range packet.StatusPacket.Want {
			ackID := gossiper.AckID(status.Identifier, status.NextID, peerAddr)
			ackChannel, waitingForIt := gossiper.AcksChannels[ackID]
			if waitingForIt {
				*ackChannel <- status
			} else {
				gossiper.CompareStatusStrict(status, peerAddr)
			}
		}

	}

}

func listenGossip(gossiper *Gossiper) {
	conn := gossiper.GossipConn
	defer conn.Close()
	for {
		message := make([]byte, 1000)
		rlen, raddr, err := conn.ReadFromUDP(message[:])
		if err != nil {
			panic(err)
		}
		go handleGossip(gossiper, message, rlen, raddr)

	}
}
