package main

import (
	//	"Peerster/nodes"
	"Peerster/nodes"
	"Peerster/packets"
	"flag"
	"fmt"
	"net"
	"protobuf"
	"strings"
)

func main() {
	uiport, gossipAddr, name, peers, simpleMode, antiEntropy := parseCmd()
	//fmt.Printf("Port %s\nGossipAddr %s\nName %s\nPeers %s\nSimpleMode %t\n", *uiport, *gossipAddr, *name, *peers, *simpleMode)
	//parse IP append uiport
	gossiper := newGossiper(*gossipAddr, *name, *uiport, peers, *simpleMode, *antiEntropy)
	go gossiper.AntiEntropyLoop()
	go listenClient(gossiper)
	listenGossip(gossiper)
}

func parseCmd() (*string, *string, *string, []string, *bool, *int64) {
	//Parse arguments
	uiport := flag.String("UIPort", "8080", "port for the UI client")
	gossipAddr := flag.String("gossipAddr", "127.0.0.1:5000", "ip:port for the gossiper")
	name := flag.String("name", "", "name of the gossiper")
	peers := flag.String("peers", "", "comma separated list of peers of the form ip:port")
	simpleMode := flag.Bool("simple", false, "run gossiper in simple broadcast mode")
	antiEntropy := flag.Int64("antiEntropy", 10, "Use the given timeout in seconds for anti-entropy (relevant only for Part 2. If the flag is absent, the default anti-entropy duration is 10 seconds")
	flag.Parse()
	peersList := []string{}
	if *peers != "" {
		peersList = strings.Split(*peers, ",")
	}
	return uiport, gossipAddr, name, peersList, simpleMode, antiEntropy
}

/*
func check(err error, msg string) {
	if err != nil {
		log.Fatal(msg)
	}
}*/

func udpConnection(address string) (*net.UDPAddr, *net.UDPConn) {
	//TODO: Check for errors
	udpAddr, _ := net.ResolveUDPAddr("udp4", address)
	udpConn, _ := net.ListenUDP("udp4", udpAddr)
	return udpAddr, udpConn
}

func newGossiper(address, namee, uiport string, peers []string, simpleMode bool, antiEntropy int64) *nodes.Gossiper {
	splitted := strings.Split(address, ":")
	ip := splitted[0]

	gossipAddr, gossipConn := udpConnection(address)
	clientAddr, clientConn := udpConnection(fmt.Sprintf("%s:%s", ip, uiport))

	return &nodes.Gossiper{
		GossipAddr:     gossipAddr,
		GossipConn:     gossipConn,
		ClientAddr:     clientAddr,
		ClientConn:     clientConn,
		Name:           namee,
		Peers:          peers,
		SimpleMode:     simpleMode,
		RumorsReceived: make(map[string][]*packets.RumorMessage),
		PendingAcks:    make(map[string][]packets.PeerStatus),
		AcksChannels:   make(map[string]*chan packets.PeerStatus),
		VectorClock:    make(map[string]*packets.PeerStatus),
		AntiEntropy:    antiEntropy,
	}
}

func handleClient(gossiper *nodes.Gossiper, message []byte, rlen int) {
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

		gossiper.StoreRumor(packet)
		//Créer fonction rumorMonger
		//Dans cette fonction on crée une goroutine et alloue un channel + wait for 10 sec
		go gossiper.RumorMonger(&packet, gossiper.RelayAddress())
	}
}

func listenClient(gossiper *nodes.Gossiper) {
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
func logPeers(gossiper *nodes.Gossiper) {
	fmt.Println("PEERS ", strings.Join(gossiper.Peers[:], ","))
}

func logStatusPacket(packet *packets.StatusPacket, address string) {
	s := fmt.Sprintf("STATUS from %s", address)
	for _, status := range packet.Want {
		s += status.String()
	}
	fmt.Println(s)
}

func logRumor(rumor *packets.RumorMessage, peerAddr string) {
	fmt.Printf("RUMOR origin %s from %s contents %s\n",
		rumor.Origin,
		peerAddr,
		rumor.Text)
}

func logSimpleMessage(packet *packets.SimpleMessage) {
	fmt.Printf("SIMPLE MESSAGE origin %s from %s contents %s\n",
		packet.OriginalName,
		packet.RelayPeerAddr,
		packet.Contents)
}

func logMongering(target string) {
	fmt.Printf("MONGERING with %s\n", target)
}

func logSync(peerAddr string) {
	fmt.Printf("IN SYNC WITH %s\n", peerAddr)
}

func handleGossip(gossiper *nodes.Gossiper, message []byte, rlen int, raddr *net.UDPAddr) {
	var packet packets.GossipPacket
	protobuf.Decode(message[:rlen], &packet)
	peerAddr := fmt.Sprintf("%s:%d", raddr.IP, raddr.Port)

	if packet.Simple != nil {
		gossiper.AddPeer(packet.Simple.RelayPeerAddr)
		gossiper.LogPeers()
		gossiper.LogSimpleMessage(packet.Simple)

		sourceAddress := packet.Simple.RelayPeerAddr
		packet.Simple.RelayPeerAddr = gossiper.RelayAddress()
		gossiper.SimpleBroadcast(packet, sourceAddress)
		logPeers(gossiper)
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

func listenGossip(gossiper *nodes.Gossiper) {
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
