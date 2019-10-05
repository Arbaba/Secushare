package main

import (
	//	"Peerster/nodes"
	"peerster/packets"
	"flag"
	"fmt"
	"net"
	"protobuf"
	"strings"
)

type Gossiper struct {
	gossipAddr *net.UDPAddr
	gossipConn *net.UDPConn
	clientAddr *net.UDPAddr
	clientConn *net.UDPConn
	name       string
	peers      string
	simpleMode bool
}

func main() {
	uiport, gossipAddr, name, peers, simpleMode := parseCmd()
	//fmt.Printf("Port %s\nGossipAddr %s\nName %s\nPeers %s\nSimpleMode %t\n", *uiport, *gossipAddr, *name, *peers, *simpleMode)
	//parse IP append uiport
	gossiper := newGossiper(*gossipAddr, "gossiper", *uiport, *name, *peers, *simpleMode)
	go listenClient(gossiper)
	listenGossip(gossiper)
}

func parseCmd() (*string, *string, *string, *string, *bool) {
	//Parse arguments
	uiport := flag.String("UIPort", "8080", "port for the UI client")
	gossipAddr := flag.String("gossipAddr", "127.0.0.1:5000", "ip:port for the gossiper")
	name := flag.String("name", "", "name of the gossiper")
	peers := flag.String("peers", "", "comma separated list of peers of the form ip:port")
	simpleMode := flag.Bool("simple", false, "run gossiper in simple broadcast mode")
	flag.Parse()
	return uiport, gossipAddr, name, peers, simpleMode
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

func newGossiper(address, namee, uiport, name, peers string, simpleMode bool) *Gossiper {
	splitted := strings.Split(address, ":")
	ip := splitted[0]
	gossipAddr, gossipConn := udpConnection(address)
	clientAddr, clientConn := udpConnection(fmt.Sprintf("%s:%s", ip, uiport))
	return &Gossiper{
		gossipAddr: gossipAddr,
		gossipConn: gossipConn,
		clientAddr: clientAddr,
		clientConn: clientConn,
		name:       namee,
		peers:      peers,
		simpleMode: simpleMode,
	}
}

func listenClient(gossiper *Gossiper) {
	conn := gossiper.clientConn
	defer conn.Close()
	for {
		message := make([]byte, 1000)
		rlen, _, err := conn.ReadFromUDP(message[:])
		if err != nil {
			panic(err)
		}
		var packet packets.GossipPacket
		protobuf.Decode(message[:rlen], &packet)
		fmt.Printf("CLIENT MESSAGE %s\n", packet.Simple.Contents)
	}
}

func listenGossip(gossiper *Gossiper) {
	conn := gossiper.gossipConn
	defer conn.Close()
	for {
		message := make([]byte, 1000)
		rlen, _, err := conn.ReadFromUDP(message[:])
		if err != nil {
			panic(err)
		}
		var packet packets.GossipPacket
		protobuf.Decode(message[:rlen], &packet)
		fmt.Printf("SIMPLE MESSAGE origin %s from %s contents %s\n",
			packet.Simple.OriginalName,
			packet.Simple.RelayPeerAddr,
			packet.Simple.Contents)
	}
}
