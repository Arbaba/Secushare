package main

import (
	"flag"
	"fmt"
	"net"

	"github.com/Arbaba/Peerster/packets"
	"github.com/dedis/protobuf"
)

func main() {
	uiport := flag.String("UIPort", "8080", "port for the UI client")
	msg := flag.String("msg", "", "message to be sent")
	dest := flag.String("dest", "", "destination for the private message")
	flag.Parse()
	target := fmt.Sprintf("127.0.0.1:%s", *uiport)
	simpleMsg := packets.Message{Text: *msg}
	if *dest != ""{
		simpleMsg.Destination = dest
	}
	encodedPacket, err := protobuf.Encode(&simpleMsg)
	if err != nil {
		fmt.Println(err)
	}
	conn, err := net.Dial("udp", target)
	_, err = conn.Write(encodedPacket)
	if err != nil {
		fmt.Println("Error : ", err)
	}
}
