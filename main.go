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
	uiport, gossipAddr, name, peers, simpleMode := parseCmd()
	//fmt.Printf("Port %s\nGossipAddr %s\nName %s\nPeers %s\nSimpleMode %t\n", *uiport, *gossipAddr, *name, *peers, *simpleMode)
	//parse IP append uiport
	gossiper := newGossiper(*gossipAddr, *name, *uiport, peers, *simpleMode)
	go listenClient(gossiper)
	listenGossip(gossiper)
}

func parseCmd() (*string, *string, *string, []string, *bool) {
	//Parse arguments
	uiport := flag.String("UIPort", "8080", "port for the UI client")
	gossipAddr := flag.String("gossipAddr", "127.0.0.1:5000", "ip:port for the gossiper")
	name := flag.String("name", "", "name of the gossiper")
	peers := flag.String("peers", "", "comma separated list of peers of the form ip:port")
	simpleMode := flag.Bool("simple", false, "run gossiper in simple broadcast mode")
	flag.Parse()
	peersList := []string{}
	if *peers != "" {
		peersList = strings.Split(*peers, ",")
	}
	return uiport, gossipAddr, name, peersList, simpleMode
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

func newGossiper(address, namee, uiport string, peers []string, simpleMode bool) *nodes.Gossiper {
	splitted := strings.Split(address, ":")
	ip := splitted[0]

	gossipAddr, gossipConn := udpConnection(address)
	clientAddr, clientConn := udpConnection(fmt.Sprintf("%s:%s", ip, uiport))
	return &nodes.Gossiper{
		GossipAddr: gossipAddr,
		GossipConn: gossipConn,
		ClientAddr: clientAddr,
		ClientConn: clientConn,
		Name:       namee,
		Peers:      peers,
		SimpleMode: simpleMode,
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
			fmt.Println(strings.Join(gossiper.Peers[:], ","))
		} else {

		}

	}
}

func listenGossip(gossiper *nodes.Gossiper) {
	conn := gossiper.GossipConn
	defer conn.Close()
	for {
		message := make([]byte, 1000)
		rlen, _, err := conn.ReadFromUDP(message[:])
		if err != nil {
			panic(err)
		}
		var packet packets.GossipPacket
		protobuf.Decode(message[:rlen], &packet)

		//if packet.Simple.OriginalName != gossiper.name {
		gossiper.AddPeer(packet.Simple.RelayPeerAddr)

		sourceAddress := packet.Simple.RelayPeerAddr
		packet.Simple.RelayPeerAddr = gossiper.RelayAddress()
		gossiper.SimpleBroadcast(packet, sourceAddress)

		fmt.Printf("SIMPLE MESSAGE origin %s from %s contents %s\n",
			packet.Simple.OriginalName,
			sourceAddress,
			packet.Simple.Contents)
		fmt.Println(strings.Join(gossiper.Peers[:], ","))
		//}

	}
}
