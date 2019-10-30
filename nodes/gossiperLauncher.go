package nodes

/*
This file serves to handle the high level behavior of the gossiper

*/
import (
	"fmt"
	"net"

	"github.com/Arbaba/Peerster/packets"

	"github.com/dedis/protobuf"
)

func (gossiper *Gossiper) LaunchGossiperCLI() {
	go listenClient(gossiper)
	go listenGossip(gossiper)
	go gossiper.SendRandomRoute()
	gossiper.AntiEntropyLoop()
}

func (gossiper *Gossiper) LaunchGossiperGUI() {
	go listenClient(gossiper)
	go listenGossip(gossiper)
	go gossiper.AntiEntropyLoop()
	if gossiper.Rtimer > 0 {
		go gossiper.RouteRumorLoop()
	}
	go gossiper.SendRandomRoute()
	RunServer(gossiper)

}

func UdpConnection(address string) (*net.UDPAddr, *net.UDPConn) {
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

//Gossiper behavior on reception on a client message
func handleClient(gossiper *Gossiper, message []byte, rlen int) {
	var msg packets.Message
	protobuf.Decode(message[:rlen], &msg)
	gossiper.LogClientMsg(msg)

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
	} else {
		if msg.Destination != nil {
			privatemsg := &packets.PrivateMessage{
				Origin:      gossiper.Name,
				ID:          0,
				Text:        msg.Text,
				Destination: *msg.Destination,
				HopLimit:    gossiper.HOPLIMIT,
			}
			gossiper.SendPrivateMsg(privatemsg)
		} else {
			//RumorMongering
			packet := packets.GossipPacket{
				Rumor: &packets.RumorMessage{
					Origin: gossiper.Name,
					ID:     gossiper.GetNextRumorID(gossiper.Name),
					Text:   msg.Text},
			}
			gossiper.StoreLastPacket(packet)

			gossiper.StoreRumor(packet)

			gossiper.RumorMonger(&packet, gossiper.RelayAddress())
		}

	}
	gossiper.LogPeers()
}

//Gossiper behavior on reception of another node packet
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
	} else if rumor := packet.Rumor; packet.Rumor != nil {
		gossiper.LogRumor(rumor, peerAddr)
		gossiper.AddPeer(peerAddr)
		gossiper.UpdateRouting(rumor.Origin, peerAddr)
		gossiper.LogDSDV(rumor, peerAddr)
		gossiper.LogPeers()

		rumor := gossiper.GetRumor(rumor.Origin, rumor.ID)
		if rumor != nil {
			//Rumor was already received, hence we discard the packet
			return
		}
		gossiper.StoreRumor(packet)
		gossiper.SendPacket(packets.GossipPacket{StatusPacket: gossiper.GetStatusPacket()}, peerAddr)
		gossiper.RumorMonger(&packet, peerAddr)

	} else if packet.StatusPacket != nil {
		gossiper.LogStatusPacket(packet.StatusPacket, peerAddr)
		gossiper.AddPeer(peerAddr)
		gossiper.LogPeers()
		for _, status := range packet.StatusPacket.Want {
			//generate and identifier to retrieve to correct ACK channel
			//A goroutine with an ackchannel is created for each rumor sent.
			//hence we pass the status via the correct channel (see rumormongering.go)
			ackID := gossiper.AckID(status.Identifier, status.NextID, peerAddr)
			ackChannel, waitingForIt := gossiper.AcksChannels[ackID]
			if waitingForIt {
				*ackChannel <- packet.StatusPacket
			} else {
				gossiper.AckRandomStatusPkt(packet.StatusPacket, peerAddr)
			}
		}

	} else if private := packet.Private; packet.Private != nil {
		if private.Destination == gossiper.Name {
			gossiper.StorePrivateMsg(private)
		} else if private.HopLimit > 0 {
			private.HopLimit -= 1
			gossiper.SendPrivateMsg(private)
		}
	}
	
}
